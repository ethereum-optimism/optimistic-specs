package l2

import (
	"context"
	"fmt"
	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"
)

// RefByL2Num fetches the L1 and L2 block IDs from the engine for the given L2 block height.
// Use a nil height to fetch the head.
func RefByL2Num(ctx context.Context, src eth.BlockByNumberSource, l2Num *big.Int, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, parentL2 common.Hash, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	refL2Block, err2 := src.BlockByNumber(ctx, l2Num) // nil for latest block
	if err2 != nil {
		err = fmt.Errorf("failed to retrieve head L2 block: %v", err2)
		return
	}
	return ParseL2Block(refL2Block, genesis)
}

// RefByL2Hash fetches the L1 and L2 block IDs from the engine for the given L2 block height.
// Use a nil height to fetch the head.
func RefByL2Hash(ctx context.Context, src eth.BlockByHashSource, l2Hash common.Hash, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, parentL2 common.Hash, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	refL2Block, err2 := src.BlockByHash(ctx, l2Hash)
	if err2 != nil {
		err = fmt.Errorf("failed to retrieve head L2 block: %v", err2)
		return
	}
	return ParseL2Block(refL2Block, genesis)
}

func DriverStep(ctx context.Context, log log.Logger, rpc DriverAPI,
	dl l1.Downloader, l1Input eth.BlockID, l2Parent eth.BlockID, l2Finalized common.Hash) (out eth.BlockID, err error) {

	logger := log.New(
		"input_l1", l1Input,
		"input_l2_parent", l2Parent,
		"finalized_l2", l2Finalized)

	fetchCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	bl, receipts, err := dl.Fetch(fetchCtx, l1Input)
	if err != nil {
		return eth.BlockID{}, fmt.Errorf("failed to fetch block with receipts: %v", err)
	}

	attrs, err := DerivePayloadAttributes(bl, receipts)
	if err != nil {
		return eth.BlockID{}, fmt.Errorf("failed to derive execution payload inputs: %v", err)
	}

	preState := &ForkchoiceState{
		HeadBlockHash:      l2Parent.Hash, // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      l2Parent.Hash,
		FinalizedBlockHash: l2Finalized,
	}
	payload, err := DeriveBlock(ctx, rpc, preState, attrs)
	if err != nil {
		return eth.BlockID{}, fmt.Errorf("failed to derive execution payload: %v", err)
	}
	l2ID := eth.BlockID{Hash: payload.BlockHash, Number: uint64(payload.BlockNumber)}
	logger = logger.New("derived_l2", l2ID)
	logger.Info("derived block")

	execCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	execRes, err := rpc.ExecutePayload(execCtx, payload)
	if err != nil {
		return l2ID, fmt.Errorf("failed to execute payload: %v", err)
	}
	switch execRes.Status {
	case ExecutionValid:
		logger.Info("Executed new payload")
	case ExecutionSyncing:
		return l2ID, fmt.Errorf("failed to execute payload %s, node is syncing, latest valid hash is %s", l2ID, execRes.LatestValidHash)
	case ExecutionInvalid:
		return l2ID, fmt.Errorf("execution payload %s was INVALID! Latest valid hash is %s, ignoring bad block: %q", l2ID, execRes.LatestValidHash, execRes.ValidationError)
	default:
		return l2ID, fmt.Errorf("unknown execution status on %s: %q, ", l2ID, string(execRes.Status))
	}

	postState := &ForkchoiceState{
		HeadBlockHash:      payload.BlockHash, // no difference yet between Head and Safe, no data ahead of L1 yet.
		SafeBlockHash:      payload.BlockHash,
		FinalizedBlockHash: l2Finalized,
	}

	fcCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	fcRes, err := rpc.ForkchoiceUpdated(fcCtx, postState, nil)
	if err != nil {
		return l2ID, fmt.Errorf("failed to update forkchoice: %v", err)
	}
	switch fcRes.Status {
	case UpdateSyncing:
		return l2ID, fmt.Errorf("updated forkchoice, but node is syncing: %v", err)
	case UpdateSuccess:
		logger.Info("updated forkchoice")
		return l2ID, nil
	default:
		return l2ID, fmt.Errorf("unknown forkchoice status on %s: %q, ", l2ID, string(fcRes.Status))
	}
}
