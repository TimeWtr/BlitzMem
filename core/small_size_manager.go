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

package core

import (
	"sync/atomic"
	"unsafe"

	"github.com/TimeWtr/TurboAlloc/common"
)

const SmallShardNums = 16

type SmallManager struct {
	// shards is a map of pointers to Shard, representing the individual memory shards managed
	// within the SizeClassUint structure. Each shard contains separate hot and cold paths for
	// memory block allocation and tracking.
	shards map[int][]*SmallSizeShard
	// Size indicates the total amount of memory allocated at the current sizeClass
	// level, in bytes
	size atomic.Uint64
	// counter is an atomic integer that tracks the total number of operations or events processed
	// by the SmallManager.
	counter atomic.Int64
}

func (s *SmallManager) OnSizeClassChange(_ common.SizeCategory, _, _ common.SizeClassDetail) {
}

func newSmallManager(shardCount int) *SmallManager {
	sm := &SmallManager{}

	singleSizeShardCount := shardCount / common.SmallSizeClassNums
	smallClasses := common.SizeClassSizes[common.SmallSizeCategory]
	sm.shards = make(map[int][]*SmallSizeShard, common.SmallSizeClassNums)
	for sizeClass, size := range smallClasses {
		shards := make([]*SmallSizeShard, 0, singleSizeShardCount)
		for i := 0; i < singleSizeShardCount; i++ {
			shards = append(shards, newSmallSizeShard(uint64(size)))
		}
		sm.shards[sizeClass.Int()] = shards
	}

	return sm
}

type SmallSizeShard struct {
	// hotTop is an atomic pointer to a block, representing the top of a hot
	// path in memory management operations.
	hotTop atomic.Pointer[block]
	// hotCount is an atomic counter that tracks the number of active or "hot"
	// memory blocks in a shard's hot path.
	hotCount atomic.Int64
	// coldTop is an atomic pointer to a block, representing the top of a cold
	// path in memory management operations.
	coldTop atomic.Pointer[block]
	// coldCount is an atomic counter that tracks the number of inactive or "cold"
	// memory blocks in a shard's cold path.
	coldCount atomic.Int64

	pages      []*unsafe.Pointer
	pagesCount atomic.Int64
	// blockSize indicates the size of a specific block in a shard, in bytes, such
	// as 8Bytes, 16Bytes
	blockSize uint64
}

func newSmallSizeShard(blockSize uint64) *SmallSizeShard {
	return &SmallSizeShard{
		blockSize: blockSize,
	}
}

type block struct {
	next *block
}
