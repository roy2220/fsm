package buddy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBAB(t *testing.T) {
	bab := blockAllocationBitmap{}
	bab.Expand()
	b := int64(0)

	for bss := minBlockSizeShift; bss < maxBlockSizeShift; bss++ {
		bab.AddBlockSize(b, bss)
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
		bab.AddBlockSize(MaxBlockSize+b, bss)
		bss2, ok := bab.GetBlockSize(b)

		if assert.True(t, ok) {
			assert.Equal(t, bss, bss2)
		}

		b = (2 << bss)
	}

	b = 0

	for bss := minBlockSizeShift; bss < maxBlockSizeShift; bss++ {
		bss2, ok := bab.DeleteBlockSize(MaxBlockSize + b)

		if assert.True(t, ok) {
			assert.Equal(t, bss, bss2)
		}

		_, ok = bab.DeleteBlockSize(MaxBlockSize + b)
		assert.False(t, ok)
		b = (2 << bss)
	}

	bab.Shrink()
	b = 0
	bss := minBlockSizeShift

	for b2 := int64(0); b2 < MaxBlockSize; b2 += MinBlockSize {
		bss2, ok := bab.DeleteBlockSize(b2)

		if ok {
			t.Log("b2=", b2)
		}

		if b2 == b {
			if assert.True(t, ok) {
				assert.Equal(t, bss, bss2)
			}

			_, ok = bab.DeleteBlockSize(b)
			assert.False(t, ok, "%v", b2)
			b = (2 << bss)
			bss++
		} else {
			assert.False(t, ok, "%v %v", b2, b)
		}
	}

	bab.Shrink()
}
