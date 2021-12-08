package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"
)

type InputItem struct {
	Block    *types.Block
	Receipts []*types.Receipt
}

func (item *InputItem) Epoch() uint64 {
	return item.Block.NumberU64()
}

// sanity check that the receipts are consistent with the block data
func (item *InputItem) CheckReceipts() bool {
	hasher := trie.NewStackTrie(nil)
	computed := types.DeriveSha(types.Receipts(item.Receipts), hasher)
	return item.Block.ReceiptHash() == computed
}

type DepositTransaction struct {
	Sender common.Address
	// TODO some flag/bool for contract-deploy transaction. Or it can be *To with nil
	Flags uint8
	To    common.Address
	Value uint256.Int
	Data  []byte
}

func (dep *DepositTransaction) Decode(data []byte) (valid bool) {
	// TODO
	return false
}

type OpaqueTransaction []byte

type OutputItem struct {
	// TODO: convert to L1 attributes deposit
	L1BlockHash   common.Hash
	L1BlockNumber uint64

	// To be converted to full block with Engine API
	UserDepositTxs []*DepositTransaction
}

var DepositGatewayAddr = common.HexToAddress("0xaaaaaaaaaTODO")

// TODO: maybe submit batches to a real deployed contract, if we can avoid the EVM running?
// To save the ETH and other things that get accidentally get here.
var BatchSubmissionAddr = common.HexToAddress("0x0000000000000000000000006f7074696d69736d") // "optimism"

func Derive(input InputItem) (out OutputItem) {
	out.L1BlockHash = input.Block.Hash()
	out.L1BlockNumber = input.Block.NumberU64()

	// UserDepositTxs
	for _, rec := range input.Receipts {
		if rec.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range rec.Logs {
			if log.Address == DepositGatewayAddr {
				var tx DepositTransaction
				if !tx.Decode(log.Data) {
					// TODO: warn about bad deposit tx
					continue
				}
				out.UserDepositTxs = append(out.UserDepositTxs, &tx)
			} else {
				// TODO: logs from other addresses may still be converted to other types of deposit txs,
				// if we implement payment for this type of L1 -> L2 subscription service.
			}
		}
	}

	for _, tx := range input.Block.Transactions() {
		if to := tx.To(); to != nil && *to == BatchSubmissionAddr {
			// TODO: parse L2 block data from transaction
		}
	}
	return
}
