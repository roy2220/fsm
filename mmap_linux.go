// +build linux

package fsm

import (
	"os"
	"syscall"
)

func mmap(file *os.File, offset int64, length int) ([]byte, error) {
	buffer, err := syscall.Mmap(
		int(file.Fd()),
		offset,
		length,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)

	if err != nil {
		return nil, err
	}

	if err := syscall.Madvise(buffer, syscall.MADV_RANDOM); err != nil {
		return nil, err
	}

	return buffer, nil
}

func munmap(buffer []byte) error {
	return syscall.Munmap(buffer)
}
