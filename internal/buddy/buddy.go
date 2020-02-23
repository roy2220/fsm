// Package buddy implements a buddy system.
package buddy

import (
	"errors"

	"github.com/roy2220/fsm/internal/rbtree"
	"github.com/roy2220/fsm/internal/spacemapper"
)

const (
	// MinBlockSize is the minimum block size of buddy systems.
	MinBlockSize = 1 << minBlockSizeShift

	// MaxBlockSize is the maximum block size of buddy systems.
	MaxBlockSize = 1 << maxBlockSizeShift

	// NumberOfFreeBlockLists is the number of free block lists of buddy systems.
	NumberOfFreeBlockLists = maxBlockSizeShift - minBlockSizeShift + 1
)

// Buddy represents a buddy system.
type Buddy struct {
	spaceMapper           spacemapper.SpaceMapper
	spaceSize             int
	usedSpaceSize         int
	mappedSpaceSize       int
	allocatedSpaceSize    int
	blockAllocationBitmap blockAllocationBitmap
	rbTreesOfFreeBlocks   [NumberOfFreeBlockLists]rbtree.RBTree
}

// Init initializes the buddy system with the given space mapper and returns it.
func (b *Buddy) Init(spaceMapper spacemapper.SpaceMapper) *Buddy {
	b.spaceMapper = spaceMapper

	for i := range b.rbTreesOfFreeBlocks {
		b.rbTreesOfFreeBlocks[i].Init()
	}

	return b
}

// Build returns a builder of the buddy system.
func (b *Buddy) Build() Builder {
	return Builder{b}
}

// AllocateBlock allocates a block with the given size
// from the buddy system and returns it and it's actual size.
func (b *Buddy) AllocateBlock(blockSize int) (int64, int, error) {
	if blockSize > MaxBlockSize {
		return 0, 0, ErrBlockTooLarge
	}

	freeBlockListIndex := locateFreeBlockList(blockSize)
	block := b.doAllocateBlock(freeBlockListIndex)
	blockSizeShift := calculateBlockSizeShift(freeBlockListIndex)
	blockSize = 1 << blockSizeShift
	b.allocatedSpaceSize += blockSize
	b.blockAllocationBitmap.AllocateBlock(block, blockSizeShift)

	if usedSpaceSize := int(block) + blockSize; usedSpaceSize > b.usedSpaceSize {
		if usedSpaceSize > b.mappedSpaceSize {
			if err := b.mapSpace(usedSpaceSize); err != nil {
				b.FreeBlock(block)
				return 0, 0, err
			}
		}

		b.usedSpaceSize = usedSpaceSize
	}

	return block, blockSize, nil
}

// FreeBlock releases the given block back to the buddy system.
func (b *Buddy) FreeBlock(block int64) error {
	if int(block) >= b.spaceSize {
		return ErrInvalidBlock
	}

	blockSizeShift, ok := b.blockAllocationBitmap.FreeBlock(block)

	if !ok {
		return ErrInvalidBlock
	}

	blockSize := 1 << blockSizeShift
	shrinkUsedSpace := int(block)+blockSize == b.usedSpaceSize
	freeBlockListIndex := calculateFreeBlockListIndex(blockSizeShift)
	b.doFreeBlock(&block, &freeBlockListIndex)
	b.allocatedSpaceSize -= blockSize

	if shrinkUsedSpace {
		if freeBlockListIndex == NumberOfFreeBlockLists-1 {
			rbTreeOfFreeBlocks := &b.rbTreesOfFreeBlocks[NumberOfFreeBlockLists-1]

			for {
				blockPrev := block - int64(MaxBlockSize)

				if !rbTreeOfFreeBlocks.FindKey(blockPrev) {
					break
				}

				block = blockPrev
			}
		}

		for freeBlockListIndex--; freeBlockListIndex >= 0; freeBlockListIndex-- {
			blockPrev := block - int64(calculateBlockSize(freeBlockListIndex))

			if b.rbTreesOfFreeBlocks[freeBlockListIndex].FindKey(blockPrev) {
				block = blockPrev
			}
		}

		b.usedSpaceSize = int(block)

		if b.usedSpaceSize < b.mappedSpaceSize/2 {
			return b.mapSpace(b.usedSpaceSize)
		}
	}

	return nil
}

// GetBlockSize returns the size of the given block of the buddy system.
func (b *Buddy) GetBlockSize(block int64) (int, error) {
	if int(block) >= b.spaceSize {
		return 0, ErrInvalidBlock
	}

	blockSizeShift, ok := b.blockAllocationBitmap.GetBlockSize(block)

	if !ok {
		return 0, ErrInvalidBlock
	}

	return 1 << blockSizeShift, nil
}

// ShrinkSpace shrink the space of the buddy system.
func (b *Buddy) ShrinkSpace() {
	rbTreeOfFreeBlocks := &b.rbTreesOfFreeBlocks[NumberOfFreeBlockLists-1]

	for {
		block := int64(b.spaceSize - MaxBlockSize)

		if !rbTreeOfFreeBlocks.DeleteKey(block) {
			return
		}

		b.spaceSize -= MaxBlockSize
		b.blockAllocationBitmap.Shrink()
	}
}

// SpaceMapper returns the space mapper of the buddy system.
func (b *Buddy) SpaceMapper() spacemapper.SpaceMapper {
	return b.spaceMapper
}

// SpaceSize returns the space size of the buddy system.
func (b *Buddy) SpaceSize() int {
	return b.spaceSize
}

// UsedSpaceSize returns the used space size of the buddy system.
func (b *Buddy) UsedSpaceSize() int {
	return b.usedSpaceSize
}

// MappedSpaceSize returns the mapped space size of the buddy system.
func (b *Buddy) MappedSpaceSize() int {
	return b.mappedSpaceSize
}

// AllocatedSpaceSize returns the allocated space size of the buddy system.
func (b *Buddy) AllocatedSpaceSize() int {
	return b.allocatedSpaceSize
}

// BlockAllocationBitmap returns block allocation bitmap of the buddy system.
func (b *Buddy) BlockAllocationBitmap() []byte {
	return b.blockAllocationBitmap
}

func (b *Buddy) doAllocateBlock(freeBlockListIndex int) int64 {
	rbTreeOfFreeBlocks := &b.rbTreesOfFreeBlocks[freeBlockListIndex]
	block, ok := rbTreeOfFreeBlocks.DeleteMinKey()

	if ok {
		return block
	}

	if freeBlockListIndex == NumberOfFreeBlockLists-1 {
		return b.expandSpace()
	}

	block = b.doAllocateBlock(freeBlockListIndex + 1)
	blockSibling := block + int64(calculateBlockSize(freeBlockListIndex))
	rbTreeOfFreeBlocks.AddKey(blockSibling)
	return block
}

func (b *Buddy) doFreeBlock(block *int64, freeBlockListIndex *int) {
	rbTreeOfFreeBlocks := &b.rbTreesOfFreeBlocks[*freeBlockListIndex]

	if *freeBlockListIndex == NumberOfFreeBlockLists-1 {
		rbTreeOfFreeBlocks.AddKey(*block)
		return
	}

	blockSibling := *block ^ int64(calculateBlockSize(*freeBlockListIndex))

	if ok := rbTreeOfFreeBlocks.DeleteKey(blockSibling); !ok {
		rbTreeOfFreeBlocks.AddKey(*block)
		return
	}

	if blockSibling < *block {
		*block = blockSibling
	}

	*freeBlockListIndex++
	b.doFreeBlock(block, freeBlockListIndex)
}

func (b *Buddy) expandSpace() int64 {
	block := int64(b.spaceSize)
	b.spaceSize += MaxBlockSize
	b.blockAllocationBitmap.Expand()
	return block
}

func (b *Buddy) mapSpace(usedSpaceSize int) error {
	mappedSpaceSize := int(nextPowerOfTwo(int64(usedSpaceSize)))

	if err := b.spaceMapper.MapSpace(mappedSpaceSize); err != nil {
		return err
	}

	b.mappedSpaceSize = mappedSpaceSize
	return nil
}

// Builder represents a builder of buddy systems.
type Builder struct {
	b *Buddy
}

// SetSpaceSize sets the space size of buddy systems to the given value.
func (b Builder) SetSpaceSize(spaceSize int) Builder {
	b.b.spaceSize = spaceSize
	return b
}

// SetUsedSpaceSize sets the used space size of buddy systems to the given value.
func (b Builder) SetUsedSpaceSize(usedSpaceSize int) Builder {
	b.b.usedSpaceSize = usedSpaceSize
	return b
}

// SetMappedSpaceSize sets the mapped space size of buddy systems to the given value.
func (b Builder) SetMappedSpaceSize(mappedSpaceSize int) Builder {
	b.b.mappedSpaceSize = mappedSpaceSize
	return b
}

// SetAllocatedSpaceSize sets the allocated space size of buddy systems to the given value.
func (b Builder) SetAllocatedSpaceSize(allocatedSpaceSize int) Builder {
	b.b.allocatedSpaceSize = allocatedSpaceSize
	return b
}

// SetBlockAllocationBitmap sets the block allocation bitmap of buddy systems to the given value.
func (b Builder) SetBlockAllocationBitmap(blockAllocationBitmap []byte) Builder {
	b.b.blockAllocationBitmap = blockAllocationBitmap

	b.b.blockAllocationBitmap.GetFreeBlocks(func(block int64, blockSizeShift int) {
		freeBlockListIndex := calculateFreeBlockListIndex(blockSizeShift)
		b.b.rbTreesOfFreeBlocks[freeBlockListIndex].AddKey(block)
	})

	return b
}

var (
	// ErrBlockTooLarge is returned when allocating a block too large
	// to allocate from buddy systems.
	ErrBlockTooLarge = errors.New("buddy: block too large")

	// ErrInvalidBlock is returned when freeing or getting size of an invalid block
	ErrInvalidBlock = errors.New("buddy: invalid block")
)

const (
	minBlockSizeShift = 12 // log2 of 4Ki
	maxBlockSizeShift = 32 // log2 of 4Gi
)

func locateFreeBlockList(blockSize int) int {
	for freeBlockListIndex := 0; ; freeBlockListIndex++ {
		if calculateBlockSize(freeBlockListIndex) >= blockSize {
			return freeBlockListIndex
		}
	}
}

func calculateBlockSize(freeBlockListIndex int) int {
	return 1 << calculateBlockSizeShift(freeBlockListIndex)
}

func calculateBlockSizeShift(freeBlockListIndex int) int {
	return minBlockSizeShift + freeBlockListIndex
}

func calculateFreeBlockListIndex(blockSizeShift int) int {
	return blockSizeShift - minBlockSizeShift
}

func nextPowerOfTwo(x int64) int64 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
}
