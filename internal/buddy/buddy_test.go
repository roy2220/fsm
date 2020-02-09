package buddy_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/roy2220/fsm/internal/buddy"
	"github.com/stretchr/testify/assert"
)

type SpaceMapper struct {
	T *testing.T
}

func (sm SpaceMapper) MapSpace(spaceSize int) error {
	sm.T.Logf("mapping space size: %d", spaceSize)
	return nil
}

func (SpaceMapper) AccessSpace() []byte { return nil }

type BlockInfo struct {
	Ptr  int64
	Size int
}

func TestBuddyInsertBlock(t *testing.T) {
	b, bis := MakeBuddy(t)

	sort.Slice(bis, func(i, j int) bool {
		return bis[i].Ptr < bis[j].Ptr
	})

	lastBlockEnd := int64(0)

	for _, bi := range bis {
		assert.GreaterOrEqual(t, bi.Ptr, lastBlockEnd)
		bs, err := b.GetBlockSize(bi.Ptr)

		if assert.NoError(t, err) {
			assert.Equal(t, bi.Size, bs)
		}

		lastBlockEnd = bi.Ptr + int64(bi.Size)
	}
}

func TestBuddyFreeBlock(t *testing.T) {
	b, bis := MakeBuddy(t)

	rand.Shuffle(len(bis), func(i, j int) {
		bis[i], bis[j] = bis[j], bis[i]
	})

	for _, bi := range bis {
		err := b.FreeBlock(bi.Ptr)
		assert.NoError(t, err)
		err = b.FreeBlock(bi.Ptr)
		assert.Error(t, err)
		_, err = b.GetBlockSize(bi.Ptr)
		assert.Error(t, err)
	}

	bptr, _, err := b.AllocateBlock(0)

	if assert.NoError(t, err) {
		assert.Equal(t, int64(0), bptr)
	}
}

func TestBuddyShrinkSpace(t *testing.T) {
	b, bis := MakeBuddy(t)

	for _, bi := range bis {
		err := b.FreeBlock(bi.Ptr)
		assert.NoError(t, err)
	}

	b.ShrinkSpace()
	assert.Equal(t, 0, b.SpaceSize())
}

func TestBuddyUsedAndAllocatedSpaceSize(t *testing.T) {
	b, bis := MakeBuddy(t)

	sort.Slice(bis, func(i, j int) bool {
		return bis[i].Ptr < bis[j].Ptr
	})

	i := len(bis) / 2

	for _, bi := range bis[:i] {
		err := b.FreeBlock(bi.Ptr)
		assert.NoError(t, err)
	}

	for _, bi := range bis[i+1:] {
		err := b.FreeBlock(bi.Ptr)
		assert.NoError(t, err)
	}

	blockEnd := bis[i].Ptr + int64(bis[i].Size)
	assert.Equal(t, int(blockEnd), b.UsedSpaceSize())
	assert.Equal(t, bis[i].Size, b.AllocatedSpaceSize())
	err := b.FreeBlock(bis[i].Ptr)
	assert.NoError(t, err)
	assert.Equal(t, 0, b.UsedSpaceSize())
	assert.Equal(t, 0, b.AllocatedSpaceSize())
}

func MakeBuddy(t *testing.T) (*buddy.Buddy, []*BlockInfo) {
	b := new(buddy.Buddy).Init(SpaceMapper{t})
	bis := make([]*BlockInfo, 10000)

	for i := range bis {
		bs := buddy.MinBlockSize + rand.Intn(buddy.MaxBlockSize-buddy.MinBlockSize+1)
		f := rand.Float64()
		f *= f
		f *= f
		f *= f
		f *= f
		bs = int(float64(bs) * f)
		bptr, bs2, err := b.AllocateBlock(bs)

		if !assert.NoError(t, err) {
			t.FailNow()
		}

		if !assert.GreaterOrEqual(t, bs2, bs) {
			t.FailNow()
		}

		bis[i] = &BlockInfo{bptr, bs2}
	}

	t.Logf("space size: %d", b.SpaceSize())
	t.Logf("used space size: %d", b.UsedSpaceSize())
	t.Logf("allocated space size: %d", b.AllocatedSpaceSize())
	t.Logf("allocated/used: %f", float64(b.AllocatedSpaceSize())/float64(b.UsedSpaceSize()))
	return b, bis
}
