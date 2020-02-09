package pool_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/roy2220/fsm/internal/buddy"
	"github.com/roy2220/fsm/internal/pool"
	"github.com/stretchr/testify/assert"
)

type SpaceMapper struct {
	buffer []byte
}

type SpaceInfo struct {
	Ptr  int64
	Size int
}

func (sm *SpaceMapper) MapSpace(spaceSize int) error {
	buffer := make([]byte, spaceSize)
	copy(buffer, sm.buffer)
	sm.buffer = buffer
	return nil
}

func (sm *SpaceMapper) AccessSpace() []byte {
	return sm.buffer
}

func TestPoolAllocateSpace(t *testing.T) {
	p, _, sis := MakePool(t)

	sort.Slice(sis, func(i, j int) bool {
		return sis[i].Ptr < sis[j].Ptr
	})

	lastSpaceEnd := int64(0)

	for _, si := range sis {
		assert.GreaterOrEqual(t, si.Ptr, lastSpaceEnd)
		ss := p.GetSpaceSize(si.Ptr)
		assert.Equal(t, si.Size, ss)
		lastSpaceEnd = si.Ptr + int64(si.Size)
	}
}

func TestPoolFreeSpace(t *testing.T) {
	p, b, sis := MakePool(t)
	t.Log(b.AllocatedSpaceSize())

	rand.Shuffle(len(sis), func(i, j int) {
		sis[i], sis[j] = sis[j], sis[i]
	})

	n := len(sis) / 2

	for _, si := range sis[:n] {
		p.FreeSpace(si.Ptr)
	}

	for i, si := range sis[:n] {
		sptr, ss := p.AllocateSpace(si.Size)
		sis[i] = &SpaceInfo{sptr, ss}
	}

	for _, si := range sis {
		p.FreeSpace(si.Ptr)
	}

	for _, si := range sis {
		assert.Panics(t, func() {
			p.FreeSpace(si.Ptr)
		})
	}

	_, ok := p.GetPooledBlocks()()
	assert.False(t, ok)
	b.ShrinkSpace()
	assert.Equal(t, 0, b.SpaceSize())
}

func MakePool(t *testing.T) (*pool.Pool, *buddy.Buddy, []*SpaceInfo) {
	spaceMapper := SpaceMapper{}
	buddy := new(buddy.Buddy).Init(&spaceMapper)
	pool1 := new(pool.Pool).Init(buddy)
	sis := make([]*SpaceInfo, 100000)
	tss2 := 0

	for i := range sis {
		ss := rand.Intn(pool.BlockSize)
		f := rand.Float64()
		f *= f
		ss = int(float64(ss) * f)
		sptr, ss2 := pool1.AllocateSpace(ss)

		if !assert.GreaterOrEqual(t, ss2, ss) {
			t.FailNow()
		}

		sis[i] = &SpaceInfo{sptr, ss2}
		tss2 += ss2
	}

	t.Logf("allocated space size: %d, total space size: %d", buddy.AllocatedSpaceSize(), tss2)
	t.Logf("total space size / allocated space size: %f", float64(tss2)/float64(buddy.AllocatedSpaceSize()))
	return pool1, buddy, sis
}
