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

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
)

type L2Chain interface {
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error)
}

var WrongChainErr = errors.New("wrong chain")
var TooDeepReorgErr = errors.New("reorg is too deep")

const MaxReorgDepth = 500

// FindUnsafeL2Head takes the supplied L2 start block and walks the L2 chain until it finds the highest L2 block who's L1 Origin is
// the same as the supplied L1 block.
func FindUnsafeL2Head(ctx context.Context, start eth.L2BlockRef, l1Base eth.BlockID, l2 L2Chain, genesis *rollup.Genesis) (eth.L2BlockRef, error) {
	// Reorg base is past L2 head
	if start.L1Origin.Number <= l1Base.Number {
		return start, nil
	}
	var err error
	reorgDepth := 0
	// Walk L2 chain from start until L1Origin matches l1Base. Bail out early of when at genesis.
	for n := start; ; {
		// Found matching block
		if n.L1Origin.Hash == l1Base.Hash {
			return n, nil
		}
		// Walking past expected L1 block.
		if n.L1Origin.Number < l1Base.Number {
			return eth.L2BlockRef{}, WrongChainErr
		}
		// Don't walk past genesis. If we were at the L2 genesis, but could not find the L1 genesis
		// pointed to from it, we are on the wrong L1 chain.
		if n.Self.Hash == genesis.L2.Hash || n.Self.Number == genesis.L2.Number {
			return eth.L2BlockRef{}, WrongChainErr
		}
		// Pull L2 parent for next iteration
		n, err = l2.L2BlockRefByHash(ctx, n.Parent.Hash)
		if err != nil {
			return eth.L2BlockRef{}, fmt.Errorf("failed to fetch L2 block by hash %v: %w", n.Parent.Hash, err)
		}
		reorgDepth++
		if reorgDepth >= MaxReorgDepth {
			return eth.L2BlockRef{}, TooDeepReorgErr
		}
	}
}

// FindSafeL2Head first finds the UnsafeL2Head and then walks L2 blocks until it passes `seqWindowSize` L1 Blocks.
func FindSafeL2Head(ctx context.Context, start eth.L2BlockRef, l1Base eth.BlockID, seqWindowSize int, l2 L2Chain, genesis *rollup.Genesis) (eth.L2BlockRef, error) {
	unsafeHead, err := FindUnsafeL2Head(ctx, start, l1Base, l2, genesis)
	if err != nil {
		return eth.L2BlockRef{}, fmt.Errorf("failed to fetch unsafe head: %w", err)
	}
	depth := 1 // SeqWindowSize is a length, but we are counting elements in the window.
	prevL1OriginHash := unsafeHead.L1Origin.Hash
	// Walk L2 chain. May walk to L2 genesis
	for n := unsafeHead; ; {
		if n.L1Origin.Hash != prevL1OriginHash {
			depth++
			prevL1OriginHash = n.L1Origin.Hash
		}
		// Advance depth if new origin
		if depth == seqWindowSize {
			return n, nil
		}
		// Genesis is always safe.
		if n.Self.Hash == genesis.L2.Hash || n.Self.Number == genesis.L2.Number {
			return eth.L2BlockRef{Self: genesis.L2, L1Origin: genesis.L1}, nil
		}
		// Pull L2 parent for next iteration
		n, err = l2.L2BlockRefByHash(ctx, n.Parent.Hash)
		if err != nil {
			return eth.L2BlockRef{}, fmt.Errorf("failed to fetch L2 block by hash %v: %w", n.Parent.Hash, err)
		}
	}
}
