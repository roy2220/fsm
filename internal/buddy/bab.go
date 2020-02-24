package buddy

const blockAllocationSubBitmapSize = (((1 << (maxBlockSizeShift - minBlockSizeShift + 1)) - 1) + 7) >> 3

type blockAllocationBitmap []uint8

func (bab *blockAllocationBitmap) Expand() {
	*bab = append(*bab, make([]uint8, blockAllocationSubBitmapSize)...)
}

func (bab *blockAllocationBitmap) Shrink() {
	*bab = (*bab)[:len(*bab)-blockAllocationSubBitmapSize]
}

func (bab blockAllocationBitmap) AllocateBlock(block int64, blockSizeShift int) {
	sub, subBlock := bab.getSub(block)
	sub.AllocateBlock(subBlock, blockSizeShift)
}

func (bab blockAllocationBitmap) FreeBlock(block int64) (int, bool) {
	sub, subBlock := bab.getSub(block)
	return sub.FreeBlock(subBlock)
}

func (bab blockAllocationBitmap) GetBlockSize(block int64) (int, bool) {
	sub, subBlock := bab.getSub(block)
	return sub.GetBlockSize(subBlock)
}

func (bab blockAllocationBitmap) GetFreeBlocks(callback func(int64, int)) {
	block := int64(0)

	for i := 0; i < len(bab); i += blockAllocationSubBitmapSize {
		sub := blockAllocationSubBitmap(bab[i : i+blockAllocationSubBitmapSize])

		sub.GetFreeBlocks(func(subBlock int64, blockSizeShift int) {
			callback(block|subBlock, blockSizeShift)
		})

		block += MaxBlockSize
	}
}

func (bab blockAllocationBitmap) getSub(block int64) (blockAllocationSubBitmap, int64) {
	i := (block >> maxBlockSizeShift) * blockAllocationSubBitmapSize
	sub := blockAllocationSubBitmap(bab[i : i+blockAllocationSubBitmapSize])
	subBlock := block & (MaxBlockSize - 1)
	return sub, subBlock
}

type blockAllocationSubBitmap []uint8

func (basb blockAllocationSubBitmap) AllocateBlock(block int64, blockSizeShift int) {
	bitPos := locateBit(block, blockSizeShift)

	for {
		basb.setBit(bitPos)

		if siblingBitPos := locateSiblingBit(bitPos); siblingBitPos < 0 || basb.testBit(siblingBitPos) {
			return
		}

		bitPos = locateParentBit(bitPos)
	}
}

func (basb blockAllocationSubBitmap) FreeBlock(block int64) (int, bool) {
	blockSizeShift, bitPos, ok := basb.doGetBlockSize(block)

	if !ok {
		return 0, false
	}

	for {
		basb.clearBit(bitPos)

		if siblingBitPos := locateSiblingBit(bitPos); siblingBitPos < 0 || basb.testBit(siblingBitPos) {
			break
		}

		bitPos = locateParentBit(bitPos)
	}

	return blockSizeShift, true
}

func (basb blockAllocationSubBitmap) GetBlockSize(block int64) (int, bool) {
	blockSizeShift, _, ok := basb.doGetBlockSize(block)
	return blockSizeShift, ok
}

func (basb blockAllocationSubBitmap) GetFreeBlocks(callback func(int64, int)) {
	basb.doGetFreeBlocks(0, maxBlockSizeShift, callback)
}

func (basb blockAllocationSubBitmap) doGetBlockSize(block int64) (int, int, bool) {
	blockSizeShift := minBlockSizeShift
	bitPos := locateBit(block, blockSizeShift)
	rightChildBitPos := -1

	for {
		if basb.testBit(bitPos) {
			if rightChildBitPos >= 0 && basb.testBit(rightChildBitPos) {
				return 0, 0, false
			}

			return blockSizeShift, bitPos, true
		}

		if !bitIsLeft(bitPos) {
			return 0, 0, false
		}

		rightChildBitPos = locateRightSiblingBit(bitPos)
		blockSizeShift++
		bitPos = locateParentBit(bitPos)
	}
}

func (basb blockAllocationSubBitmap) doGetFreeBlocks(bitPos int, blockSizeShift int, callback func(int64, int)) {
	if !basb.testBit(bitPos) {
		if siblingBitPos := locateSiblingBit(bitPos); siblingBitPos < 0 || basb.testBit(siblingBitPos) {
			block := convertBitPosToBlock(bitPos, blockSizeShift)
			callback(block, blockSizeShift)
		}

		return
	}

	if blockSizeShift == minBlockSizeShift {
		return
	}

	leftChildBitPos := locateLeftChildBit(bitPos)
	rightChildBitPos := locateRightSiblingBit(leftChildBitPos)
	basb.doGetFreeBlocks(leftChildBitPos, blockSizeShift-1, callback)
	basb.doGetFreeBlocks(rightChildBitPos, blockSizeShift-1, callback)
}

func (basb blockAllocationSubBitmap) setBit(bitPos int) {
	basb[bitPos>>3] |= 1 << (bitPos & 7)
}

func (basb blockAllocationSubBitmap) clearBit(bitPos int) {
	basb[bitPos>>3] &^= 1 << (bitPos & 7)
}

func (basb blockAllocationSubBitmap) testBit(bitPos int) bool {
	return basb[bitPos>>3]&(1<<(bitPos&7)) != 0
}

func locateBit(block int64, blockSizeShift int) int {
	// return ((1 << (maxBlockSizeShift - blockSizeShift)) - 1) + int(block>>blockSizeShift)
	return int((block+(1<<maxBlockSizeShift))>>blockSizeShift) - 1
}

func locateParentBit(bitPos int) int {
	return ((bitPos + 1) >> 1) - 1
}

func locateSiblingBit(bitPos int) int {
	return ((bitPos + 1) ^ 1) - 1
}

func bitIsLeft(bitPos int) bool {
	return bitPos&1 == 1
}

func locateRightSiblingBit(bitPos int) int {
	return bitPos + 1
}

func locateLeftChildBit(bitPos int) int {
	return ((bitPos + 1) << 1) - 1
}

func convertBitPosToBlock(bitPos int, blockSizeShift int) int64 {
	// return int64(bitPos-((1<<(maxBlockSizeShift-blockSizeShift))-1)) << blockSizeShift
	return (int64(bitPos+1) << blockSizeShift) - (1 << maxBlockSizeShift)
}
