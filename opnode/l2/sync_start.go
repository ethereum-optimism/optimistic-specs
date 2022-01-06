package l2

import (
	"context"
	"fmt"
	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

func FindSyncStart(ctx context.Context, l1Src eth.BlockHashByNumber, l2Src eth.BlockSource, genesis *Genesis) (refL1 eth.BlockID, refL2 eth.BlockID, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var parentL2 common.Hash
	// Start at L2 head
	refL1, refL2, parentL2, err = RefByL2Num(ctx, l2Src, nil, genesis)
	if err != nil {
		err = fmt.Errorf("failed to fetch L2 head: %v", err)
		return
	}
	// Check if L1 source has the block
	var l1Hash common.Hash
	l1Hash, err = l1Src.BlockHashByNumber(ctx, refL1.Number)
	if err != nil {
		err = fmt.Errorf("failed to lookup block %d in L1: %v", refL1.Number, err)
		return
	}
	if l1Hash == refL1.Hash {
		return
	}

	// Search back: linear walk back from engine head. Should only be as deep as the reorg.
	for refL2.Number > 0 {
		refL1, refL2, parentL2, err = RefByL2Hash(ctx, l2Src, parentL2, genesis)
		if err != nil {
			// TODO: re-attempt look-up, now that we already traversed previous history?
			err = fmt.Errorf("failed to lookup block %d in L1: %v", refL1.Number, err)
			return
		}
		// Check if L1 source has the block
		l1Hash, err = l1Src.BlockHashByNumber(ctx, refL1.Number)
		if err != nil {
			err = fmt.Errorf("failed to lookup block %d in L1: %v", refL1.Number, err)
			return
		}
		if l1Hash == refL1.Hash {
			return
		}
		// TODO: after e.g. initial N steps, use binary search instead
		// (relies on block numbers, not great for tip of chain, but nice-to-have in deep reorgs)
	}
	// Enforce that we build on the desired genesis block.
	// The engine might be configured for a different chain or older testnet.
	if refL2 != genesis.L2 {
		err = fmt.Errorf("engine was anchored at unexpected block: %s, expected %s", refL2, genesis.L2)
		return
	}
	// we got the correct genesis, all good, but a lot to sync!
	return
}
