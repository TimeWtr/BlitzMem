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

package weight

import (
	"context"
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/TimeWtr/TurboAlloc/utils/atomicx"
)

const (
	chunkSize = 256
	minChunks = 2
	maxChunks = 50
)

type (
	DLQImpl struct {
		q             []*QueueChunk
		headChunk     *atomicx.Int32
		tailChunk     *atomicx.Int32
		headIdx       *atomicx.Int32
		tailIdx       *atomicx.Int32
		totalCount    *atomicx.Int32
		resizePending *atomicx.Bool
		state         *atomicx.Bool
		ticker        *time.Ticker
		mu            sync.RWMutex
		closeCh       chan struct{}
	}

	QueueChunk struct {
		q    []unsafe.Pointer
		size int
	}
)

func newQueueChunk(size int) *QueueChunk {
	return &QueueChunk{
		q:    make([]unsafe.Pointer, size),
		size: size,
	}
}

func newDLQImpl() DLQ {
	d := &DLQImpl{
		q:             make([]*QueueChunk, 0, minChunks),
		headChunk:     atomicx.NewInt32(0),
		tailChunk:     atomicx.NewInt32(0),
		headIdx:       atomicx.NewInt32(0),
		tailIdx:       atomicx.NewInt32(0),
		totalCount:    atomicx.NewInt32(0),
		resizePending: atomicx.NewBool(),
		state:         atomicx.NewBool(),
		ticker:        time.NewTicker(time.Millisecond * 10),
		closeCh:       make(chan struct{}),
	}

	for i := 0; i < minChunks; i++ {
		d.q = append(d.q, newQueueChunk(chunkSize))
	}

	return d
}

func (d *DLQImpl) Push(ctx context.Context, event *DLQEvent) error {
	for {
		if !d.state.Load() {
			return errors.New("queue closed")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.resizePending.Load() {
			time.Sleep(time.Millisecond)
			continue
		}

		if d.tailIdx.Load() == int32(d.q[d.tailChunk.Load()].size) {
			if !d.resizePending.CompareAndSwap(false, true) {
				continue
			}

			newChunkIdx := d.tailChunk.Add(1) % int32(len(d.q))
			if newChunkIdx == d.headChunk.Load() {
				// exec expansion operation
				newChunk := newQueueChunk(chunkSize)
				d.q = append(d.q, newChunk)
			} else {
				d.tailChunk.Store(newChunkIdx)
			}

			d.tailIdx.Store(0)
			d.resizePending.SetFalse()
		}

		tail := d.tailIdx.Load()
		if !d.tailIdx.CompareAndSwap(tail, tail+1) {
			continue
		}

		chunk := d.q[d.tailChunk.Load()]
		chunk.q[tail] = unsafe.Pointer(event)
		d.totalCount.Add(1)
		return nil
	}
}

func (d *DLQImpl) Pop(ctx context.Context) (event *DLQEvent, err error) {
	for {
		if !d.state.Load() {
			return &DLQEvent{}, errors.New("queue closed")
		}

		select {
		case <-ctx.Done():
			return &DLQEvent{}, ctx.Err()
		default:
		}

		if len(d.q) == 0 {
			return &DLQEvent{}, nil
		}

		if d.resizePending.Load() {
			time.Sleep(time.Millisecond)
			continue
		}

		if d.headIdx.Load() == int32(d.q[d.headChunk.Load()].size) {
			if !d.resizePending.CompareAndSwap(false, true) {
				continue
			}

			// newChunkIdx := d.headChunk.Add(1) % int32(len(d.q))
			if len(d.q) > minChunks {
				// calculate free chunk counts
				headChunk := d.headChunk.Load()
				tailChunk := d.tailChunk.Load()
				freeChunks := d.calculateFreeChunks(headChunk, tailChunk)
				if freeChunks > len(d.q)/2 {
					d.shrink(headChunk, tailChunk, freeChunks)
				}
			}

			d.headIdx.Store(0)
			d.resizePending.SetFalse()
		}
	}
}

func (d *DLQImpl) calculateFreeChunks(headChunk, tailChunk int32) (freeChunks int) {
	switch {
	case headChunk == tailChunk:
		freeChunks = len(d.q) - 1
	case headChunk < tailChunk:
		freeChunks = int(int32(len(d.q)-1) - tailChunk + headChunk)
	default:
		freeChunks = int(headChunk - tailChunk - 1)
	}

	return freeChunks
}

func (d *DLQImpl) shrink(headChunk, tailChunk int32, freeChunks int) {
	// exec shrink operation
	copyCount := len(d.q) - freeChunks
	newSize := max(minChunks, copyCount)
	newQ := make([]*QueueChunk, newSize)

	if headChunk < tailChunk {
		copy(newQ, d.q[headChunk:headChunk+int32(copyCount)])
	} else {
		firstPart := int32(len(d.q)) - headChunk
		copy(newQ, d.q[headChunk:headChunk+firstPart])

		if firstPart < int32(copyCount) {
			copy(newQ[firstPart:], d.q[:int32(copyCount)-firstPart])
		}
	}

	d.q = newQ
	d.headChunk.Store(0)
	d.tailChunk.Store(int32(copyCount - 1))
}

func (d *DLQImpl) GetSize() (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return int(d.totalCount.Load()), nil
}

func (d *DLQImpl) GetAll(ctx context.Context) ([]*DLQEvent, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.state.Load() {
		return nil, errors.New("queue closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return nil, nil
}

func (d *DLQImpl) Remove(ctx context.Context, _ int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Load() {
		return errors.New("queue closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func (d *DLQImpl) removeAt() {
}

func (d *DLQImpl) Close() {
	if !d.state.CompareAndSwap(true, false) {
		return
	}

	close(d.closeCh)
	if d.ticker != nil {
		d.ticker.Stop()
	}
}
