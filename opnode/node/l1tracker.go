package node

import (
	"context"
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

type L1Tracker struct {
	sync.Mutex

	// Parent hash (value) of each block (key)
	parents map[BlockID]BlockID
	// last seen head height, may not be the max height (total difficult != height)
	head BlockID
}

func (l1t *L1Tracker) HeadSignal(id BlockID) {
	l1t.Lock()
	defer l1t.Unlock()

	l1t.head = id
}

func (l1t *L1Tracker) Parent(id BlockID, parent BlockID) {
	l1t.Lock()
	defer l1t.Unlock()

	l1t.parents[id] = parent
}

func (l1t *L1Tracker) Pull(lastLocal BlockID) (next BlockID, mode ChainMode) {
	l1t.Lock()
	defer l1t.Unlock()

	// find the common block between the last view and the canonical chain
	canon := l1t.head
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
		p, ok := l1t.parents[last]
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
		p, ok := l1t.parents[canon]
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
		p, ok := l1t.parents[last]
		if !ok {
			// Our local view does not reach the canonical chain history.
			// Traverse back canonical chain history to find the lowest
			// canonical block the local view needs to be extended to.
			for {
				p, ok := l1t.parents[canon]
				if !ok {
					return canon, ChainMissing
				}
				canon = p
			}
		}
		last = p

		p, ok = l1t.parents[canon]
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

func (l1t *L1Tracker) Prune(number uint64) {
	l1t.Lock()
	defer l1t.Unlock()

	// Don't bother pruning parallel branches,
	// they will go out of scope by height eventually.
	for k := range l1t.parents {
		if k.Number < number {
			delete(l1t.parents, k)
		}
	}
}

// L1Maintainer is a *cache* of chain connections.
// It helps quickly decide which blocks to pull down, and resolves reorgs.
// It does not track the past *processed* L1 blocks, it is robust against change.
type L1Maintainer interface {
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
}

// SubL1Node wraps a new-head subscription from ChainReader to feed the given L1Maintainer
func SubL1Node(ctx context.Context, chr ethereum.ChainReader, l1m L1Maintainer) (ethereum.Subscription, error) {
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
					l1m.Parent(self, BlockID{Hash: header.ParentHash, Number: height - 1})
				}
				l1m.HeadSignal(self)
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// Confirmation depth: how far reorgs *of L1 blocks* are supported.
// TODO: move to configuration? Setting to 4 hours of 14 second blocks for now.
const PruneL1Distance = 4 * 60 * 60 / 14

func NewTracker(lastL1Head BlockID) *L1Tracker {
	return &L1Tracker{
		parents: make(map[BlockID]BlockID),
		head:    lastL1Head,
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
	ProcessL1(id BlockID) // TODO: maybe add ChainMode arg, but keeping simplicity here is nice too
}

// L1HeaderSync fetches missing data for L1 maintainer, and detects head changes / reorgs,
// to then instruct the local view to process a given L1 block.
func L1HeaderSync(ctx context.Context, chr ethereum.ChainReader, l1m L1Maintainer, log Logger, local LocalView) error {
	sub, err := SubL1Node(ctx, chr, l1m)
	if err != nil {
		return err
	}

	syncStep := func(lastLocal BlockID) (nextLocal BlockID, fastNext bool) {
		id, mode := l1m.Pull(lastLocal)
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
				l1m.Parent(self, BlockID{Hash: header.ParentHash, Number: height - 1})
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

	// Pruning saves memory, we don't need to cache block data deeper than the maximum expected L1 reorg depth
	pruneTicker := time.NewTicker(time.Minute * 2)
	defer pruneTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-sub.Err():
			return err
		case <-pullTicker.C:
			lastLocal := local.LastL1()
			nextLocal, fastNext := syncStep(lastLocal)
			if fastNext {
				pullTicker.Reset(pullHot)
			} else {
				pullTicker.Reset(pullCold)
			}
			if lastLocal != nextLocal {
				local.ProcessL1(nextLocal)
			}
		case <-pruneTicker.C:
			height := local.LastL1().Number
			if height > PruneL1Distance {
				l1m.Prune(height - PruneL1Distance)
			}
		}
	}
}
