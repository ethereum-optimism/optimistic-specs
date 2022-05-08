package p2p

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/snappy"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

func init() {
	// TODO: a PR is open to make this configurable upstream as option instead of having to override a global
	pubsub.TimeCacheDuration = 80 * pubsub.GossipSubHeartbeatInterval
}

const maxGossipSize = 1 << 20
const maxOutboundQueue = 256
const maxValidateQueue = 256
const globalValidateThrottle = 512

// Message domains, the msg id function uncompresses to keep data monomorphic,
// but invalid compressed data will need a unique different id.

var MessageDomainInvalidSnappy = [4]byte{0, 0, 0, 0}
var MessageDomainValidSnappy = [4]byte{1, 0, 0, 0}

const MaxGossipSize = 1 << 20

func blocksTopicV1(cfg *rollup.Config) string {
	return fmt.Sprintf("/optimism/%s/0/blocks", cfg.L2ChainID.String())
}

// BuildSubscriptionFilter builds a simple subscription filter,
// to help protect against peers spamming useless subscriptions.
func BuildSubscriptionFilter(cfg *rollup.Config) pubsub.SubscriptionFilter {
	return pubsub.NewAllowlistSubscriptionFilter(blocksTopicV1(cfg)) // add more topics here in the future, if any.
}

var msgBufPool = sync.Pool{New: func() any {
	// note: the topic validator concurrency is limited, so pool won't blow up, even with large pre-allocation.
	x := make([]byte, 0, MaxGossipSize)
	return &x
}}

// BuildMsgIdFn builds a generic message ID function for gossipsub that can handle compressed payloads,
// mirroring the eth2 p2p gossip spec.
func BuildMsgIdFn(cfg *rollup.Config) pubsub.MsgIdFunction {
	return func(pmsg *pb.Message) string {
		valid := false
		var data []byte
		// If it's a valid compressed snappy data, then hash the uncompressed contents.
		// The validator can throw away the message later when recognized as invalid,
		// and the unique hash helps detect duplicates.
		dLen, err := snappy.DecodedLen(pmsg.Data)
		if err == nil && dLen <= MaxGossipSize {
			res := msgBufPool.Get().(*[]byte)
			defer msgBufPool.Put(res)
			if data, err = snappy.Decode((*res)[:0], pmsg.Data); err == nil {
				*res = data // if we ended up growing the slice capacity, fine, keep the larger one.
				valid = true
			}
		}
		if data == nil {
			data = pmsg.Data
		}
		h := sha256.New()
		if valid {
			h.Write(MessageDomainValidSnappy[:])
		} else {
			h.Write(MessageDomainInvalidSnappy[:])
		}
		// The chain ID is part of the gossip topic, making the msg id unique
		topic := pmsg.GetTopic()
		var topicLen [8]byte
		binary.LittleEndian.PutUint64(topicLen[:], uint64(len(topic)))
		h.Write(topicLen[:])
		h.Write([]byte(topic))
		h.Write(data)
		// the message ID is shortened to save space, a lot of these may be gossiped.
		return string(h.Sum(nil)[:20])
	}
}

func BuildGlobalGossipParams(cfg *rollup.Config) pubsub.GossipSubParams {
	params := pubsub.DefaultGossipSubParams()
	params.D = 8                                      // topic stable mesh target count
	params.Dlo = 6                                    // topic stable mesh low watermark
	params.Dhi = 12                                   // topic stable mesh high watermark
	params.Dlazy = 6                                  // gossip target
	params.HeartbeatInterval = 500 * time.Millisecond // frequency of heartbeat, seconds
	params.FanoutTTL = 24 * time.Second               // ttl for fanout maps for topics we are not subscribed to but have published to, seconds
	params.HistoryLength = 12                         // number of windows to retain full messages in cache for IWANT responses
	params.HistoryGossip = 3                          // number of windows to gossip about

	return params
}

func NewGossipSub(p2pCtx context.Context, h host.Host, cfg *rollup.Config) (*pubsub.PubSub, error) {
	denyList, err := pubsub.NewTimeCachedBlacklist(30 * time.Second)
	if err != nil {
		return nil, err
	}
	return pubsub.NewGossipSub(p2pCtx, h,
		pubsub.WithMaxMessageSize(maxGossipSize),
		pubsub.WithMessageIdFn(BuildMsgIdFn(cfg)),
		pubsub.WithNoAuthor(),
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign),
		pubsub.WithSubscriptionFilter(BuildSubscriptionFilter(cfg)),
		pubsub.WithValidateQueueSize(maxValidateQueue),
		pubsub.WithPeerOutboundQueueSize(maxOutboundQueue),
		pubsub.WithValidateThrottle(globalValidateThrottle),
		pubsub.WithPeerExchange(false),
		pubsub.WithBlacklist(denyList),
		pubsub.WithGossipSubParams(BuildGlobalGossipParams(cfg)))
	// TODO: pubsub.WithDiscovery(discover) to search for peers instead of randomly grabbing from open connections
	// TODO: pubsub.WithPeerScoreInspect(inspect, InspectInterval) to update peerstore scores with gossip scores
}

func BuildBlocksValidator(log log.Logger, cfg *rollup.Config) pubsub.ValidatorEx {

	// Seen block hashes per block height
	// uint64 -> []common.Hash
	blockHeightLRU, err := lru.New(100)
	if err != nil {
		panic(fmt.Errorf("failed to set up block height LRU cache: %v", err))
	}

	return func(ctx context.Context, id peer.ID, message *pubsub.Message) pubsub.ValidationResult {
		// [REJECT] if the compression is not valid
		outLen, err := snappy.DecodedLen(message.Data)
		if err != nil {
			log.Warn("invalid snappy compression length data", "err", err, "peer", id)
			return pubsub.ValidationReject
		}
		if outLen > maxGossipSize {
			log.Warn("possible snappy zip bomb, decoded length is too large", "decoded_length", outLen, "peer", id)
			return pubsub.ValidationReject
		}

		res := msgBufPool.Get().(*[]byte)
		defer msgBufPool.Put(res)
		data, err := snappy.Decode((*res)[:0], message.Data)
		if err != nil {
			log.Warn("invalid snappy compression", "err", err, "peer", id)
			return pubsub.ValidationReject
		}
		*res = data // if we ended up growing the slice capacity, fine, keep the larger one.

		// message starts with compact-encoding secp256k1 encoded signature
		signatureBytes, payloadBytes := data[:65], data[65:]

		// [REJECT] if the block encoding is not valid
		var payload l2.ExecutionPayload
		if err := payload.UnmarshalSSZ(uint32(len(payloadBytes)), bytes.NewReader(payloadBytes)); err != nil {
			log.Warn("invalid payload", "err", err, "peer", id)
			return pubsub.ValidationReject
		}

		// rounding down to seconds is fine here.
		now := uint64(time.Now().Unix())

		// [REJECT] if the `payload.timestamp` is older than 20 seconds in the past
		if uint64(payload.Timestamp) < now-20 {
			log.Warn("payload is too old", "timestamp", uint64(payload.Timestamp))
			return pubsub.ValidationReject
		}

		// [REJECT] if the `payload.timestamp` is more than 5 seconds into the future
		if uint64(payload.Timestamp) > now+5 {
			log.Warn("payload is too new", "timestamp", uint64(payload.Timestamp))
			return pubsub.ValidationReject
		}

		// [REJECT] if the `block_hash` in the `payload` is not valid
		if !payload.CheckBlockHash() {
			log.Warn("payload has bad block hash", "bad_hash", payload.BlockHash)
			return pubsub.ValidationReject
		}

		// [REJECT] if more than 5 blocks have been seen with the same block height
		seen, ok := blockHeightLRU.Get(uint64(payload.BlockNumber))
		if !ok {
			seen = []common.Hash{}
		}
		if len(seen.([]common.Hash)) > 5 {
			log.Warn("seen too many different blocks at same height", "height", payload.BlockNumber)
			return pubsub.ValidationReject
		}
		for _, prev := range seen.([]common.Hash) {
			// [IGNORE] if the block has already been seen
			if prev == payload.BlockHash {
				log.Warn("validated already seen message again")
				return pubsub.ValidationIgnore
			}
		}

		// [REJECT] if the signature by the sequencer is not valid
		signingHash := BlockSigningHash(cfg, payloadBytes)

		pub, err := crypto.SigToPub(signingHash[:], signatureBytes)
		if err != nil {
			log.Warn("invalid block signature", "err", err, "peer", id)
			return pubsub.ValidationReject
		}
		addr := crypto.PubkeyToAddress(*pub)

		// TODO: in the future we can support multiple valid p2p addresses.
		if addr != cfg.P2PSequencerAddress {
			log.Warn("unexpected block author", "err", err, "peer", id)
			return pubsub.ValidationReject
		}

		seen = append(seen.([]common.Hash), payload.BlockHash)
		blockHeightLRU.Add(uint64(payload.BlockNumber), seen)

		// remember the decoded payload for later usage in topic subscriber.
		message.ValidatorData = payload
		return pubsub.ValidationAccept
	}
}

func BlockSigningHash(cfg *rollup.Config, payloadBytes []byte) common.Hash {
	var msgInput [32 + 32 + 32]byte
	// domain: first 32 bytes, always zero for now
	// chain_id: second 32 bytes
	cfg.L2ChainID.FillBytes(msgInput[32:64])
	// payload_hash: third 32 bytes, hash of encoded payload
	copy(msgInput[32:], crypto.Keccak256(payloadBytes))

	return crypto.Keccak256Hash(msgInput[:])
}

type BlockMessage struct {
	ReceivedFrom peer.ID
	Block        *l2.ExecutionPayload
}

func JoinGossip(p2pCtx context.Context, ps *pubsub.PubSub, log log.Logger, cfg *rollup.Config, blocks chan<- BlockMessage) error {
	val := BuildBlocksValidator(log, cfg)
	topicName := blocksTopicV1(cfg)
	err := ps.RegisterTopicValidator(topicName,
		val,
		pubsub.WithValidatorTimeout(3*time.Second),
		pubsub.WithValidatorConcurrency(4))
	if err != nil {
		return fmt.Errorf("failed to register blocks gossip topic: %v", err)
	}
	topic, err := ps.Join(topicName)
	if err != nil {
		return fmt.Errorf("failed to join blocks gossip topic: %v", err)
	}
	// TODO: block topic scoring parameters
	// See prysm: https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/p2p/gossip_scoring_params.go
	// And research from lighthouse: https://gist.github.com/blacktemplar/5c1862cb3f0e32a1a7fb0b25e79e6e2c
	// And docs: https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#topic-parameter-calculation-and-decay
	//err := topic.SetScoreParams(&pubsub.TopicScoreParams{......})

	subscription, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to blocks gossip topic: %v", err)
	}

	subscriber := MakeSubscriber(log, BlocksHandler(blocks))
	go subscriber(p2pCtx, subscription)
	return nil
}

type TopicSubscriber func(ctx context.Context, sub *pubsub.Subscription)
type MessageHandler func(ctx context.Context, from peer.ID, msg interface{}) error

func BlocksHandler(out chan<- BlockMessage) MessageHandler {
	return func(ctx context.Context, from peer.ID, msg interface{}) error {
		payload, ok := msg.(*l2.ExecutionPayload)
		if !ok {
			return fmt.Errorf("expected topic validator to parse and validate data into execution payload, but got %T", msg)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- BlockMessage{ReceivedFrom: from, Block: payload}:
			return nil
		}
	}
}

func MakeSubscriber(log log.Logger, msgHandler MessageHandler) TopicSubscriber {
	return func(ctx context.Context, sub *pubsub.Subscription) {
		topicLog := log.New("topic", sub.Topic())
		for {
			msg, err := sub.Next(ctx)
			if err != nil { // ctx was closed, or subscription was closed
				topicLog.Debug("stopped subscriber")
				return
			}
			if msg.ValidatorData == nil {
				topicLog.Error("gossip message with no data", "from", msg.ReceivedFrom)
				continue
			}
			if err := msgHandler(ctx, msg.ReceivedFrom, msg.ValidatorData); err != nil {
				topicLog.Error("failed to process gossip message", "err", err)
			}
		}
	}
}