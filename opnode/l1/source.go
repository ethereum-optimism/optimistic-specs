package l1

import (
	"context"
	"sync/atomic"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type NewHeadSource interface {
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
}

type HeaderSource interface {
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
}

type ReceiptSource interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type BlockSource interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
}

type Source interface {
	NewHeadSource
	HeaderSource
	ReceiptSource
	BlockSource
	Close()
}

// For test instances, composition etc. we implement the interfaces with equivalent function types

type NewHeadFn func(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)

func (fn NewHeadFn) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return fn(ctx, ch)
}

type HeaderFn func(ctx context.Context, hash common.Hash) (*types.Header, error)

func (fn HeaderFn) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return fn(ctx, hash)
}

type ReceiptFn func(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

func (fn ReceiptFn) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return fn(ctx, txHash)
}

type BlockFn func(ctx context.Context, hash common.Hash) (*types.Block, error)

func (fn BlockFn) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return fn(ctx, hash)
}

// CombinedSource balances multiple L1 sources, to shred concurrent requests to multiple endpoints
type CombinedSource struct {
	i       uint64
	sources []Source
}

func NewCombinedL1Source(sources []Source) Source {
	if len(sources) == 0 {
		panic("need at least 1 source")
	}
	return &CombinedSource{i: 0, sources: sources}
}

func (cs *CombinedSource) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return cs.sources[atomic.AddUint64(&cs.i, 1)%uint64(len(cs.sources))].HeaderByHash(ctx, hash)
}

func (cs *CombinedSource) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	// TODO: can't use multiple sources as consensus, or head may be conflicting too much
	return cs.sources[0].SubscribeNewHead(ctx, ch)
}

func (cs *CombinedSource) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return cs.sources[atomic.AddUint64(&cs.i, 1)%uint64(len(cs.sources))].TransactionReceipt(ctx, txHash)
}

func (cs *CombinedSource) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return cs.sources[atomic.AddUint64(&cs.i, 1)%uint64(len(cs.sources))].BlockByHash(ctx, hash)
}

func (cs *CombinedSource) Close() {
	for _, src := range cs.sources {
		src.Close()
	}
}
