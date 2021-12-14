package l1

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"sync"
	"sync/atomic"
	"time"
)

type BlockAndReceipts struct {
	// Count receipts to track status
	// First field of struct for memory aligned atomic access
	DownloadedReceipts uint32

	Block *types.Block
	// allocated in advance, one for each transaction, nil until downloaded
	Receipts []*types.Receipt
}

type downloaderProcess struct {
	// Don't download the same block in parallel
	sync.Mutex
	cache        map[common.Hash]*BlockAndReceipts
	chr          ChainReaderWithReceipts
	log          log.Logger
	requests     <-chan BlockID
	results      event.Feed
	receiptTasks chan *ReceiptTask
	quitWg       sync.WaitGroup
	stop         func()
}

type ChainReaderWithReceipts interface {
	ethereum.ChainReader
	ethereum.TransactionReader
}

type ReceiptTask struct {
	BlockHash common.Hash
	TxHash    common.Hash
	TxIndex   uint64
	// Avoid concurrent Downloader cache access and pruning edge cases with receipts
	// Keep a pointer to insert the receipt at
	Dest *BlockAndReceipts
}

func (d *downloaderProcess) fetchBlock(ctx context.Context, id BlockID, receiptTasks chan<- *ReceiptTask) {
	// check if we are already working on it
	d.Lock()
	if _, ok := d.cache[id.Hash]; ok {
		d.Unlock()
		return
	}
	bnr := new(BlockAndReceipts)
	d.cache[id.Hash] = bnr
	d.Unlock()

	ctx, _ = context.WithTimeout(ctx, time.Second*10)
	bl, err := d.chr.BlockByHash(ctx, id.Hash)
	if err != nil {
		// Failing to fetch the block just means we'll be requested again later to fetch it.
		d.log.Warn("Failed to fetch full L1 block, skipping", "hash", id.Hash, "err", err)
		d.Lock()
		d.cache[id.Hash] = nil
		d.Unlock()
		return
	}
	h := bl.Hash()
	if h != id.Hash {
		// sanity check
		d.log.Error("Retrieved wrong block from RPC", "requested", id.Hash, "got", h)
		d.Lock()
		d.cache[id.Hash] = nil
		d.Unlock()
		return
	}

	txs := bl.Transactions()
	bnr.Block = bl
	bnr.Receipts = make([]*types.Receipt, len(txs), len(txs))

	for i, tx := range txs {
		receiptTasks <- &ReceiptTask{BlockHash: h, TxHash: tx.Hash(), TxIndex: uint64(i), Dest: bnr}
	}
}

// returns the block with all other receipts, or nil if there are more receipts left to fetch.
func (d *downloaderProcess) fetchReceipt(ctx context.Context, task *ReceiptTask) *BlockAndReceipts {
	ctx, _ = context.WithTimeout(ctx, time.Second*10)
	receipt, err := d.chr.TransactionReceipt(ctx, task.TxHash)
	if err != nil {

	}
	task.Dest.Receipts[task.TxIndex] = receipt
	total := atomic.AddUint32(&task.Dest.DownloadedReceipts, 1)
	if total == uint32(len(task.Dest.Receipts)) {
		return task.Dest
	}
	return nil
}

// fetch blocks and receipts in parallel (we can't fetch all receipts in one request, RPC design problem)
func (d *downloaderProcess) blockWorker(ctx context.Context) {
	defer d.quitWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-d.requests:
			if !ok {
				// exit worker if requests stopped
				return
			}
			d.fetchBlock(ctx, id, d.receiptTasks)
		}
	}
}

func (d *downloaderProcess) receiptWorker(ctx context.Context) {
	defer d.quitWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-d.receiptTasks:
			if !ok {
				// exit worker if task pipeline closed
				return
			}
			if res := d.fetchReceipt(ctx, task); res != nil {
				// finished downloading full block with receipts!
				d.results.Send(res)
			}
		}
	}
}

func (d *downloaderProcess) Listen(ch chan<- *BlockAndReceipts) {
	d.results.Subscribe(ch)
}

// TODO: there must be something more elegant than absuing a subscription to manage lifetime of workers
func (d *downloaderProcess) run(ctx context.Context) ethereum.Subscription {
	return event.NewSubscription(func(quit <-chan struct{}) error {
		workersCtx, workersStop := context.WithCancel(ctx)
		defer workersStop()

		defer func() {
			go func() {
				for range d.receiptTasks {
					// Drain the remaining receiptTasks to unblock any workers
				}
			}()
			d.quitWg.Wait()
		}()

		for i := 0; i < 4; i++ {
			go d.blockWorker(workersCtx)
		}

		for i := 0; i < 30; i++ {
			go d.receiptWorker(workersCtx)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-quit:
			return nil
		}
	})
}

func (d *downloaderProcess) Close() {
	d.stop()
}

type Downloader interface {
	Listen(ch chan<- *BlockAndReceipts)
	Close()
}

func NewDownloader(ctx context.Context, chr ChainReaderWithReceipts, log log.Logger, requests <-chan BlockID) *downloaderProcess {
	// buffer requests, and back-pressure the block workers if the receipt workers are busy
	// TODO: can do more concurrent requests, but API would get overwhelmed quickly.
	// Maybe multiply size by number of L1 sources, and cycle between the sources for requests?

	d := &downloaderProcess{
		cache:        make(map[common.Hash]*BlockAndReceipts),
		chr:          chr,
		log:          log,
		requests:     requests,
		receiptTasks: make(chan *ReceiptTask, 100),
	}
	d.stop = d.run(ctx).Unsubscribe
	return d
}
