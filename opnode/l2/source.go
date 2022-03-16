package l2

import (
	"context"
	"math/big"
	"time"

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
	rpc    *rpc.Client       // raw RPC client. Used for the consensus namespace
	client *ethclient.Client // go-ethereum's wrapper around the rpc client for the eth namespace
	log    log.Logger
}

func NewSource(l2addr string, log log.Logger) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rpc, err := rpc.DialContext(ctx, l2addr)
	if err != nil {
		return nil, err
	}
	return &Source{
		rpc:    rpc,
		client: ethclient.NewClient(rpc),
		log:    log,
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
