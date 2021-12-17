package l1

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

type ChainMode uint

const (
	// ChainNoop indicates the local view is complete: nothing to do.
	// block: current head
	ChainNoop ChainMode = iota
	// ChainExtend indicates to add a canonical block to local view.
	// block: next to extend the tip of the chain with
	ChainExtend
	// ChainUndo indicates to form of reorg, without new blocks: wind back to older block as head
	// block: block to rewind to as head
	ChainUndo
	// ChainReorg indicates to form of reorg, and orphan previous alternative chain.
	// block: next to extend a previous part of the chain with
	ChainReorg
	// ChainMissing indicates the local view is incomplete, with a gap to canonical chain.
	// block: lowest-height canonical block, to find the parent of to be able to complete the local view.
	ChainMissing
)

type BlockID struct {
	Hash   common.Hash
	Number uint64
}

type tracker struct {
	sync.Mutex

	// Parent hash (value) of each block (key)
	parents map[BlockID]BlockID
	// last seen head height, may not be the max height (total difficult != height)
	head BlockID

	// feed of BlockID
	headChanges event.Feed
}

func (tr *tracker) HeadSignal(id BlockID) {
	tr.Lock()
	defer tr.Unlock()

	tr.head = id
	tr.headChanges.Send(id)
}

func (tr *tracker) WatchHeads(ch chan<- BlockID) ethereum.Subscription {
	return tr.headChanges.Subscribe(ch)
}

func (tr *tracker) Parent(id BlockID, parent BlockID) {
	tr.Lock()
	defer tr.Unlock()

	tr.parents[id] = parent
}

func (tr *tracker) Pull(lastLocal BlockID) (next BlockID, mode ChainMode) {
	tr.Lock()
	defer tr.Unlock()

	// find the common block between the last view and the canonical chain
	canon := tr.head
	last := lastLocal

	// if we don't have any canonical view at all, then stay at the local view
	if canon == (BlockID{}) {
		return last, ChainNoop
	}

	// Any difference to resolve at all?
	if canon == last {
		return last, ChainNoop
	}

	// traverse back last
	for canon.Number < last.Number {
		p, ok := tr.parents[last]
		if !ok {
			// can't track last view back to canonical chain
			return canon, ChainMissing
		}
		// did we undo a canonical block? (i.e. reorg to previous block in same chain)
		if p == canon {
			// return the block to rewind back to
			return canon, ChainUndo
		}
		// if we have history of last viewed chain, go back,
		// to find where things are canonical again.
		last = p
	}

	// traverse back canon
	for canon.Number > last.Number {
		p, ok := tr.parents[canon]
		if !ok {
			// can't track last view back to canonical chain
			return canon, ChainMissing
		}
		// does the canonical chain extend the last chain?
		if p == last {
			// return the very first block to extend with
			return canon, ChainExtend
		}
		canon = p
	}

	// Now we have two equal-height chains, that do not match.
	// Find the common block, if any
	for {
		p, ok := tr.parents[last]
		if !ok {
			// Our local view does not reach the canonical chain history.
			// Traverse back canonical chain history to find the lowest
			// canonical block the local view needs to be extended to.
			for {
				p, ok := tr.parents[canon]
				if !ok {
					return canon, ChainMissing
				}
				canon = p
			}
		}
		last = p

		p, ok = tr.parents[canon]
		if !ok {
			// If we don't have the canonical history to connect back with
			// the local view, then we will need that first.
			return canon, ChainMissing
		}
		if p.Number != last.Number {
			panic("sanity check: equal height canonical chain and local chain view in reorg search")
		}
		// Check if the parent matches the local view at this height.
		// If yes, then we found the block to reorg with
		if p == last {
			return canon, ChainReorg
		}
		canon = p
	}
}

func (tr *tracker) Prune(number uint64) {
	tr.Lock()
	defer tr.Unlock()

	// Don't bother pruning parallel branches,
	// they will go out of scope by height eventually.
	for k := range tr.parents {
		if k.Number < number {
			delete(tr.parents, k)
		}
	}
}

// Tracker is a *cache* of chain connections.
// It helps quickly decide which blocks to pull down, and resolves reorgs.
// It does not track the past *processed* L1 blocks, it is robust against change.
type Tracker interface {
	// Parent inserts a link into the chain between the block and its parent.
	Parent(id BlockID, parent BlockID)
	// HeadSignal informs future Pull calls which chain to follow
	HeadSignal(id BlockID)
	// Pull the block to process on top of the last chain view.
	//
	// Depending on the returned chain-mode, this may mean extension, reorg,
	// or filling missing data first (possible long range sync)
	Pull(lastSeen BlockID) (next BlockID, mode ChainMode)
	// Prune everything older than the given block number
	Prune(number uint64)
	// WatchHeads subscribes to get head updates
	WatchHeads(chan<- BlockID) ethereum.Subscription
}

// WatchHeadChanges wraps a new-head subscription from ChainReader to feed the given Tracker
func WatchHeadChanges(ctx context.Context, chr ethereum.ChainReader, tr Tracker) (ethereum.Subscription, error) {
	headChanges := make(chan *types.Header)
	sub, err := chr.SubscribeNewHead(ctx, headChanges)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case header := <-headChanges:
				hash := header.Hash()
				height := header.Number.Uint64()
				self := BlockID{Hash: hash, Number: height}
				if height > 0 {
					tr.Parent(self, BlockID{Hash: header.ParentHash, Number: height - 1})
				}
				tr.HeadSignal(self)
			case err := <-sub.Err():
				return err
			case <-ctx.Done():
				return ctx.Err()
			case <-quit:
				return nil
			}
		}
	}), nil
}

// Confirmation depth: how far reorgs *of L1 blocks* are supported.
// TODO: move to configuration? Setting to 4 hours of 14 second blocks for now.
const PruneL1Distance = 4 * 60 * 60 / 14

func NewTracker() Tracker {
	return &tracker{
		parents: make(map[BlockID]BlockID),
		// head is zeroed and will be filled later
	}
}

type LocalView interface {
	// LastL1 returns the local L1 block in view
	LastL1() BlockID
	// ProcessL1 instructs to update the local view by fetching and processing the block.
	// - If id == LastL1(): no-op
	// - If id.Number <= LastL1().Number && id.Hash != LastL1().Hash: reorg or unwind
	// - If id.Number == LastL1().Number + 1: extend chain if parent matches, or reorg
	// - If id.Number > LastL1().Number + 1: far out of sync, including possible reorg,
	//   try catching up through other means if possible, headers are being synced to find 1-by-1 sync start.
	ProcessL1(dl Downloader, finalized common.Hash, id BlockID) // TODO: maybe add ChainMode arg, but keeping simplicity here is nice too
}

// HeaderSync fetches missing data for L1 tracker, and detects head changes / reorgs,
// to then instruct the local view to process a given L1 block.
func HeaderSync(ctx context.Context, chr ethereum.ChainReader, tr Tracker,
	log log.Logger, dl Downloader, local LocalView) ethereum.Subscription {

	syncStep := func(lastLocal BlockID) (nextLocal BlockID, fastNext bool) {
		id, mode := tr.Pull(lastLocal)
		if mode == ChainNoop {
			log.Info("fully synced", "hash", id.Hash, "height", id.Number)
			return lastLocal, false
		}

		switch mode {
		case ChainExtend:
			log.Info("fetching extended L1 chain", "hash", id.Hash, "height", id.Number)
		case ChainUndo:
			log.Info("rewinding L1 chain", "hash", id.Hash, "height", id.Number)
		case ChainReorg:
			log.Info("reorging L1 chain", "hash", id.Hash, "height", id.Number)
		case ChainMissing:
			log.Info("searching for history of L1 chain", "hash", id.Hash, "height", id.Number)
		}

		header, err := chr.HeaderByHash(ctx, id.Hash)
		if err != nil {
			log.Error("failed to fetch L1 header", "hash", id.Hash, "height", id.Number, "err", err)

			// Don't continue, wait for next cold tick.
			// Maybe we get a reorg event or other change that affects the next step,
			// which would explain the retrieval error.
			return lastLocal, false
		} else {
			// We retrieved the missing block, update the L1 data maintainer
			hash := header.Hash()
			height := header.Number.Uint64()
			self := BlockID{Hash: hash, Number: height}
			if height > 0 {
				tr.Parent(self, BlockID{Hash: header.ParentHash, Number: height - 1})
			}

			// let's continue with the next sync step quickly!
			return id, true
		}
	}

	// Every so often, attempt to sync. If sync is making progress, keep up a faster pace.
	pullCold := time.Second * 6
	pullHot := 30 * time.Millisecond
	pullTicker := time.NewTicker(pullCold)
	defer pullTicker.Stop()

	// Whenever we get a new head, start syncing at faster pace again
	heads := make(chan BlockID, 10)
	headsSub := tr.WatchHeads(heads)

	// Pruning saves memory, we don't need to cache block data deeper than the maximum expected L1 reorg depth
	pruneTicker := time.NewTicker(time.Minute * 2)
	defer pruneTicker.Stop()

	// TODO: track finalized hash (long distance from head)
	finalized := BlockID{}

	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer headsSub.Unsubscribe()
		for {
			select {
			case err := <-headsSub.Err():
				return err
			case <-heads:
				// new head, resume sync
				pullTicker.Reset(pullHot)
			case <-pullTicker.C:
				lastLocal := local.LastL1()
				nextLocal, fastNext := syncStep(lastLocal)
				if lastLocal != nextLocal {
					// If we know the block is already finalized we share that,
					// but not with a hash beyond the block itself.
					// *This finalized field does not affect the block hashes or execution,
					// only the syncing / user RPC: sync edge-cases may not affect the fraud proof*
					relativeFinalized := finalized.Hash
					if finalized.Number > nextLocal.Number {
						relativeFinalized = nextLocal.Hash
					}
					local.ProcessL1(dl, relativeFinalized, nextLocal)
				}
				// after completing processing, schedule the next step
				if fastNext {
					pullTicker.Reset(pullHot)
				} else {
					pullTicker.Reset(pullCold)
				}
			case <-pruneTicker.C:
				height := local.LastL1().Number
				if height > PruneL1Distance {
					tr.Prune(height - PruneL1Distance)
				}
			case <-ctx.Done():
				return ctx.Err()
			case <-quit:
				return nil
			}
		}
	})
}
