package node

import (
	"context"
	"fmt"
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
	L1NodeAddrs   []string `ask:"--l1" help:"Addresses of L1 User JSON-RPC endpoints to use (eth namespace required)"`
	L2EngineAddrs []string `ask:"--l2" help:"Addresses of L2 Engine JSON-RPC endpoints to use (engine and eth namespace required)"`

	LogCmd `ask:".log" help:"Log configuration"`

	// TODO: multi-addrs option
	// TODO: bootnodes option

	log log.Logger

	// sources to fetch data from
	l1Sources []l1.Source

	// engines to keep synced
	l2Engines []*l2.Engine

	l1Tracker    l1.Tracker
	l1Downloader l1.Downloader

	ctx   context.Context
	close chan chan error
}

func (c *OpNodeCmd) Default() {
	c.L1NodeAddrs = []string{"http://127.0.0.1:8545"}
	c.L2EngineAddrs = []string{"http://127.0.0.1:8551"}
}

func (c *OpNodeCmd) Help() string {
	return "Run optimism node"
}

func (c *OpNodeCmd) Run(ctx context.Context, args ...string) error {
	logger := c.LogCmd.Create()
	c.log = logger
	c.ctx = ctx

	for i, addr := range c.L1NodeAddrs {
		// L1 exec engine: read-only, to update L2 consensus with
		l1Node, err := rpc.DialContext(ctx, addr)
		if err != nil {
			// HTTP or WS RPC may create a disconnected client, RPC over IPC may fail directly
			if l1Node == nil {
				return fmt.Errorf("failed to dial L1 address %d (%s): %v", i, addr, err)
			}
			c.log.Warn("failed to dial L1 address, but may connect later", "i", i, "addr", addr, "err", err)
		}
		// TODO: we may need to authenticate the connection with L1
		// l1Node.SetHeader()
		cl := ethclient.NewClient(l1Node)
		c.l1Sources = append(c.l1Sources, cl)
	}
	if len(c.l1Sources) == 0 {
		return fmt.Errorf("need at least one L1 source endpoint, see --l1")
	}

	for i, addr := range c.L2EngineAddrs {
		// L2 exec engine: updated by this OpNode (L2 consensus layer node)
		backend, err := rpc.DialContext(ctx, addr)
		if err != nil {
			if backend == nil {
				return fmt.Errorf("failed to dial L2 address %d (%s): %v", i, addr, err)
			}
			c.log.Warn("failed to dial L2 address, but may connect later", "i", i, "addr", addr, "err", err)
		}
		// TODO: we may need to authenticate the connection with L2
		// backend.SetHeader()
		client := &l2.EngineClient{
			RPCBackend: backend,
			Log:        c.log.New("engine_client", i),
		}
		engine := &l2.Engine{
			Ctx: c.ctx,
			Log: c.log.New("engine", i),
			RPC: client,
		}
		c.l2Engines = append(c.l2Engines, engine)
	}

	// TODO: maybe spin up an API server
	//  (to get debug data, change runtime settings like logging, serve pprof, get peering info, node health, etc.)

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

	// Combine L1 sources, so work can be balanced between them
	l1Source := l1.NewCombinedL1Source(c.l1Sources)

	// We download receipts in parallel
	c.l1Downloader.AddReceiptWorkers(4)

	for _, eng := range c.l2Engines {
		// start syncing headers in the background, based on head signals
		l1HeaderSyncSub := l1.HeaderSync(c.ctx, l1Source, c.l1Tracker, c.log, c.l1Downloader, eng)
		mergeSub(l1HeaderSyncSub, "header sync unexpectedly failed")
	}

	// Keep subscribed to the L1 heads, which keeps the L1 maintainer pointing to the best headers to sync
	l1HeadsSub := event.ResubscribeErr(time.Second*10, func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			c.log.Warn("resubscribing after failed L1 subscription", "err", err)
		}
		return l1.WatchHeadChanges(c.ctx, l1Source, c.l1Tracker)
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
			// close L1 data sources
			l1Source.Close()
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
