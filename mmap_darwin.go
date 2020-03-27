// +build darwin

package fsm

import (
	"os"
	"syscall"
)

func mmap(file *os.File, offset int64, length int) ([]byte, error) {
	return syscall.Mmap(
		int(file.Fd()),
		offset,
		length,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
}

func munmap(buffer []byte) error {
	return syscall.Munmap(buffer)
}
