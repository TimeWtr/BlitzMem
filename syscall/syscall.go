// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syscall

import (
	"fmt"
	"syscall"
	"unsafe"
)

// pageSize defines the fixed size of a memory page in bytes,
// typically used for memory allocation and management.
const pageSize = 4096

func allocPages(numPages int) (unsafe.Pointer, error) {
	if numPages <= 0 {
		return nil, fmt.Errorf("invalid number of pages: %d", numPages)
	}

	allocSize := numPages * pageSize
	memPtr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(allocSize),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
		^uintptr(0),
		0)
	if errno != 0 {
		return nil, fmt.Errorf("failed to alloc pages, errno: %w", errno)
	}

	if memPtr%uintptr(syscall.Getpagesize()) != 0 {
		return nil, fmt.Errorf("memory not page-aligned: %x", memPtr)
	}

	return unsafe.Pointer(memPtr), nil
}

func freePages(ptr unsafe.Pointer, pageSize int) error {
	if ptr == nil {
		return fmt.Errorf("invalid pointer")
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_MUNMAP,
		uintptr(ptr),
		uintptr(pageSize),
		0)
	if errno != 0 {
		return fmt.Errorf("failed to free pages, errno: %w", errno)
	}

	return nil
}
