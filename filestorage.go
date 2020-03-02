// Package fsm implements file space management.
package fsm

import (
	"os"

	"github.com/roy2220/fsm/internal/buddy"
	"github.com/roy2220/fsm/internal/pool"
)

// FileStorage represents a file storage.
type FileStorage struct {
	spaceMapper  spaceMapper
	buddy        buddy.Buddy
	pool         pool.Pool
	primarySpace int64
}

// Init initializes the file storage and returns it.
func (fs *FileStorage) Init() *FileStorage {
	fs.buddy.Init(&fs.spaceMapper)
	fs.pool.Init(&fs.buddy)
	fs.primarySpace = -1
	return fs
}

// Open opens a file storage on the given file.
func (fs *FileStorage) Open(fileName string, createFileIfNotExists bool) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0666)

	if err != nil {
		if !(createFileIfNotExists && os.IsNotExist(err)) {
			return err
		}

		file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)

		if err != nil {
			return err
		}

		if _, err := file.WriteAt(make([]byte, fileHeaderSize), 0); err != nil {
			file.Close()
			return err
		}
	}

	fs.spaceMapper.File = file

	if err := fs.loadFile(); err != nil {
		file.Close()
		return err
	}

	return nil
}

// Close closes the file storage.
func (fs *FileStorage) Close() error {
	if err := fs.storeFile(); err != nil {
		return err
	}

	if err := fs.spaceMapper.File.Close(); err != nil {
		return err
	}

	return nil
}

// AllocateSpace allocates space with the given size on the file,
// returns the space allocated and an ephemeral space accessor
// (a byte slice for reading/writing space,
// may get *INVALIDATED* after calling AllocateSpace/FreeSpace).
func (fs *FileStorage) AllocateSpace(spaceSize int) (int64, []byte) {
	space, spaceSize := fs.pool.AllocateSpace(spaceSize)
	spaceAccessor := fs.spaceMapper.AccessSpace()[space : space+int64(spaceSize)]
	return space, spaceAccessor
}

// FreeSpace releases the given space back to the file.
func (fs *FileStorage) FreeSpace(space int64) {
	fs.pool.FreeSpace(space)
}

// AccessSpace returns an ephemeral space accessor of the given space on the file
// (a byte slice for reading/writing space,
// may get *INVALIDATED* after calling AllocateSpace/FreeSpace).
func (fs *FileStorage) AccessSpace(space int64) []byte {
	spaceSize := fs.pool.GetSpaceSize(space)
	spaceAccessor := fs.spaceMapper.AccessSpace()[space : space+int64(spaceSize)]
	return spaceAccessor
}

// SetPrimarySpace set the primary space on the file.
// The primary space is allocated by user and serves for user-defined metadata.
func (fs *FileStorage) SetPrimarySpace(primarySpace int64) {
	fs.primarySpace = primarySpace
}

// PrimarySpace returns the primary space on the file.
// The primary space is allocated by user and serves for user-defined metadata.
func (fs *FileStorage) PrimarySpace() int64 {
	return fs.primarySpace
}

// Stats returns the stats of the file.
func (fs *FileStorage) Stats() Stats {
	return Stats{
		SpaceSize:                 fs.buddy.SpaceSize(),
		UsedSpaceSize:             fs.buddy.UsedSpaceSize(),
		MappedSpaceSize:           fs.buddy.MappedSpaceSize(),
		AllocatedSpaceSize:        fs.buddy.AllocatedSpaceSize(),
		BlockAllocationBitmapSize: len(fs.buddy.BlockAllocationBitmap()),
		DismissedSpaceSize:        fs.pool.DismissedSpaceSize(),
	}
}

func (fs *FileStorage) loadFile() error {
	buffer := [fileHeaderSize]byte{}

	if _, err := fs.spaceMapper.File.ReadAt(buffer[:], 0); err != nil {
		return err
	}

	var fileHeader fileHeader
	fileHeader.Deserialize(buffer[:])
	blockAllocationBitmap := make([]byte, fileHeader.BlockAllocationBitmapSize)

	if _, err := fs.spaceMapper.File.ReadAt(
		blockAllocationBitmap,
		int64(fileHeaderSize+int(fileHeader.UsedSpaceSize)),
	); err != nil {
		return err
	}

	if err := fs.spaceMapper.MapSpace(int(fileHeader.MappedSpaceSize)); err != nil {
		return err
	}

	buddyBuilder := fs.buddy.Build()
	buddyBuilder.SetSpaceSize(int(fileHeader.SpaceSize)).
		SetUsedSpaceSize(int(fileHeader.UsedSpaceSize)).
		SetMappedSpaceSize(int(fileHeader.MappedSpaceSize)).
		SetAllocatedSpaceSize(int(fileHeader.AllocatedSpaceSize)).
		SetBlockAllocationBitmap(blockAllocationBitmap)
	poolBuilder := fs.pool.Build()
	poolBuilder.LoadPooledBlockList(fileHeader.PooledBlockList[:]).
		SetDismissedSpaceSize(int(fileHeader.DismissedSpaceSize))
	fs.primarySpace = fileHeader.PrimarySpace
	return nil
}

func (fs *FileStorage) storeFile() error {
	fs.buddy.ShrinkSpace()

	fileHeader := fileHeader{
		SpaceSize:                 int64(fs.buddy.SpaceSize()),
		UsedSpaceSize:             int64(fs.buddy.UsedSpaceSize()),
		MappedSpaceSize:           int64(fs.buddy.MappedSpaceSize()),
		AllocatedSpaceSize:        int64(fs.buddy.AllocatedSpaceSize()),
		BlockAllocationBitmapSize: int64(len(fs.buddy.BlockAllocationBitmap())),
		DismissedSpaceSize:        int64(fs.pool.DismissedSpaceSize()),
		PrimarySpace:              fs.primarySpace,
	}

	fs.pool.StorePooledBlockList(fileHeader.PooledBlockList[:])

	if err := fs.spaceMapper.Close(); err != nil {
		return err
	}

	buffer := [fileHeaderSize]byte{}
	fileHeader.Serialize(buffer[:])

	if _, err := fs.spaceMapper.File.WriteAt(buffer[:], 0); err != nil {
		return err
	}

	if _, err := fs.spaceMapper.File.WriteAt(
		fs.buddy.BlockAllocationBitmap(),
		int64(fileHeaderSize+fs.buddy.UsedSpaceSize()),
	); err != nil {
		return err
	}

	return nil
}

// Stats represents the stats about file space management.
type Stats struct {
	SpaceSize                 int
	UsedSpaceSize             int
	MappedSpaceSize           int
	AllocatedSpaceSize        int
	BlockAllocationBitmapSize int
	DismissedSpaceSize        int
}

const pageSize = 4096
