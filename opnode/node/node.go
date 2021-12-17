package node

import (
	"context"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum"
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
	// TODO: multiple L1 data sources
	l1Node *ethclient.Client

	// engines to keep synced
	l2Engines []*l2.L2Engine

	l1Tracker    l1.Tracker
	l1Downloader l1.Downloader

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
	c.l2Engines = append(c.l2Engines, &l2.L2Engine{RPC: engine, L1Head: l1.BlockID{}})

	// TODO: maybe spin up an API server
	//  (to get debug data, change runtime settings like logging, serve pprof, get peering info, node health, etc.)

	// TODO: determine L1 starting point from L2 node
	c.l1Tracker = l1.NewTracker()

	c.close = make(chan chan error)

	go c.RunNode()

	return nil
}

func (c *OpNodeCmd) RunNode() {
	c.log.Info("started OpNode")

	heartbeat := time.NewTicker(time.Millisecond * 700)
	defer heartbeat.Stop()

	var unsub []func()
	mergeSub := func(sub ethereum.Subscription, errMsg string) {
		unsub = append(unsub, sub.Unsubscribe)
		go func() {
			err, ok := <-sub.Err()
			if !ok {
				return
			}
			c.log.Error(errMsg, "err", err)
		}()
	}

	// We download receipts in parallel
	c.l1Downloader.AddReceiptWorkers(4)

	for _, eng := range c.l2Engines {
		// start syncing headers in the background, based on head signals
		l1HeaderSyncSub := l1.HeaderSync(c.ctx, c.l1Node, c.l1Tracker, c.log, c.l1Downloader, eng)
		mergeSub(l1HeaderSyncSub, "header sync unexpectedly failed")
	}

	// Keep subscribed to the L1 heads, which keeps the L1 maintainer pointing to the best headers to sync
	l1HeadsSub := event.ResubscribeErr(time.Second*10, func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			c.log.Warn("resubscribing after failed L1 subscription", "err", err)
		}
		return l1.WatchHeadChanges(c.ctx, c.l1Node, c.l1Tracker)
	})
	mergeSub(l1HeadsSub, "l1 heads subscription failed")

	// feed from tracker, as fed with head events from above subscription
	l1Heads := make(chan l1.BlockID)
	mergeSub(c.l1Tracker.WatchHeads(l1Heads), "l1 heads info feed unexpectedly failed")

	for {
		select {
		case l1Head := <-l1Heads:
			c.log.Info("New Layer1 head: nr %10d, hash %s", l1Head.Number, l1Head.Hash)
		case <-heartbeat.C:
			// TODO log info like latest L1/L2 head of engines

		case done := <-c.close:
			c.log.Info("Closing OpNode")
			// close all tasks
			for _, f := range unsub {
				f()
			}
			// close L1 data source
			c.l1Node.Close()
			// close L2 engines
			for _, eng := range c.l2Engines {
				eng.Close()
			}
			// signal back everything closed without error
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
