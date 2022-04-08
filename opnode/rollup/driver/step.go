package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type outputImpl struct {
	dl     Downloader
	l2     Engine
	log    log.Logger
	Config rollup.Config
}

// isDepositTx checks an opaqueTx to determine if it is a Deposit Trransaction
// It has to return an error in the case the transaction is empty
func isDepositTx(opaqueTx l2.Data) (bool, error) {
	if len(opaqueTx) == 0 {
		return false, errors.New("empty transaction")
	}
	return opaqueTx[0] == types.DepositTxType, nil
}

// lastDeposit finds the index of last deposit at the start of the transactions.
// It walks the transactions from the start until it finds a non-deposit tx.
// An error is returned if any looked at transaction cannot be decoded
func lastDeposit(txns []l2.Data) (int, error) {
	var lastDeposit int
	for i, tx := range txns {
		deposit, err := isDepositTx(tx)
		if err != nil {
			return 0, fmt.Errorf("invalid transaction at idx %d", i)
		}
		if deposit {
			lastDeposit = i
		} else {
			break
		}
	}
	return lastDeposit, nil
}

func (d *outputImpl) createNewBlock(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.BlockID, l2Finalized eth.BlockID, l1Origin eth.L1BlockRef) (eth.L2BlockRef, *derive.BatchData, error) {
	d.log.Info("creating new block", "l2Head", l2Head)

	// If the L1 origin changed this block, then we are in the first block of the epoch
	firstEpochBlock := l2Head.L1Origin.Number != l1Origin.Number

	fetchCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	var l1Info derive.L1Info
	var receipts types.Receipts
	var err error
	// Include deposits if this is the first block of an epoch
	if firstEpochBlock {
		l1Info, _, receipts, err = d.dl.Fetch(fetchCtx, l1Origin.Hash)
	} else {
		l1Info, err = d.dl.InfoByHash(fetchCtx, l1Origin.Hash)
		// don't fetch receipts if we do not process deposits
	}
	if err != nil {
		return l2Head, nil, fmt.Errorf("failed to fetch L1 block info of %s: %v", l1Origin, err)
	}

	seqNumber := l2Head.Number + 1 - l2SafeHead.Number
	l1InfoTx, err := derive.L1InfoDepositBytes(seqNumber, l1Info)
	if err != nil {
		return l2Head, nil, err
	}
	var txns []l2.Data
	txns = append(txns, l1InfoTx)
	deposits, err := derive.DeriveDeposits(receipts)
	d.log.Info("Derived deposits", "deposits", deposits, "l2Parent", l2Head, "l1Origin", l1Origin)
	if err != nil {
		return l2Head, nil, fmt.Errorf("failed to derive deposits: %v", err)
	}
	txns = append(txns, deposits...)

	attrs := &l2.PayloadAttributes{
		Timestamp:             hexutil.Uint64(l2Head.Time + d.Config.BlockTime),
		PrevRandao:            l2.Bytes32(l1Info.MixDigest()),
		SuggestedFeeRecipient: d.Config.FeeRecipientAddress,
		Transactions:          txns,
		NoTxPool:              false,
	}
	fc := l2.ForkchoiceState{
		HeadBlockHash:      l2Head.Hash,
		SafeBlockHash:      l2SafeHead.Hash,
		FinalizedBlockHash: l2Finalized.Hash,
	}

	payload, lastDeposit, err := d.insertHeadBlock(ctx, fc, attrs, false)
	if err != nil {
		return l2Head, nil, fmt.Errorf("failed to extend L2 chain: %v", err)
	}

	batch := &derive.BatchData{
		BatchV1: derive.BatchV1{
			Epoch:        rollup.Epoch(l1Info.NumberU64()),
			Timestamp:    uint64(payload.Timestamp),
			Transactions: payload.Transactions[lastDeposit+1:],
		},
	}
	ref, err := l2.PayloadToBlockRef(payload, &d.Config.Genesis)
	return ref, batch, err
}

// insertEpoch creates and inserts one epoch on top of the safe head. It prefers blocks it creates to what is recorded in the unsafe chain.
// It returns the new L2 head and L2 Safe head and if there was a reorg. This function must return if there was a reorg otherwise the L2 chain must be traversed.
func (d *outputImpl) insertEpoch(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.L2BlockRef, l2Finalized eth.BlockID, l1Input []eth.BlockID) (eth.L2BlockRef, eth.L2BlockRef, bool, error) {
	// Sanity Checks
	if len(l1Input) <= 1 {
		return l2Head, l2SafeHead, false, fmt.Errorf("too small L1 sequencing window for L2 derivation on %s: %v", l2SafeHead, l1Input)
	}
	if len(l1Input) != int(d.Config.SeqWindowSize) {
		return l2Head, l2SafeHead, false, errors.New("invalid sequencing window size")
	}

	logger := d.log.New("input_l1_first", l1Input[0], "input_l1_last", l1Input[len(l1Input)-1], "input_l2_parent", l2SafeHead, "finalized_l2", l2Finalized)
	logger.Trace("Running update step on the L2 node")

	// Get inputs from L1 and L2
	epoch := rollup.Epoch(l1Input[0].Number)
	fetchCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	l2Info, err := d.l2.PayloadByHash(fetchCtx, l2SafeHead.Hash)
	if err != nil {
		return l2Head, l2SafeHead, false, fmt.Errorf("failed to fetch L2 block info of %s: %w", l2SafeHead, err)
	}
	l1Info, _, receipts, err := d.dl.Fetch(fetchCtx, l1Input[0].Hash)
	if err != nil {
		return l2Head, l2SafeHead, false, fmt.Errorf("failed to fetch L1 block info of %s: %w", l1Input[0], err)
	}
	if l2SafeHead.L1Origin.Hash != l1Info.ParentHash() {
		return l2Head, l2SafeHead, false, fmt.Errorf("l1Info %v does not extend L1 Origin (%v) of L2 Safe Head (%v)", l1Info.Hash(), l2SafeHead.L1Origin, l2SafeHead)
	}
	nextL1Block, err := d.dl.InfoByHash(ctx, l1Input[1].Hash)
	if err != nil {
		return l2Head, l2SafeHead, false, fmt.Errorf("failed to get L1 timestamp of next L1 block: %v", err)
	}
	deposits, err := derive.DeriveDeposits(receipts)
	if err != nil {
		return l2Head, l2SafeHead, false, fmt.Errorf("failed to derive deposits: %w", err)
	}
	// TODO: with sharding the blobs may be identified in more detail than L1 block hashes
	transactions, err := d.dl.FetchAllTransactions(fetchCtx, l1Input)
	if err != nil {
		return l2Head, l2SafeHead, false, fmt.Errorf("failed to fetch transactions from %s: %v", l1Input, err)
	}
	batches, err := derive.BatchesFromEVMTransactions(&d.Config, transactions)
	if err != nil {
		return l2Head, l2SafeHead, false, fmt.Errorf("failed to fetch create batches from transactions: %w", err)
	}
	// Make batches contiguous
	minL2Time := uint64(l2Info.Timestamp) + d.Config.BlockTime
	maxL2Time := l1Info.Time() + d.Config.MaxSequencerDrift
	if minL2Time+d.Config.BlockTime > maxL2Time {
		maxL2Time = minL2Time + d.Config.BlockTime
	}
	batches = derive.FilterBatches(&d.Config, epoch, minL2Time, maxL2Time, batches)
	batches = derive.FillMissingBatches(batches, uint64(epoch), d.Config.BlockTime, minL2Time, nextL1Block.Time())

	fc := l2.ForkchoiceState{
		HeadBlockHash:      l2Head.Hash,
		SafeBlockHash:      l2SafeHead.Hash,
		FinalizedBlockHash: l2Finalized.Hash,
	}
	// Execute each L2 block in the epoch
	lastHead := l2Head
	lastSafeHead := l2SafeHead
	didReorg := false
	var payload *l2.ExecutionPayload
	var reorg bool
	for i, batch := range batches {
		var txns []l2.Data
		l1InfoTx, err := derive.L1InfoDepositBytes(uint64(i), l1Info)
		if err != nil {
			return l2Head, l2SafeHead, false, fmt.Errorf("failed to create l1InfoTx: %w", err)
		}
		txns = append(txns, l1InfoTx)
		if i == 0 {
			txns = append(txns, deposits...)
		}
		txns = append(txns, batch.Transactions...)
		attrs := &l2.PayloadAttributes{
			Timestamp:             hexutil.Uint64(batch.Timestamp),
			PrevRandao:            l2.Bytes32(l1Info.MixDigest()),
			SuggestedFeeRecipient: d.Config.FeeRecipientAddress,
			Transactions:          txns,
			// we are verifying, not sequencing, we've got all transactions and do not pull from the tx-pool
			// (that would make the block derivation non-deterministic)
			NoTxPool: true,
		}

		// We are either verifying blocks (with a potential for a reorg) or inserting a safe head to the chain
		if lastHead.Hash != lastSafeHead.Hash {
			payload, reorg, err = d.verifySafeBlock(ctx, fc, attrs, lastSafeHead.ID())

		} else {
			payload, _, err = d.insertHeadBlock(ctx, fc, attrs, true)
		}
		if err != nil {
			return lastHead, lastSafeHead, didReorg, fmt.Errorf("failed to extend L2 chain at block %d/%d of epoch %d: %w", i, len(batches), epoch, err)
		}

		newLast, err := l2.PayloadToBlockRef(payload, &d.Config.Genesis)
		if err != nil {
			return lastHead, lastSafeHead, didReorg, fmt.Errorf("failed to derive block references: %w", err)
		}
		if reorg {
			didReorg = true
		}
		// If reorg or the L2 Head is not ahead of the safe head, bump the head block.
		if reorg || lastHead.Hash == lastSafeHead.Hash {
			lastHead = newLast
		}
		lastSafeHead = newLast

		fc.HeadBlockHash = lastHead.Hash
		fc.SafeBlockHash = lastSafeHead.Hash
	}

	return lastHead, lastSafeHead, didReorg, nil
}

// attributesMatchBlock checks if the L2 attributes pre-inputs match the output
// nil if it is a match. If err is not nil, the error contains the reason for the mismatch
func attributesMatchBlock(attrs *l2.PayloadAttributes, parentHash common.Hash, block *l2.ExecutionPayload) error {
	if parentHash != block.ParentHash {
		return fmt.Errorf("parent hash field does not match. expected: %v. got: %v", parentHash, block.ParentHash)
	}
	if attrs.Timestamp != block.Timestamp {
		return fmt.Errorf("timestamp field does not match. expected: %v. got: %v", uint64(attrs.Timestamp), block.Timestamp)
	}
	if attrs.PrevRandao != block.PrevRandao {
		return fmt.Errorf("random field does not match. expected: %v. got: %v", attrs.PrevRandao, block.PrevRandao)
	}
	if len(attrs.Transactions) != len(block.Transactions) {
		return fmt.Errorf("transaction count does not match. expected: %v. got: %v", len(attrs.Transactions), block.Transactions)
	}
	for i, otx := range attrs.Transactions {
		if expect := block.Transactions[i]; !bytes.Equal(otx, expect) {
			return fmt.Errorf("transaction %d does not match. expected: %x. got: %x", i, expect, otx)
		}
	}
	return nil
}

// verifySafeBlock reconciles the supplied payload attributes against the actual L2 block.
// If they do not match, it inserts the new block and sets the head and safe head to the new block in the FC.
func (d *outputImpl) verifySafeBlock(ctx context.Context, fc l2.ForkchoiceState, attrs *l2.PayloadAttributes, parent eth.BlockID) (*l2.ExecutionPayload, bool, error) {
	payload, err := d.l2.PayloadByNumber(ctx, new(big.Int).SetUint64(parent.Number+1))
	if err != nil {
		return nil, false, fmt.Errorf("failed to get L2 block: %w", err)
	}
	err = attributesMatchBlock(attrs, parent.Hash, payload)
	if err != nil {
		// Have reorg
		d.log.Warn("Detected L2 reorg when verifying L2 safe head", "parent", parent, "prev_block", payload.BlockHash, "mismatch", err)
		fc.HeadBlockHash = parent.Hash
		fc.SafeBlockHash = parent.Hash
		payload, _, err := d.insertHeadBlock(ctx, fc, attrs, true)
		return payload, true, err
	}
	// If the attributes match, just bump the safe head
	d.log.Debug("Verified L2 block", "number", payload.BlockNumber, "hash", payload.BlockHash)
	fc.SafeBlockHash = payload.BlockHash
	_, err = d.l2.ForkchoiceUpdate(ctx, &fc, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute ForkchoiceUpdated: %w", err)
	}
	return payload, false, nil
}

// insertHeadBlock creates, executes, and inserts the specified block as the head block.
// It first uses the given FC to start the block creation process and then after the payload is executed,
// sets the FC to the same safe and finalized hashes, but updates the head hash to the new block.
// If updateSafe is true, the head block is considered to be the safe head as well as the head.
// It returns the payload, the count of deposits, and an error.
func (d *outputImpl) insertHeadBlock(ctx context.Context, fc l2.ForkchoiceState, attrs *l2.PayloadAttributes, updateSafe bool) (*l2.ExecutionPayload, int, error) {
	fcRes, err := d.l2.ForkchoiceUpdate(ctx, &fc, attrs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create new block via forkchoice: %w", err)
	}
	id := fcRes.PayloadID
	if id == nil {
		return nil, 0, errors.New("nil id in forkchoice result when expecting a valid ID")
	}
	payload, err := d.l2.GetPayload(ctx, *id)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get execution payload: %w", err)
	}
	// Sanity check payload before inserting it
	if len(payload.Transactions) == 0 {
		return nil, 0, errors.New("no transactions in returned payload")
	}
	if payload.Transactions[0][0] != types.DepositTxType {
		return nil, 0, fmt.Errorf("first transaction was not deposit tx. Got %v", payload.Transactions[0][0])
	}
	// Ensure that the deposits are first
	lastDeposit, err := lastDeposit(payload.Transactions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find last deposit: %w", err)
	}
	// Ensure no deposits after last deposit
	for i := lastDeposit + 1; i < len(payload.Transactions); i++ {
		tx := payload.Transactions[i]
		deposit, err := isDepositTx(tx)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decode transaction idx %d: %w", i, err)
		}
		if deposit {
			d.log.Error("Produced an invalid block where the deposit txns are not all at the start of the block", "tx_idx", i, "lastDeposit", lastDeposit)
			return nil, 0, fmt.Errorf("deposit tx (%d) after other tx in l2 block with prev deposit at idx %d", i, lastDeposit)
		}
	}
	// If this is an unsafe block, it has deposits & transactions included from L2.
	// Record if the execution engine dropped deposits. The verification process would see a mismatch
	// between attributes and the block, but then execute the correct block.
	if !updateSafe && lastDeposit+1 != len(attrs.Transactions) {
		d.log.Error("Dropped deposits when executing L2 block")
	}

	err = d.l2.NewPayload(ctx, payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to insert execution payload: %w", err)
	}
	fc.HeadBlockHash = payload.BlockHash
	if updateSafe {
		fc.SafeBlockHash = payload.BlockHash
	}
	d.log.Debug("Inserted L2 head block", "number", uint64(payload.BlockNumber), "hash", payload.BlockHash, "update_safe", updateSafe)
	_, err = d.l2.ForkchoiceUpdate(ctx, &fc, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to make the new L2 block canonical via forkchoice: %w", err)
	}
	return payload, lastDeposit, nil
}
