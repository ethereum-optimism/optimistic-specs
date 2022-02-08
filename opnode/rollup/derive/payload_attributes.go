package derive

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
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

type ReceiptHash interface {
	ReceiptHash() common.Hash
}

// CheckReceipts sanity checks that the receipts are consistent with the block data.
func CheckReceipts(block ReceiptHash, receipts []*types.Receipt) bool {
	hasher := trie.NewStackTrie(nil)
	computed := types.DeriveSha(types.Receipts(receipts), hasher)
	return block.ReceiptHash() == computed
}

// UserDeposits transforms a L1 block and corresponding receipts into the transaction inputs for a full L2 block
func UserDeposits(height uint64, receipts []*types.Receipt) ([]*types.DepositTx, error) {
	var out []*types.DepositTx

	for _, rec := range receipts {
		if rec.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range rec.Logs {
			if log.Address == DepositContractAddr {
				// offset transaction index by 1, the first is the l1-info tx
				dep, err := UnmarshalLogEvent(height, uint64(len(out))+1, log)
				if err != nil {
					return nil, fmt.Errorf("malformatted L1 deposit log: %v", err)
				}
				out = append(out, dep)
			}
		}
	}
	return out, nil
}

type BlockInput interface {
	ReceiptHash
	ParentHash() common.Hash
	L1Info
	MixDigest() common.Hash
	Transactions() []*types.Transaction
	Receipts() []*types.Receipt
}

// PayloadAttributes derives the pre-execution payload from the L1 block info and deposit receipts.
// This is a pure function.
func PayloadAttributes(config *rollup.Config, ontoL1 eth.BlockID, lastL2Time uint64, blocks []BlockInput) ([]*l2.PayloadAttributes, error) {
	// check if we have the full sequencing window
	if uint64(len(blocks)) != config.SeqWindowSize {
		return nil, fmt.Errorf("expected %d blocks in sequencing window, got %d", config.SeqWindowSize, len(blocks))
	}
	// check if our inputs are consistent
	for i, block := range blocks {
		if block.NumberU64() != ontoL1.Number+1 {
			return nil, fmt.Errorf("sanity check failed, sequencing window blocks are not consecutive (index %d)", i)
		}
		// sanity check the parent hash
		if block.ParentHash() != ontoL1.Hash {
			return nil, fmt.Errorf("sanity check failed, sequencing window blocks are not chained  (index %d)", i)
		}
		ontoL1 = eth.BlockID{Hash: block.Hash(), Number: block.NumberU64()}
	}

	// Retrieve the deposits of this epoch (all deposits from the first block)
	deposits, err := DeriveDeposits(blocks[0])
	if err != nil {
		return nil, fmt.Errorf("failed to derive deposits: %v", err)
	}

	// All block times in L2 are config.BlockTime seconds apart. If L2 time moves faster than L1, use that.
	// L1 is expected to have larger increments per block, L2 time will get back in sync eventually.
	firstTimestamp := blocks[0].Time()
	if firstTimestamp >= config.Genesis.L2Time { // if not, then lastL2Time will be larger
		// round down, only create a trailing empty block if the L1 time is at least a full block span ahead of L2
		firstTimestamp = firstTimestamp - (firstTimestamp-config.Genesis.L2Time)%config.BlockTime
	}
	if lastL2Time+config.BlockTime > firstTimestamp {
		firstTimestamp = lastL2Time + config.BlockTime
	}

	// Sequencers may not use timestamps beyond the end of the sequencing window (with some configurable margin)
	maxTimestamp := blocks[len(blocks)-1].Time() + config.MaxSequencerTimeDiff

	l1Signer := config.L1Signer()

	// copy L1 randomness (mix-digest becomes randao field post-merge)
	// TODO: we don't have a randomness oracle on L2, what should sequencing randomness look like.
	// Repeating the latest randomness of L1 might not be ideal.
	randomnessSeed := l2.Bytes32(blocks[0].MixDigest())

	// Collect all L2 batches, the bathes may be out-of-order, or possibly missing.
	l2Blocks := make(map[uint64]*l2.PayloadAttributes)
	highestSeenTimestamp := firstTimestamp
	for _, block := range blocks {
		// scan the block for batches that match this epoch
		for _, tx := range block.Transactions() {
			if to := tx.To(); to != nil && *to == config.BatchInboxAddress {
				seqDataSubmitter, err := l1Signer.Sender(tx)
				if err != nil {
					continue // bad signature, ignore
				}
				// some random L1 user might have sent a transaction to our batch inbox, ignore them
				if seqDataSubmitter != config.BatchSenderAddress {
					continue // not an authorized batch submitter, ignore
				}
				batches := ParseBatches(tx.Data())
				for _, batch := range batches {
					if batch.Epoch != blocks[0].NumberU64() {
						continue // batch was tagged for future epoch (i.e. it depends on the given L1 block to be processed first)
					}
					if (batch.Timestamp-config.Genesis.L2Time)%config.BlockTime != 0 {
						continue // bad timestamp, not a multiple of the block time
					}
					if batch.Timestamp < firstTimestamp {
						continue // old batch
					}
					// limit timestamp upper bound to avoid huge amount of empty blocks
					if batch.Timestamp > maxTimestamp {
						continue // too far in future
					}
					// Check if we have already seen a batch for this L2 block
					if _, ok := l2Blocks[batch.Timestamp]; ok {
						// block already exists, batch is duplicate (first batch persists, others are ignored)
						continue
					}

					// Track the last batch we've seen (gaps will be filled with empty L2 blocks)
					if batch.Timestamp > highestSeenTimestamp {
						highestSeenTimestamp = batch.Timestamp
					}

					l2Blocks[batch.Timestamp] = &l2.PayloadAttributes{
						Timestamp:             l2.Uint64Quantity(batch.Timestamp),
						Random:                randomnessSeed,
						SuggestedFeeRecipient: config.FeeRecipientAddress,
						Transactions:          batch.Transactions,
					}
				}
			}
		}
	}

	// fill the gaps and always ensure at least one L2 block
	var out []*l2.PayloadAttributes
	for t := firstTimestamp; t <= highestSeenTimestamp; t += config.BlockTime {
		if bl, ok := l2Blocks[t]; ok {
			out = append(out, bl)
		} else {
			// skipped/missing L2 block, create an empty block instead
			out = append(out, &l2.PayloadAttributes{
				Timestamp:             l2.Uint64Quantity(t),
				Random:                randomnessSeed,
				SuggestedFeeRecipient: config.FeeRecipientAddress,
				Transactions:          nil,
			})
		}
	}

	// Force deposits into the first block
	out[0].Transactions = append(append(make([]l2.Data, 0), deposits...), out[0].Transactions...)

	return out, nil
}

func DeriveDeposits(block BlockInput) ([]l2.Data, error) {
	receipts := block.Receipts()
	if !CheckReceipts(block, receipts) {
		return nil, fmt.Errorf("receipts are not consistent with the receipts root %s of block %s (%d)", block.ReceiptHash(), block.Hash(), block.NumberU64())
	}
	l1Tx := types.NewTx(L1InfoDeposit(block))
	opaqueL1Tx, err := l1Tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode L1 info tx")
	}
	userDeposits, err := UserDeposits(block.NumberU64(), receipts)
	if err != nil {
		return nil, fmt.Errorf("failed to derive user deposits: %v", err)
	}
	encodedTxs := make([]l2.Data, 0, len(userDeposits)+1)
	encodedTxs = append(encodedTxs, opaqueL1Tx)
	for i, tx := range userDeposits {
		opaqueTx, err := types.NewTx(tx).MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to encode user tx %d", i)
		}
		encodedTxs = append(encodedTxs, opaqueTx)
	}
	return encodedTxs, nil
}

type BatchData struct {
	Epoch     uint64 // aka l1 num
	Timestamp uint64
	// no feeRecipient address input, all fees go to a L2 contract
	Transactions []l2.Data
}

func ParseBatches(data l2.Data) []BatchData {
	return nil // TODO
}
