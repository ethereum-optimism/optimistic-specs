package driver

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/sync"
	"github.com/ethereum/go-ethereum/log"
)

type state struct {
	// Chain State
	l1Head      eth.L1BlockRef // Latest recorded head of the L1 Chain
	l2Head      eth.L2BlockRef // L2 Unsafe Head
	l2SafeHead  eth.L2BlockRef // L2 Safe Head - this is the head of the L2 chain as derived from L1 (thus it is Sequencer window blocks behind)
	l2Finalized eth.BlockID    // L2 Block that will never be reversed
	l1WindowBuf []eth.BlockID  // l1WindowBuf buffers the next L1 block IDs to derive new L2 blocks from, with increasing block height.

	// Rollup config
	Config    rollup.Config
	sequencer bool

	// Connections (in/out)
	l1Heads <-chan eth.L1BlockRef
	l1      L1Chain
	l2      L2Chain
	output  outputInterface
	bss     BatchSubmitter

	log  log.Logger
	done chan struct{}

	closed uint32 // non-zero when closed
}

func NewState(log log.Logger, config rollup.Config, l1 L1Chain, l2 L2Chain, output outputInterface, submitter BatchSubmitter, sequencer bool) *state {
	return &state{
		Config:    config,
		done:      make(chan struct{}),
		log:       log,
		l1:        l1,
		l2:        l2,
		output:    output,
		bss:       submitter,
		sequencer: sequencer,
	}
}

// Start starts up the state loop. The context is only for initilization.
// The loop will have been started iff err is not nil.
func (s *state) Start(ctx context.Context, l1Heads <-chan eth.L1BlockRef) error {
	l1Head, err := s.l1.L1HeadBlockRef(ctx)
	if err != nil {
		return err
	}

	// Check that we are past the genesis
	if l1Head.Number > s.Config.Genesis.L1.Number {
		l2Head, err := s.l2.L2BlockRefByNumber(ctx, nil)
		if err != nil {
			return err
		}
		// Ensure that we are on the correct chain. Note that we cannot rely on rely on the UnsafeHead being more than
		// a sequence window behind the L1 Head and must walk back 1 sequence window as we do not track the end L1 block
		// hash of the sequence window when we derive an L2 block.
		unsafeHead, safeHead, err := sync.FindL2Heads(ctx, l2Head, s.Config.SeqWindowSize, s.l1, s.l2, &s.Config.Genesis)
		if err != nil {
			return err
		}
		s.l2Head = unsafeHead
		s.l2SafeHead = safeHead

	} else {
		// Not yet reached genesis block
		// TODO: Test this codepath. That requires setting up L1, letting it run, and then creating the L2 genesis from there.
		// Note: This will not work for setting the the genesis normally, but if the L1 node is not yet synced we could get this case.
		l2genesis := eth.L2BlockRef{
			Hash:     s.Config.Genesis.L2.Hash,
			Number:   s.Config.Genesis.L2.Number,
			Time:     s.Config.Genesis.L2Time,
			L1Origin: s.Config.Genesis.L1,
		}
		s.l2Head = l2genesis
		s.l2SafeHead = l2genesis
	}

	s.l1Head = l1Head
	s.l1Heads = l1Heads

	go s.loop()
	return nil
}

func (s *state) Close() error {
	close(s.done)
	return nil
}

// l1WindowBufEnd returns the last block that should be used as `base` to L1ChainWindow.
// This is either the last block of the window, or the L1 base block if the window is not populated.
func (s *state) l1WindowBufEnd() eth.BlockID {
	if len(s.l1WindowBuf) == 0 {
		return s.l2Head.L1Origin
	}
	return s.l1WindowBuf[len(s.l1WindowBuf)-1]
}

func (s *state) handleNewL1Block(ctx context.Context, newL1Head eth.L1BlockRef) error {
	if s.l1Head.Hash == newL1Head.Hash {
		log.Trace("Received L1 head signal that is the same as the current head", "l1Head", newL1Head)
		return nil
	}

	if s.l1Head.Hash == newL1Head.ParentHash {
		s.log.Trace("Linear extension", "l1Head", newL1Head)
		s.l1Head = newL1Head
		if s.l1WindowBufEnd().Hash == newL1Head.ParentHash {
			s.l1WindowBuf = append(s.l1WindowBuf, newL1Head.ID())
		}
		return nil
	}
	// New L1 Head is not the same as the current head or a single step linear extension.
	// This could either be a long L1 extension, or a reorg. Both can be handled the same way.
	s.log.Warn("L1 Head signal indicates an L1 re-org", "old_l1_head", s.l1Head, "new_l1_head_parent", newL1Head.ParentHash, "new_l1_head", newL1Head)
	unsafeL2Head, safeL2Head, err := sync.FindL2Heads(ctx, s.l2Head, s.Config.SeqWindowSize, s.l1, s.l2, &s.Config.Genesis)
	if err != nil {
		s.log.Error("Could not get new unsafe L2 head when trying to handle a re-org", "err", err)
		return err
	}
	// Update forkchoice
	fc := l2.ForkchoiceState{
		HeadBlockHash:      unsafeL2Head.Hash,
		SafeBlockHash:      safeL2Head.Hash,
		FinalizedBlockHash: s.l2Finalized.Hash,
	}
	_, err = s.l2.ForkchoiceUpdate(ctx, &fc, nil)
	if err != nil {
		s.log.Error("Could not set new forkchoice when trying to handle a re-org", "err", err)
		return err
	}
	// State Update
	s.l1Head = newL1Head
	s.l1WindowBuf = nil
	s.l2Head = unsafeL2Head
	// Don't advance l2SafeHead past it's current value
	if s.l2SafeHead.Number >= safeL2Head.Number {
		s.l2SafeHead = safeL2Head
	}

	return nil
}

// findNextL1Origin determines what the next L1 Origin should be.
// The L1 Origin is either the L2 Head's Origin, or the following L1 block
// if the next L2 block's time is greater than or equal to the L2 Head's Origin.
func (s *state) findNextL1Origin(ctx context.Context) (eth.L1BlockRef, error) {
	// If we are at the head block, don't do a lookup.
	// Don't do a timestamp check either as we are unable to get the next block even if we wanted to.
	if s.l2Head.L1Origin.Hash == s.l1Head.Hash {
		return s.l1Head, nil
	}

	// Grab the block ref
	curr, err := s.l1.L1BlockRefByHash(ctx, s.l2Head.L1Origin.Hash)
	if err != nil {
		return eth.L1BlockRef{}, err
	}
	// Somehow reorg'd. Will let the state loop take care of it.
	if curr.Hash != s.l2Head.L1Origin.Hash {
		return eth.L1BlockRef{}, errors.New("Unknown L1Origin")
	}

	// TODO: There is an interaction with not using the L1 Genesis as an L1 Origin and
	// the fact that the L2 Genesis time needs to be set around the L1 Genesis such
	// that this check will return true.
	if s.l2Head.Time+s.Config.BlockTime >= curr.Time {
		// TODO: Need to walk more?
		ref, err := s.l1.L1BlockRefByNumber(ctx, curr.Number+1)
		s.log.Debug("Advancing L1 Origin", "l2Head", s.l2Head, "previous_l1Origin", s.l2Head.L1Origin, "l1Origin", ref, "err", err)
		return ref, err
	}
	s.log.Debug("Next L1 Origin is the same as the previous", "l2Head", s.l2Head, "l1Origin", curr)
	return curr, nil
}

// createNewL2Block builds a L2 block on top of the L2 Head (unsafe)
func (s *state) createNewL2Block(ctx context.Context) (eth.L1BlockRef, error) {
	nextOrigin, err := s.findNextL1Origin(context.Background())
	if err != nil {
		s.log.Error("Error finding next L1 Origin", "err", err)
		return eth.L1BlockRef{}, err
	}
	if nextOrigin.Time <= s.Config.BlockTime+s.l2Head.Time {
		s.log.Trace("Skipping block production because the next block time is behind the next L1 Origin", "l2Head", s.l2Head, "l1Origin", nextOrigin)
		return eth.L1BlockRef{}, nil
	}
	// Don't produce blocks until past the L1 genesis
	if nextOrigin.Number <= s.Config.Genesis.L1.Number {
		s.log.Trace("Skipping block production because the next L1 Origin is behind the L1 genesis")
		return eth.L1BlockRef{}, nil
	}
	// Actually create the new block
	newUnsafeL2Head, batch, err := s.output.createNewBlock(context.Background(), s.l2Head, s.l2SafeHead.ID(), s.l2Finalized, nextOrigin.ID())
	if err != nil {
		s.log.Error("Could not extend chain as sequencer", "err", err, "l2UnsafeHead", s.l2Head, "l1Origin", nextOrigin)
		return eth.L1BlockRef{}, err
	}
	// State update
	s.l2Head = newUnsafeL2Head
	s.log.Info("Sequenced new l2 block", "l2Head", s.l2Head, "l1Origin", s.l2Head.L1Origin)
	//Submit batch
	go func() {
		_, err := s.bss.Submit(&s.Config, []*derive.BatchData{batch}) // TODO: submit multiple batches
		if err != nil {
			s.log.Error("Error submitting batch", "err", err)
		}
	}()
	return nextOrigin, nil
}

// handleEpoch attempts to insert a full L2 epoch on top of the L2 Safe Head.
// It ensures that a full sequencing window is available and updates the state as needed.
func (s *state) handleEpoch(ctx context.Context) (bool, error) {
	s.log.Trace("Handling epoch", "l2Head", s.l2Head, "l2SafeHead", s.l2SafeHead)
	// Extend cached window if we do not have enough saved blocks
	if len(s.l1WindowBuf) < int(s.Config.SeqWindowSize) {
		// attempt to buffer up to 2x the size of a sequence window of L1 blocks, to speed up later handleEpoch calls
		nexts, err := s.l1.L1Range(ctx, s.l1WindowBufEnd(), 2*s.Config.SeqWindowSize)
		if err != nil {
			s.log.Error("Could not extend the cached L1 window", "err", err, "l2Head", s.l2Head, "l2SafeHead", s.l2SafeHead, "l1Head", s.l1Head, "window_end", s.l1WindowBufEnd())
			return false, err
		}
		s.l1WindowBuf = append(s.l1WindowBuf, nexts...)

	}
	// Ensure that there are enough blocks in the cached window
	if len(s.l1WindowBuf) < int(s.Config.SeqWindowSize) {
		s.log.Debug("Not enough cached blocks to run step", "cached_window_len", len(s.l1WindowBuf))
		return false, nil
	}

	// Insert the epoch
	window := s.l1WindowBuf[:int(s.Config.SeqWindowSize)]
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	newL2Head, newL2SafeHead, reorg, err := s.output.insertEpoch(ctx, s.l2Head, s.l2SafeHead, s.l2Finalized, window)
	cancel()
	if err != nil {
		s.log.Error("Error in running the output step.", "err", err, "l2Head", s.l2Head, "l2SafeHead", s.l2SafeHead)
		return false, err
	}

	// State update
	s.l2Head = newL2Head
	s.l2SafeHead = newL2SafeHead
	s.l1WindowBuf = s.l1WindowBuf[1:]
	s.log.Info("Inserted a new epoch", "l2Head", s.l2Head, "l2SafeHead", s.l2SafeHead, "reorg", reorg)
	// TODO: l2Finalized
	return reorg, nil

}

// loop is the event loop that responds to L1 changes and internal timers to produce L2 blocks.
func (s *state) loop() {
	s.log.Info("State loop started")
	ctx := context.Background()
	var l2BlockCreation <-chan time.Time
	if s.sequencer {
		l2BlockCreationTicker := time.NewTicker(time.Duration(s.Config.BlockTime) * time.Second)
		defer l2BlockCreationTicker.Stop()
		l2BlockCreation = l2BlockCreationTicker.C
	}

	stepRequest := make(chan struct{}, 1)
	l2BlockCreationReq := make(chan struct{}, 1)

	createBlock := func() {
		select {
		case l2BlockCreationReq <- struct{}{}:
		// Don't deadlock if the channel is already full
		default:
		}
	}

	requestStep := func() {
		select {
		case stepRequest <- struct{}{}:
		// Don't deadlock if the channel is already full
		default:
		}
	}

	requestStep()

	for {
		select {
		case <-s.done:
			atomic.AddUint32(&s.closed, 1)
			return
		case <-l2BlockCreation:
			s.log.Trace("L2 Creation Ticker")
			createBlock()
		case <-l2BlockCreationReq:
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			nextOrigin, err := s.createNewL2Block(ctx)
			cancel()
			if err != nil {
				s.log.Error("Error creating new L2 block", "err", err)
			}
			if nextOrigin.Time > s.l2Head.Time+s.Config.BlockTime {
				s.log.Trace("Asking for a second L2 block asap", "l2Head", s.l2Head)
				createBlock()
			}

		case newL1Head := <-s.l1Heads:
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := s.handleNewL1Block(ctx, newL1Head)
			cancel()
			if err != nil {
				s.log.Error("Error in handling new L1 Head", "err", err)
			}
			// Run step if we are able to
			if s.l1Head.Number-s.l2SafeHead.L1Origin.Number >= s.Config.SeqWindowSize {
				s.log.Trace("Requesting next step", "l1Head", s.l1Head, "l2Head", s.l2Head, "l1Origin", s.l2Head.L1Origin)
				requestStep()
			}
		case <-stepRequest:
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			reorg, err := s.handleEpoch(ctx)
			cancel()
			if err != nil {
				s.log.Error("Error in handling epoch", "err", err)
			}
			if reorg {
				s.log.Warn("Got reorg")
				if s.sequencer {
					createBlock()
				}
			}

			// Immediately run next step if we have enough blocks.
			if s.l1Head.Number-s.l2Head.L1Origin.Number >= s.Config.SeqWindowSize {
				s.log.Trace("Requesting next step", "l1Head", s.l1Head, "l2Head", s.l2Head, "l1Origin", s.l2Head.L1Origin)
				requestStep()
			}
		}
	}

}
