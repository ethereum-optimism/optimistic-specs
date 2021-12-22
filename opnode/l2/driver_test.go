package l2

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	// testnet chain ID, not meaningful here
	chaindID = big.NewInt(69)
)

type deriveTxsTestInput struct {
	block    *types.Block
	receipts []*types.Receipt
}

// Generates an test case with nSuccess successful transactions and nFailed failed transactions,
// as well as corresponding deposits.
// The transactions are dummy transactions, that do not match the log entries contained the deposits.
func GenerateTest(nSuccess uint64, nFailed uint64) *deriveTxsTestInput {
	nTxs := nSuccess + nFailed
	txs := make(types.Transactions, nTxs)
	for i := range txs {
		txs[i] = GenerateTransaction()
	}

	// TODO: test receipts with multiple deposit logs.

	receipts := make(types.Receipts, nTxs)
	for i := uint64(0); i < nSuccess; i++ {
		receipts[i] = GenerateDepositReceipt(i)
	}
	for i := nSuccess; i < nTxs; i++ {
		receipts[i] = GenerateFailedDepositReceipt(i)
	}

	block := GenerateBlock(txs, receipts, time.Now())
	return &deriveTxsTestInput{
		block:    block,
		receipts: receipts,
	}
}

// Generates a dummy transaction with most fields zeroed.
func GenerateTransaction() *types.Transaction {
	txData := &types.DynamicFeeTx{
		ChainID: chaindID,
		Data:    []byte{},

		// ignored (zeroed):
		Nonce:      0,
		GasTipCap:  new(big.Int),
		GasFeeCap:  new(big.Int),
		Gas:        0,
		To:         &common.Address{},
		Value:      new(big.Int),
		AccessList: types.AccessList{},
		V:          new(big.Int),
		R:          new(big.Int),
		S:          new(big.Int),
	}
	return types.NewTx(txData)
}

// Returns a DepositEvent customized on the basis of the id parameter.
func GenerateDeposit(blockNum uint64, txIndex uint64, id uint64) *types.DepositTx {
	to := common.HexToAddress(fmt.Sprintf("0x69%16x", id))
	return &types.DepositTx{
		BlockHeight:      blockNum,
		TransactionIndex: txIndex,
		From:             common.HexToAddress(fmt.Sprintf("0x42%16x", id)),
		To:               &to,
		Value:            big.NewInt(1000 + int64(id)),
		Gas:              42000 + id,
		Data:             []byte{byte(id), 8, 9, 10},
		// TODO Mint field
	}
}

// Generates an EVM log entry that encodes a TransactionDeposited event from the deposit contract.
// Calls GenerateDeposit with `id` to generate the deposit.
func GenerateDepositLog(id uint64) *types.Log {
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

	deposit := GenerateDeposit(0, 0, id)

	toBytes := common.Hash{}
	if deposit.To != nil {
		toBytes = deposit.To.Hash()
	}
	topics := []common.Hash{
		DepositEventABIHash,
		deposit.From.Hash(),
		toBytes,
	}

	data := make([]byte, 128)
	offset := 0
	copy(data[0:offset+32], deposit.Value.Bytes())
	offset += 32
	binary.BigEndian.PutUint64(data[offset+24:offset+32], deposit.Gas)
	offset += 32
	if deposit.To == nil { // isCreation
		data[offset+31] = 1
	}
	offset += 32
	binary.BigEndian.PutUint64(data[offset+24:offset+32], 128)
	offset += 32
	binary.BigEndian.PutUint64(data[offset+24:offset+32], uint64(len(deposit.Data)))
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
func GenerateDepositReceipt(id uint64) *types.Receipt {
	return GenerateReceipt(types.ReceiptStatusSuccessful, []*types.Log{GenerateDepositLog(id)})
}

// Generates a receipt for a failed transaction with a single log entry for a deposit.
// Calls GenerateDeposit with `id` to generate the deposit.
func GenerateFailedDepositReceipt(id uint64) *types.Receipt {
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
		BlockNumber:       new(big.Int),
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
		Difficulty: new(big.Int),
		Number:     new(big.Int),
		GasLimit:   0,
		MixDigest:  common.Hash{},
		Nonce:      types.EncodeNonce(0),
		BaseFee:    new(big.Int),

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

/*
type DeriveUserDepositsTestCase struct {
	name     string
	input    *deriveTxsTestInput
	expected []*types.Transaction
}

func TestDeriveUserDeposits(t *testing.T) {
	testCases := []DeriveUserDepositsTestCase{
		{"no deposits", GenerateTest(0, 0), []*types.Transaction{}},
		{"success deposit", GenerateTest(1, 0), []*types.Transaction{nil}}, // TODO
		{"failed deposit", GenerateTest(0, 1), []*types.Transaction{}},
		{"many deposits", nil, nil}, // TODO
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := DeriveUserDeposits(testCase.input.block, testCase.input.receipts)
			assert.NoError(t, err)
			assert.Equal(t, got, testCase.expected)
		})
	}
}

type DeriveL1InfoDepositTestCase struct {
	name     string
	input    *types.Block
	expected *types.DepositTx
}

func TestDeriveL1InfoDeposit(t *testing.T) {
	testCases := []DeriveL1InfoDepositTestCase{
		// TODO
		{"random block", GenerateBlock()},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := DeriveL1InfoDeposit(testCase.input)
			assert.Equal(t, got, testCase.expected)
		})
	}
}

type DerivePayloadAttributesTestCase struct {
	name     string
	input    *deriveTxsTestInput
	expected *PayloadAttributes
}

func TestDerivePayloadAttributes(t *testing.T) {
	testCases := []DerivePayloadAttributesTestCase{
		// TODO
		{"random block", GenerateBlock(), &PayloadAttributes{}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := DerivePayloadAttributes(testCase.input)
			assert.NoError(t, err)
			assert.Equal(t, got, testCase.expected)
		})
	}
}
*/
