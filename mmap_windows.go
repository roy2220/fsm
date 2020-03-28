// +build windows

package fsm

import (
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

var allocationGranularity = func() int {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getSystemInfo := kernel32.NewProc("GetSystemInfo")

	systemInfo := struct {
		wProcessorArchitecture      uint16
		wReserved                   uint16
		dwPageSize                  uint32
		lpMinimumApplicationAddress uintptr
		lpMaximumApplicationAddress uintptr
		dwActiveProcessorMask       uintptr
		dwNumberOfProcessors        uint32
		dwProcessorType             uint32
		dwAllocationGranularity     uint32
		wProcessorLevel             uint16
		wProcessorRevision          uint16
	}{}

	syscall.Syscall(getSystemInfo.Addr(), 1, uintptr(unsafe.Pointer(&systemInfo)), 0, 0)
	return int(systemInfo.dwAllocationGranularity)
}()

func mmap(file *os.File, offset int64, length int) ([]byte, error) {
	fileMappingHandle, err := syscall.CreateFileMapping(
		syscall.Handle(file.Fd()),
		nil,
		syscall.PAGE_READWRITE,
		0,
		0,
		nil,
	)

	if err != nil {
		return nil, err
	}

	defer syscall.CloseHandle(fileMappingHandle)
	var i int

	if alignedOffset := offset &^ int64(allocationGranularity-1); offset > alignedOffset {
		i = int(offset - alignedOffset)
		offset = alignedOffset
		length += i
	} else {
		i = 0
	}

	bufferPtr, err := syscall.MapViewOfFile(
		fileMappingHandle,
		syscall.FILE_MAP_READ|syscall.FILE_MAP_WRITE,
		uint32(offset>>32),
		uint32(offset),
		uintptr(length),
	)

	if err != nil {
		return nil, err
	}

	buffer := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{bufferPtr + uintptr(i), length - i, length - i}))
	return buffer, nil
}

func munmap(buffer []byte) error {
	bufferPtr := (*reflect.SliceHeader)(unsafe.Pointer(&buffer)).Data &^ uintptr(allocationGranularity-1)
	return syscall.UnmapViewOfFile(bufferPtr)
}
