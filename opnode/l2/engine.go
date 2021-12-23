package l2

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type DriverAPI interface {
	EngineAPI

	eth.NewHeadSource
	eth.HeaderSource
}

type Engine struct {
	Ctx context.Context
	Log log.Logger
	// API bindings to execution engine
	RPC EngineAPI

	// Locks the L1 and L2 head changes, to keep a consistent view of the engine
	headLock sync.RWMutex
	// l1Head tracks both number and hash, this simplifies L1 reorg detection
	l1Head eth.BlockID
	// l2Head tracks the head-block of the engine
	l2Head common.Hash
}

// L1 returns the block-id (hash and number) of the last L1 block that was derived into the L2 block
func (e *Engine) L1Head() eth.BlockID {
	e.headLock.RLock()
	defer e.headLock.RUnlock()
	return e.l1Head
}

// L2 returns the block-hash of the L2 chain head
func (e *Engine) L2Head() common.Hash {
	e.headLock.RLock()
	defer e.headLock.RUnlock()
	return e.l2Head
}

func (e *Engine) Head() (l1Head eth.BlockID, l2Head common.Hash) {
	e.headLock.RLock()
	defer e.headLock.RUnlock()
	return e.l1Head, e.l2Head
}

func (e *Engine) UpdateHead(l1Head eth.BlockID, l2Head common.Hash) {
	e.headLock.Lock()
	defer e.headLock.Unlock()
	e.l1Head = l1Head
	e.l2Head = l2Head
}

func (e *Engine) ProcessL1(dl l1.Downloader, newL1Head eth.BlockID, finalizedL2Block common.Hash) {
	e.headLock.Lock()
	defer e.headLock.Unlock()

	logger := e.Log.New(
		"prev_l1_nr", e.l1Head.Number, "prev_l1_hash", e.l1Head.Hash,
		"prev_l2_hash", e.l2Head,
		"l1_nr", newL1Head.Number, "l1_hash", newL1Head.Hash)

	if newL1Head == e.l1Head {
		// no-op, already processed it
		logger.Debug("skipping, engine already processed block")
		return
	}

	// check if it's in the far distance.
	// Header sync will need to move closer before we start fetching/caching full blocks.
	if newL1Head.Number > e.l1Head.Number+5 {
		logger.Info("Deferring processing, block too far out")
		return
	}
	ctx, cancel := context.WithTimeout(e.Ctx, time.Second*20)
	defer cancel()
	bl, receipts, err := dl.Fetch(ctx, newL1Head)
	if err != nil {
		logger.Warn("failed to fetch block with receipts")
		return
	}

	parentL2Hash := e.l2Head
	if bl.ParentHash() != e.l1Head.Hash {
		// not a simple extension, try fetch the corresponding L2 parent hash to this reorg
		// TODO
		//parentL2Hash = common.Hash{}
		return
	}

	if newL1Head.Number+1 == e.l1Head.Number { // extension or reorg
		if bl.ParentHash() != e.l1Head.Hash {
			// still good we fetched the block,
			// cache will make the reorg faster once we do get the connecting block.
			return
		}
	} else if newL1Head.Number > e.l1Head.Number {
		// Block is farther out, if far enough it would make sense to trigger a state-sync,
		// instead of waiting for header-sync to figure out the first block.

		// No state-sync for now
		// TODO: can trigger state-sync if we have full L2 head block (e.g. retrieved via p2p in sequencer rollup)
		return
	}

	attrs, err := DerivePayloadAttributes(bl, receipts)
	if err != nil {
		logger.Error("failed to derive execution payload inputs")
		return
	}

	preState := &ForkchoiceState{
		HeadBlockHash:      parentL2Hash, // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      parentL2Hash,
		FinalizedBlockHash: finalizedL2Block,
	}
	payload, err := DeriveBlock(ctx, e.RPC, preState, attrs)
	if err != nil {
		logger.Error("failed to derive execution payload")
		return
	}
	logger = logger.New("l2_nr", payload.BlockNumber, "l2_hash", payload.BlockHash)
	logger.Info("derived block")

	ctx, cancel = context.WithTimeout(e.Ctx, time.Second*5)
	defer cancel()
	execRes, err := e.RPC.ExecutePayload(ctx, payload)
	if err != nil {
		logger.Error("failed to execute payload")
		return
	}
	switch execRes.Status {
	case ExecutionValid:
		logger.Info("Executed new payload")
	case ExecutionSyncing:
		logger.Info("Failed to execute payload, node is syncing")
		return
	case ExecutionInvalid:
		logger.Error("Execution payload was INVALID! Ignoring bad block")
		return
	default:
		logger.Error("Unknown execution status")
		return
	}

	postState := &ForkchoiceState{
		HeadBlockHash:      payload.BlockHash, // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      payload.BlockHash,
		FinalizedBlockHash: finalizedL2Block,
	}

	ctx, cancel = context.WithTimeout(e.Ctx, time.Second*5)
	defer cancel()
	fcRes, err := e.RPC.ForkchoiceUpdated(ctx, postState, nil)
	if err != nil {
		logger.Error("failed to update forkchoice")
		return
	}
	switch fcRes.Status {
	case UpdateSyncing:
		logger.Info("updated forkchoice, but node is syncing")
		return
	case UpdateSuccess:
		logger.Info("updated forkchoice")
		e.l1Head = newL1Head
		e.l2Head = payload.BlockHash
		return
	}
}

func (e *Engine) Close() {
	e.RPC.Close()
}
