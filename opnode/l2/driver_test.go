package l2

import (
	"fmt"
	"github.com/holiman/uint256"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/assert"
)

// ================================================================================================
// INPUT GENERATION HELPERS

var (
	// big integer zero
	zero = big.NewInt(0)

	// testnet chain ID, not meaningful here
	chaindID = big.NewInt(69)
)

// Generates an InputItem with nSuccess successful transactions and nFailed failed transactions,
// as well as corresponding deposits.
// The transactions are dummy transactions, that do not match the log entries contained the
// deposits.
func GenerateInputItem(nSuccess uint, nFailed uint) InputItem {
	nTxs := nSuccess + nFailed
	txs := make(types.Transactions, nTxs)
	for i := range txs {
		txs[i] = GenerateTransaction()
	}

	receipts := make(types.Receipts, nTxs)
	for i := uint(0); i < nSuccess; i++ {
		receipts[i] = GenerateDepositReceipt(i)
	}
	for i := nSuccess; i < nTxs; i++ {
		receipts[i] = GenerateFailedDepositReceipt(i)
	}

	block := GenerateBlock(txs, receipts, time.Now())
	return InputItem{
		Block:    block,
		Receipts: receipts,
	}
}

// Generates a dummy transaction with most fields zeroed.
func GenerateTransaction() *types.Transaction {
	txData := &types.DynamicFeeTx{
		ChainID: chaindID,
		Data:    []byte{},

		// ignored (zeroed):
		Nonce:      0,
		GasTipCap:  zero,
		GasFeeCap:  zero,
		Gas:        0,
		To:         &common.Address{},
		Value:      zero,
		AccessList: types.AccessList{},
		V:          zero,
		R:          zero,
		S:          zero,
	}
	return types.NewTx(txData)
}

// Creates a Hash from an Address by left-padding it to 32 bytes.
func extendAddress(address common.Address) (out common.Hash) {
	copy(out[12:], address[:])
	return out
}

// Returns a DepositEvent customized on the basis of the id parameter.
func GenerateDeposit(id uint) DepositEvent {
	return DepositEvent{
		From:       common.HexToAddress(fmt.Sprintf("0x42%d", id)),
		To:         common.HexToAddress(fmt.Sprintf("0x69%d", id)),
		Value:      newInt256(uint64(1000 + id)),
		GasLimit:   newInt256(uint64(42000 + id)),
		IsCreation: false,
		Data:       []byte{byte(id), 8, 9, 10},
	}
}

func bytes32(x *uint256.Int) []byte {
	bytes32 := x.Bytes32()
	return bytes32[:]
}

func newInt256(x uint64) (out uint256.Int) {
	out.SetUint64(x)
	return out
}

func toInt(b bool) uint64 {
	if b {
		return 1
	} else {
		return 0
	}
}

// Generates an EVM log entry that encodes a TransactionDeposited event from the deposit contract.
// Calls GenerateDeposit with `id` to generate the deposit.
func GenerateDepositLog(id uint) *types.Log {
	// generate topics & data for:
	//     event TransactionDeposited(
	//    	 address indexed from,
	//    	 address indexed to,
	//    	 uint256 value,
	//    	 uint256 gasLimit,
	//    	 bool isCreation,
	//    	 data data
	//     );
	// as specified by https://docs.soliditylang.org/en/v0.8.10/abi-spec.html?highlight=topics#events

	deposit := GenerateDeposit(id)

	topics := []common.Hash{
		DepositEventABIHash,
		extendAddress(deposit.From),
		extendAddress(deposit.To),
	}

	data := []byte{}
	data = append(data, bytes32(&deposit.Value)...)
	data = append(data, bytes32(&deposit.GasLimit)...)
	data = append(data, bytes32(uint256.NewInt(toInt(deposit.IsCreation)))...)
	data = append(data, bytes32(uint256.NewInt(uint64(128)))...) // (4 * 32 bytes) is the eventData offset
	data = append(data, bytes32(uint256.NewInt(uint64(len(deposit.Data))))...)
	data = append(data, deposit.Data...)

	return GenerateLog(DepositContractAddr, topics, data)
}

// Generates an EVM log entry with the given topics and data.
func GenerateLog(addr common.Address, topics []common.Hash, data []byte) *types.Log {
	return &types.Log{
		Address: addr,
		Topics:  topics,
		Data:    data,
		Removed: false,

		// ignored (zeroed):
		BlockNumber: 0,
		TxHash:      common.Hash{},
		TxIndex:     0,
		BlockHash:   common.Hash{},
		Index:       0,
	}
}

// Generates a receipt for a successful transaction with a single log entry for a deposit.
// Calls GenerateDeposit with `id` to generate the deposit.
func GenerateDepositReceipt(id uint) *types.Receipt {
	return GenerateReceipt(types.ReceiptStatusSuccessful, []*types.Log{GenerateDepositLog(id)})
}

// Generates a receipt for a failed transaction with a single log entry for a deposit.
// Calls GenerateDeposit with `id` to generate the deposit.
func GenerateFailedDepositReceipt(id uint) *types.Receipt {
	return GenerateReceipt(types.ReceiptStatusFailed, []*types.Log{GenerateDepositLog(id)})
}

// Generates a receipt with the given status and the given log entries.
func GenerateReceipt(status uint64, logs []*types.Log) *types.Receipt {
	receipt := &types.Receipt{
		Type:   types.DynamicFeeTxType,
		Status: status,
		Logs:   logs,

		// ignored (zeroed):
		PostState:         []byte{},
		CumulativeGasUsed: 0,
		TxHash:            common.Hash{},
		ContractAddress:   common.Address{},
		GasUsed:           0,
		BlockHash:         common.Hash{},
		BlockNumber:       zero,
		TransactionIndex:  0,
	}
	// not really needed
	receipt.Bloom = types.CreateBloom([]*types.Receipt{receipt})
	return receipt
}

// Generate an L1 block with the given transactions, receipts and timetamp.
func GenerateBlock(txs types.Transactions, receipts types.Receipts, _time time.Time) *types.Block {

	header := &types.Header{
		Time:  uint64(_time.Unix()),
		Extra: []byte{},

		// ignored (zeroed):
		ParentHash: common.Hash{},
		UncleHash:  common.Hash{},
		Coinbase:   common.Address{},
		Root:       common.Hash{},
		Difficulty: zero,
		Number:     zero,
		GasLimit:   0,
		MixDigest:  common.Hash{},
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
	input := GenerateInputItem(0, 0)
	output := Derive(input)
	assert.Equal(t, output.L1BlockHash, input.Block.Hash())
	assert.Equal(t, output.L1BlockTime, input.Block.Time())
	assert.Equal(t, output.L1Random, Bytes32(input.Block.MixDigest()))
	assert.Equal(t, len(output.UserDepositEvents), 0)
}

func assertDepositsEqual(t *testing.T, in *DepositEvent, out *DepositEvent) {
	assert.Equal(t, out.From, in.From)
	assert.Equal(t, out.To, in.To)
	assert.Equal(t, out.Value, in.Value)
	assert.Equal(t, out.GasLimit, in.GasLimit)
	assert.Equal(t, out.IsCreation, in.IsCreation)
	assert.Equal(t, out.Data, in.Data)
}

func TestSingleDeposit(t *testing.T) {
	input := GenerateInputItem(1, 0)
	output := Derive(input)
	assert.Equal(t, len(output.UserDepositEvents), 1)
	inDeposit := GenerateDeposit(0)
	outDeposit := output.UserDepositEvents[0]
	assertDepositsEqual(t, &inDeposit, outDeposit)
}

func TestFailedDeposit(t *testing.T) {
	input := GenerateInputItem(0, 1)
	output := Derive(input)
	assert.Equal(t, len(output.UserDepositEvents), 0)
}

func TestManyDeposits(t *testing.T) {
	nDeposits := uint(3)
	input := GenerateInputItem(nDeposits, nDeposits)
	output := Derive(input)
	assert.Equal(t, len(output.UserDepositEvents), 3)

	for id := uint(0); id < nDeposits; id++ {
		inDeposit := GenerateDeposit(id)
		outDeposit := output.UserDepositEvents[id]
		assertDepositsEqual(t, &inDeposit, outDeposit)
	}
}

// ================================================================================================
