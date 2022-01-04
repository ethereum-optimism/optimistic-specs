package l2

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/event"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type DriverAPI interface {
	EngineAPI
	EthBackend
}

type Genesis struct {
	L1 eth.BlockID
	L2 eth.BlockID
}

type Engine struct {
	Ctx context.Context
	Log log.Logger
	// API bindings to execution engine
	RPC DriverAPI

	// The current driving force, to shutdown before closing the engine.
	driveSub ethereum.Subscription
	// There may only be 1 driver at a time
	driveLock sync.Mutex

	// Locks the L1 and L2 head changes, to keep a consistent view of the engine
	headLock sync.RWMutex

	// l1Head tracks the L1 block corresponding to the l2Head
	l1Head eth.BlockID

	// l2Head tracks the head-block of the engine
	l2Head eth.BlockID

	// l2Finalized tracks the block the engine can safely regard as irreversible
	// (a week for disputes, or maybe shorter if we see L1 finalize and take the derived L2 chain up till there)
	l2Finalized eth.BlockID

	// The L1 block we are syncing towards, may be ahead of l1Head
	l1Target eth.BlockID

	// Genesis starting point
	Genesis Genesis
}

// L1Head returns the block-id (hash and number) of the last L1 block that was derived into the L2 block
func (e *Engine) L1Head() eth.BlockID {
	e.headLock.RLock()
	defer e.headLock.RUnlock()
	return e.l1Head
}

// L2Head returns the block-id (hash and number) of the L2 chain head
func (e *Engine) L2Head() eth.BlockID {
	e.headLock.RLock()
	defer e.headLock.RUnlock()
	return e.l2Head
}

func (e *Engine) Head() (l1Head eth.BlockID, l2Head eth.BlockID) {
	e.headLock.RLock()
	defer e.headLock.RUnlock()
	return e.l1Head, e.l2Head
}

func (e *Engine) UpdateHead(l1Head eth.BlockID, l2Head eth.BlockID) {
	e.headLock.Lock()
	defer e.headLock.Unlock()
	e.l1Head = l1Head
	e.l2Head = l2Head
}

func (e *Engine) RequestHeadUpdate() error {
	e.headLock.Lock()
	defer e.headLock.Unlock()
	refL1, refL2, _, err := RefByL2Num(e.Ctx, e.RPC, nil, &e.Genesis)
	if err != nil {
		return err
	}
	e.l1Head = refL1
	e.l2Head = refL2
	return nil
}

// RefByL2Num fetches the L1 and L2 block IDs from the engine for the given L2 block height.
// Use a nil height to fetch the head.
func RefByL2Num(ctx context.Context, src eth.BlockByNumberSource, l2Num *big.Int, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, parentL2 common.Hash, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	refL2Block, err2 := src.BlockByNumber(ctx, l2Num) // nil for latest block
	if err2 != nil {
		err = fmt.Errorf("failed to retrieve head L2 block: %v", err2)
		return
	}
	return ParseL2Block(refL2Block, genesis)
}

// RefByL2Hash fetches the L1 and L2 block IDs from the engine for the given L2 block height.
// Use a nil height to fetch the head.
func RefByL2Hash(ctx context.Context, src eth.BlockByHashSource, l2Hash common.Hash, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, parentL2 common.Hash, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	refL2Block, err2 := src.BlockByHash(ctx, l2Hash)
	if err2 != nil {
		err = fmt.Errorf("failed to retrieve head L2 block: %v", err2)
		return
	}
	return ParseL2Block(refL2Block, genesis)
}

func ParseL2Block(refL2Block *types.Block, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, parentL2 common.Hash, err error) {
	refL2 = eth.BlockID{Hash: refL2Block.Hash(), Number: refL2Block.NumberU64()}
	if refL2.Number <= genesis.L2.Number {
		if refL2.Hash != genesis.L2.Hash {
			err = fmt.Errorf("unexpected L2 genesis block: %s, expected %s", refL2, genesis.L2)
			return
		}
		refL1 = genesis.L1
		refL2 = genesis.L2
		parentL2 = common.Hash{}
		return
	}

	parentL2 = refL2Block.ParentHash()

	if refL2Block.NumberU64() == 0 {
		// TODO: set to genesis L1 block ID (after sanity checking we got the right L2 genesis block from the engine)
		refL1 = eth.BlockID{}
		return
	}
	txs := refL2Block.Transactions()
	if len(txs) == 0 || txs[0].Type() != types.DepositTxType {
		err = fmt.Errorf("l2 block is missing L1 info deposit tx, block hash: %s", refL2Block.Hash())
		return
	}
	refL1Nr, _, _, refL1Hash, err := ParseL1InfoDepositTxData(txs[0].Data())
	if err != nil {
		err = fmt.Errorf("failed to parse L1 info deposit tx from L2 block: %v", err)
		return
	}
	refL1 = eth.BlockID{Hash: refL1Hash, Number: refL1Nr}
	return
}

func (e *Engine) Drive(dl l1.Downloader, canonicalL1 eth.BlockHashByNumber, l1Heads <-chan eth.HeadSignal) ethereum.Subscription {
	e.driveLock.Lock()
	defer e.driveLock.Unlock()
	if e.driveSub != nil {
		return e.driveSub
	}
	e.driveSub = event.NewSubscription(func(quit <-chan struct{}) error {
		// keep making many sync steps if we can make sync progress
		hot := time.Millisecond * 30
		// check on sync regularly, but prioritize sync triggers with head updates etc.
		cold := time.Second * 15
		// at least try every minute to sync, even if things are going well
		max := time.Minute

		syncTicker := time.NewTicker(cold)
		defer syncTicker.Stop()

		// backoff sync attempts if we are not making progress
		backoff := cold
		syncQuickly := func() {
			syncTicker.Reset(hot)
			backoff = cold
		}
		// exponential backoff, add 10% each step, up to max.
		syncBackoff := func() {
			backoff += backoff / 10
			if backoff > max {
				backoff = max
			}
			syncTicker.Reset(backoff)
		}

		l2HeadPollTicker := time.NewTicker(time.Second * 14)
		defer l2HeadPollTicker.Stop()

		onL2Update := func() {
			// When we updated L2, we want to continue sync quickly
			syncQuickly()
			// And we want to slow down requesting the L2 engine for its head (we just changed it ourselves)
			// Request head if we don't successfully change it in the next 14 seconds.
			l2HeadPollTicker.Reset(time.Second * 14)
		}

		for {
			select {
			case <-e.Ctx.Done():
				return e.Ctx.Err()
			case <-quit:
				return nil
			case <-l2HeadPollTicker.C:
				if err := e.RequestHeadUpdate(); err != nil {
					e.Log.Error("failed to fetch L2 head info", "err", err)
				}
				continue
			case l1HeadSig := <-l1Heads:
				if e.l1Head == l1HeadSig.Self {
					e.Log.Debug("Received L1 head signal, already synced to it, ignoring event", "l1_head", e.l1Head)
					continue
				}
				if e.l1Head == l1HeadSig.Parent {
					// Simple extend, a linear life is easy
					if err := e.DriveStep(dl, l1HeadSig.Self, e.l2Head, e.l2Finalized.Hash); err != nil {
						e.Log.Error("Failed to extend L2 chain with new L1 block", "l1", l1HeadSig.Self, "l2", e.l2Head, "err", err)
						// Retry sync later
						e.l1Target = l1HeadSig.Self
						onL2Update()
						continue
					}
					continue
				}
				if e.l1Head.Number < l1HeadSig.Parent.Number {
					e.Log.Debug("Received new L1 head, engine is out of sync, cannot immediately process", "l1", l1HeadSig.Self, "l2", e.l2Head)
				} else {
					e.Log.Warn("Received a L1 reorg, syncing new alternative chain", "l1", l1HeadSig.Self, "l2", e.l2Head)
				}

				e.l1Target = l1HeadSig.Self
				syncQuickly()
				continue
			case <-syncTicker.C:
				// If already synced, or in case of failure, we slow down
				syncBackoff()
				if e.l1Head == e.l1Target {
					e.Log.Debug("Engine is fully synced", "l1_head", e.l1Head, "l2_head", e.l2Head)
					// TODO: even though we are fully synced, it may be worth attempting anyway,
					// in case the e.l1Head is not updating (failed/broken L1 head subscription)
					continue
				}
				refL1, refL2, err := FindSyncStart(e.Ctx, canonicalL1, e.RPC, &e.Genesis)
				if err != nil {
					e.Log.Error("Failed to find sync starting point", "err", err)
					continue
				}
				if refL1 == e.l1Head {
					e.Log.Debug("Engine is already synced, aborting sync", "l1_head", e.l1Head, "l2_head", e.l2Head)
					continue
				}
				if err := e.DriveStep(dl, refL1, refL2, e.l2Finalized.Hash); err != nil {
					e.Log.Error("Failed to sync L2 chain with new L1 block", "l1", refL1, "onto_l2", refL2, "err", err)
					continue
				}
				// Successfully stepped toward target. Continue quickly if we are not there yet
				if e.l1Head != e.l1Target {
					onL2Update()
				}
			}
		}
	})
	return e.driveSub
}

func (e *Engine) DriveStep(dl l1.Downloader, l1Input eth.BlockID, l2Parent eth.BlockID, l2Finalized common.Hash) error {
	e.headLock.Lock()
	defer e.headLock.Unlock()

	logger := e.Log.New(
		"eng_l1", e.l1Head,
		"eng_l2", e.l2Head,
		"input_l1", l1Input,
		"input_l2_parent", l2Parent,
		"finalized_l2", l2Finalized)

	ctx, cancel := context.WithTimeout(e.Ctx, time.Second*20)
	defer cancel()
	bl, receipts, err := dl.Fetch(ctx, l1Input)
	if err != nil {
		return fmt.Errorf("failed to fetch block with receipts: %v", err)
	}

	attrs, err := DerivePayloadAttributes(bl, receipts)
	if err != nil {
		return fmt.Errorf("failed to derive execution payload inputs: %v", err)
	}

	preState := &ForkchoiceState{
		HeadBlockHash:      l2Parent.Hash, // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      l2Parent.Hash,
		FinalizedBlockHash: l2Finalized,
	}
	payload, err := DeriveBlock(ctx, e.RPC, preState, attrs)
	if err != nil {
		return fmt.Errorf("failed to derive execution payload: %v", err)
	}
	l2ID := eth.BlockID{Hash: payload.BlockHash, Number: uint64(payload.BlockNumber)}
	logger = logger.New("derived_l2", l2ID)
	logger.Info("derived block")

	ctx, cancel = context.WithTimeout(e.Ctx, time.Second*5)
	defer cancel()
	execRes, err := e.RPC.ExecutePayload(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to execute payload: %v", err)
	}
	switch execRes.Status {
	case ExecutionValid:
		logger.Info("Executed new payload")
	case ExecutionSyncing:
		return fmt.Errorf("failed to execute payload %s, node is syncing, latest valid hash is %s", l2ID, execRes.LatestValidHash)
	case ExecutionInvalid:
		return fmt.Errorf("execution payload %s was INVALID! Latest valid hash is %s, ignoring bad block: %q", l2ID, execRes.LatestValidHash, execRes.ValidationError)
	default:
		return fmt.Errorf("unknown execution status on %s: %q, ", l2ID, string(execRes.Status))
	}

	postState := &ForkchoiceState{
		HeadBlockHash:      payload.BlockHash, // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      payload.BlockHash,
		FinalizedBlockHash: l2Finalized,
	}

	ctx, cancel = context.WithTimeout(e.Ctx, time.Second*5)
	defer cancel()
	fcRes, err := e.RPC.ForkchoiceUpdated(ctx, postState, nil)
	if err != nil {
		return fmt.Errorf("failed to update forkchoice: %v", err)
	}
	switch fcRes.Status {
	case UpdateSyncing:
		return fmt.Errorf("updated forkchoice, but node is syncing: %v", err)
	case UpdateSuccess:
		logger.Info("updated forkchoice")
		e.l1Head = l1Input
		e.l2Head = l2ID
		return nil
	default:
		return fmt.Errorf("unknown forkchoice status on %s: %q, ", l2ID, string(fcRes.Status))
	}
}

func (e *Engine) Close() {
	e.RPC.Close()
	e.driveSub.Unsubscribe()
}

func FindSyncStart(ctx context.Context, l1Src eth.BlockHashByNumber, l2Src eth.BlockSource, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var parentL2 common.Hash
	// Start at L2 head
	refL1, refL2, parentL2, err = RefByL2Num(ctx, l2Src, nil, genesis)
	if err != nil {
		err = fmt.Errorf("failed to fetch L2 head: %v", err)
		return
	}
	// Check if L1 source has the block
	var l1Hash common.Hash
	l1Hash, err = l1Src.BlockHashByNumber(ctx, refL1.Number)
	if err != nil {
		err = fmt.Errorf("failed to lookup block %d in L1: %v", refL1.Number, err)
		return
	}
	if l1Hash == refL1.Hash {
		return
	}

	// Search back: linear walk back from engine head. Should only be as deep as the reorg.
	for refL2.Number > 0 {
		refL1, refL2, parentL2, err = RefByL2Hash(ctx, l2Src, parentL2, genesis)
		if err != nil {
			// TODO: re-attempt look-up, now that we already traversed previous history?
			err = fmt.Errorf("failed to lookup block %d in L1: %v", refL1.Number, err)
			return
		}
		// Check if L1 source has the block
		l1Hash, err = l1Src.BlockHashByNumber(ctx, refL1.Number)
		if err != nil {
			err = fmt.Errorf("failed to lookup block %d in L1: %v", refL1.Number, err)
			return
		}
		if l1Hash == refL1.Hash {
			return
		}
		// TODO: after e.g. initial N steps, use binary search instead
		// (relies on block numbers, not great for tip of chain, but nice-to-have in deep reorgs)
	}
	// Enforce that we build on the desired genesis block.
	// The engine might be configured for a different chain or older testnet.
	if refL2 != genesis.L2 {
		err = fmt.Errorf("engine was anchored at unexpected block: %s, expected %s", refL2, genesis.L2)
		return
	}
	// we got the correct genesis, all good, but a lot to sync!
	return
}
