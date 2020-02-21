package buddy

const blockAllocationSubBitmapSize = (((1 << (maxBlockSizeShift - minBlockSizeShift + 1)) - 1) + 7) >> 3

type blockAllocationBitmap []uint8

func (bab *blockAllocationBitmap) Expand() {
	*bab = append(*bab, make([]uint8, blockAllocationSubBitmapSize)...)
}

func (bab *blockAllocationBitmap) Shrink() {
	*bab = (*bab)[:len(*bab)-blockAllocationSubBitmapSize]
}

func (bab blockAllocationBitmap) AddBlockSize(block int64, blockSizeShift int) {
	i := (block >> maxBlockSizeShift) * blockAllocationSubBitmapSize
	blockAllocationSubBitmap(bab[i:i+blockAllocationSubBitmapSize]).AddBlockSize(block&(MaxBlockSize-1), blockSizeShift)
}

func (bab blockAllocationBitmap) DeleteBlockSize(block int64) (int, bool) {
	i := (block >> maxBlockSizeShift) * blockAllocationSubBitmapSize
	return blockAllocationSubBitmap(bab[i : i+blockAllocationSubBitmapSize]).DeleteBlockSize(block & (MaxBlockSize - 1))
}

func (bab blockAllocationBitmap) GetBlockSize(block int64) (int, bool) {
	i := (block >> maxBlockSizeShift) * blockAllocationSubBitmapSize
	return blockAllocationSubBitmap(bab[i : i+blockAllocationSubBitmapSize]).GetBlockSize(block & (MaxBlockSize - 1))
}

type blockAllocationSubBitmap []uint8

func (bab blockAllocationSubBitmap) AddBlockSize(block int64, blockSizeShift int) {
	pos := ((1 << (maxBlockSizeShift - blockSizeShift)) - 1) + (int(block) >> blockSizeShift)
	bab.setBit(pos)
}

func (bab blockAllocationSubBitmap) DeleteBlockSize(block int64) (int, bool) {
	for blockSizeShift := minBlockSizeShift; ; blockSizeShift++ {
		pos := ((1 << (maxBlockSizeShift - blockSizeShift)) - 1) + (int(block) >> blockSizeShift)

		if bab.testBit(pos) {
			bab.clearBit(pos)
			return blockSizeShift, true
		}

		if pos&1 == 0 {
			return 0, false
		}
	}
}

func (bab blockAllocationSubBitmap) GetBlockSize(block int64) (int, bool) {
	for blockSizeShift := minBlockSizeShift; ; blockSizeShift++ {
		pos := ((1 << (maxBlockSizeShift - blockSizeShift)) - 1) + (int(block) >> blockSizeShift)

		if bab.testBit(pos) {
			return blockSizeShift, true
		}

		if pos&1 == 0 {
			return 0, false
		}
	}
}

func (bab blockAllocationSubBitmap) setBit(pos int) {
	bab[pos>>3] |= 1 << (pos & 7)
}

func (bab blockAllocationSubBitmap) clearBit(pos int) {
	bab[pos>>3] &^= 1 << (pos & 7)
}

func (bab blockAllocationSubBitmap) testBit(pos int) bool {
	return bab[pos>>3]&(1<<(pos&7)) != 0
}
