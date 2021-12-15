package l2

import (
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum/go-ethereum/rpc"
)

type L2Engine struct {
	// Raw RPC client, separate bindings
	RPC *rpc.Client
	// track where the l2 engine is at
	L1Head l1.BlockID
}

func (c *L2Engine) LastL1() l1.BlockID {
	return c.L1Head
}

func (c *L2Engine) ProcessL1(l1m l1.L1Maintainer, id l1.BlockID) {
	if id == c.L1Head {
		// no-op, already processed it
		return
	} else if id.Number+1 == c.L1Head.Number {
		// extension or reorg
		// TODO: fetch block and run driver

	} else if id.Number <= c.L1Head.Number {
		// reorg
		// TODO: fetch block and run driver
	} else {
		// Block is farther out, if far enough it would
		// make sense to trigger a state-sync.
		// TODO: need previous L2 state to derive full block.
		// Triggering sync with just a hash would be better.
	}
}

func (c *L2Engine) Close() {
	c.RPC.Close()
}
