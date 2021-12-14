package node

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/rpc"
)

type OpNodeCmd struct {
	L1NodeAddr   string `ask:"--l1" help:"Address of L1 User JSON-RPC endpoint to use (eth namespace required)"`
	L2EngineAddr string `ask:"--l2" help:"Address of L2 Engine JSON-RPC endpoint to use (engine and eth namespace required)"`

	LogCmd `ask:".log" help:"Log configuration"`

	// TODO: multi-addrs option
	// TODO: bootnodes option

	log Logger

	// wraps *rpc.Client, provides eth namespace
	l1Node *ethclient.Client
	// Raw RPC client, separate bindings
	l2Engine *rpc.Client

	ctx   context.Context
	close chan chan error
}

func (c *OpNodeCmd) Default() {
	c.L1NodeAddr = "http://127.0.0.1:8545"
	c.L2EngineAddr = "http://127.0.0.1:8551"
}

func (c *OpNodeCmd) Help() string {
	return "Run optimism node"
}

func (c *OpNodeCmd) Run(ctx context.Context, args ...string) error {
	log := c.LogCmd.Create()
	c.log = log
	c.ctx = ctx

	// L1 exec engine: read-only, to update L2 consensus with
	l1Node, err := rpc.DialContext(ctx, c.L1NodeAddr)
	if err != nil {
		return err
	}
	// TODO: we may need to authenticate the connection with L1
	// l1Node.SetHeader()
	c.l1Node = ethclient.NewClient(l1Node)

	// L2 exec engine: updated by this OpNode (L2 consensus layer node)
	engine, err := rpc.DialContext(ctx, c.L2EngineAddr)
	if err != nil {
		return err
	}
	// TODO: we may need to authenticate the connection with L2
	// engine.SetHeader()
	c.l2Engine = engine

	// TODO: maybe spin up an API server
	//  (to get debug data, change runtime settings like logging, serve pprof, get peering info, node health, etc.)

	c.close = make(chan chan error)

	go c.RunNode()

	return nil
}

// background process
func (c *OpNodeCmd) RunNode() {
	c.log.Info("started OpNode")

	heartbeat := time.NewTicker(time.Millisecond * 700)
	defer heartbeat.Stop()

	for {
		select {
		case <-heartbeat.C:
			// TODO poll data, process blocks, etc.

		// TODO: open a channel with L1 RPC to listen for new blocks and reorgs?

		case done := <-c.close:
			c.log.Info("Closing OpNode")
			c.l1Node.Close()
			c.l2Engine.Close()
			done <- nil
			return
		}
	}
}

func (c *OpNodeCmd) Close() error {
	if c.close != nil {
		done := make(chan error)
		c.close <- done
		err := <-done
		return err
	}
	return nil
}
