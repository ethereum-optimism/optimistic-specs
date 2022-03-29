package l2

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type Source struct {
	rpc     *rpc.Client       // raw RPC client. Used for the consensus namespace
	client  *ethclient.Client // go-ethereum's wrapper around the rpc client for the eth namespace
	genesis *rollup.Genesis
	log     log.Logger
}

func NewSource(ll2Node *rpc.Client, genesis *rollup.Genesis, log log.Logger) (*Source, error) {
	return &Source{
		rpc:     ll2Node,
		client:  ethclient.NewClient(ll2Node),
		genesis: genesis,
		log:     log,
	}, nil
}

func (s *Source) Close() {
	s.rpc.Close()
}

func (s *Source) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return s.client.BlockByHash(ctx, hash)
}

func (s *Source) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return s.client.BlockByNumber(ctx, number)
}

// ForkchoiceUpdate updates the forkchoice on the execution client. If attributes is not nil, the engine client will also begin building a block
// based on attributes after the new head block and return the payload ID.
// May return an error in ForkChoiceResult, but the error is marshalled into the error return
func (s *Source) ForkchoiceUpdate(ctx context.Context, fc *ForkchoiceState, attributes *PayloadAttributes) (*ForkchoiceUpdatedResult, error) {
	e := s.log.New("state", fc, "attr", attributes)
	e.Debug("Sharing forkchoice-updated signal")
	fcCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var result ForkchoiceUpdatedResult
	err := s.rpc.CallContext(fcCtx, &result, "engine_forkchoiceUpdatedV1", fc, attributes)
	if err == nil {
		e.Debug("Shared forkchoice-updated signal")
		if attributes != nil {
			e.Debug("Received payload id", "payloadId", result.PayloadID)
		}
	} else {
		e = e.New("err", err)
		if rpcErr, ok := err.(rpc.Error); ok {
			code := ErrorCode(rpcErr.ErrorCode())
			e.Warn("Unexpected error code in forkchoice-updated response", "code", code)
		} else {
			e.Error("Failed to share forkchoice-updated signal")
		}
	}
	switch result.Status {
	case UpdateSyncing:
		return nil, fmt.Errorf("updated forkchoice, but node is syncing: %v", err)
	case UpdateSuccess:
		return &result, nil
	default:
		return nil, fmt.Errorf("unknown forkchoice status on %s: %q, ", fc.SafeBlockHash, string(result.Status))
	}
}

// ExecutePayload executes a built block on the execution engine and returns an error if it was not successful.
func (s *Source) ExecutePayload(ctx context.Context, payload *ExecutionPayload) error {
	e := s.log.New("block_hash", payload.BlockHash)
	e.Debug("sending payload for execution")

	execCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var result ExecutePayloadResult
	err := s.rpc.CallContext(execCtx, &result, "engine_executePayloadV1", payload)
	e.Debug("Received payload execution result", "status", result.Status, "latestValidHash", result.LatestValidHash, "message", result.ValidationError)
	if err != nil {
		e.Error("Payload execution failed", "err", err)
		return fmt.Errorf("failed to execute payload: %v", err)
	}

	switch result.Status {
	case ExecutionValid:
		return nil
	case ExecutionSyncing:
		return fmt.Errorf("failed to execute payload %s, node is syncing, latest valid hash is %s", payload.ID(), result.LatestValidHash)
	case ExecutionInvalid:
		return fmt.Errorf("execution payload %s was INVALID! Latest valid hash is %s, ignoring bad block: %q", payload.ID(), result.LatestValidHash, result.ValidationError)
	default:
		return fmt.Errorf("unknown execution status on %s: %q, ", payload.ID(), string(result.Status))
	}
}

// GetPayload gets the execution payload associated with the PayloadId
func (s *Source) GetPayload(ctx context.Context, payloadId PayloadID) (*ExecutionPayload, error) {
	e := s.log.New("payload_id", payloadId)
	e.Debug("getting payload")
	var result ExecutionPayload
	err := s.rpc.CallContext(ctx, &result, "engine_getPayloadV1", payloadId)
	if err != nil {
		e = log.New("payload_id", "err", err)
		if rpcErr, ok := err.(rpc.Error); ok {
			code := ErrorCode(rpcErr.ErrorCode())
			if code != UnavailablePayload {
				e.Warn("unexpected error code in get-payload response", "code", code)
			} else {
				e.Warn("unavailable payload in get-payload request")
			}
		} else {
			e.Error("failed to get payload")
		}
		return nil, err
	}
	e.Debug("Received payload")
	return &result, nil
}

// L2BlockRefByNumber returns the canonical block and parent ids.
func (s *Source) L2BlockRefByNumber(ctx context.Context, l2Num *big.Int) (eth.L2BlockRef, error) {
	block, err := s.client.BlockByNumber(ctx, l2Num)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Num, err)
	}
	return derive.BlockReferences(block, s.genesis)
}

// L2BlockRefByHash returns the block & parent ids based on the supplied hash. The returned BlockRef may not be in the canonical chain
func (s *Source) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	block, err := s.client.BlockByHash(ctx, l2Hash)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Hash, err)
	}
	return derive.BlockReferences(block, s.genesis)
}

type ReadOnlySource struct {
	rpc     *rpc.Client       // raw RPC client. Used for methods that do not already have bindings
	client  *ethclient.Client // go-ethereum's wrapper around the rpc client for the eth namespace
	genesis *rollup.Genesis
	log     log.Logger
}

func NewReadOnlySource(ll2Node *rpc.Client, genesis *rollup.Genesis, log log.Logger) (*ReadOnlySource, error) {
	return &ReadOnlySource{
		rpc:     ll2Node,
		client:  ethclient.NewClient(ll2Node),
		genesis: genesis,
		log:     log,
	}, nil
}

// TODO: de-duplicate Source and ReadOnlySource.
// We should really have a L1-downloader like binding that is more configurable and has caching.

// L2BlockRefByNumber returns the canonical block and parent ids.
func (s *ReadOnlySource) L2BlockRefByNumber(ctx context.Context, l2Num *big.Int) (eth.L2BlockRef, error) {
	block, err := s.client.BlockByNumber(ctx, l2Num)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Num, err)
	}
	return derive.BlockReferences(block, s.genesis)
}

// L2BlockRefByHash returns the block & parent ids based on the supplied hash. The returned BlockRef may not be in the canonical chain
func (s *ReadOnlySource) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	block, err := s.client.BlockByHash(ctx, l2Hash)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Hash, err)
	}
	return derive.BlockReferences(block, s.genesis)
}

func (s *ReadOnlySource) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return s.client.BlockByNumber(ctx, number)
}

func (s *ReadOnlySource) GetBlockHeader(ctx context.Context, blockTag string) (*types.Header, error) {
	var head *types.Header
	err := s.rpc.CallContext(ctx, &head, "eth_getBlockByNumber", blockTag, false)
	return head, err
}

func (s *ReadOnlySource) GetProof(ctx context.Context, address common.Address, blockTag string) (*AccountResult, error) {
	var getProofResponse *AccountResult
	err := s.rpc.CallContext(ctx, &getProofResponse, "eth_getProof", address, []common.Hash{}, blockTag)
	if err == nil && getProofResponse == nil {
		err = ethereum.NotFound
	}
	return getProofResponse, err
}
