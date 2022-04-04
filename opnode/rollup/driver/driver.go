package driver

import (
	"context"
	"math/big"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type Driver struct {
	s *state
}

type BatchSubmitter interface {
	Submit(config *rollup.Config, batches []*derive.BatchData) (common.Hash, error)
}

type Downloader interface {
	InfoByHash(ctx context.Context, hash common.Hash) (derive.L1Info, error)
	Fetch(ctx context.Context, blockHash common.Hash) (derive.L1Info, types.Transactions, types.Receipts, error)
	FetchAllTransactions(ctx context.Context, window []eth.BlockID) ([]types.Transactions, error)
}

type Engine interface {
	GetPayload(ctx context.Context, payloadId l2.PayloadID) (*l2.ExecutionPayload, error)
	ForkchoiceUpdate(ctx context.Context, state *l2.ForkchoiceState, attr *l2.PayloadAttributes) (*l2.ForkchoiceUpdatedResult, error)
	ExecutePayload(ctx context.Context, payload *l2.ExecutionPayload) error
	BlockByHash(context.Context, common.Hash) (*types.Block, error)
	BlockByNumber(context.Context, *big.Int) (*types.Block, error)
}

type L1Chain interface {
	L1BlockRefByNumber(context.Context, uint64) (eth.L1BlockRef, error)
	L1BlockRefByHash(context.Context, common.Hash) (eth.L1BlockRef, error)
	L1HeadBlockRef(context.Context) (eth.L1BlockRef, error)
	L1Range(ctx context.Context, base eth.BlockID, max uint64) ([]eth.BlockID, error)
}

// TODO: Extend L2 Interface to get safe/unsafe blocks (specifically for Unsafe L2 head)
type L2Chain interface {
	L2BlockRefByNumber(ctx context.Context, l2Num *big.Int) (eth.L2BlockRef, error)
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error)
}

type outputInterface interface {
	// insertEpoch creates and inserts one epoch on top of the safe head. It prefers blocks it creates to what is recorded in the unsafe chain.
	// It returns the new L2 head and L2 Safe head and if there was a reorg. This function must return if there was a reorg otherwise the L2 chain must be traversed.
	insertEpoch(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.L2BlockRef, l2Finalized eth.BlockID, l1Input []eth.BlockID) (eth.L2BlockRef, eth.L2BlockRef, bool, error)

	// createNewBlock builds a new block based on the L2 Head, L1 Origin, and the current mempool.
	createNewBlock(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.BlockID, l2Finalized eth.BlockID, l1Origin eth.BlockID) (eth.L2BlockRef, *derive.BatchData, error)
}

func NewDriver(cfg rollup.Config, l2 *l2.Source, l1 *l1.Source, log log.Logger, submitter BatchSubmitter, sequencer bool) *Driver {
	if sequencer && submitter == nil {
		log.Error("Bad configuration")
		// TODO: return error
	}
	output := &outputImpl{
		Config: cfg,
		dl:     l1,
		l2:     l2,
		log:    log,
	}
	return &Driver{
		s: NewState(log, cfg, l1, l2, output, submitter, sequencer),
	}
}

func (d *Driver) Start(ctx context.Context, l1Heads <-chan eth.L1BlockRef) error {
	return d.s.Start(ctx, l1Heads)
}
func (d *Driver) Close() error {
	return d.s.Close()
}
