package fsm

import (
	"encoding/binary"
	"unsafe"

	"github.com/roy2220/fsm/internal/list"
)

const fileHeaderSize = (int(unsafe.Sizeof(fileHeader{})) + (pageSize - 1)) &^ (pageSize - 1)

type fileHeader struct {
	SpaceSize                 int64
	UsedSpaceSize             int64
	MappedSpaceSize           int64
	AllocatedSpaceSize        int64
	BlockAllocationBitmapSize int64
	PooledBlockList           [list.Size64]byte
	DismissedSpaceSize        int64
	PrimarySpace              int64
}

func (fh *fileHeader) Serialize(buffer []byte) {
	_ = buffer[fileHeaderSize-1]
	i := 0
	binary.BigEndian.PutUint64(buffer[i:], uint64(fh.SpaceSize))
	i += 8
	binary.BigEndian.PutUint64(buffer[i:], uint64(fh.UsedSpaceSize))
	i += 8
	binary.BigEndian.PutUint64(buffer[i:], uint64(fh.MappedSpaceSize))
	i += 8
	binary.BigEndian.PutUint64(buffer[i:], uint64(fh.AllocatedSpaceSize))
	i += 8
	binary.BigEndian.PutUint64(buffer[i:], uint64(fh.BlockAllocationBitmapSize))
	i += 8
	copy(buffer[i:], fh.PooledBlockList[:])
	i += len(fh.PooledBlockList)
	binary.BigEndian.PutUint64(buffer[i:], uint64(fh.DismissedSpaceSize))
	i += 8
	binary.BigEndian.PutUint64(buffer[i:], ^uint64(fh.PrimarySpace))
	i += 8

	for ; i < fileHeaderSize; i++ {
		buffer[i] = 0
	}
}

func (fh *fileHeader) Deserialize(data []byte) {
	_ = data[fileHeaderSize-1]
	i := 0
	fh.SpaceSize = int64(binary.BigEndian.Uint64(data[i:]))
	i += 8
	fh.UsedSpaceSize = int64(binary.BigEndian.Uint64(data[i:]))
	i += 8
	fh.MappedSpaceSize = int64(binary.BigEndian.Uint64(data[i:]))
	i += 8
	fh.AllocatedSpaceSize = int64(binary.BigEndian.Uint64(data[i:]))
	i += 8
	fh.BlockAllocationBitmapSize = int64(binary.BigEndian.Uint64(data[i:]))
	i += 8
	copy(fh.PooledBlockList[:], data[i:])
	i += len(fh.PooledBlockList)
	fh.DismissedSpaceSize = int64(binary.BigEndian.Uint64(data[i:]))
	i += 8
	fh.PrimarySpace = int64(^binary.BigEndian.Uint64(data[i:]))
	i += 8
}
