package fsm

import (
	"os"
	"syscall"

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
		if err := syscall.Munmap(sm.buffer); err != nil {
			return err
		}

		sm.buffer = nil
	}

	if err := sm.File.Truncate(int64(fileHeaderSize + spaceSize)); err != nil {
		return err
	}

	if spaceSize >= 1 {
		buffer, err := syscall.Mmap(
			int(sm.File.Fd()),
			int64(fileHeaderSize),
			spaceSize,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED,
		)

		if err != nil {
			return err
		}

		sm.buffer = buffer

		if err := syscall.Madvise(sm.buffer, syscall.MADV_RANDOM); err != nil {
			return err
		}
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

	return syscall.Munmap(sm.buffer)
}

var _ = spacemapper.SpaceMapper(&spaceMapper{})
