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
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	// L1Sizes Each large size class level in the L1 cache contains 128 memory entries
	L1Sizes = 128
)

type systemCache struct {
	l1 *L1Cache
	l2 *L2Cache
	l3 *L3Cache
}

func newSystemCache() *systemCache {
	return &systemCache{
		l1: &L1Cache{},
		l2: &L2Cache{
			slots: [1024]cacheEntry{},
		},
		l3: &L3Cache{
			slots: make(map[SizeClass][]*Slab),
		},
	}
}

func (c *systemCache) tryAlloc(size int) unsafe.Pointer {
	sc := getClassSize(size)
	if ptr := c.l1.tryAlloc(sc); ptr != nil {
		return ptr
	}

	if ptr := c.l2.tryAlloc(size); ptr != nil {
		return ptr
	}

	if ptr := c.l3.tryAlloc(size); ptr != nil {
		return ptr
	}

	return nil
}

func (c *systemCache) tryPut(sc SizeClass, ptr unsafe.Pointer) bool {
	if c.l1.tryPut(sc, ptr) {
		return true
	}

	if c.l2.tryPut(sc, ptr) {
		return true
	}

	return c.l3.tryPut(sc, ptr)
}

type cacheEntry struct {
	ptr       unsafe.Pointer
	sizeClass SizeClass
}

// L1Cache Each physical CPU has its own cache, which stores 32 pointers to the location
// of available blocks in memory. This is not a direct operation of the CPU L1 cache, but a
// benchmark L1 cache.
type L1Cache struct {
	// 19 * 128 slots, each storing a pointer to a block
	slots [numSizeClasses][L1Sizes]unsafe.Pointer
	// counters
	counters [numSizeClasses]uint32
	// aligned memory
	_pad [cacheLineSize -
		(unsafe.Sizeof([numSizeClasses]unsafe.Pointer{})+
			unsafe.Sizeof([numSizeClasses]uint32{}))%cacheLineSize]byte
}

func (l *L1Cache) tryPut(sc SizeClass, ptr unsafe.Pointer) bool {
	idx := sizeClassSizes[sc]
	if atomic.LoadUint32(&l.counters[idx]) >= L1Sizes {
		return false
	}

	pos := atomic.LoadUint32(&l.counters[idx])
	if !atomic.CompareAndSwapUint32(&l.counters[idx], pos, pos+1) {
		return false
	}

	return true
}

func (l *L1Cache) tryAlloc(sc SizeClass) unsafe.Pointer {
	idx := sizeClassSizes[sc]
	if atomic.LoadUint32(&l.counters[idx]) == 0 {
		return nil
	}

	//oldCounter := l.counters[idx]

	return nil
}

// L2Cache Each physical CPU has its own cache, which stores 256 pointers to the location
// of available blocks in memory. This is not a direct operation of the CPU L2 cache, but a
// benchmark L2 cache.
type L2Cache struct {
	// 256 slots, each storing a pointer to a block
	slots [1024]cacheEntry
	// The number of pointers stored
	size int
	// aligned memory
	_pad [cacheLineSize - (unsafe.Sizeof([256]unsafe.Pointer{})+4)%cacheLineSize]byte
	// the lock
	mu sync.Mutex
}

func (l *L2Cache) tryPut(sc SizeClass, ptr unsafe.Pointer) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.size >= len(l.slots) {
		return false
	}

	l.slots[l.size] = cacheEntry{
		ptr:       ptr,
		sizeClass: sc,
	}
	l.size++

	return true
}

func (l *L2Cache) tryAlloc(size int) unsafe.Pointer {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.size <= 0 {
		return nil
	}

	for i := l.size - 1; i >= 0; i-- {
		if sizeClassSizes[l.slots[i].sizeClass] >= size {
			l.size--
			l.slots[i] = cacheEntry{}
			return l.slots[i].ptr
		}
	}

	return nil
}

type L3Cache struct {
	slots map[SizeClass][]*Slab
	size  int
	mu    sync.Mutex
}

func (l *L3Cache) tryPut(sc SizeClass, ptr unsafe.Pointer) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return false
}

func (l *L3Cache) tryAlloc(size int) unsafe.Pointer {
	l.mu.Lock()
	defer l.mu.Unlock()

	classSize := getClassSize(size)
	entries := l.slots[classSize]
	if entries == nil {
		return nil
	}

	for i := 0; i < len(entries); i++ {
		if entries[i].freeCount > 0 {
			l.size--
			return entries[i].allocBlock()
		}
	}

	return nil
}
