// +build windows

package fsm

import (
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

func mmap(file *os.File, offset int64, length int) ([]byte, error) {
	fileMappingHandle, err := syscall.CreateFileMapping(
		syscall.Handle(file.Fd()),
		nil,
		syscall.PAGE_READWRITE,
		uint32(length>>32),
		uint32(length),
		nil,
	)

	if err != nil {
		return nil, err
	}

	defer syscall.CloseHandle(fileMappingHandle)

	bufferPtr, err := syscall.MapViewOfFile(
		fileMappingHandle,
		syscall.FILE_MAP_READ|syscall.FILE_MAP_WRITE,
		uint32(offset>>32),
		uint32(offset),
		uintptr(length),
	)

	buffer := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{bufferPtr, length, length}))
	return buffer, nil
}

func munmap(buffer []byte) error {
	bufferPtr := (*reflect.SliceHeader)(unsafe.Pointer(&buffer)).Data
	return syscall.UnmapViewOfFile(bufferPtr)
}
