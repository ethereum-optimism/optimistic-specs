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
	// FetchL1Info fetches the L1 header information corresponding to a L1 block ID
	FetchL1Info(ctx context.Context, id eth.BlockID) (derive.L1Info, error)
	// FetchReceipts of a L1 block
	FetchReceipts(ctx context.Context, id eth.BlockID) ([]*types.Receipt, error)
	// FetchTransactions from the given window of L1 blocks
	FetchTransactions(ctx context.Context, window []eth.BlockID) ([]*types.Transaction, error)
}

type BlockPreparer interface {
	GetPayload(ctx context.Context, payloadId l2.PayloadID) (*l2.ExecutionPayload, error)
	ForkchoiceUpdate(ctx context.Context, state *l2.ForkchoiceState, attr *l2.PayloadAttributes) (*l2.ForkchoiceUpdatedResult, error)
	ExecutePayload(ctx context.Context, payload *l2.ExecutionPayload) error
	BlockByHash(context.Context, common.Hash) (*types.Block, error)
}

type L1Chain interface {
	L1BlockRefByNumber(ctx context.Context, l1Num uint64) (eth.L1BlockRef, error)
	L1HeadBlockRef(ctx context.Context) (eth.L1BlockRef, error)
	L1Range(ctx context.Context, base eth.BlockID) ([]eth.BlockID, error)
}

type L2Chain interface {
	L2BlockRefByNumber(ctx context.Context, l2Num *big.Int) (eth.L2BlockRef, error)
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error)
}

type outputInterface interface {
	step(ctx context.Context, l2Head eth.BlockID, l2Finalized eth.BlockID, unsafeL2Head eth.BlockID, l1Input []eth.BlockID) (eth.BlockID, error)
	newBlock(ctx context.Context, l2Finalized eth.BlockID, l2Parent eth.BlockID, l2Safe eth.BlockID, l1Origin eth.BlockID, includeDeposits bool) (eth.BlockID, *derive.BatchData, error)
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
