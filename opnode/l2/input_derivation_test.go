package l2

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

// Generates a test case with nSuccess successful transactions and nFailed failed transactions,
// as well as corresponding deposits.
// The transactions are dummy transactions, that do not match the log entries contained the deposits.
func GenerateTest(nSuccess uint64, nFailed uint64, rng *rand.Rand) *deriveTxsTestInput {
	nTxs := nSuccess + nFailed
	txs := make(types.Transactions, nTxs)
	for i := range txs {
		txs[i] = GenerateTransaction()
	}

	// TODO: test receipts with multiple deposit logs.

	receipts := make(types.Receipts, nTxs)
	for i := uint64(0); i < nSuccess; i++ {
		receipts[i] = GenerateDepositReceipt(rng)
	}
	for i := nSuccess; i < nTxs; i++ {
		receipts[i] = GenerateFailedDepositReceipt(rng)
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

func GenerateAddress(rng *rand.Rand) (out common.Address) {
	rng.Read(out[:])
	return
}

func RandETH(rng *rand.Rand, max int64) *big.Int {
	x := big.NewInt(rng.Int63n(max))
	x = new(big.Int).Mul(x, big.NewInt(1e18))
	return x
}

// Returns a DepositEvent customized on the basis of the id parameter.
func GenerateDeposit(blockNum uint64, txIndex uint64, rng *rand.Rand) *types.DepositTx {
	dataLen := rng.Int63n(10_000)
	data := make([]byte, dataLen)
	rng.Read(data)

	var to *common.Address
	if rng.Intn(2) == 0 {
		x := GenerateAddress(rng)
		to = &x
	}
	var mint *big.Int
	if rng.Intn(2) == 0 {
		mint = RandETH(rng, 200)
	}

	dep := &types.DepositTx{
		BlockHeight:      blockNum,
		TransactionIndex: txIndex,
		From:             GenerateAddress(rng),
		To:               to,
		Value:            RandETH(rng, 200),
		Gas:              uint64(rng.Int63n(10 * 1e6)), // 10 M gas max
		Data:             data,
		Mint:             mint,
	}
	return dep
}

// Generates an EVM log entry that encodes a TransactionDeposited event from the deposit contract.
// Calls GenerateDeposit with random number generator to generate the deposit.
func GenerateDepositLog(deposit *types.DepositTx) *types.Log {

	toBytes := common.Hash{}
	if deposit.To != nil {
		toBytes = deposit.To.Hash()
	}
	topics := []common.Hash{
		DepositEventABIHash,
		deposit.From.Hash(),
		toBytes,
	}

	data := make([]byte, 6*32)
	offset := 0
	deposit.Value.FillBytes(data[offset : offset+32])
	offset += 32

	if deposit.Mint != nil {
		deposit.Mint.FillBytes(data[offset : offset+32])
	}
	offset += 32

	binary.BigEndian.PutUint64(data[offset+24:offset+32], deposit.Gas)
	offset += 32
	if deposit.To == nil { // isCreation
		data[offset+31] = 1
	}
	offset += 32
	binary.BigEndian.PutUint64(data[offset+24:offset+32], 5*32)
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
func GenerateDepositReceipt(rng *rand.Rand) *types.Receipt {
	return GenerateReceipt(types.ReceiptStatusSuccessful, []*types.Log{
		GenerateDepositLog(GenerateDeposit(rng.Uint64(), rng.Uint64(), rng)),
	})
}

// Generates a receipt for a failed transaction with a single log entry for a deposit.
// Calls GenerateDeposit with `id` to generate the deposit.
func GenerateFailedDepositReceipt(rng *rand.Rand) *types.Receipt {
	return GenerateReceipt(types.ReceiptStatusFailed, []*types.Log{
		GenerateDepositLog(GenerateDeposit(rng.Uint64(), rng.Uint64(), rng)),
	})
}

// Generates a receipt with the given status and the given log entries.
func GenerateReceipt(status uint64, logs []*types.Log) *types.Receipt {
	return &types.Receipt{
		Type:   types.DynamicFeeTxType,
		Status: status,
		Logs:   logs,
	}
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

func TestUnmarshalLogEvent(t *testing.T) {
	for i := int64(0); i < 100; i++ {
		t.Run(fmt.Sprintf("random_deposit_%d", i), func(t *testing.T) {
			rng := rand.New(rand.NewSource(1234 + i))
			blockNum := rng.Uint64()
			txIndex := uint64(rng.Intn(10000))
			depInput := GenerateDeposit(blockNum, txIndex, rng)
			log := GenerateDepositLog(depInput)
			depOutput, err := UnmarshalLogEvent(blockNum, txIndex, log)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, depInput, depOutput)
		})
	}
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
			got, err := DeriveBlockInputs(testCase.input)
			assert.NoError(t, err)
			assert.Equal(t, got, testCase.expected)
		})
	}
}
*/
