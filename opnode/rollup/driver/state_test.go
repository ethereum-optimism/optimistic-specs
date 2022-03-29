package driver

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/internal/testlog"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

type testID string

func (id testID) ID() eth.BlockID {
	parts := strings.Split(string(id), ":")
	if len(parts) != 2 {
		panic("bad id")
	}
	if len(parts[0]) > 32 {
		panic("test ID hash too long")
	}
	var h common.Hash
	copy(h[:], parts[0])
	v, err := strconv.ParseUint(parts[1], 0, 64)
	if err != nil {
		panic(err)
	}
	return eth.BlockID{
		Hash:   h,
		Number: v,
	}
}

type outputHandlerFn func(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.L2BlockRef, l2Finalized eth.BlockID, l1Input []eth.BlockID) (eth.L2BlockRef, eth.L2BlockRef, bool, error)

func (fn outputHandlerFn) insertEpoch(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.L2BlockRef, l2Finalized eth.BlockID, l1Input []eth.BlockID) (eth.L2BlockRef, eth.L2BlockRef, bool, error) {
	return fn(ctx, l2Head, l2SafeHead, l2Finalized, l1Input)
}

func (fn outputHandlerFn) createNewBlock(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.BlockID, l2Finalized eth.BlockID, l1Origin eth.BlockID) (eth.L2BlockRef, *derive.BatchData, error) {
	panic("Unimplemented")
}

type outputArgs struct {
	l2Head      eth.BlockID
	l2Finalized eth.BlockID
	l1Window    []eth.BlockID
}

type outputReturnArgs struct {
	l2Head eth.L2BlockRef
	err    error
}

type stateTestCaseStep struct {
	// Expect l1head, l2head, and sequence window
	l1head testID
	l2head testID
	window []testID

	// l1act and l2act are ran at each step
	l1act func(t *testing.T, s *state, src *fakeChainSource, l1Heads chan eth.L1BlockRef)
	l2act func(t *testing.T, expectedWindow []testID, s *state, src *fakeChainSource, outputIn chan outputArgs, outputReturn chan outputReturnArgs)
	reorg bool
}

func advanceL1(t *testing.T, s *state, src *fakeChainSource, l1Heads chan eth.L1BlockRef) {
	l1Heads <- src.advanceL1()
}

func stutterL1(t *testing.T, s *state, src *fakeChainSource, l1Heads chan eth.L1BlockRef) {
	l1Heads <- src.l1Head()
}

func stutterAdvance(t *testing.T, s *state, src *fakeChainSource, l1Heads chan eth.L1BlockRef) {
	l1Heads <- src.l1Head()
	l1Heads <- src.l1Head()
	l1Heads <- src.l1Head()
	l1Heads <- src.advanceL1()
	l1Heads <- src.l1Head()
	l1Heads <- src.l1Head()
	l1Heads <- src.l1Head()
}

func stutterL2(t *testing.T, expectedWindow []testID, s *state, src *fakeChainSource, outputIn chan outputArgs, outputReturn chan outputReturnArgs) {
	select {
	case <-outputIn:
		t.Error("Got a step when no step should have occurred (l1 only advance)")
	default:
	}
}

func advanceL2(t *testing.T, expectedWindow []testID, s *state, src *fakeChainSource, outputIn chan outputArgs, outputReturn chan outputReturnArgs) {
	args := <-outputIn
	assert.Equal(t, int(s.Config.SeqWindowSize), len(args.l1Window), "Invalid L1 window size")
	assert.Equal(t, len(expectedWindow), len(args.l1Window), "L1 Window size does not match expectedWindow")
	for i := range expectedWindow {
		assert.Equal(t, expectedWindow[i].ID(), args.l1Window[i], "Window elements must match")
	}
	outputReturn <- outputReturnArgs{l2Head: src.setL2Head(int(args.l2Head.Number) + 1), err: nil}
}

func reorg__L2(t *testing.T, expectedWindow []testID, s *state, src *fakeChainSource, outputIn chan outputArgs, outputReturn chan outputReturnArgs) {
	args := <-outputIn
	assert.Equal(t, int(s.Config.SeqWindowSize), len(args.l1Window), "Invalid L1 window size")
	assert.Equal(t, len(expectedWindow), len(args.l1Window), "L1 Window size does not match expectedWindow")
	for i := range expectedWindow {
		assert.Equal(t, expectedWindow[i].ID(), args.l1Window[i], "Window elements must match")
	}
	src.reorgL2()

	outputReturn <- outputReturnArgs{l2Head: src.setL2Head(int(args.l2Head.Number) + 1), err: nil}
}

type stateTestCase struct {
	name      string
	l1Chains  []string
	l2Chains  []string
	steps     []stateTestCaseStep
	seqWindow int
	genesis   rollup.Genesis
}

func (tc *stateTestCase) Run(t *testing.T) {
	log := testlog.Logger(t, log.LvlTrace)
	chainSource := NewFakeChainSource(tc.l1Chains, tc.l2Chains, log)
	l1headsCh := make(chan eth.L1BlockRef, 10)
	// Unbuffered channels to force a sync point between the test and the state loop.
	outputIn := make(chan outputArgs)
	outputReturn := make(chan outputReturnArgs)
	outputHandler := func(ctx context.Context, l2Head eth.L2BlockRef, l2SafeHead eth.L2BlockRef, l2Finalized eth.BlockID, l1Input []eth.BlockID) (eth.L2BlockRef, eth.L2BlockRef, bool, error) {
		// TODO: Not sequencer, but need to pass unsafeL2Head here for the test.
		outputIn <- outputArgs{l2Head: l2Head.ID(), l2Finalized: l2Finalized, l1Window: l1Input}
		r := <-outputReturn
		return r.l2Head, r.l2Head, false, r.err
	}
	config := rollup.Config{SeqWindowSize: uint64(tc.seqWindow), Genesis: tc.genesis, BlockTime: 2}
	state := NewState(log, config, chainSource, chainSource, outputHandlerFn(outputHandler), nil, false)
	defer func() {
		assert.NoError(t, state.Close(), "Error closing state")
	}()

	err := state.Start(context.Background(), l1headsCh)
	assert.NoError(t, err, "Error starting the state object")

	for _, step := range tc.steps {
		if step.reorg {
			chainSource.reorgL1()
		}
		step.l1act(t, state, chainSource, l1headsCh)
		<-time.After(5 * time.Millisecond)
		step.l2act(t, step.window, state, chainSource, outputIn, outputReturn)
		<-time.After(5 * time.Millisecond)

		assert.Equal(t, step.l1head.ID(), state.l1Head.ID(), "l1 head")
		assert.Equal(t, step.l2head.ID(), state.l2SafeHead.ID(), "l2 head")
	}
}

func TestDriver(t *testing.T) {
	cases := []stateTestCase{
		{
			name:      "Simple extensions",
			l1Chains:  []string{"abcdefgh"},
			l2Chains:  []string{"ABCDEF"},
			seqWindow: 2,
			genesis:   fakeGenesis('a', 'A', 0),
			steps: []stateTestCaseStep{
				{l1act: stutterL1, l2act: stutterL2, l1head: "a:0", l2head: "A:0"},
				{l1act: advanceL1, l2act: stutterL2, l1head: "b:1", l2head: "A:0", window: []testID{"a:0", "b:1"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "c:2", l2head: "B:1", window: []testID{"b:1", "c:2"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "d:3", l2head: "C:2", window: []testID{"c:2", "d:3"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "e:4", l2head: "D:3", window: []testID{"d:3", "e:4"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "f:5", l2head: "E:4", window: []testID{"e:4", "f:5"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "g:6", l2head: "F:5", window: []testID{"f:5", "g:6"}},
			},
		},
		{
			name:      "Reorg",
			l1Chains:  []string{"abcdefg", "abcxyzw"},
			l2Chains:  []string{"ABCDEF", "ABCXYZ"},
			seqWindow: 2,
			genesis:   fakeGenesis('a', 'A', 0),
			steps: []stateTestCaseStep{
				{l1act: stutterL1, l2act: stutterL2, l1head: "a:0", l2head: "A:0"},
				{l1act: advanceL1, l2act: stutterL2, l1head: "b:1", l2head: "A:0", window: []testID{"a:0", "b:1"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "c:2", l2head: "B:1", window: []testID{"b:1", "c:2"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "d:3", l2head: "C:2", window: []testID{"c:2", "d:3"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "e:4", l2head: "D:3", window: []testID{"d:3", "e:4"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "f:5", l2head: "E:4", window: []testID{"e:4", "f:5"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "g:6", l2head: "F:5", window: []testID{"f:5", "g:6"}},
				{l1act: stutterL1, l2act: reorg__L2, l1head: "w:6", l2head: "X:3", window: []testID{"x:3", "y:4"}, reorg: true},
				{l1act: stutterL1, l2act: advanceL2, l1head: "w:6", l2head: "Y:4", window: []testID{"y:4", "z:5"}},
				{l1act: stutterL1, l2act: advanceL2, l1head: "w:6", l2head: "Z:5", window: []testID{"z:5", "w:6"}},
				{l1act: stutterL1, l2act: stutterL2, l1head: "w:6", l2head: "Z:5", window: []testID{"z:5", "w:6"}},
				{l1act: stutterL1, l2act: stutterL2, l1head: "w:6", l2head: "Z:5", window: []testID{"z:5", "w:6"}},
			},
		},
		{
			name:      "Simple extensions with multi-step stutter",
			l1Chains:  []string{"abcdefgh"},
			l2Chains:  []string{"ABCDEF"},
			seqWindow: 2,
			genesis:   fakeGenesis('a', 'A', 0),
			steps: []stateTestCaseStep{
				{l1act: stutterL1, l2act: stutterL2, l1head: "a:0", l2head: "A:0"},
				{l1act: advanceL1, l2act: stutterL2, l1head: "b:1", l2head: "A:0", window: []testID{"a:0", "b:1"}},
				{l1act: stutterAdvance, l2act: advanceL2, l1head: "c:2", l2head: "B:1", window: []testID{"b:1", "c:2"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "d:3", l2head: "C:2", window: []testID{"c:2", "d:3"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "e:4", l2head: "D:3", window: []testID{"d:3", "e:4"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "f:5", l2head: "E:4", window: []testID{"e:4", "f:5"}},
				{l1act: advanceL1, l2act: advanceL2, l1head: "g:6", l2head: "F:5", window: []testID{"f:5", "g:6"}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.Run)
	}

}
