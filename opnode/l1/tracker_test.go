package l1

import (
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestTracker_HeadSignal(t *testing.T) {
	tr := NewTracker()
	assert.Equal(t, BlockID{}, tr.Head(), "expecting")
	a := BlockID{Hash: common.Hash{0xaa}, Number: 123}
	tr.HeadSignal(a)
	assert.Equal(t, a, tr.Head(), "expecting a")
	// note: lower number, head changes can decrease height
	b := BlockID{Hash: common.Hash{0xbb}, Number: 100}
	tr.HeadSignal(b)
	assert.Equal(t, b, tr.Head(), "expecting b")
}

func TestTracker_WatchHeads(t *testing.T) {
	tr := NewTracker()
	ids := []BlockID{
		{Hash: common.Hash{0xaa}, Number: 123},
		{Hash: common.Hash{0xbb}, Number: 100},
		{Hash: common.Hash{0xcc}, Number: 140},
		{Hash: common.Hash{0xbb}, Number: 100}, // back and forth head changes can happen in PoS L1
		{Hash: common.Hash{0xbb}, Number: 100}, // re-announcing too
		{Hash: common.Hash{0xdd}, Number: 150},
	}
	recorder := make(chan BlockID, len(ids))
	sub := tr.WatchHeads(recorder)
	for _, id := range ids {
		tr.HeadSignal(id)
	}
	for i, id := range ids {
		assert.Equal(t, ids[i], id)
	}
	// unsubscribe: can still change heads, without recording the changes
	sub.Unsubscribe()
	// remaining sends would panic, if they were sent
	close(recorder)
	tr.HeadSignal(BlockID{Hash: common.Hash{0xff}, Number: 9000})

	// Open two new watchers, and check we receive heads in both
	recA := make(chan BlockID, 1)
	recB := make(chan BlockID, 1)
	tr.WatchHeads(recA)
	tr.WatchHeads(recB)
	exp := BlockID{Hash: common.Hash{0x42}, Number: 1337}
	tr.HeadSignal(exp)
	a, b := <-recA, <-recB
	assert.Equal(t, exp, a)
	assert.Equal(t, exp, b)
}

func rndID(n uint64) (out BlockID) {
	_, _ = rand.Read(out.Hash[:])
	out.Number = n
	return
}

type pullCase struct {
	name   string
	start  BlockID
	head   BlockID
	pulled BlockID
	mode   ChainMode
}

func TestTracker_Pull(t *testing.T) {
	tr := NewTracker()
	// graph:
	//  a - b - d - x
	//    \ c - e - y
	//      z - zz   (not connected to main)
	a, b, c, d, e, x, y := rndID(0), rndID(1), rndID(1), rndID(2), rndID(2), rndID(3), rndID(3)
	tr.Parent(b, a)
	tr.Parent(d, b)
	tr.Parent(c, a)
	tr.Parent(e, c)
	tr.Parent(x, d)
	tr.Parent(y, e)
	z, zz := rndID(1), rndID(2)
	tr.Parent(zz, z)

	t.Run("no head", func(t *testing.T) {
		last := rndID(123)
		r, m := tr.Pull(last)
		assert.Equal(t, last, r, "nothing to pull without head, stay where you are")
		assert.Equal(t, ChainNoop, m, "nothing to pull without head")
	})

	cases := []pullCase{
		{"Already synced", a, a, a, ChainNoop},
		{"Already synced far", x, x, x, ChainNoop},
		{"Extend the chain from a to c", a, c, c, ChainExtend},
		{"Extend the chain from a to d, first apply b", a, d, b, ChainExtend},
		{"Orphan block b in favor of c", b, c, c, ChainReorg},
		{"Orphan block c in favor of d, first b", c, d, b, ChainReorg},
		{"Go to head, not absolute tip", c, c, c, ChainNoop},
		{"Deep reorg", x, y, c, ChainReorg},
		{"Deep uneven behind reorg", d, y, c, ChainReorg},
		{"Deep uneven ahead reorg", x, e, c, ChainReorg},
		{"Walk back, one bad block", x, d, d, ChainUndo},
		{"Walk back, two bad blocks", x, b, b, ChainUndo},
		{"Disconnected chain", b, zz, z, ChainMissing},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tr.HeadSignal(c.head)
			r, m := tr.Pull(c.start)
			assert.Equal(t, c.pulled, r)
			assert.Equal(t, c.mode, m)
		})
	}
}

func TestTracker_Prune(t *testing.T) {
	tr := NewTracker()
	a, b, c, d := rndID(0), rndID(1), rndID(2), rndID(3)
	tr.Parent(b, a)
	tr.Parent(c, b)
	tr.Parent(d, c)
	z, zz := rndID(1), rndID(2)
	tr.Parent(zz, z)

	// everything with height < 2 gets removed (those with height 1 will still be known as parents of height 2)
	tr.Prune(2)
	tr.HeadSignal(d)

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
