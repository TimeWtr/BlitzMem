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

package atomicx

import (
	"sync/atomic"
	"unsafe"
)

type Pointer struct {
	value unsafe.Pointer
}

func NewPointer(value unsafe.Pointer) *Pointer {
	return &Pointer{value: value}
}

func (p *Pointer) Load() unsafe.Pointer {
	return atomic.LoadPointer(&p.value)
}

func (p *Pointer) Store(value unsafe.Pointer) {
	atomic.StorePointer(&p.value, value)
}

func (p *Pointer) Swap(newPtr unsafe.Pointer) unsafe.Pointer {
	return atomic.SwapPointer(&p.value, newPtr)
}

func (p *Pointer) CompareAndSwap(oldPtr, newPtr unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&p.value, oldPtr, newPtr)
}
