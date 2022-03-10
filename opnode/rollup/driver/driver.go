package driver

import (
	"context"

	"github.com/ethereum-optimism/optimistic-specs/opnode/chain"
	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum/go-ethereum/log"
)

type Driver struct {
	s *state
}

func NewDriver(cfg rollup.Config, l2 L2Client, engine l2.EngineAPI, dl Downloader, chain chain.ChainSource, log log.Logger) *Driver {
	output := &outputImpl{
		Config: cfg,
		dl:     dl,
		log:    log,
		l2:     l2,
		engine: engine,
	}
	return &Driver{
		s: NewState(log, cfg, chain, output),
	}
}

func (d *Driver) Start(ctx context.Context, l1Heads <-chan eth.L1BlockRef) error {
	return d.s.Start(ctx, l1Heads)
}
func (d *Driver) Close() error {
	return d.s.Close()
}
