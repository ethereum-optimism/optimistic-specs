package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BlockHashByNumber Retrieves the *currently* canonical block-hash at the given block-height.
// The results of this should not be cached, or the cache needs to be reorg-aware.
type BlockHashByNumber interface {
	BlockHashByNumber(ctx context.Context, num uint64) (common.Hash, error)
}

type BlockHashByNumberFn func(ctx context.Context, num uint64) (common.Hash, error)

func (fn BlockHashByNumberFn) BlockHashByNumber(ctx context.Context, num uint64) (common.Hash, error) {
	return fn(ctx, num)
}

// CanonicalChain presents the block-hashes by height by wrapping a header-source
// (useful due to lack of a direct JSON RPC endpoint)
func CanonicalChain(l1Src HeaderByNumberSource) BlockHashByNumber {
	return BlockHashByNumberFn(func(ctx context.Context, num uint64) (common.Hash, error) {
		header, err := l1Src.HeaderByNumber(ctx, big.NewInt(int64(num)))
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to determine block-hash of height %d, could not get header: %v", num, err)
		}
		return header.Hash(), nil
	})
}
