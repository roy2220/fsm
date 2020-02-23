package buddy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBABAllocateAndFreeBlock(t *testing.T) {
	bab := blockAllocationBitmap{}
	bab.Expand()
	b := int64(0)

	for bss := minBlockSizeShift; bss < maxBlockSizeShift; bss++ {
		bab.AllocateBlock(b, bss)
		bss2, ok := bab.GetBlockSize(b)

		if assert.True(t, ok) {
			assert.Equal(t, bss, bss2)
		}

		t.Log("b=", b)
		b = (2 << bss)
	}

	bab.Expand()
	b = 0

	for bss := minBlockSizeShift; bss < maxBlockSizeShift; bss++ {
		bab.AllocateBlock(MaxBlockSize+b, bss)
		bss2, ok := bab.GetBlockSize(b)

		if assert.True(t, ok) {
			assert.Equal(t, bss, bss2)
		}

		b = (2 << bss)
	}

	b = 0

	for bss := minBlockSizeShift; bss < maxBlockSizeShift; bss++ {
		bss2, ok := bab.FreeBlock(MaxBlockSize + b)

		if assert.True(t, ok) {
			assert.Equal(t, bss, bss2)
		}

		_, ok = bab.FreeBlock(MaxBlockSize + b)
		assert.False(t, ok)
		b = (2 << bss)
	}

	bab.Shrink()
	b = 0
	bss := minBlockSizeShift

	for b2 := int64(0); b2 < MaxBlockSize; b2 += MinBlockSize {
		bss2, ok := bab.FreeBlock(b2)

		if ok {
			t.Log("b2=", b2)
		}

		if b2 == b {
			if assert.True(t, ok) {
				assert.Equal(t, bss, bss2)
			}

			_, ok = bab.FreeBlock(b)
			assert.False(t, ok, "%v", b2)
			b = (2 << bss)
			bss++
		} else {
			assert.False(t, ok, "%v %v", b2, b)
		}
	}

	bab.Shrink()
}

func TestBABGetFreeBlocks(t *testing.T) {
	bab := blockAllocationBitmap{}
	bab.Expand()
	n := 0

	bab.GetFreeBlocks(func(block int64, blockSizeShift int) {
		assert.Equal(t, int64(0), block)
		assert.Equal(t, maxBlockSizeShift, blockSizeShift)
		n++
	})

	assert.Equal(t, 1, n)
	bab.AllocateBlock(0, minBlockSizeShift)
	bss := minBlockSizeShift
	n = 0

	bab.GetFreeBlocks(func(block int64, blockSizeShift int) {
		b := int64(1 << bss)
		assert.Equal(t, b, block)
		assert.Equal(t, bss, blockSizeShift)
		bss++
		n++
	})

	assert.Equal(t, maxBlockSizeShift-minBlockSizeShift, n)
	b := int64(2 * MinBlockSize)

	for bss := minBlockSizeShift + 1; bss < maxBlockSizeShift; bss++ {
		bab.AllocateBlock(b, bss)
		b = (2 << bss)
	}

	n = 0

	bab.GetFreeBlocks(func(block int64, blockSizeShift int) {
		assert.Equal(t, int64(MinBlockSize), block)
		assert.Equal(t, minBlockSizeShift, blockSizeShift)
		n++
	})

	assert.Equal(t, 1, n)
	b = 0

	for bss := minBlockSizeShift; bss < maxBlockSizeShift; bss++ {
		bab.FreeBlock(b)
		b = (2 << bss)
	}

	n = 0

	bab.GetFreeBlocks(func(block int64, blockSizeShift int) {
		assert.Equal(t, int64(0), block)
		assert.Equal(t, maxBlockSizeShift, blockSizeShift)
		n++
	})

	assert.Equal(t, 1, n)
}
