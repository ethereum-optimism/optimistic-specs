package l1

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TODO: Make these configurable and part of a configuration object
const MaxConcurrentFetchesPerCall = 10
const MaxReceiptRetry = 3
const MaxBlocksInL1Range = uint64(100)

type DownloaderClient interface {
	BlockByHash(context.Context, common.Hash) (*types.Block, error)
	BlockByNumber(context.Context, *big.Int) (*types.Block, error)
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
}

// Block
// Receipts
// Transactions (from range of blocks)

type Downloader struct {
	client DownloaderClient
	// log    log.Logger

	// Block Cache
	// Receipt Cache
}

func NewDownloader(client DownloaderClient) *Downloader {
	return &Downloader{client: client}
}

func (dl *Downloader) FetchBlockAndReceipts(ctx context.Context, id eth.BlockID) (*types.Block, []*types.Receipt, error) {
	block, err := dl.client.BlockByHash(ctx, id.Hash)
	if err != nil {
		return nil, nil, err
	}
	txs := block.Transactions()
	receipts := make([]*types.Receipt, len(txs))

	semaphoreChan := make(chan struct{}, MaxConcurrentFetchesPerCall)
	defer close(semaphoreChan)
	var retErr error
	var errMu sync.Mutex
	var wg sync.WaitGroup
	for idx, tx := range txs {
		wg.Add(1)
		i := idx
		hash := tx.Hash()
		go func() {
			semaphoreChan <- struct{}{}
			for j := 0; j < MaxReceiptRetry; j++ {
				receipt, err := dl.client.TransactionReceipt(ctx, hash)
				if err != nil && j == MaxReceiptRetry-1 {
					// dl.log.Error("Got error in final retry of fetch", "err", err)
					errMu.Lock()
					retErr = err
					errMu.Unlock()
				} else if err == nil {
					receipts[i] = receipt
					break
				} else {
					time.Sleep(20 * time.Millisecond)
				}
			}
			wg.Done()
			<-semaphoreChan
		}()
	}
	wg.Wait()
	if retErr != nil {
		return nil, nil, retErr
	}
	return block, receipts, nil
}

func (dl *Downloader) FetchReceipts(ctx context.Context, id eth.BlockID) ([]*types.Receipt, error) {
	_, receipts, err := dl.FetchBlockAndReceipts(ctx, id)
	return receipts, err
}

func (dl *Downloader) FetchBlock(ctx context.Context, id eth.BlockID) (*types.Block, error) {
	return dl.client.BlockByHash(ctx, id.Hash)
}

func (dl *Downloader) FetchTransactions(ctx context.Context, window []eth.BlockID) ([]*types.Transaction, error) {
	var txns []*types.Transaction
	for _, id := range window {
		block, err := dl.client.BlockByHash(ctx, id.Hash)
		if err != nil {
			return nil, err
		}
		txns = append(txns, block.Transactions()...)
	}
	return txns, nil

}
