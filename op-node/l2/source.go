package l2

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
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

func NewSource(l2Node *rpc.Client, genesis *rollup.Genesis, log log.Logger) (*Source, error) {
	return &Source{
		rpc:     l2Node,
		client:  ethclient.NewClient(l2Node),
		genesis: genesis,
		log:     log,
	}, nil
}

func (s *Source) Close() {
	s.rpc.Close()
}

func (s *Source) PayloadByHash(ctx context.Context, hash common.Hash) (*ExecutionPayload, error) {
	// TODO: we really do not need to parse every single tx and block detail, keeping transactions encoded is faster.
	block, err := s.client.BlockByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve L2 block by hash: %v", err)
	}
	payload, err := BlockAsPayload(block)
	if err != nil {
		return nil, fmt.Errorf("failed to read L2 block as payload: %w", err)
	}
	return payload, nil
}

func (s *Source) PayloadByNumber(ctx context.Context, number *big.Int) (*ExecutionPayload, error) {
	// TODO: we really do not need to parse every single tx and block detail, keeping transactions encoded is faster.
	block, err := s.client.BlockByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve L2 block by number: %v", err)
	}
	payload, err := BlockAsPayload(block)
	if err != nil {
		return nil, fmt.Errorf("failed to read L2 block as payload: %w", err)
	}
	return payload, nil
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
	switch result.PayloadStatus.Status {
	case ExecutionSyncing:
		return nil, fmt.Errorf("updated forkchoice, but node is syncing: %v", err)
	case ExecutionAccepted, ExecutionInvalidTerminalBlock, ExecutionInvalidBlockHash:
		// ACCEPTED, INVALID_TERMINAL_BLOCK, INVALID_BLOCK_HASH are only for execution
		return nil, fmt.Errorf("unexpected %s status, could not update forkchoice: %v", result.PayloadStatus.Status, err)
	case ExecutionInvalid:
		return nil, fmt.Errorf("cannot update forkchoice, block is invalid: %v", err)
	case ExecutionValid:
		return &result, nil
	default:
		return nil, fmt.Errorf("unknown forkchoice status on %s: %q, ", fc.SafeBlockHash, string(result.PayloadStatus.Status))
	}
}

// ExecutePayload executes a built block on the execution engine and returns an error if it was not successful.
func (s *Source) NewPayload(ctx context.Context, payload *ExecutionPayload) error {
	e := s.log.New("block_hash", payload.BlockHash)
	e.Debug("sending payload for execution")

	execCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var result PayloadStatusV1
	err := s.rpc.CallContext(execCtx, &result, "engine_newPayloadV1", payload)
	e.Debug("Received payload execution result", "status", result.Status, "latestValidHash", result.LatestValidHash, "message", result.ValidationError)
	if err != nil {
		e.Error("Payload execution failed", "err", err)
		return fmt.Errorf("failed to execute payload: %v", err)
	}

	switch result.Status {
	case ExecutionValid:
		return nil
	case ExecutionSyncing:
		return fmt.Errorf("failed to execute payload %s, node is syncing", payload.ID())
	case ExecutionInvalid:
		return fmt.Errorf("execution payload %s was INVALID! Latest valid hash is %s, ignoring bad block: %q", payload.ID(), result.LatestValidHash, result.ValidationError)
	case ExecutionInvalidBlockHash:
		return fmt.Errorf("execution payload %s has INVALID BLOCKHASH! %v", payload.BlockHash, result.ValidationError)
	case ExecutionInvalidTerminalBlock:
		return fmt.Errorf("engine is misconfigured. Received invalid-terminal-block error while engine API should be active at genesis. err: %v", result.ValidationError)
	case ExecutionAccepted:
		return fmt.Errorf("execution payload cannot be validated yet, latest valid hash is %s", result.LatestValidHash)
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
		e = e.New("payload_id", "err", err)
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
	return blockToBlockRef(block, s.genesis)
}

// L2BlockRefByHash returns the block & parent ids based on the supplied hash. The returned BlockRef may not be in the canonical chain
func (s *Source) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	block, err := s.client.BlockByHash(ctx, l2Hash)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Hash, err)
	}
	return blockToBlockRef(block, s.genesis)
}

// blockToBlockRef extracts the essential L2BlockRef information from a block,
// falling back to genesis information if necessary.
func blockToBlockRef(block *types.Block, genesis *rollup.Genesis) (eth.L2BlockRef, error) {
	var l1Origin eth.BlockID
	var sequenceNumber uint64
	if block.NumberU64() == genesis.L2.Number {
		if block.Hash() != genesis.L2.Hash {
			return eth.L2BlockRef{}, fmt.Errorf("expected L2 genesis hash to match L2 block at genesis block number %d: %s <> %s", genesis.L2.Number, block.Hash(), genesis.L2.Hash)
		}
		l1Origin = genesis.L1
		sequenceNumber = 0
	} else {
		txs := block.Transactions()
		if len(txs) == 0 {
			return eth.L2BlockRef{}, fmt.Errorf("l2 block is missing L1 info deposit tx, block hash: %s", block.Hash())
		}
		tx := txs[0]
		if tx.Type() != types.DepositTxType {
			return eth.L2BlockRef{}, fmt.Errorf("first block tx has unexpected tx type: %d", tx.Type())
		}
		info, err := derive.L1InfoDepositTxData(tx.Data())
		if err != nil {
			return eth.L2BlockRef{}, fmt.Errorf("failed to parse L1 info deposit tx from L2 block: %v", err)
		}
		l1Origin = eth.BlockID{Hash: info.BlockHash, Number: info.Number}
		sequenceNumber = info.SequenceNumber
	}
	return eth.L2BlockRef{
		Hash:           block.Hash(),
		Number:         block.NumberU64(),
		ParentHash:     block.ParentHash(),
		Time:           block.Time(),
		L1Origin:       l1Origin,
		SequenceNumber: sequenceNumber,
	}, nil
}

// PayloadToBlockRef extracts the essential L2BlockRef information from an execution payload,
// falling back to genesis information if necessary.
func PayloadToBlockRef(payload *ExecutionPayload, genesis *rollup.Genesis) (eth.L2BlockRef, error) {
	var l1Origin eth.BlockID
	var sequenceNumber uint64
	if uint64(payload.BlockNumber) == genesis.L2.Number {
		if payload.BlockHash != genesis.L2.Hash {
			return eth.L2BlockRef{}, fmt.Errorf("expected L2 genesis hash to match L2 block at genesis block number %d: %s <> %s", genesis.L2.Number, payload.BlockHash, genesis.L2.Hash)
		}
		l1Origin = genesis.L1
		sequenceNumber = 0
	} else {
		if len(payload.Transactions) == 0 {
			return eth.L2BlockRef{}, fmt.Errorf("l2 block is missing L1 info deposit tx, block hash: %s", payload.BlockHash)
		}
		var tx types.Transaction
		if err := tx.UnmarshalBinary(payload.Transactions[0]); err != nil {
			return eth.L2BlockRef{}, fmt.Errorf("failed to decode first tx to read l1 info from: %v", err)
		}
		if tx.Type() != types.DepositTxType {
			return eth.L2BlockRef{}, fmt.Errorf("first payload tx has unexpected tx type: %d", tx.Type())
		}
		info, err := derive.L1InfoDepositTxData(tx.Data())
		if err != nil {
			return eth.L2BlockRef{}, fmt.Errorf("failed to parse L1 info deposit tx from L2 block: %v", err)
		}
		l1Origin = eth.BlockID{Hash: info.BlockHash, Number: info.Number}
		sequenceNumber = info.SequenceNumber
	}

	return eth.L2BlockRef{
		Hash:           payload.BlockHash,
		Number:         uint64(payload.BlockNumber),
		ParentHash:     payload.ParentHash,
		Time:           uint64(payload.Timestamp),
		L1Origin:       l1Origin,
		SequenceNumber: sequenceNumber,
	}, nil
}

type ReadOnlySource struct {
	rpc     *rpc.Client       // raw RPC client. Used for methods that do not already have bindings
	client  *ethclient.Client // go-ethereum's wrapper around the rpc client for the eth namespace
	genesis *rollup.Genesis
	log     log.Logger
}

func NewReadOnlySource(l2Node *rpc.Client, genesis *rollup.Genesis, log log.Logger) (*ReadOnlySource, error) {
	return &ReadOnlySource{
		rpc:     l2Node,
		client:  ethclient.NewClient(l2Node),
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
	return blockToBlockRef(block, s.genesis)
}

// L2BlockRefByHash returns the block & parent ids based on the supplied hash. The returned BlockRef may not be in the canonical chain
func (s *ReadOnlySource) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	block, err := s.client.BlockByHash(ctx, l2Hash)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Hash, err)
	}
	return blockToBlockRef(block, s.genesis)
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
