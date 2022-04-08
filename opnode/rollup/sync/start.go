// The sync package is responsible for reconciling L1 and L2.
//
// The ethereum chain is a DAG of blocks with the root block being the genesis block. At any given
// time, the head (or tip) of the chain can change if an offshoot of the chain has a higher number.
// This is known as a re-organization of the canonical chain. Each block points to a parent block
// and the node is responsible for deciding which block is the head and thus the mapping from block
// number to canonical block.
//
// The Optimism chain has similar properties, but also retains references to the ethereum chain.
// Each Optimism block retains a reference to an L1 block (its "L1 origin", i.e. L1 block associated
// with the epoch that the L2 block belongs to) and to its parent L2 block. The L2 chain node must
// satisfy the following validity rules:
//     1. l2block.number == l2block.l2parent.block.number + 1
//     2. l2block.l1Origin.number >= l2block.l2parent.l1Origin.number
//     3. l2block.l1Origin is in the canonical chain on L1
//     4. l1_rollup_genesis is an ancestor of l2block.l1Origin
//
// During normal operation, both the L1 and L2 canonical chains can change, due to a re-organisation
// or due to an extension (new L1 or L2 block).
//
// When one of these changes occurs, the rollup node needs to determine what the new L2 head blocks
// should be. We track two L2 head blocks:
//     - The *unsafe L2 block*: This is the highest L2 block whose L1 origin is a plausible (1)
//       extension of the canonical L1 chain (as known to the opnode).
//     - The *safe L2 block*: This is the highest L2 block whose epoch's sequencing window is
//       complete within the canonical L1 chain (as known to the opnode).
//
// (1) Plausible meaning that the blockhash of the L2 block's L1 origin (as reported in the L1
//     Attributes deposit within the L2 block) is not canonical at another height in the L1 chain.
//
// In particular, in the case of L1 extension, the L2 unsafe head will generally remain the same,
// but in the case of an L1 re-org, we need to search for the new safe and unsafe L2 block.
package sync

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
)

type L1Chain interface {
	L1HeadBlockRef(ctx context.Context) (eth.L1BlockRef, error)
	L1BlockRefByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error)
}

type L2Chain interface {
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error)
}

var WrongChainErr = errors.New("wrong chain")
var TooDeepReorgErr = errors.New("reorg is too deep")

const MaxReorgDepth = 500

// isCanonical returns true if the supplied block ID is canonical in the L1 chain.
// If the L1 block is not found by number (ethereum.NotFound), the error will be nil.
func isCanonical(ctx context.Context, l1 L1Chain, block eth.BlockID) (bool, error) {
	canonical, err := l1.L1BlockRefByNumber(ctx, block.Number)
	if err != nil && !errors.Is(err, ethereum.NotFound) {
		return false, err
	} else if err != nil {
		return false, nil
	}
	return canonical.Hash == block.Hash, nil
}

// FindL2Heads walks back from `start` (the previous unsafe L2 block) and finds the unsafe and safe
// L2 blocks.
//
//     - The *unsafe L2 block*: This is the highest L2 block whose L1 origin is a plausible (1)
//       extension of the canonical L1 chain (as known to the opnode).
//     - The *safe L2 block*: This is the highest L2 block whose epoch's sequencing window is
//       complete within the canonical L1 chain (as known to the opnode).
//
// (1) Plausible meaning that the blockhash of the L2 block's L1 origin (as reported in the L1
//     Attributes deposit within the L2 block) is not canonical at another height in the L1 chain.
func FindL2Heads(ctx context.Context, start eth.L2BlockRef, seqWindowSize uint64,
	l1 L1Chain, l2 L2Chain, genesis *rollup.Genesis) (unsafe eth.L2BlockRef, safe eth.L2BlockRef, err error) {

	// First check if the L1 origin of the start block is ahead of the current L1 head.
	// If so, we assume that this should be the next unsafe head for the sequencing window.
	// We still need to walk back the safe head because we don't know where the reorg started.
	l1Head, err := l1.L1HeadBlockRef(ctx)
	if err != nil {
		return eth.L2BlockRef{}, eth.L2BlockRef{}, err
	}
	l2Ahead := start.L1Origin.Number > l1Head.Number

	// The loop that follows walks the L2 chain until we find the "latest" L2 block.
	// This the first L2 block whose L1 origin is canonical.
	var latest eth.L2BlockRef

	// L1 origin hash during for the L2 block during the previous iteration, 0 for first iteration.
	// When this changes as we walk the L2 chain backwards, it means we're seeing a different
	// (earlier) epoch.
	var prevL1OriginHash common.Hash

	// Current L2 block.
	n := start

	// Number of blocks between n and start.
	reorgDepth := 0

	for {
		// Check if l1Origin is canonical when we get to a new epoch.
		if prevL1OriginHash != n.L1Origin.Hash {
			if ok, err := isCanonical(ctx, l1, n.L1Origin); err != nil {
				return eth.L2BlockRef{}, eth.L2BlockRef{}, err
			} else if ok {
				// L1 block is canonical, "latest" L2 block found.
				latest = n
				prevL1OriginHash = n.L1Origin.Hash
				break
			} else {
				// L1 block not canonical (either we don't know an L1 block at this height, or
				// have another block hash for it), keep looking.
				prevL1OriginHash = n.L1Origin.Hash
			}
		}

		// Don't walk past genesis. If we were at the L2 genesis, but could not find its L1 origin,
		// the L2 chain is building on the wrong L1 branch.
		if n.Hash == genesis.L2.Hash || n.Number == genesis.L2.Number {
			return eth.L2BlockRef{}, eth.L2BlockRef{}, WrongChainErr
		}

		// Pull L2 parent for next iteration
		n, err = l2.L2BlockRefByHash(ctx, n.ParentHash)
		if err != nil {
			return eth.L2BlockRef{}, eth.L2BlockRef{},
				fmt.Errorf("failed to fetch L2 block by hash %v: %w", n.ParentHash, err)
		}

		reorgDepth++
		if reorgDepth >= MaxReorgDepth {
			// If the reorg depth is too large, something is fishy.
			// This can legitimately happen if L1 goes down for a while. But in that case,
			// restarting the L2 node with a bigger configured MaxReorgDepth is an acceptable
			// stopgap solution.
			// Currently this can also happen if the L2 node is down for a while, but in the future
			// state sync should prevent this issue.
			return eth.L2BlockRef{}, eth.L2BlockRef{}, TooDeepReorgErr
		}
	}

	// Walk from the L1 origin of the "latest" block back to the L1 block that starts the sequencing
	// window ending at that block. Instead of iterating on L1 blocks, we actually iterate on L2
	// blocks, because we want to find the safe head, i.e. the highest L2 block whose L1 origin
	// is the start of the sequencing window.

	// Depth counter: we need to walk back `seqWindowSize` L1 blocks in order to find the start
	// of the sequencing window.
	depth := uint64(1)

	// Before entering the loop: n == latest && prevL1OriginHash == n.L1Origin.Hash
	for {
		// Advance depth if we change to a different (earlier) epoch.
		if n.L1Origin.Hash != prevL1OriginHash {
			depth++
			prevL1OriginHash = n.L1Origin.Hash
		}

		// Found an L2 block whose L1 origin is the start of the sequencing window.
		if depth == seqWindowSize {
			if l2Ahead {
				return start, n, nil
			} else {
				return latest, n, nil
			}
		}

		// Genesis is always safe.
		if n.Hash == genesis.L2.Hash || n.Number == genesis.L2.Number {
			safe = eth.L2BlockRef{Hash: genesis.L2.Hash, Number: genesis.L2.Number,
				Time: genesis.L2Time, L1Origin: genesis.L1}
			return start, safe, nil
		}

		// Pull L2 parent for next iteration.
		n, err = l2.L2BlockRefByHash(ctx, n.ParentHash)
		if err != nil {
			return eth.L2BlockRef{}, eth.L2BlockRef{},
				fmt.Errorf("failed to fetch L2 block by hash %v: %w", n.ParentHash, err)
		}
	}
}
