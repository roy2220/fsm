// Package pool implements a pool of space.
package pool

import (
	"errors"
	"fmt"
	"io"

	"github.com/roy2220/fsm/internal/buddy"
	"github.com/roy2220/fsm/internal/list"
)

// Pool represents a pool of space.
type Pool struct {
	buddy              *buddy.Buddy
	listOfPooledBlocks list.List64
	dismissedSpaceSize int
}

// Init initializes the pool with the given buddy system and returns it.
func (p *Pool) Init(buddy *buddy.Buddy) *Pool {
	p.buddy = buddy
	p.listOfPooledBlocks.Init()
	return p
}

// Build returns a builder of the pool.
func (p *Pool) Build() Builder {
	return Builder{p}
}

// AllocateSpace allocates space with the given size
// from the pool and returns it and it's actual size.
func (p *Pool) AllocateSpace(spaceSize int) (int64, int) {
	if chunkSize := chunkHeaderSize + spaceSize; chunkSize <= maxChunkSize {
		if chunkSize < minChunkSize {
			chunkSize = minChunkSize
		}

		block, chunk, chunkSize := p.allocateChunk(chunkSize)
		return makeChunkSpace(block, chunk), calculateChunkSpaceSize(chunkSize)
	}

	unmanagedBlock, unmanagedBlockSize, err := p.buddy.AllocateBlock(spaceSize)

	if err != nil {
		panic(err)
	}

	return unmanagedBlock, unmanagedBlockSize
}

// FreeSpace releases the given space back to the pool.
func (p *Pool) FreeSpace(space int64) {
	if block, chunk, ok := parseChunkSpace(space); ok {
		p.freeChunk(block, chunk)
		return
	}

	if err := p.buddy.FreeBlock(space); err != nil {
		panic(err)
	}
}

// GetSpaceSize returns the size of the given space of the pool.
func (p *Pool) GetSpaceSize(space int64) int {
	if block, chunk, ok := parseChunkSpace(space); ok {
		return calculateChunkSpaceSize(p.getChunkSize(block, chunk))
	}

	unmanagedBlockSize, err := p.buddy.GetBlockSize(space)

	if err != nil {
		panic(err)
	}

	return unmanagedBlockSize
}

// StorePooledBlockList stores the pooled block list of the pool to the given buffer.
func (p *Pool) StorePooledBlockList(buffer []byte) {
	p.listOfPooledBlocks.Store(buffer)
}

// DismissedSpaceSize returns the dismissed space size of the pool.
func (p *Pool) DismissedSpaceSize() int {
	return p.dismissedSpaceSize
}

// Fprint dumps the pool tree as plain text for debugging purposes
func (p *Pool) Fprint(writer io.Writer) error {
	getBlock := p.listOfPooledBlocks.GetItems()
	spaceAccessor := p.buddy.SpaceMapper().AccessSpace()

	for block, ok := getBlock(spaceAccessor); ok; block, ok = getBlock(spaceAccessor) {
		p.doFprint(writer, spaceAccessor, block)
	}

	return nil
}

func (p *Pool) allocateChunk(chunkSize int) (int64, int32, int) {
	getBlock := p.listOfPooledBlocks.GetItems()
	spaceAccessor := p.buddy.SpaceMapper().AccessSpace()

	for block, ok := getBlock(spaceAccessor); ok; block, ok = getBlock(spaceAccessor) {
		if chunk, chunkSize, ok := p.splitChunk(spaceAccessor, block, chunkSize); ok {
			return block, chunk, chunkSize
		}
	}

	block, chunk := p.allocateBlock(chunkSize)
	return block, chunk, chunkSize
}

func (p *Pool) freeChunk(block int64, chunk int32) {
	spaceAccessor := p.accessSpace()

	if chunkSize := p.mergeChunk(spaceAccessor, block, chunk); chunkSize == blockPayloadSize {
		p.freeBlock(spaceAccessor, block)
	}
}

func (p *Pool) getChunkSize(block int64, chunk int32) int {
	chunkController := chunkController{accessBlock(p.accessSpace(), block), chunk}

	if !chunkController.IsUsed() {
		panic(errInvalidChunk)
	}

	return int(chunkController.Size())
}

func (p *Pool) splitChunk(spaceAccessor []byte, block int64, chunkSize int) (int32, int, bool) {
	blockAccessor := accessBlock(spaceAccessor, block)
	blockHeader := blockHeader(blockAccessor)
	listOfFreeChunks := blockHeader.ListOfFreeChunks()
	getChunk := getFreeChunks(listOfFreeChunks)

	for chunk, ok := getChunk(blockAccessor); ok; chunk, ok = getChunk(blockAccessor) {
		chunkController1 := chunkController{blockAccessor, chunk}
		chunkSize2 := int(chunkController1.Size())

		if remainingChunkSize := chunkSize2 - chunkSize; remainingChunkSize >= 0 {
			if remainingChunkSize < minChunkSize {
				chunkSize = chunkSize2
				remainingChunkSize = 0
			} else {
				if chunkIsViolated(chunk + int32(chunkSize)) {
					if remainingChunkSize == minChunkSize {
						chunkSize = chunkSize2
						remainingChunkSize = 0
					} else {
						chunkSize++
						remainingChunkSize--
					}
				}
			}

			if remainingChunkSize >= 1 {
				remainingChunkController := chunkController{blockAccessor, chunk + int32(chunkSize)}
				remainingChunkController.SetUsed(false)
				listOfChunks := blockHeader.ListOfChunks()
				remainingChunkController.InsertAfter(&listOfChunks, chunk)
				blockHeader.SetListOfChunks(listOfChunks)
				remainingChunkController.SetMissCount(0)
				remainingChunkController.InsertFreeAfter(&listOfFreeChunks, chunk)
			}

			chunkController1.SetUsed(true)
			chunkController1.SetAsFirstFree(&listOfFreeChunks)
			chunkController1.RemoveFree(&listOfFreeChunks)
			blockHeader.SetListOfFreeChunks(listOfFreeChunks)
			p.listOfPooledBlocks.SetHead(spaceAccessor, block)

			if listOfFreeChunks.IsEmpty() {
				p.listOfPooledBlocks.RemoveItem(spaceAccessor, block)
			}

			return chunk, chunkSize, true
		}

		missCount := int(chunkController1.MissCount()) + 1
		chunkController1.SetMissCount(int8(missCount))

		if missCount == maxMissCount {
			chunkController1.RemoveFree(&listOfFreeChunks)
			p.dismissedSpaceSize += chunkSize2
		}
	}

	blockHeader.SetListOfFreeChunks(listOfFreeChunks)

	if listOfFreeChunks.IsEmpty() {
		p.listOfPooledBlocks.RemoveItem(spaceAccessor, block)
	}

	return 0, 0, false
}

func (p *Pool) mergeChunk(spaceAccessor []byte, block int64, chunk int32) int {
	blockAccessor := accessBlock(spaceAccessor, block)
	chunkController1 := chunkController{blockAccessor, chunk}

	if !chunkController1.IsUsed() {
		panic(errInvalidChunk)
	}

	blockHeader := blockHeader(blockAccessor)
	listOfChunks := blockHeader.ListOfChunks()
	listOfFreeChunks := blockHeader.ListOfFreeChunks()
	listOfFreeChunksWasEmpty := listOfFreeChunks.IsEmpty()

	if chunkPrev := chunkController1.Prev(); chunkPrev < chunk {
		if chunkPrevController := (chunkController{blockAccessor, chunkPrev}); !chunkPrevController.IsUsed() {
			if chunkPrevController.MissCount() == maxMissCount {
				p.dismissedSpaceSize -= int(chunkPrevController.Size())
			} else {
				chunkPrevController.RemoveFree(&listOfFreeChunks)
			}

			chunkController1.Remove(&listOfChunks)
			chunkController1 = chunkPrevController
		}
	}

	if chunkNext := chunkController1.Next(); chunkNext > chunk {
		if chunkNextController := (chunkController{blockAccessor, chunkNext}); !chunkNextController.IsUsed() {
			if chunkNextController.MissCount() == maxMissCount {
				p.dismissedSpaceSize -= int(chunkNextController.Size())
			} else {
				chunkNextController.RemoveFree(&listOfFreeChunks)
			}

			chunkNextController.Remove(&listOfChunks)
		}
	}

	chunkController1.SetUsed(false)
	chunkController1.SetMissCount(0)
	chunkController1.PrependFree(&listOfFreeChunks)
	blockHeader.SetListOfChunks(listOfChunks)
	blockHeader.SetListOfFreeChunks(listOfFreeChunks)

	if !listOfFreeChunksWasEmpty {
		p.listOfPooledBlocks.RemoveItem(spaceAccessor, block)
	}

	p.listOfPooledBlocks.PrependItem(spaceAccessor, block)
	return int(chunkController1.Size())
}

func (p *Pool) allocateBlock(chunkSize int) (int64, int32) {
	block, _, _ := p.buddy.AllocateBlock(blockSize)
	spaceAccessor := p.accessSpace()
	blockAccessor := accessBlock(spaceAccessor, block)
	chunk := int32(blockHeaderSize)

	if chunkIsViolated(chunk + int32(chunkSize)) {
		chunkSize++
	}

	chunkController1 := chunkController{blockAccessor, chunk}
	chunkController1.SetUsed(true)
	listOfChunks := new(list.List32).Init()
	chunkController1.Prepend(listOfChunks)
	remainingChunkController := chunkController{blockAccessor, chunk + int32(chunkSize)}
	remainingChunkController.SetUsed(false)
	remainingChunkController.InsertAfter(listOfChunks, chunk)
	remainingChunkController.SetMissCount(0)
	listOfFreeChunks := new(list.List32).Init()
	remainingChunkController.PrependFree(listOfFreeChunks)
	blockHeader := blockHeader(blockAccessor)
	blockHeader.SetListOfChunks(*listOfChunks)
	blockHeader.SetListOfFreeChunks(*listOfFreeChunks)
	p.listOfPooledBlocks.PrependItem(spaceAccessor, block)
	return block, chunk
}

func (p *Pool) freeBlock(spaceAccessor []byte, block int64) {
	p.listOfPooledBlocks.RemoveItem(spaceAccessor, block)
	p.buddy.FreeBlock(block)
}

func (p *Pool) accessSpace() []byte {
	return p.buddy.SpaceMapper().AccessSpace()
}

func (p *Pool) doFprint(writer io.Writer, spaceAccessor []byte, block int64) error {
	if _, err := fmt.Fprintf(writer, "pooled block %d:", block); err != nil {
		return err
	}

	blockAccessor := accessBlock(spaceAccessor, block)
	blockHeader := blockHeader(blockAccessor)
	listOfChunks := blockHeader.ListOfChunks()
	getChunk := listOfChunks.GetItems()

	for chunk, ok := getChunk(blockAccessor); ok; chunk, ok = getChunk(blockAccessor) {
		chunkController := chunkController{blockAccessor, chunk}
		var err error

		if chunkController.IsUsed() {
			_, err = fmt.Fprintf(writer, " (%d, %d)", chunk, chunk+chunkController.Size())
		} else {
			_, err = fmt.Fprintf(writer, " [%d, %d]", chunk, chunk+chunkController.Size())
		}

		if err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(writer, ""); err != nil {
		return err
	}

	return nil
}

// Builder represents a builder of pools of space.
type Builder struct {
	p *Pool
}

// LoadPooledBlockList loads the pooled block list from the given data.
func (b Builder) LoadPooledBlockList(data []byte) Builder {
	b.p.listOfPooledBlocks.Load(data)
	return b
}

// SetDismissedSpaceSize sets the dismissed space size.
func (b Builder) SetDismissedSpaceSize(dismissedSpaceSize int) Builder {
	b.p.dismissedSpaceSize = dismissedSpaceSize
	return b
}

const (
	blockSize             = 1 << 20
	blockPayloadSize      = blockSize - blockHeaderSize
	minChunkSize          = freeChunkHeaderSize
	maxChunkSize          = blockPayloadSize / 16
	minUnmanagedBlockSize = (maxChunkSize + (buddy.MinBlockSize - 1)) &^ (buddy.MinBlockSize - 1)
	maxMissCount          = 3
)

type blockHeader []byte

func (bh blockHeader) SetListOfChunks(listOfChunks list.List32) {
	listOfChunks.Store(bh[list.ItemSize64:])
}

func (bh blockHeader) ListOfChunks() list.List32 {
	var listOfChunks list.List32
	listOfChunks.Load(bh[list.ItemSize64:])
	return listOfChunks
}

func (bh blockHeader) SetListOfFreeChunks(listOfFreeChunks list.List32) {
	listOfFreeChunks.Store(bh[list.ItemSize64+list.Size32:])
}

func (bh blockHeader) ListOfFreeChunks() list.List32 {
	var listOfFreeChunks list.List32
	listOfFreeChunks.Load(bh[list.ItemSize64+list.Size32:])
	return listOfFreeChunks
}

const blockHeaderSize = list.ItemSize64 + 2*list.Size32

type chunkController struct {
	blockAccessor []byte
	c             int32
}

func (cc chunkController) SetUsed(isUsed bool) {
	var flags int8

	if isUsed {
		flags = int8(cc.c%3 + 1)
	} else {
		flags = 0
	}

	list.SetItem32Flags(cc.blockAccessor, cc.c, flags)
}

func (cc chunkController) Prepend(listOfChunks *list.List32) {
	listOfChunks.PrependItem(cc.blockAccessor, cc.c)
}

func (cc chunkController) InsertAfter(listOfChunks *list.List32, other int32) {
	listOfChunks.InsertItemAfter(cc.blockAccessor, cc.c, other)
}

func (cc chunkController) Remove(listOfChunks *list.List32) {
	listOfChunks.RemoveItem(cc.blockAccessor, cc.c)
}

func (cc chunkController) IsUsed() bool {
	return list.Item32Flags(cc.blockAccessor, cc.c) == int8(cc.c%3+1)
}

func (cc chunkController) Prev() int32 {
	return list.Item32Prev(cc.blockAccessor, cc.c)
}

func (cc chunkController) Next() int32 {
	return list.Item32Next(cc.blockAccessor, cc.c)
}

func (cc chunkController) Size() int32 {
	chunkEnd := cc.Next()

	if chunkEnd <= cc.c {
		chunkEnd = blockSize
	}

	return chunkEnd - cc.c
}

const (
	chunkHeaderSize           = list.ItemSize32
	freeListItemOffsetOfChunk = chunkHeaderSize
)

func (cc chunkController) SetMissCount(missCount int8) {
	list.SetItem32Flags(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk, missCount)
}

func (cc chunkController) PrependFree(listOfFreeChunks *list.List32) {
	listOfFreeChunks.PrependItem(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk)
}

func (cc chunkController) InsertFreeAfter(listOfFreeChunks *list.List32, other int32) {
	listOfFreeChunks.InsertItemAfter(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk, other+freeListItemOffsetOfChunk)
}

func (cc chunkController) RemoveFree(listOfFreeChunks *list.List32) {
	listOfFreeChunks.RemoveItem(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk)
}

func (cc chunkController) SetAsFirstFree(listOfFreeChunks *list.List32) {
	listOfFreeChunks.SetHead(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk)
}

func (cc chunkController) MissCount() int8 {
	return list.Item32Flags(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk)
}

func (cc chunkController) FreePrev() int32 {
	return list.Item32Prev(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk) - freeListItemOffsetOfChunk
}

func (cc chunkController) FreeNext() int32 {
	return list.Item32Next(cc.blockAccessor, cc.c+freeListItemOffsetOfChunk) - freeListItemOffsetOfChunk
}

const freeChunkHeaderSize = freeListItemOffsetOfChunk + list.ItemSize32

var errInvalidChunk = errors.New("pool: invalid chunk")

func makeChunkSpace(block int64, chunk int32) int64 {
	return block | int64(chunk+chunkHeaderSize)
}

func parseChunkSpace(chunkSpace int64) (int64, int32, bool) {
	if chunkSpace&(minUnmanagedBlockSize-1) == 0 {
		return 0, 0, false
	}

	block := chunkSpace &^ (blockSize - 1)
	chunk := int32((chunkSpace & (blockSize - 1))) - chunkHeaderSize
	return block, chunk, true
}

func chunkIsViolated(chunk int32) bool {
	return (chunk+chunkHeaderSize)&(minUnmanagedBlockSize-1) == 0
}

func calculateChunkSpaceSize(chunkSize int) int {
	return chunkSize - chunkHeaderSize
}

func accessBlock(spaceAccessor []byte, block int64) []byte {
	return spaceAccessor[block : block+blockSize]
}

func getFreeChunks(listOfFreeChunks list.List32) func([]byte) (int32, bool) {
	getListItemOfFreeChunk := listOfFreeChunks.GetItems()

	return func(spaceAccessor []byte) (int32, bool) {
		freeListItemOfChunk, ok := getListItemOfFreeChunk(spaceAccessor)

		if !ok {
			return 0, false
		}

		chunk := freeListItemOfChunk - freeListItemOffsetOfChunk
		return chunk, true
	}
}
