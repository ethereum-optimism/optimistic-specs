package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum/go-ethereum/common"
)

var WrongChainErr = errors.New("wrong chain")

func V2FindSyncStart(ctx context.Context, reference SyncReference, genesis *rollup.Genesis) (nextL1s []eth.BlockID, refL2 eth.BlockID, err error) {
	// Start the search for a matching chain between L1 and L2 with the L2 head
	refL1, refL2, parentL2, err := reference.RefByL2Num(ctx, nil, genesis)
	if err != nil {
		return nil, eth.BlockID{}, fmt.Errorf("failed to fetch L2 head: %v", err)
	}

	maxBuffer := uint64(100)

	// holds the traversed L1 hashes in a circular buffer, to build nextL1s from, as far as we can
	cachedL1s := make([]eth.BlockID, maxBuffer)
	// we cannot use the cache beyond the first block we fill it with, or any reorg or missing block we find along the way
	minCacheL1Num := refL1.Number + 1
	var currentL1, parentL1 eth.BlockID
	for i := refL1.Number; i > genesis.L1.Number; i-- {
		prevParent := parentL1
		// 1: check if L1 has it (by number)
		currentL1, parentL1, err = reference.RefByL1Num(ctx, refL1.Number)
		if err != nil {
			if !errors.Is(err, ethereum.NotFound) {
				return nil, eth.BlockID{}, fmt.Errorf("failed to lookup block %d in L1: %w", refL1.Number, err)
			}
			// L1 doesn't have it, keep going
			// Don't recognize this or later hashes by number though, it may have been reorged.
			minCacheL1Num = i
		} else {
			// If this does not align with later blocks, then don't recognize the later blocks
			if prevParent != (eth.BlockID{}) && prevParent != currentL1 {
				minCacheL1Num = i + 1 // +1 because we can still recognize the current block
			}
			if currentL1 == refL1 {
				break
			}
		}
		// buffer actual L1
		cachedL1s[i%maxBuffer] = currentL1

		// continue with next L2 block
		refL1, refL2, parentL2, err = reference.RefByL2Hash(ctx, parentL2, genesis)
		if err != nil {
			// TODO: re-attempt look-up, now that we already traversed previous history?
			err = fmt.Errorf("failed to lookup block %s in L2: %w", refL2, err) // refL2 is previous parentL2
			return
		}
	}
	// if we have different genesis, then stop
	if (refL2.Number <= genesis.L2.Number && refL2.Hash != genesis.L2.Hash) ||
		(refL1.Number <= genesis.L1.Number && refL1.Hash != genesis.L1.Hash) {
		return nil, genesis.L2, fmt.Errorf("cannot find L1 blocks building on L2 genesis")
	}

	// Collect the L1 chain that builds on the common point, to derive new L2 blocks from.
	nextL1s = make([]eth.BlockID, 1, maxBuffer)
	nextL1s[0] = refL1
	// 3: walk forward L1 from common point to get nextL1s
	for i := uint64(1); i < maxBuffer; i++ {
		// if it's at or before the first L1 block we tried to match, then we cached it
		if nextL1s[0].Number+i < minCacheL1Num {
			nextL1s = append(nextL1s, cachedL1s[i%maxBuffer])
		} else {
			currentL1, parentL1, err = reference.RefByL1Num(ctx, nextL1s[0].Number+i)
			if nextL1s[i-1] != parentL1 { // check if it extends the last block
				// even though we hit an error (e.g. commonly block not found),
				// there is no need to error, we just return less results,
				// to avoid an inconsistent nextL1s list.
				break
			}
			nextL1s = append(nextL1s, currentL1)
		}
	}
	// skip the first L1 block that we already have in common
	return nextL1s[1:], refL2, nil
}

// FindSyncStart finds nextL1s: the L1 blocks needed next for sync, to derive into a L2 block on top of refL2.
// If the L1 reorgs then this will find the common history to build on top of and then follow the first step of the reorg.
func FindSyncStart(ctx context.Context, reference SyncReference, genesis *rollup.Genesis) (nextL1s []eth.BlockID, refL2 eth.BlockID, err error) {
	var refL1 eth.BlockID    // the L1 block the refL2 was derived from
	var parentL2 common.Hash // the parent of refL2
	// Start at L2 head
	refL1, refL2, parentL2, err = reference.RefByL2Num(ctx, nil, genesis)
	if err != nil {
		err = fmt.Errorf("failed to fetch L2 head: %v", err)
		return
	}
	// Check if L1 source has the block
	var currentL1 eth.BlockID // the expected L1 block at the height of refL1
	currentL1, _, err = reference.RefByL1Num(ctx, refL1.Number)
	if err != nil {
		if !errors.Is(err, ethereum.NotFound) {
			err = fmt.Errorf("failed to lookup block %d in L1: %w", refL1.Number, err)
			return
		}
		// If the L1 did not find the block, it might be out of sync.
		// We cannot sync from L1 in this case, but we still traverse back to
		// make sure we are not just in a reorg to a L1 chain with fewer blocks.
		err = nil
		currentL1 = eth.BlockID{} // empty = not found
	}
	if currentL1 == refL1 {
		// TODO: try get the next N blocks, instead of next 1 ...
		// L1 node has head-block of execution-engine, so we should fetch the L1 block that builds on top.
		var nextRefL1, ontoL1 eth.BlockID // ontoL1 is the parent, to make sure we got a nextRefL1 that connects as expected.
		nextRefL1, ontoL1, err = reference.RefByL1Num(ctx, refL1.Number+1)
		if err != nil {
			// If refL1 is the head block, then we might not have a next block to build on the head
			if errors.Is(err, ethereum.NotFound) {
				// return the same as the engine head was already built on, no error.
				nextRefL1 = refL1
				refL2 = eth.BlockID{Hash: parentL2, Number: refL2.Number}
				if refL2.Number > 0 {
					refL2.Number -= 1
				}
				err = nil
				return
			}
			return
		}
		nextL1s = append(nextL1s, nextRefL1)
		// The L1 source might rug us with a reorg between API calls, catch that.
		if ontoL1 != currentL1 {
			err = fmt.Errorf("the L1 source reorged, the block for N+1 %s doesn't have the previously fetched block N %s as parent, but builds on %s instead", nextRefL1, currentL1, ontoL1)
		}
		return
	}

	// Search back: linear walk back from engine head. Should only be as deep as the reorg.
	for refL2.Number > 0 {
		// remember the canonical L1 block that builds on top of the L1 source block of the L2 parent block.
		nextRefL1 = currentL1
		refL1, refL2, parentL2, err = reference.RefByL2Hash(ctx, parentL2, genesis)
		if err != nil {
			// TODO: re-attempt look-up, now that we already traversed previous history?
			err = fmt.Errorf("failed to lookup block %s in L2: %w", refL2, err) // refL2 is previous parentL2
			return
		}
		// Check if L1 source has the block that derived the L2 block we are planning to build on
		currentL1, _, err = reference.RefByL1Num(ctx, refL1.Number)
		if err != nil {
			if !errors.Is(err, ethereum.NotFound) {
				err = fmt.Errorf("failed to lookup block %d in L1: %w", refL1.Number, err)
				return
			}
			// again, if L1 does not have the block, then we just search if we are reorging.
			err = nil
			currentL1 = eth.BlockID{} // empty = not found
		}
		if currentL1 == refL1 {
			// check if we had a L1 block to build on top of the common chain with
			if nextRefL1 == (eth.BlockID{}) {
				err = ethereum.NotFound
			}
			return
		}
		// TODO: after e.g. initial N steps, use binary search instead
		// (relies on block numbers, not great for tip of chain, but nice-to-have in deep reorgs)
	}
	// Enforce that we build on the desired genesis block.
	// The engine might be configured for a different chain or older testnet.
	if refL2 != genesis.L2 {
		err = fmt.Errorf("unexpected L2 genesis block: %s, expected %s, %w", refL2, genesis.L2, WrongChainErr)
		return
	}
	if currentL1 != genesis.L1 {
		err = fmt.Errorf("unexpected L1 anchor block: %s, expected %s, %w", currentL1, genesis.L1, WrongChainErr)
		return
	}
	// we got the correct genesis, all good, but a lot to sync!
	return
}
