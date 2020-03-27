package fsm

import (
	"os"

	"github.com/roy2220/fsm/internal/spacemapper"
)

type spaceMapper struct {
	File *os.File

	buffer []byte
}

func (sm *spaceMapper) MapSpace(spaceSize int) error {
	if spaceSize >= 1 && spaceSize < pageSize {
		spaceSize = pageSize
	}

	if len(sm.buffer) == spaceSize {
		return nil
	}

	if sm.buffer != nil {
		if err := munmap(sm.buffer); err != nil {
			return err
		}

		sm.buffer = nil
	}

	if err := sm.File.Truncate(int64(fileHeaderSize + spaceSize)); err != nil {
		return err
	}

	if spaceSize >= 1 {
		buffer, err := mmap(sm.File, int64(fileHeaderSize), spaceSize)

		if err != nil {
			return err
		}

		sm.buffer = buffer
	}

	return nil
}

func (sm *spaceMapper) AccessSpace() []byte {
	return sm.buffer
}

func (sm *spaceMapper) Close() error {
	if sm.buffer == nil {
		return nil
	}

	return munmap(sm.buffer)
}

var _ = spacemapper.SpaceMapper(&spaceMapper{})
