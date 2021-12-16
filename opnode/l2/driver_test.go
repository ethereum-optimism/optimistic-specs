package l2

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
)

// ================================================================================================
// INPUT GENERATION HELPERS

var (
	zero        = big.NewInt(0)
	zeroHash    = common.BigToHash(zero)
	zeroAddress = common.BigToAddress(zero)
	chaindID    = big.NewInt(69) // testnet chain ID, not meaningful here
)

var depositEventABI = "TransactionDeposited(address,address,uint256,uint256,bool,bytes)"

func GenerateInputItem() InputItem {
	txs := types.Transactions{GenerateTransaction()}
	receipts := types.Receipts{}
	block := GenerateBlock(txs, receipts, time.Now())
	return InputItem{
		Block:    block,
		Receipts: receipts,
	}
}

func GenerateTransaction() *types.Transaction {
	txData := &types.DynamicFeeTx{
		ChainID: chaindID,
		Data:    []byte{},

		// ignored (zeroed):
		Nonce:      0,
		GasTipCap:  zero,
		GasFeeCap:  zero,
		Gas:        0,
		To:         &zeroAddress,
		Value:      zero,
		AccessList: types.AccessList{},
		V:          zero,
		R:          zero,
		S:          zero,
	}
	return types.NewTx(txData)
}

func GenerateDepositLog() *types.Log {
	// as specified by https://docs.soliditylang.org/en/v0.8.10/abi-spec.html?highlight=topics#events

	topics := []common.Hash{
		common.BytesToHash(crypto.Keccak256([]byte(depositEventABI))),
	}
	data := []byte{}
	return GenerateLog(DepositContractAddr, topics, data)
}

func GenerateLog(addr common.Address, topics []common.Hash, data []byte) *types.Log {
	return &types.Log{
		Address: addr,
		Topics:  topics,
		Data:    data,
		Removed: false,

		// ignored (zeroed):
		BlockNumber: 0,
		TxHash:      zeroHash,
		TxIndex:     0,
		BlockHash:   zeroHash,
		Index:       0,
	}
}

func GenerateReceipt() *types.Receipt {
	logs := []*types.Log{}
	receipt := &types.Receipt{
		Type:   types.DynamicFeeTxType,
		Status: types.ReceiptStatusSuccessful,
		Logs:   logs,

		// ignored (zeroed):
		PostState:         []byte{},
		CumulativeGasUsed: 0,
		TxHash:            zeroHash,
		ContractAddress:   zeroAddress,
		GasUsed:           0,
		BlockHash:         zeroHash,
		BlockNumber:       zero,
		TransactionIndex:  0,
	}
	// not really needed
	receipt.Bloom = types.CreateBloom([]*types.Receipt{receipt})
	return receipt
}

func GenerateBlock(txs types.Transactions, receipts types.Receipts, _time time.Time) *types.Block {

	header := &types.Header{
		Time:  uint64(_time.Unix()),
		Extra: []byte{},

		// ignored (zeroed):
		ParentHash: zeroHash,
		UncleHash:  zeroHash,
		Coinbase:   zeroAddress,
		Root:       zeroHash,
		Difficulty: zero,
		Number:     zero,
		GasLimit:   0,
		MixDigest:  zeroHash,
		Nonce:      types.EncodeNonce(0),
		BaseFee:    zero,

		// not supplied (computed by NewBlock):
		// - TxHash
		// - ReceiptHash
		// - Bloom
		// - UncleHash
	}

	uncles := []*types.Header{}
	hasher := trie.NewStackTrie(nil)
	return types.NewBlock(header, txs, uncles, receipts, hasher)
}

// ================================================================================================
// TESTS

func TestNoDeposits(t *testing.T) {
	input := GenerateInputItem()
	output := Derive(input)
	if output.L1BlockHash != input.Block.Hash() {
		t.Errorf("wrong L1 hash")
	} else if output.L1BlockTime != input.Block.Time() {
		t.Errorf("wrong L1 block time")
	} else if output.L1Random != Bytes32(input.Block.MixDigest()) {
		t.Errorf("wrong L1 randomness")
	} else if len(output.UserDepositTxs) != 0 {
		t.Errorf("expected empty deposit transaction array")
	}
}

func TestSingleDeposit(t *testing.T) {
	// TODO
}

// TODO test failed receipts

// ================================================================================================
