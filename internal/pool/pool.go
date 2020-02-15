// Package pool implements a pool of space.
package pool

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/roy2220/fsm/internal/buddy"
	"github.com/roy2220/fsm/internal/rbtree"
)

// BlockSize is the block size of the pool.
const BlockSize = buddy.MinBlockSize

// Pool represents a pool of space.
type Pool struct {
	buddy                *buddy.Buddy
	rbTreeOfPooledBlocks rbtree.RBTree
}

// Init initializes the pool with the given buddy system and returns it.
func (p *Pool) Init(buddy *buddy.Buddy) *Pool {
	p.buddy = buddy
	p.rbTreeOfPooledBlocks.Init()
	return p
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

	block, blockSize, err := p.buddy.AllocateBlock(spaceSize)

	if err != nil {
		panic(err)
	}

	return block, blockSize
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

	blockSize, err := p.buddy.GetBlockSize(space)

	if err != nil {
		panic(err)
	}

	return blockSize
}

// GetPooledBlocks returns the blocks of the pooled list of the pool.
func (p *Pool) GetPooledBlocks() func() (int64, bool) {
	return p.rbTreeOfPooledBlocks.GetKeys()
}

// PutPooledBlocks puts the given block to the pooled list of the pool.
func (p *Pool) PutPooledBlocks(block int64) {
	p.rbTreeOfPooledBlocks.InsertKey(block)
}

// Fprint dumps the pool tree as plain text for debugging purposes
func (p *Pool) Fprint(writer io.Writer) error {
	getPooledBlock := p.rbTreeOfPooledBlocks.GetKeys()

	for block, ok := getPooledBlock(); ok; block, ok = getPooledBlock() {
		p.doFprint(writer, block)
	}

	return nil
}

func (p *Pool) allocateChunk(chunkSize int) (int64, int32, int) {
	getBlock := p.rbTreeOfPooledBlocks.GetKeys()
	blocksToDismiss := make([]int64, 0, 1<<maxMissCount)

	dismissBlocks := func() {
		for _, block := range blocksToDismiss {
			blockHeader(p.accessBlock(block)).SetState(blockDismissed)
			p.rbTreeOfPooledBlocks.RemoveKey(block)
		}
	}

	for block, ok := getBlock(); ok; block, ok = getBlock() {
		blockAccessor := p.accessBlock(block)
		blockHeader := blockHeader(blockAccessor)

		if chunkSize2 := int(blockHeader.MaxFreeChunkSize()); !(chunkSize2 >= 1 && chunkSize > chunkSize2) {
			if chunk, chunkSize, ok := p.splitChunk(block, blockAccessor, chunkSize); ok {
				blockHeader.SetMissCount(0)
				dismissBlocks()
				return block, chunk, chunkSize
			}
		}

		missCount := blockHeader.MissCount() + 1
		blockHeader.SetMissCount(missCount)

		if missCount < maxMissCount {
			continue
		}

		blocksToDismiss = append(blocksToDismiss, block)
	}

	dismissBlocks()
	block, chunk := p.allocateBlock(chunkSize)
	return block, chunk, chunkSize
}

func (p *Pool) freeChunk(block int64, chunk int32) {
	blockAccessor := p.accessBlock(block)

	if chunkSize := p.mergeChunk(block, blockAccessor, chunk); chunkSize > maxChunkSize {
		p.freeBlock(block)
	}
}

func (p *Pool) getChunkSize(block int64, chunk int32) int {
	blockAccessor := p.accessBlock(block)
	chunkHeader := chunkHeader(blockAccessor[chunk:])

	if chunkHeader.Next() != chunkAllocationMark {
		panic(errInvalidChunk)
	}

	return int(chunkHeader.Size())
}

func (p *Pool) splitChunk(block int64, blockAccessor []byte, chunkSize int) (int32, int, bool) {
	blockHeader := blockHeader(blockAccessor)
	chunk := blockHeader.FirstFreeChunk()
	lastChunkHeader := chunkHeader(nil)
	maxChunkSize2 := 0

	for {
		chunkHeader1 := chunkHeader(blockAccessor[chunk:])
		chunkSize2 := int(chunkHeader1.Size())
		chunkNext := chunkHeader1.Next()

		if remainingChunkSize := chunkSize2 - chunkSize; remainingChunkSize >= 0 {
			chunkHeader1.SetNext(chunkAllocationMark)

			if remainingChunkSize < minChunkSize {
				chunkSize = chunkSize2
			} else {
				chunkHeader1.SetSize(int32(chunkSize))
				remainingChunk := chunk + int32(chunkSize)
				remainingChunkHeader := chunkHeader(blockAccessor[remainingChunk:])
				remainingChunkHeader.SetNext(chunkNext)
				remainingChunkHeader.SetSize(int32(remainingChunkSize))
				chunkNext = remainingChunk
			}

			blockIsExhausted := false

			if lastChunkHeader == nil {
				blockHeader.SetFirstFreeChunk(chunkNext)
				blockIsExhausted = chunkNext < 1
			} else {
				lastChunkHeader.SetNext(chunkNext)
			}

			if blockIsExhausted {
				blockHeader.SetMaxFreeChunkSize(0)
				blockHeader.SetState(blockExhausted)
				p.rbTreeOfPooledBlocks.RemoveKey(block)
			} else {
				if chunkSize3 := int(blockHeader.MaxFreeChunkSize()); chunkSize3 >= 1 && chunkSize2 == chunkSize3 {
					blockHeader.SetMaxFreeChunkSize(0)
				}
			}

			return chunk, chunkSize, true
		}

		if chunkSize2 > maxChunkSize2 {
			maxChunkSize2 = chunkSize2
		}

		lastChunkHeader = chunkHeader1
		chunk = chunkNext

		if chunk < 1 {
			break
		}
	}

	blockHeader.SetMaxFreeChunkSize(int32(maxChunkSize2))
	return 0, 0, false
}

func (p *Pool) mergeChunk(block int64, blockAccessor []byte, chunk int32) int {
	chunkHeader1 := chunkHeader(blockAccessor[chunk:])

	if chunkHeader1.Next() != chunkAllocationMark {
		panic(errInvalidChunk)
	}

	chunkEnd := chunk + int32(chunkHeader1.Size())
	blockHeader := blockHeader(blockAccessor)
	chunk2 := blockHeader.FirstFreeChunk()
	lastChunkHeader2 := chunkHeader(nil)

	for {
		if chunk2 < 1 {
			chunkHeader1.SetNext(0)

			if lastChunkHeader2 == nil {
				blockHeader.SetFirstFreeChunk(chunk)
			} else {
				lastChunkHeader2.SetNext(chunk)
			}

			break
		}

		chunkHeader2 := chunkHeader(blockAccessor[chunk2:])
		chunkEnd2 := chunk2 + int32(chunkHeader2.Size())
		chunkNext2 := chunkHeader2.Next()

		if chunkEnd <= chunk2 {
			if chunkEnd < chunk2 {
				chunkHeader1.SetNext(chunk2)
			} else {
				chunkHeader1.SetNext(chunkNext2)
				chunkHeader1.SetSize(chunkEnd2 - chunk)
				chunkEnd = chunkEnd2
			}

			if lastChunkHeader2 == nil {
				blockHeader.SetFirstFreeChunk(chunk)
			} else {
				lastChunkHeader2.SetNext(chunk)
			}

			break
		}

		if chunk == chunkEnd2 {
			chunkHeader2.SetSize(chunkEnd - chunk2)
			chunkHeader1.SetNext(0)
			chunk, chunkHeader1 = chunk2, chunkHeader2
		} else {
			lastChunkHeader2 = chunkHeader2
		}

		chunk2 = chunkNext2
	}

	chunkSize := int(chunkEnd - chunk)

	if blockState := blockHeader.State(); blockState == blockExhausted {
		blockHeader.SetMaxFreeChunkSize(int32(chunkSize))
		blockHeader.SetState(blockPooled)
		p.rbTreeOfPooledBlocks.InsertKey(block)
	} else {
		if chunkSize2 := int(blockHeader.MaxFreeChunkSize()); chunkSize2 >= 1 && chunkSize > chunkSize2 {
			blockHeader.SetMaxFreeChunkSize(int32(chunkSize))
		}

		if missCount := blockHeader.MissCount(); missCount >= 1 {
			blockHeader.SetMissCount(missCount - 1)
		}

		if blockState == blockDismissed {
			blockHeader.SetState(blockPooled)
			p.rbTreeOfPooledBlocks.InsertKey(block)
		}
	}

	return chunkSize
}

func (p *Pool) allocateBlock(chunkSize int) (int64, int32) {
	block, _, _ := p.buddy.AllocateBlock(BlockSize)
	blockAccessor := p.accessBlock(block)
	chunk := int32(blockHeaderSize)
	chunkHeader1 := chunkHeader(blockAccessor[chunk:])
	chunkHeader1.SetNext(chunkAllocationMark)
	chunkHeader1.SetSize(int32(chunkSize))
	remainingChunk := chunk + int32(chunkSize)
	remainingChunkHeader := chunkHeader(blockAccessor[remainingChunk:])
	remainingChunkHeader.SetNext(0)
	remainingChunkSize := BlockSize - int(remainingChunk)
	remainingChunkHeader.SetSize(int32(remainingChunkSize))
	blockHeader := blockHeader(blockAccessor)
	blockHeader.SetFirstFreeChunk(remainingChunk)
	blockHeader.SetMaxFreeChunkSize(int32(remainingChunkSize))
	blockHeader.SetMissCount(0)
	blockHeader.SetState(blockPooled)
	p.rbTreeOfPooledBlocks.InsertKey(block)
	return block, chunk
}

func (p *Pool) freeBlock(block int64) {
	p.rbTreeOfPooledBlocks.RemoveKey(block)
	p.buddy.FreeBlock(block)
}

func (p *Pool) accessBlock(block int64) []byte {
	return p.buddy.SpaceMapper().AccessSpace()[block:]
}

func (p *Pool) doFprint(writer io.Writer, block int64) error {
	if _, err := fmt.Fprintf(writer, "pooled block %d:", block); err != nil {
		return err
	}

	blockAccessor := p.accessBlock(block)
	blockHeader := blockHeader(blockAccessor)
	chunk := blockHeader.FirstFreeChunk()

	for chunk >= 1 {
		chunkHeader := chunkHeader(blockAccessor[chunk:])

		if _, err := fmt.Fprintf(writer, " [%d, %d]", chunk, chunk+chunkHeader.Size()); err != nil {
			return err
		}

		chunk = chunkHeader.Next()
	}

	if _, err := fmt.Fprintln(writer, ""); err != nil {
		return err
	}

	return nil
}

const (
	minChunkSize        = chunkHeaderSize + 8
	maxChunkSize        = BlockSize - blockHeaderSize - minChunkSize
	maxMissCount        = 3
	chunkAllocationMark = -0xBADBEEF
)

const (
	blockPooled = blockState(iota)
	blockDismissed
	blockExhausted
)

type blockHeader []byte // reserve the first 8 bytes for pooled block linked-list.

func (bh blockHeader) SetFirstFreeChunk(firstFreeChunk int32) {
	binary.BigEndian.PutUint32(bh[8:], uint32(firstFreeChunk))
}

func (bh blockHeader) FirstFreeChunk() int32 {
	return int32(binary.BigEndian.Uint32(bh[8:]))
}

func (bh blockHeader) SetMaxFreeChunkSize(maxFreeChunkSize int32) {
	binary.BigEndian.PutUint32(bh[12:], uint32(maxFreeChunkSize))
}

func (bh blockHeader) MaxFreeChunkSize() int32 {
	return int32(binary.BigEndian.Uint32(bh[12:]))
}

func (bh blockHeader) SetMissCount(missCount int8) {
	bh[16] = byte(missCount)
}

func (bh blockHeader) MissCount() int8 {
	return int8(bh[16])
}

func (bh blockHeader) SetState(state blockState) {
	bh[17] = byte(state)
}

func (bh blockHeader) State() blockState {
	return blockState(bh[17])
}

const blockHeaderSize = 18

type blockState int

type chunkHeader []byte

func (ch chunkHeader) SetNext(next int32) {
	binary.BigEndian.PutUint32(ch[0:], uint32(next))
}

func (ch chunkHeader) Next() int32 {
	return int32(binary.BigEndian.Uint32(ch[0:]))
}

func (ch chunkHeader) SetSize(size int32) {
	binary.BigEndian.PutUint32(ch[4:], uint32(size))
}

func (ch chunkHeader) Size() int32 {
	return int32(binary.BigEndian.Uint32(ch[4:]))
}

const chunkHeaderSize = 8

var errInvalidChunk = errors.New("pool: invalid chunk")

func makeChunkSpace(block int64, chunk int32) int64 {
	return block | int64(chunk+chunkHeaderSize)
}

func parseChunkSpace(chunkSpace int64) (int64, int32, bool) {
	block := chunkSpace &^ (BlockSize - 1)

	if chunkSpace == block {
		return 0, 0, false
	}

	chunk := int32((chunkSpace & (BlockSize - 1))) - chunkHeaderSize
	return block, chunk, true
}

func calculateChunkSpaceSize(chunkSize int) int {
	return chunkSize - chunkHeaderSize
}
