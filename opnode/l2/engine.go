package l2

import (
	"context"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"time"
)

type L2Engine struct {
	Ctx context.Context
	Log log.Logger
	// Raw RPC client, separate bindings
	RPC *rpc.Client
	// track where the l2 engine is at
	L1Head l1.BlockID
}

func (c *L2Engine) LastL1() l1.BlockID {
	return c.L1Head
}

func (c *L2Engine) ProcessL1(dl l1.Downloader, finalized common.Hash, id l1.BlockID) {
	if id == c.L1Head {
		// no-op, already processed it
		return
	}
	// check if it's in the far distance.
	// Header sync will need to move closer before we start fetching/caching full blocks.
	if id.Number > c.L1Head.Number+5 {
		c.Log.Info("Deferring processing, block too far out", "last", c.L1Head.Number, "nr", id.Number, "hash", id.Hash)
		return
	}
	ctx, _ := context.WithTimeout(c.Ctx, time.Second*20)
	bl, receipts, err := dl.Fetch(ctx, id)
	if err != nil {
		c.Log.Warn("failed to fetch block with receipts", "nr", id.Number, "hash", id.Hash)
		return
	}

	if id.Number+1 == c.L1Head.Number { // extension or reorg
		if bl.ParentHash() != c.L1Head.Hash {
			// still good we fetched the block,
			// cache will make the reorg faster once we do get the connecting block.
			return
		}
	} else if id.Number > c.L1Head.Number {
		// Block is farther out, if far enough it would make sense to trigger a state-sync,
		// instead of waiting for header-sync to figure out the first block.

		// No state-sync for now
		// TODO: can trigger state-sync if we have full L2 head block (e.g. retrieved via p2p in sequencer rollup)
		return
	}

	payload, err := PlaceholderDerive(ctx, bl, receipts)
	if err != nil {
		c.Log.Error("failed to derive execution payload", "nr", id.Number, "hash", id.Hash)
		return
	}

	ctx, _ = context.WithTimeout(c.Ctx, time.Second*5)
	execRes, err := ExecutePayload(ctx, c.RPC, c.Log, payload)
	if err != nil {
		c.Log.Error("failed to execute payload", "nr", id.Number, "hash", id.Hash)
		return
	}
	switch execRes.Status {
	case ExecutionValid:
		c.Log.Info("Executed new payload", "nr", id.Number, "hash", id.Hash)
		break
	case ExecutionSyncing:
		c.Log.Info("Failed to execute payload, node is syncing", "nr", id.Number, "hash", id.Hash)
		return
	case ExecutionInvalid:
		c.Log.Error("Execution payload was INVALID! Ignoring bad block", "nr", id.Number, "hash", id.Hash)
		return
	}

	ctx, _ = context.WithTimeout(c.Ctx, time.Second*5)
	fcRes, err := ForkchoiceUpdated(ctx, c.RPC, c.Log, &ForkchoiceState{
		HeadBlockHash:      Bytes32(id.Hash), // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      Bytes32(id.Hash),
		FinalizedBlockHash: Bytes32(finalized),
	}, nil)
	if err != nil {
		c.Log.Error("failed to update forkchoice", "nr", id.Number, "hash", id.Hash)
		return
	}
	switch fcRes.Status {
	case UpdateSyncing:
		c.Log.Info("updated forkchoice, but node is syncing", "nr", id.Number, "hash", id.Hash)
		return
	case UpdateSuccess:
		c.Log.Info("updated forkchoice", "nr", id.Number, "hash", id.Hash)
		c.L1Head = id
		return
	}
}

func (c *L2Engine) Close() {
	c.RPC.Close()
}
