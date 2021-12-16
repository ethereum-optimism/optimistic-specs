package l2

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"
)

const (
	DepositTxTypePrefix = 0x7E
)

var (
	ErrTxTypePrefix     = errors.New("wrong transaction type prefix")
	DepositContractAddr = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
	BatchSubmissionAddr = common.HexToAddress("0x0000000000000000000000006f7074696d69736d") // TODO pick address
	SequencerAddr       = common.HexToAddress("0x4242424242424242424242424242424242424242") // TODO pick address
)

type InputItem struct {
	Block    *types.Block
	Receipts []*types.Receipt
}

// sanity check that the receipts are consistent with the block data
func (item *InputItem) CheckReceipts() bool {
	hasher := trie.NewStackTrie(nil)
	computed := types.DeriveSha(types.Receipts(item.Receipts), hasher)
	return item.Block.ReceiptHash() == computed
}

// NOTE: It's impossible to define DecodeRLP on a pointer of pointer, so we need the struct.
//       Defining DecodeRLP on DepositTransaction is too much of a hassle due to manual error handling.
type AddressOrNil struct {
	Ptr *common.Address
}

func (addr *AddressOrNil) DecodeRLP(s *rlp.Stream) error {
	var bytes []byte
	var err error
	if bytes, err = s.Bytes(); err != nil {
		return err
	} else if len(bytes) == 0 {
		addr.Ptr = nil
		return nil
	} else if len(bytes) != common.AddressLength {
		// we can't use `s.Decode(addr.Ptr)` because we already read the bytes
		return fmt.Errorf("rlp: wrong address size: %d", len(bytes))
	} else {
		addr.Ptr = new(common.Address)
		copy(addr.Ptr[:], bytes)
		return nil
	}
}

type DepositTransaction struct {
	From     common.Address
	To       AddressOrNil // nil for the empty/nil address (∅)
	Value    uint256.Int
	GasLimit uint256.Int
	Data     []byte
}

func (dep *DepositTransaction) Decode(data []byte) error {
	if len(data) == 0 || data[0] != DepositTxTypePrefix {
		return ErrTxTypePrefix
	}
	return rlp.Decode(bytes.NewReader(data), dep)
}

type OpaqueTransaction []byte

type OutputItem struct {
	// TODO: convert to L1 attributes deposit
	L1BlockHash   common.Hash
	L1BlockNumber uint64
	L1BlockTime   uint64
	L1Random      Bytes32

	// To be converted to full block with Engine API
	UserDepositTxs []*DepositTransaction
}

func Derive(input InputItem) (out OutputItem) {
	if !input.CheckReceipts() {
		panic("Receipts are not consistent with the block's receipts root.")
	}
	out.L1BlockHash = input.Block.Hash()
	out.L1BlockNumber = input.Block.NumberU64()
	out.L1BlockTime = input.Block.Time()
	out.L1Random = Bytes32(input.Block.MixDigest()) // TODO change to Random (needs post-merge geth)

	// fill in out.UserDepositTxs
	for _, rec := range input.Receipts {
		if rec.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range rec.Logs {
			if log.Address == DepositContractAddr {
				var tx DepositTransaction
				if err := tx.Decode(log.Data); err != nil {
					// TODO use propper logging
					fmt.Println("bad deposit transaction: " + err.Error())
					continue
				}
				out.UserDepositTxs = append(out.UserDepositTxs, &tx)
			}
		}
	}
	return
}

func OutputItemToPayloadAttributes(item OutputItem) (attr *PayloadAttributes) {
	attr = new(PayloadAttributes)
	attr.Timestamp = Uint64Quantity(item.L1BlockTime)
	attr.Random = item.L1Random
	attr.SuggestedFeeRecipient = SequencerAddr

	return attr
}

/*
func CreateBlock(item OutputItem) {
	// TODO pick appropriate ctx, client
	ctx := context.Background()
	var client *rpc.Client
	var logger log.Logger
	var state *ForkchoiceState
	var attributes *PayloadAttributes
	ForkchoiceUpdated(ctx, client, logger, state, attributes)
}
*/
