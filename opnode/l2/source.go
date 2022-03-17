package l2

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type internalRPC interface {
	executePayload(ctx context.Context, payload *ExecutionPayload) (*ExecutePayloadResult, error)
	forkchoiceUpdated(ctx context.Context, state *ForkchoiceState, attr *PayloadAttributes) (ForkchoiceUpdatedResult, error)
}

type Source struct {
	rpc     *rpc.Client       // raw RPC client. Used for the consensus namespace
	client  *ethclient.Client // go-ethereum's wrapper around the rpc client for the eth namespace
	genesis *rollup.Genesis   // Genesis to enable derive.BlockReference
	log     log.Logger
}

func NewSource(l2addr string, genesis *rollup.Genesis, log log.Logger) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rpc, err := rpc.DialContext(ctx, l2addr)
	if err != nil {
		return nil, err
	}
	return &Source{
		rpc:     rpc,
		client:  ethclient.NewClient(rpc),
		genesis: genesis,
		log:     log,
	}, nil
}

func (s *Source) Close() {
	s.rpc.Close()
}

func (s *Source) ForkchoiceUpdate(ctx context.Context, fc *ForkchoiceState, attributes *PayloadAttributes) (*ForkchoiceUpdatedResult, error) {
	return ForkchoiceUpdate(ctx, s, fc, attributes)
}

func (s *Source) ExecutePayload(ctx context.Context, payload *ExecutionPayload) error {
	return ExecutePayloadStatic(ctx, s, payload)
}

func (s *Source) GetPayload(ctx context.Context, payloadId PayloadID) (*ExecutionPayload, error) {
	return s.getPayload(ctx, payloadId)
}

func (s *Source) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return s.client.BlockByHash(ctx, hash)
}

func (s *Source) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return s.client.BlockByNumber(ctx, number)
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
