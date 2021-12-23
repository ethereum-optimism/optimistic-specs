package l1

import (
	"crypto/rand"
	"testing"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestTracker_HeadSignal(t *testing.T) {
	tr := NewTracker()
	assert.Equal(t, eth.BlockID{}, tr.Head(), "expecting")
	aP := eth.BlockID{Hash: common.Hash{0xa0}, Number: 122}
	a := eth.BlockID{Hash: common.Hash{0xaa}, Number: 123}
	tr.HeadSignal(aP, a)
	assert.Equal(t, a, tr.Head(), "expecting a")
	assert.Equal(t, aP, tr.(*tracker).parents[a], "expecting parent of a")
	// note: lower number, head changes can decrease height
	bP := eth.BlockID{Hash: common.Hash{0xb0}, Number: 99}
	b := eth.BlockID{Hash: common.Hash{0xbb}, Number: 100}
	tr.HeadSignal(bP, b)
	assert.Equal(t, b, tr.Head(), "expecting b")
	assert.Equal(t, bP, tr.(*tracker).parents[b], "expecting parent of b")
}

func TestTracker_WatchHeads(t *testing.T) {
	tr := NewTracker()
	g := eth.BlockID{Hash: common.Hash{0x11}, Number: 0}
	a0 := eth.BlockID{Hash: common.Hash{0xa0}, Number: 1}
	a := eth.BlockID{Hash: common.Hash{0xaa}, Number: 2}
	b := eth.BlockID{Hash: common.Hash{0xbb}, Number: 1}
	c := eth.BlockID{Hash: common.Hash{0xcc}, Number: 3}
	d := eth.BlockID{Hash: common.Hash{0xdd}, Number: 4}
	// 2 ids: parent and self
	edges := [][2]eth.BlockID{
		{a0, a}, // head without history to genesis
		{g, a0}, // rewind head
		{g, b},
		{a, c},
		{g, b}, // back and forth head changes can happen in PoS L1
		{g, b}, // re-announcing too
		{c, d},
	}
	recorder := make(chan eth.BlockID, len(edges))
	sub := tr.WatchHeads(recorder)
	for _, edge := range edges {
		tr.HeadSignal(edge[0], edge[1])
	}
	close(recorder)
	i := 0
	for id := range recorder {
		assert.Equal(t, edges[i][1], id)
		i += 1
	}
	// unsubscribe: can still change heads, without recording the changes
	sub.Unsubscribe()
	// remaining sends would panic, if they were sent to the closed recorder
	tr.HeadSignal(g, eth.BlockID{Hash: common.Hash{0xff}, Number: 1})

	// Open two new watchers, and check we receive heads in both
	recA := make(chan eth.BlockID, 1)
	recB := make(chan eth.BlockID, 1)
	tr.WatchHeads(recA)
	tr.WatchHeads(recB)
	exp := eth.BlockID{Hash: common.Hash{0x42}, Number: 1}
	tr.HeadSignal(g, exp)
	ra, rb := <-recA, <-recB
	assert.Equal(t, exp, ra)
	assert.Equal(t, exp, rb)
}

func rndID(n uint64) (out eth.BlockID) {
	_, _ = rand.Read(out.Hash[:])
	out.Number = n
	return
}

type pullCase struct {
	name       string
	start      eth.BlockID
	headParent eth.BlockID
	head       eth.BlockID
	pulled     eth.BlockID
	mode       ChainMode
}

func TestTracker_Pull(t *testing.T) {
	tr := NewTracker()
	// graph:
	//  a - b - d - x
	//    \ c - e - y
	//      z - zz   (not connected to main)
	a, b, c, d, e, x, y := rndID(0), rndID(1), rndID(1), rndID(2), rndID(2), rndID(3), rndID(3)
	tr.Parent(a, b)
	tr.Parent(b, d)
	tr.Parent(a, c)
	tr.Parent(c, e)
	tr.Parent(d, x)
	tr.Parent(e, y)
	z, zz := rndID(1), rndID(2)
	tr.Parent(z, zz)

	t.Run("no head", func(t *testing.T) {
		last := rndID(123)
		r, m := tr.Pull(last)
		assert.Equal(t, last, r, "nothing to pull without head, stay where you are")
		assert.Equal(t, ChainNoop, m, "nothing to pull without head")
	})

	cases := []pullCase{
		{"Already synced", a, eth.BlockID{}, a, a, ChainNoop},
		{"Already synced far", x, d, x, x, ChainNoop},
		{"Extend the chain from a to c", a, a, c, c, ChainExtend},
		{"Extend the chain from a to d, first apply b", a, b, d, b, ChainExtend},
		{"Orphan block b in favor of c", b, a, c, c, ChainReorg},
		{"Orphan block c in favor of d, first b", c, b, d, b, ChainReorg},
		{"Go to head, not absolute tip", c, a, c, c, ChainNoop},
		{"Deep reorg", x, e, y, c, ChainReorg},
		{"Deep uneven behind reorg", d, e, y, c, ChainReorg},
		{"Deep uneven ahead reorg", x, c, e, c, ChainReorg},
		{"Walk back, one bad block", x, b, d, d, ChainUndo},
		{"Walk back, two bad blocks", x, a, b, b, ChainUndo},
		{"Disconnected chain", b, z, zz, z, ChainMissing},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tr.HeadSignal(c.headParent, c.head)
			r, m := tr.Pull(c.start)
			assert.Equal(t, c.pulled, r)
			assert.Equal(t, c.mode, m)
		})
	}
}

func TestTracker_Prune(t *testing.T) {
	tr := NewTracker()
	a, b, c, d := rndID(0), rndID(1), rndID(2), rndID(3)
	tr.Parent(a, b)
	tr.Parent(b, c)
	tr.Parent(c, d)
	z, zz := rndID(1), rndID(2)
	tr.Parent(z, zz)

	// everything with height < 2 gets removed (those with height 1 will still be known as parents of height 2)
	tr.Prune(2)
	tr.HeadSignal(c, d)

	t.Run("from pruned info", func(t *testing.T) {
		p := tr.(*tracker).parents
		t.Logf("%v", p)
		r, m := tr.Pull(a)
		assert.Equal(t, b, r, "get informed of first canonical unpruned node towards head")
		assert.Equal(t, ChainMissing, m, "missing chain due to pruning")
	})

	t.Run("pruned but known parent", func(t *testing.T) {
		r, m := tr.Pull(b)
		assert.Equal(t, c, r, "found extending block id on border")
		assert.Equal(t, ChainExtend, m, "found extension in unpruned part")
	})

	t.Run("pruned, known as parent, but not canonical", func(t *testing.T) {
		r, m := tr.Pull(z)
		assert.Equal(t, b, r, "get informed of first canonical unpruned node towards head")
		assert.Equal(t, ChainMissing, m, "missing chain due to pruning")
	})
}
