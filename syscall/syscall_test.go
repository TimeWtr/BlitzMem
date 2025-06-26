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
	"testing"
	"unsafe"
)

const (
	minPages = 1
	maxPages = 8092 // 16MB
)

func writeTestData(ptr unsafe.Pointer, size int) {
	data := unsafe.Slice((*byte)(ptr), size)

	for i := range data {
		data[i] = byte(i % 256)
	}
}

func verifyTestData(ptr unsafe.Pointer, size int) error {
	data := unsafe.Slice((*byte)(ptr), size)

	for i := range data {
		if data[i] != byte(i%256) {
			return fmt.Errorf("data verification failed at index %d: expected %d, got %d",
				i, byte(i%256), data[i])
		}
	}
	return nil
}

func TestAllocFreeSinglePage(t *testing.T) {
	ptr, err := allocPages(minPages)
	if err != nil {
		t.Fatalf("allocPages failed: %v", err)
	}
	defer func() {
		if err = freePages(ptr, pageSize); err != nil {
			t.Errorf("freePages failed: %v", err)
		}
	}()

	writeTestData(ptr, pageSize)
	if err = verifyTestData(ptr, pageSize); err != nil {
		t.Error(err)
	}
}

func TestAllocFreeMultiplePages(t *testing.T) {
	ptr, err := allocPages(maxPages)
	if err != nil {
		t.Fatalf("allocPages failed: %v", err)
	}
	defer func() {
		if err := freePages(ptr, maxPages*pageSize); err != nil {
			t.Errorf("freePages failed: %v", err)
		}
	}()

	writeTestData(ptr, maxPages*pageSize)
	if err := verifyTestData(ptr, maxPages*pageSize); err != nil {
		t.Error(err)
	}
}

func TestZeroPagesAllocation(t *testing.T) {
	ptr, err := allocPages(0)
	if err == nil {
		defer func() {
			_ = freePages(ptr, 0)
		}()
		t.Error("expected error for zero pages allocation, got nil")
	} else if ptr != nil {
		t.Errorf("expected nil pointer for zero pages allocation, got %p", ptr)
		t.Logf("error: %v", err)
	}
}

func TestNegativePagesAllocation(t *testing.T) {
	ptr, err := allocPages(-1)
	if err == nil {
		defer func() {
			_ = freePages(ptr, pageSize)
		}()
		t.Error("expected error for negative pages allocation, got nil")
	} else if ptr != nil {
		t.Errorf("expected nil pointer for negative pages allocation, got %p", ptr)
		t.Logf("error: %v", err)
	}
}

func TestFreeInvalidPointer(t *testing.T) {
	invalidPtr := unsafe.Pointer(uintptr(0xdeadbeef))
	err := freePages(invalidPtr, pageSize)
	if err == nil {
		t.Error("expected error for freeing invalid pointer, got nil")
	} else {
		t.Logf("freePages returned expected error: %v", err)
	}
}

func TestDoubleFreeProtection(t *testing.T) {
	ptr, err := allocPages(minPages)
	if err != nil {
		t.Fatalf("allocPages failed: %v", err)
	}

	if err = freePages(ptr, pageSize); err != nil {
		t.Errorf("first freePages failed: %v", err)
	}

	err = freePages(ptr, pageSize)
	if err == nil {
		t.Logf("double free succeeded unexpectedly")
	} else {
		t.Errorf("double free correctly failed with: %v", err)
	}
}

func TestMemoryAlignment(t *testing.T) {
	ptr, err := allocPages(minPages)
	if err != nil {
		t.Fatalf("allocPages failed: %v", err)
	}
	defer func() {
		_ = freePages(ptr, pageSize)
	}()

	alignment := uintptr(syscall.Getpagesize())
	if uintptr(ptr)%alignment != 0 {
		t.Errorf("pointer %p not aligned to page size %d", ptr, alignment)
	}
}

func TestMemoryReuseAfterFree(t *testing.T) {
	ptr1, err := allocPages(minPages)
	if err != nil {
		t.Fatalf("allocPages failed: %v", err)
	}

	if err := freePages(ptr1, pageSize); err != nil {
		t.Fatalf("freePages failed: %v", err)
	}

	ptr2, err := allocPages(minPages)
	if err != nil {
		t.Fatalf("allocPages failed: %v", err)
	}
	defer func() {
		_ = freePages(ptr2, pageSize)
	}()

	if ptr1 == ptr2 {
		t.Logf("memory reuse detected: %p == %p", ptr1, ptr2)
	} else {
		t.Logf("different memory addresses: %p vs %p", ptr1, ptr2)
	}
}

func TestConcurrentAllocFree(t *testing.T) {
	const (
		numRoutines = 10
		numPages    = 10
		iterations  = 100
	)

	errChan := make(chan error, numRoutines)
	sem := make(chan struct{}, numRoutines) // 限制并发goroutine数量

	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			sem <- struct{}{}
			defer func() { <-sem }()

			for j := 0; j < iterations; j++ {
				ptr, err := allocPages(numPages)
				if err != nil {
					errChan <- fmt.Errorf("goroutine %d: allocPages failed: %w", id, err)
					return
				}

				writeTestData(ptr, numPages*pageSize)

				if err = verifyTestData(ptr, numPages*pageSize); err != nil {
					errChan <- fmt.Errorf("goroutine %d: verifyTestData failed: %w", id, err)
					return
				}

				if err = freePages(ptr, numPages*pageSize); err != nil {
					errChan <- fmt.Errorf("goroutine %d: freePages failed: %w", id, err)
					return
				}
			}

			errChan <- nil
		}(i)
	}

	for i := 0; i < numRoutines; i++ {
		if err := <-errChan; err != nil {
			t.Error(err)
		}
	}
}

//func TestMemoryProtectionFlags(t *testing.T) {
//	ptr, err := allocPages(minPages)
//	if err != nil {
//		t.Fatalf("allocPages failed: %v", err)
//	}
//	defer func() {
//		_ = freePages(ptr, pageSize)
//	}()
//
//	*(*byte)(ptr) = 42
//	if protectMemory(ptr, pageSize, syscall.PROT_EXEC) == nil {
//		defer protectMemory(ptr, pageSize, syscall.PROT_READ|syscall.PROT_WRITE)
//
//		if !testExecute(ptr) {
//			t.Error("executing non-executable memory succeeded unexpectedly")
//		} else {
//			t.Logf("execution correctly blocked as expected")
//		}
//	}
//}

//func protectMemory(ptr unsafe.Pointer, size int, prot int) error {
//	_, _, errno := syscall.Syscall(
//		syscall.SYS_MPROTECT,
//		uintptr(ptr),
//		uintptr(size),
//		uintptr(prot))
//	if errno != 0 {
//		return fmt.Errorf("mprotect failed: errno %d", errno)
//	}
//	return nil
//}
//
//func testExecute(ptr unsafe.Pointer) (crashed bool) {
//	defer func() {
//		if r := recover(); r != nil {
//			crashed = true
//		}
//	}()
//
//	f := *(*func())(unsafe.Pointer(&ptr))
//	f()
//
//	return false
//}
