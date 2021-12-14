package node

import (
	"context"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
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

	log log.Logger

	// wraps *rpc.Client, provides eth namespace
	l1Node *ethclient.Client
	// Raw RPC client, separate bindings
	l2Engine *rpc.Client

	l1Maintainer l1.L1Maintainer

	downloadRequests chan l1.BlockID

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

	// TODO: determine L1 starting point from L2 node
	c.l1Maintainer = l1.NewTracker()

	// TODO improve request scheduling
	c.downloadRequests = make(chan common.Hash, 100)

	downloader := l1.NewDownloader(c.ctx, c.l1Node, c.log, c.downloadRequests)

	// TODO: transform and apply blocks to L2
	//downloader.Listen()

	c.close = make(chan chan error)

	go c.RunNode()

	return nil
}

func (c *OpNodeCmd) RunNode() {
	c.log.Info("started OpNode")

	heartbeat := time.NewTicker(time.Millisecond * 700)
	defer heartbeat.Stop()

	// start syncing headers in the background, based on head signals
	headerSyncSub := l1.L1HeaderSync(c.ctx, c.l1Node, c.l1Maintainer, c.log, c)

	// Keep subscribed to the L1 heads, which keeps the L1 maintainer pointing to the best headers to sync
	l1HeadsSub := event.ResubscribeErr(time.Second*10, func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			c.log.Warn("resubscribing after failed L1 subscription", "err", err)
		}
		return l1.SubL1Node(c.ctx, c.l1Node, c.l1Maintainer)
	})

	for {
		select {
		case err := <-headerSyncSub.Err():
			c.log.Error("header sync unexpectedly failed", "err", err)
		case err := <-l1HeadsSub.Err():
			c.log.Error("l1 heads subscription failed", "err", err)
		case <-heartbeat.C:
			// TODO poll data, process blocks, etc.

		// TODO: open a channel with L1 RPC to listen for new blocks and reorgs?

		case done := <-c.close:
			c.log.Info("Closing OpNode")
			headerSyncSub.Unsubscribe()
			l1HeadsSub.Unsubscribe()
			c.l1Node.Close()
			c.l2Engine.Close()
			done <- nil
			return
		}
	}
}

func (c *OpNodeCmd) LastL1() l1.BlockID {

}

func (c *OpNodeCmd) ProcessL1(id l1.BlockID) {

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
