// The sync package is responsible for reconciling L1 and L2.
//
//
// The ethereum chain is a DAG of blocks with the root block being the genesis block.
// At any given time, the head (or tip) of the chain can change if an offshoot of the chain
// has a higher number. This is known as a re-organization of the canonical chain.
// Each block points to a parent block and the node is responsible for deciding which block is the head
// and thus the mapping from block number to canonical block.
//
// The optimism chain has similar properties, but also retains references to the ethereum chain.
// Each optimism block retains a reference to an L1 block and to its parent L2 block.
// The L2 chain node must satisfy the following validity rules
//     1. l2block.height == l2parent.block.height + 1
//     2. l2block.l1parent.height >= l2block.l2parent.l1parent.height
//     3. l2block.l1parent is in the canonical chain on L1
//     4. l1_rollup_genesis is reachable from l2block.l1parent
//
//
// During normal operation, both the L1 and L2 canonical chains can change, due to a reorg
// or an extension (new block).
//     - L1 reorg
//     - L1 extension
//     - L2 reorg
//     - L2 extension
//
// When one of these changes occurs, the rollup node needs to determine what the new L2 Head should be.
// In a simple extension case, the L2 head remains the same, but in the case of a re-org on L1, it needs
// to find the first L2 block where the l1parent is in the L1 canonical chain.
// In the case of a re-org, it is also helpful to obtain the L1 blocks after the L1 base to re-start the
// chain derivation process.
//
// FindUnsafeL2Head finds the first L2 block that has an L1 block that is canonical (supply the reorg base as l1base)
// FindSafeL2Head finds the first L2 block that can be fully derived from a sequencing window that has not changed with the reorg.

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
// It will suppress ethereum.NotFound errors
func isCanonical(ctx context.Context, l1 L1Chain, block eth.BlockID) (bool, error) {
	canonical, err := l1.L1BlockRefByNumber(ctx, block.Number)
	if err != nil && !errors.Is(err, ethereum.NotFound) {
		return false, err
	} else if err != nil {
		return false, nil
	}
	return canonical.Hash == block.Hash, nil
}

// FindL2Heads walks back from the supplied L2 blocks and finds the unsafe and safe L2 heads.
// Unsafe Head: The first L2 block who's L1 Origin is canonical (determined via block by number).
// Safe Head: Walk back 1 sequence window of epochs from the Unsafe Head.
func FindL2Heads(ctx context.Context, start eth.L2BlockRef, seqWindowSize int,
	l1 L1Chain, l2 L2Chain, genesis *rollup.Genesis) (unsafeHead eth.L2BlockRef, safeHead eth.L2BlockRef, err error) {
	reorgDepth := 0
	var prevL1OriginHash common.Hash
	// First check if the L1 Origin of the start block is ahead of the current L1 head
	// If so, we assume that this should be the next unsafe head for the sequencing window
	// We still need to walk back the safe head because we don't know where the reorg started.
	l1Head, err := l1.L1HeadBlockRef(ctx)
	if err != nil {
		return eth.L2BlockRef{}, eth.L2BlockRef{}, err
	}
	l2Ahead := start.L1Origin.Number > l1Head.Number

	// Walk L2 chain from start until L1Origin matches l1Base. Bail out early of when at genesis.
	for n := start; ; {
		// Check if l1Origin is canonical when we get to a new epoch
		if prevL1OriginHash != n.L1Origin.Hash {
			if ok, err := isCanonical(ctx, l1, n.L1Origin); err != nil {
				return eth.L2BlockRef{}, eth.L2BlockRef{}, err
			} else if ok {
				unsafeHead = n
				break
			}
			prevL1OriginHash = n.L1Origin.Hash
		}
		// Don't walk past genesis. If we were at the L2 genesis, but could not find the L1 genesis
		// pointed to from it, we are on the wrong L1 chain.
		if n.Hash == genesis.L2.Hash || n.Number == genesis.L2.Number {
			return eth.L2BlockRef{}, eth.L2BlockRef{}, WrongChainErr
		}
		// Pull L2 parent for next iteration
		n, err = l2.L2BlockRefByHash(ctx, n.ParentHash)
		if err != nil {
			return eth.L2BlockRef{}, eth.L2BlockRef{}, fmt.Errorf("failed to fetch L2 block by hash %v: %w", n.ParentHash, err)
		}
		reorgDepth++
		if reorgDepth >= MaxReorgDepth {
			return eth.L2BlockRef{}, eth.L2BlockRef{}, TooDeepReorgErr
		}
	}
	depth := 1 // SeqWindowSize is a length, but we are counting elements in the window.
	prevL1OriginHash = unsafeHead.L1Origin.Hash
	// Walk L2 chain. May walk to L2 genesis
	for n := unsafeHead; ; {
		// Advance depth if new origin
		if n.L1Origin.Hash != prevL1OriginHash {
			depth++
			prevL1OriginHash = n.L1Origin.Hash
		}
		// Walked sufficiently far
		if depth == seqWindowSize {
			if l2Ahead {
				return start, n, nil
			} else {
				return unsafeHead, n, nil
			}

		}
		// Genesis is always safe.
		if n.Hash == genesis.L2.Hash || n.Number == genesis.L2.Number {
			safeHead = eth.L2BlockRef{Hash: genesis.L2.Hash, Number: genesis.L2.Number, Time: genesis.L2Time, L1Origin: genesis.L1}
			if l2Ahead {
				return start, safeHead, nil
			} else {
				return unsafeHead, safeHead, nil
			}

		}
		// Pull L2 parent for next iteration
		n, err = l2.L2BlockRefByHash(ctx, n.ParentHash)
		if err != nil {
			return eth.L2BlockRef{}, eth.L2BlockRef{}, fmt.Errorf("failed to fetch L2 block by hash %v: %w", n.ParentHash, err)
		}
	}

}
