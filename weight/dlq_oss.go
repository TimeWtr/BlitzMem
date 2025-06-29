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

	"github.com/TimeWtr/TurboAlloc/utils/atomicx"
)

type DeadQueueImpl struct {
	q        []*DLQEvent
	head     int
	tail     int
	count    int
	capacity int
	mu       sync.RWMutex
	state    *atomicx.Bool
	ticker   *time.Ticker
	closeCh  chan struct{}
}

func newDeadQueueImpl() DLQ {
	return &DeadQueueImpl{
		q: make([]*DLQEvent, deadQueueSize),
		// current read index
		head: 0,
		// current write index
		tail:     0,
		count:    0,
		capacity: deadQueueSize,
		state:    atomicx.NewBool(),
		ticker:   time.NewTicker(time.Second * 10),
		closeCh:  make(chan struct{}),
	}
}

func (d *DeadQueueImpl) Push(ctx context.Context, event DLQEvent) error {
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

	if d.count == d.capacity {
		if d.count >= maxQueueSize {
			return errors.New("queue is full")
		}

		// execute resize operation
		newCapacity := min(d.capacity*2, maxQueueSize)
		d.resize(newCapacity)
	}

	d.q[d.tail] = &event
	d.tail++
	return nil
}

func (d *DeadQueueImpl) Pop(ctx context.Context) (event DLQEvent, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.state.Load() {
		return DLQEvent{}, errors.New("queue closed")
	}

	select {
	case <-ctx.Done():
		return DLQEvent{}, ctx.Err()
	default:

	}

	return DLQEvent{}, nil
}

func (d *DeadQueueImpl) GetSize() (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.q), nil
}

func (d *DeadQueueImpl) GetAll(ctx context.Context) ([]DLQEvent, error) {
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

func (d *DeadQueueImpl) Remove(ctx context.Context, _ int64) error {
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

func (d *DeadQueueImpl) resize(capacity int) {
	newBuffer := make([]*DLQEvent, capacity)
	if d.head < d.tail {
		copy(newBuffer, d.q[d.head:d.tail])
	} else {
	}
}

func (d *DeadQueueImpl) Close() {
	if !d.state.CompareAndSwap(true, false) {
		return
	}

	close(d.closeCh)
	if d.ticker != nil {
		d.ticker.Stop()
	}
}
