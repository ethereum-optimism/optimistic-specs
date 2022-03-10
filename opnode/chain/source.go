package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ChainSource provides access to the L1 and L2 block graph
type ChainSource interface {
	L1BlockRefByNumber(ctx context.Context, number *big.Int) (eth.L1BlockRef, error) // TODO: Implement by Hash Version as well
	L2BlockRefByNumber(ctx context.Context, number *big.Int, genesis *rollup.Genesis) (eth.L2BlockRef, error)
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash, genesis *rollup.Genesis) (eth.L2BlockRef, error)
}

var WrongChainErr = errors.New("wrong chain")
var TooDeepReorgErr = errors.New("reorg is too deep")

// TODO: Make configurable
var MaxReorgDepth = 500

// L1ChainClient is the subset of methods that Chain needs to determine the L1 block graph
type L1ChainClient interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

// L2ChainClient is the subset of methods that Chain needs to determine the L2 block graph
type L2ChainClient interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
}

func NewChainSource(l1 L1ChainClient, l2 L2ChainClient) *Source {
	return &Source{l1: l1, l2: l2}
}

type Source struct {
	l1 L1ChainClient
	l2 L2ChainClient
}

// L1BlockRefByNumber returns the canonical block and parent ids.
func (src Source) L1BlockRefByNumber(ctx context.Context, number *big.Int) (eth.L1BlockRef, error) {
	header, err := src.l1.HeaderByNumber(ctx, number)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L1BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", number, err)
	}
	parentNum := header.Number.Uint64()
	if parentNum > 0 {
		parentNum -= 1
	}
	return eth.L1BlockRef{
		Self:   eth.BlockID{Hash: header.Hash(), Number: header.Number.Uint64()},
		Parent: eth.BlockID{Hash: header.ParentHash, Number: parentNum},
	}, nil
}

// L2BlockRefByNumber returns the canonical block and parent ids.
func (src Source) L2BlockRefByNumber(ctx context.Context, number *big.Int, genesis *rollup.Genesis) (eth.L2BlockRef, error) {
	block, err := src.l2.BlockByNumber(ctx, number)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", number, err)
	}
	return derive.BlockReferences(block, genesis)
}

// L2BlockRefByHash returns the block & parent ids based on the supplied hash. The returned BlockRef may not be in the canonical chain
func (src Source) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash, genesis *rollup.Genesis) (eth.L2BlockRef, error) {
	block, err := src.l2.BlockByHash(ctx, l2Hash)
	if err != nil {
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Hash, err)
	}
	return derive.BlockReferences(block, genesis)
}
