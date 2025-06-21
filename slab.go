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

package slab

import (
	"container/list"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	cacheLineSize = 128 // go official used
	numClasses    = 6
	maxEmptySlabs = 5
	reclaimPeriod = time.Minute * 5
	idleThreshold = 10 * time.Minute
)

type Allocator struct {
	conf           *Config
	sizeClassSlabs map[SizeClass][]*Slab
	state          struct {
		Allocs     atomic.Uint64
		Frees      atomic.Uint64
		L1Hits     atomic.Uint64
		L2Hits     atomic.Uint64
		L3Hits     atomic.Uint64
		NewSlabs   atomic.Uint64
		CacheFlush atomic.Uint64
		Prefetch   atomic.Uint64
	}
}

type Slab struct {
	// The starting position of the current slab in memory
	start unsafe.Pointer
	// the size of the current slab
	size int
	// The size of the small block in the slab
	blockSize int
	// Doubly linked list of free blocks
	freeList *list.List
	// The number of free blocks in the doubly linked list
	freeCount int
	// The total number of blocks
	totalSize int
}

func (s *Slab) allocBlock() unsafe.Pointer {
	return nil
}

func alloc(size int) unsafe.Pointer {
	return nil
}

func AllocBytes(size int) []byte {
	ptr := alloc(size)
	if ptr == nil {
		return nil
	}

	rawPtr := (*byte)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocInts(count int) []int {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(0))
	rawPtr := (*int)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocInt32s(count int) []int32 {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(int32(0)))
	rawPtr := (*int32)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocInt64s(count int) []int64 {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(int64(0)))
	rawPtr := (*int64)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocUint32s(count int) []uint32 {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(uint32(0)))
	rawPtr := (*uint32)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocUint64s(count int) []uint64 {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(uint64(0)))
	rawPtr := (*uint64)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocFloat32s(count int) []float32 {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(float32(0)))
	rawPtr := (*float32)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func AllocFloat64s(count int) []float64 {
	ptr := alloc(count)
	if ptr == nil {
		return nil
	}

	size := count * int(unsafe.Sizeof(float64(0)))
	rawPtr := (*float64)(ptr)
	return unsafe.Slice(rawPtr, size)
}

func getClassSize(size int) SizeClass {
	if size < SizeClass8B.int() {
		return 0
	}

	for sc := SizeClass8B; sc < SizeClassMax; sc++ {
		if size <= sizeClassSizes[sc] {
			return sc
		}
	}

	return SizeClassMax
}
