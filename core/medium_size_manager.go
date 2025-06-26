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
)

type MediumManager struct {
	shards  []*MediumSizeShard
	counter atomic.Int64
}

func newMediumManager(shards int) *MediumManager {
	return &MediumManager{
		shards: make([]*MediumSizeShard, shards),
	}
}

type MediumSizeShard struct {
	freeList  atomic.Pointer[block]
	freeCount atomic.Uint32

	pages      []unsafe.Pointer
	pagesCount atomic.Uint32
}
