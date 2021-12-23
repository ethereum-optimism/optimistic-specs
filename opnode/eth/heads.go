package eth

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

type HeadSignalTracker interface {
	HeadSignal(parent BlockID, id BlockID)
}

// HeadSignalFn is called to inform a chain with a new head
type HeadSignalFn func(parent BlockID, id BlockID)

func (fn HeadSignalFn) HeadSignal(parent BlockID, id BlockID) {
	fn(parent, id)
}

// WatchHeadChanges wraps a new-head subscription from ChainReader to feed the given Tracker
func WatchHeadChanges(ctx context.Context, src NewHeadSource, tr HeadSignalTracker) (ethereum.Subscription, error) {
	headChanges := make(chan *types.Header, 10)
	sub, err := src.SubscribeNewHead(ctx, headChanges)
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
				parent := BlockID{}
				if height > 0 {
					parent = BlockID{Hash: header.ParentHash, Number: height - 1}
				}
				tr.HeadSignal(parent, self)
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
