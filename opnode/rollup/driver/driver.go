package driver

import (
	"context"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l1"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/sync"
	rollupSync "github.com/ethereum-optimism/optimistic-specs/opnode/rollup/sync"

	"github.com/ethereum/go-ethereum/log"
)

// keep making many sync steps if we can make sync progress
const hot = time.Millisecond * 30

// check on sync regularly, but prioritize sync triggers with head updates etc.
const cold = time.Second * 8

// at least try every minute to sync, even if things are going well
const max = time.Minute

type Driver struct {
	log     log.Logger
	rpc     DriverAPI
	syncRef rollupSync.SyncReference
	dl      Downloader
	l1Heads <-chan eth.HeadSignal
	done    chan struct{}
	state   // embedded engine state
}

func NewDriver(cfg rollup.Config, l2 DriverAPI, l1 l1.Source, log log.Logger) *Driver {
	return &Driver{
		log:     log,
		rpc:     l2,
		syncRef: rollupSync.SyncSource{L1: l1, L2: l2},
		dl:      l1,
		done:    make(chan struct{}),
		state:   state{Config: cfg},
	}
}

func (d *Driver) Start(ctx context.Context, l1Heads <-chan eth.HeadSignal) error {
	d.l1Heads = l1Heads
	if !d.requestUpdate(ctx, d.log, d) {
		d.log.Error("failed to fetch engine head, defaulting to genesis")
		d.updateHead(d.state.Config.Genesis.L1, d.state.Config.Genesis.L2)
	}
	go d.loop()
	return nil
}
func (d *Driver) Close() error {
	close(d.done)
	return nil
}

func (d *Driver) loop() {
	ctx := context.Background()
	backoff := cold
	syncTicker := time.NewTicker(cold)
	l2HeadPoll := time.NewTicker(time.Second * 14)

	// exponential backoff, add 10% each step, up to max.
	syncBackoff := func() {
		backoff += backoff / 10
		if backoff > max {
			backoff = max
		}
		syncTicker.Reset(backoff)
	}
	syncQuickly := func() {
		syncTicker.Reset(hot)
		backoff = cold
	}
	onL2Update := func() {
		// And we want to slow down requesting the L2 engine for its head (we just changed it ourselves)
		// Request head if we don't successfully change it in the next 14 seconds.
		l2HeadPoll.Reset(time.Second * 14)
	}
	defer syncTicker.Stop()
	defer l2HeadPoll.Stop()

	for {
		select {
		case <-d.done:
			return
		case <-l2HeadPoll.C:
			ctx, cancel := context.WithTimeout(ctx, time.Second*4)
			if d.requestUpdate(ctx, d.log, d) {
				onL2Update()
			}
			cancel()
			continue
		case l1HeadSig := <-d.l1Heads:
			ctx, cancel := context.WithTimeout(ctx, time.Second*4)
			if d.notifyL1Head(ctx, d.log, l1HeadSig, d) {
				syncQuickly()
			}
			cancel()
			continue
		case <-syncTicker.C:
			// If already synced, or in case of failure, we slow down
			syncBackoff()
			ctx, cancel := context.WithTimeout(ctx, time.Second*4)
			if d.requestSync(ctx, d.log, d) {
				// Successfully stepped toward target. Continue quickly if we are not there yet
				syncQuickly()
				onL2Update()
			}
			cancel()
		}
	}

}

// Fulfill the `internalDriver` interface

func (e *Driver) requestEngineHead(ctx context.Context) (refL1 eth.BlockID, refL2 eth.BlockID, err error) {
	refL1, refL2, _, err = e.syncRef.RefByL2Num(ctx, nil, &e.Config.Genesis)
	return
}

func (e *Driver) findSyncStart(ctx context.Context) (nextRefL1s []eth.BlockID, refL2 eth.BlockID, err error) {
	return rollupSync.V3FindSyncStart(ctx, sync.SyncSourceV2{e.syncRef}, &e.Config.Genesis)
}

func (e *Driver) driverStep(ctx context.Context, l1Input []eth.BlockID, l2Parent eth.BlockID, l2Finalized eth.BlockID) (l2ID eth.BlockID, err error) {
	return e.step(ctx, l1Input, l2Parent, l2Finalized.Hash)
}
