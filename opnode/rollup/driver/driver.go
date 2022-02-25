package driver

import (
	"context"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/sync"
	"github.com/ethereum/go-ethereum/log"
)

type Driver struct {
	s *state
}

func NewDriver(cfg rollup.Config, l2 DriverAPI, l1 l1.Source, log log.Logger) *Driver {
	input := &inputImpl{l1: l1,
		l2:         l2,
		syncSource: sync.SyncSourceV2{sync.SyncSource{L1: l1, L2: l2}}}
	output := &outputImpl{
		Config: cfg,
		dl:     l1,
		log:    log,
		rpc:    l2,
	}
	return &Driver{
		s: NewState(log, cfg, input, output),
	}
}

func (d *Driver) Start(ctx context.Context, l1Heads <-chan eth.HeadSignal) error {
	return d.s.Start(ctx, l1Heads)
}
func (d *Driver) Close() error {
	return d.s.Close()
}

type DriverAPI interface {
	l2.EngineAPI
	l2.EthBackend
}

type inputImpl struct {
	l1         l1.Source
	l2         DriverAPI
	syncSource sync.SyncReferenceV2
}

func (i *inputImpl) L1Head(ctx context.Context) (eth.BlockID, error) {
	header, err := i.l1.HeaderByNumber(ctx, nil)
	if err != nil {
		return eth.BlockID{}, err
	}
	return eth.BlockID{Hash: header.Hash(), Number: header.Number.Uint64()}, nil
}

func (i *inputImpl) L2Head(ctx context.Context) (eth.BlockID, error) {
	block, err := i.l2.BlockByNumber(ctx, nil)
	if err != nil {
		return eth.BlockID{}, err
	}
	return eth.BlockID{Hash: block.Hash(), Number: block.Number().Uint64()}, nil

}

func (i *inputImpl) L1ChainWindow(ctx context.Context, base eth.BlockID) ([]eth.BlockID, error) {
	return sync.FindL1Range(ctx, i.syncSource, base)
}
