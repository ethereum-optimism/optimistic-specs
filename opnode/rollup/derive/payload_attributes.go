package derive

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

var (
	DepositEventABI     = "TransactionDeposited(address,address,uint256,uint256,uint256,bool,bytes)"
	DepositEventABIHash = crypto.Keccak256Hash([]byte(DepositEventABI))
	DepositContractAddr = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
	L1InfoFuncSignature = "setL1BlockValues(uint256 _number, uint256 _timestamp, uint256 _basefee, bytes32 _hash)"
	L1InfoFuncBytes4    = crypto.Keccak256([]byte(L1InfoFuncSignature))[:4]
	L1InfoPredeployAddr = common.HexToAddress("0x4242424242424242424242424242424242424242")
)

// UnmarshalLogEvent decodes an EVM log entry emitted by the deposit contract into typed deposit data.
//
// parse log data for:
//     event TransactionDeposited(
//    	 address indexed from,
//    	 address indexed to,
//       uint256 mint,
//    	 uint256 value,
//    	 uint256 gasLimit,
//    	 bool isCreation,
//    	 data data
//     );
//
// Deposits additionally get:
//  - blockNum matching the L1 block height
//  - txIndex: matching the deposit index, not L1 transaction index, since there can be multiple deposits per L1 tx
func UnmarshalLogEvent(blockNum uint64, txIndex uint64, ev *types.Log) (*types.DepositTx, error) {
	if len(ev.Topics) != 3 {
		return nil, fmt.Errorf("expected 3 event topics (event identity, indexed from, indexed to)")
	}
	if ev.Topics[0] != DepositEventABIHash {
		return nil, fmt.Errorf("invalid deposit event selector: %s, expected %s", ev.Topics[0], DepositEventABIHash)
	}
	if len(ev.Data) < 6*32 {
		return nil, fmt.Errorf("deposit event data too small (%d bytes): %x", len(ev.Data), ev.Data)
	}

	var dep types.DepositTx

	dep.BlockHeight = blockNum
	dep.TransactionIndex = txIndex

	// indexed 0
	dep.From = common.BytesToAddress(ev.Topics[1][12:])
	// indexed 1
	to := common.BytesToAddress(ev.Topics[2][12:])

	// unindexed data
	offset := uint64(0)
	dep.Value = new(big.Int).SetBytes(ev.Data[offset : offset+32])
	offset += 32

	dep.Mint = new(big.Int).SetBytes(ev.Data[offset : offset+32])
	// 0 mint is represented as nil to skip minting code
	if dep.Mint.Cmp(new(big.Int)) == 0 {
		dep.Mint = nil
	}
	offset += 32

	gas := new(big.Int).SetBytes(ev.Data[offset : offset+32])
	if !gas.IsUint64() {
		return nil, fmt.Errorf("bad gas value: %x", ev.Data[offset:offset+32])
	}
	offset += 32
	dep.Gas = gas.Uint64()
	// isCreation: If the boolean byte is 1 then dep.To will stay nil,
	// and it will create a contract using L2 account nonce to determine the created address.
	if ev.Data[offset+31] == 0 {
		dep.To = &to
	}
	offset += 32
	var dataOffset uint256.Int
	dataOffset.SetBytes(ev.Data[offset : offset+32])
	offset += 32
	if dataOffset.Eq(uint256.NewInt(128)) {
		return nil, fmt.Errorf("incorrect data offset: %v", dataOffset[0])
	}

	var dataLen uint256.Int
	dataLen.SetBytes(ev.Data[offset : offset+32])
	offset += 32

	if !dataLen.IsUint64() {
		return nil, fmt.Errorf("data too large: %s", dataLen.String())
	}
	// The data may be padded to a multiple of 32 bytes
	maxExpectedLen := uint64(len(ev.Data)) - offset
	dataLenU64 := dataLen.Uint64()
	if dataLenU64 > maxExpectedLen {
		return nil, fmt.Errorf("data length too long: %d, expected max %d", dataLenU64, maxExpectedLen)
	}

	// remaining bytes fill the data
	dep.Data = ev.Data[offset : offset+dataLenU64]

	return &dep, nil
}

type L1Info interface {
	NumberU64() uint64
	Time() uint64
	Hash() common.Hash
	BaseFee() *big.Int
	// MixDigest field, reused for randomness after The Merge (Bellatrix hardfork)
	MixDigest() common.Hash
	ReceiptHash() common.Hash
	ParentHash() common.Hash
}

// L1InfoDeposit creats a L1 Info deposit transaction based on the L1 block
func L1InfoDeposit(block L1Info) *types.DepositTx {
	data := make([]byte, 4+8+8+32+32)
	offset := 0
	copy(data[offset:4], L1InfoFuncBytes4)
	offset += 4
	binary.BigEndian.PutUint64(data[offset:offset+8], block.NumberU64())
	offset += 8
	binary.BigEndian.PutUint64(data[offset:offset+8], block.Time())
	offset += 8
	block.BaseFee().FillBytes(data[offset : offset+32])
	offset += 32
	copy(data[offset:offset+32], block.Hash().Bytes())

	return &types.DepositTx{
		BlockHeight:      block.NumberU64(),
		TransactionIndex: 0, // always the first transaction
		From:             DepositContractAddr,
		To:               &L1InfoPredeployAddr,
		Mint:             nil,
		Value:            big.NewInt(0),
		Gas:              99_999_999,
		Data:             data,
	}
}

// UserDeposits transforms the L2 block-height and L1 receipts into the transaction inputs for a full L2 block
func UserDeposits(l2BlockHeight uint64, receipts []*types.Receipt) ([]*types.DepositTx, error) {
	var out []*types.DepositTx

	for _, rec := range receipts {
		if rec.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range rec.Logs {
			if log.Address == DepositContractAddr {
				// offset transaction index by 1, the first is the l1-info tx
				dep, err := UnmarshalLogEvent(l2BlockHeight, uint64(len(out))+1, log)
				if err != nil {
					return nil, fmt.Errorf("malformatted L1 deposit log: %v", err)
				}
				out = append(out, dep)
			}
		}
	}
	return out, nil
}

func BatchesFromEVMTransactions(config *rollup.Config, txs []*types.Transaction) ([]*BatchData, error) {
	var out []*BatchData
	l1Signer := config.L1Signer()
	for _, tx := range txs {
		if to := tx.To(); to != nil && *to == config.BatchInboxAddress {
			seqDataSubmitter, err := l1Signer.Sender(tx)
			if err != nil {
				// TODO: log error
				continue // bad signature, ignore
			}
			// some random L1 user might have sent a transaction to our batch inbox, ignore them
			if seqDataSubmitter != config.BatchSenderAddress {
				// TODO: log/record metric
				continue // not an authorized batch submitter, ignore
			}
			batches, err := DecodeBatches(config, bytes.NewReader(tx.Data()))
			if err != nil {
				// TODO: log/record metric
				continue
			}
			out = append(out, batches...)
		}
	}
	return out, nil
}

func FilterBatches(config *rollup.Config, epoch rollup.Epoch, minL2Time uint64, maxL2Time uint64, batches []*BatchData) (out []*BatchData) {
	uniqueTime := make(map[uint64]struct{})
	for _, batch := range batches {
		if !ValidBatch(batch, config, epoch, minL2Time, maxL2Time) {
			continue
		}
		// Check if we have already seen a batch for this L2 block
		if _, ok := uniqueTime[batch.Timestamp]; ok {
			// block already exists, batch is duplicate (first batch persists, others are ignored)
			continue
		}
		uniqueTime[batch.Timestamp] = struct{}{}
		out = append(out, batch)
	}
	return
}

func ValidBatch(batch *BatchData, config *rollup.Config, epoch rollup.Epoch, minL2Time uint64, maxL2Time uint64) bool {
	if batch.Epoch != epoch {
		// Batch was tagged for past or future epoch,
		// i.e. it was included too late or depends on the given L1 block to be processed first.
		return false
	}
	if (batch.Timestamp-config.Genesis.L2Time)%config.BlockTime != 0 {
		return false // bad timestamp, not a multiple of the block time
	}
	if batch.Timestamp < minL2Time {
		return false // old batch
	}
	// limit timestamp upper bound to avoid huge amount of empty blocks
	if batch.Timestamp >= maxL2Time {
		return false // too far in future
	}
	for _, txBytes := range batch.Transactions {
		if len(txBytes) == 0 {
			return false // transaction data must not be empty
		}
		if txBytes[0] == types.DepositTxType {
			return false // sequencers may not embed any deposits into batch data
		}
	}
	return true
}

type L2Info interface {
	Time() uint64
}

// SortedAndPreparedBatches turns a collection of batches to the input batches for a series of blocks
func SortedAndPreparedBatches(batches []*BatchData, epoch, blockTime, minL2Time, maxL2Time uint64) []*BatchData {
	m := make(map[uint64]*BatchData)
	for _, b := range batches {
		m[b.BatchV1.Timestamp] = b
	}
	var out []*BatchData
	for t := minL2Time; t < maxL2Time; t += blockTime {
		b, ok := m[t]
		if ok {
			out = append(out, b)
		} else {
			out = append(out, &BatchData{
				BatchV1{
					Epoch:     rollup.Epoch(epoch),
					Timestamp: t,
				},
			})
		}
	}
	return out
}

func L1InfoDepositBytes(l1Info L1Info) (hexutil.Bytes, error) {
	l1Tx := types.NewTx(L1InfoDeposit(l1Info))
	opaqueL1Tx, err := l1Tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode L1 info tx")
	}
	return opaqueL1Tx, nil
}

func DeriveDeposits(l2BlockHeight uint64, receipts []*types.Receipt) ([]hexutil.Bytes, error) {
	userDeposits, err := UserDeposits(l2BlockHeight, receipts)
	if err != nil {
		return nil, fmt.Errorf("failed to derive user deposits: %v", err)
	}
	encodedTxs := make([]hexutil.Bytes, 0, len(userDeposits))
	for i, tx := range userDeposits {
		opaqueTx, err := types.NewTx(tx).MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to encode user tx %d", i)
		}
		encodedTxs = append(encodedTxs, opaqueTx)
	}
	return encodedTxs, nil
}
