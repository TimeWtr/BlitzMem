package weight

import (
	"context"
	"errors"
	"sync"
	"unsafe"

	"github.com/TimeWtr/TurboAlloc/utils/atomicx"
)

const (
	minCapacity  = 32
	maxCapacity  = 1 << 20
	shrinkFactor = 4
	growFactor   = 2
)

var (
	ErrQueueClosed = errors.New("queue closed")
	ErrBufFull     = errors.New("buf full")
	ErrBufClosed   = errors.New("buf closed")
)

type (
	DLQOss struct {
		buf     *atomicx.Pointer
		m       *sync.Map
		mu      sync.Mutex
		closeCh chan struct{}
		eventID *atomicx.Int64
		once    sync.Once
	}

	RingBuffer struct {
		head     *atomicx.Int32
		tail     *atomicx.Int32
		capacity *atomicx.Int32
		count    *atomicx.Int32
		buf      []unsafe.Pointer
		state    *atomicx.Bool
	}
)

func newRingBuffer(capacity int32) *RingBuffer {
	ringBuf := &RingBuffer{
		head:     atomicx.NewInt32(0),
		tail:     atomicx.NewInt32(0),
		capacity: atomicx.NewInt32(capacity),
		count:    atomicx.NewInt32(0),
		buf:      make([]unsafe.Pointer, capacity),
		state:    atomicx.NewBool(),
	}
	ringBuf.state.SetTrue()

	return ringBuf
}

func (r *RingBuffer) push(event *DLQEvent) (int32, error) {
	capacity := r.capacity.Load()
	if r.count.Load() >= capacity {
		return 0, ErrBufFull
	}

	for {
		if !r.state.Load() {
			return 0, ErrBufClosed
		}

		tail := r.tail.Load()
		if tail >= capacity {
			return 0, ErrBufFull
		}

		if !r.tail.CompareAndSwap(tail, tail+1) {
			continue
		}

		r.buf[tail] = unsafe.Pointer(event)
		r.count.Add(1)
		return tail, nil
	}
}

func (r *RingBuffer) pop() (*DLQEvent, error) {
	if r.count.Load() == 0 {
		return &DLQEvent{}, nil
	}

	for {
		if !r.state.Load() {
			return &DLQEvent{}, ErrBufClosed
		}
	}
}

func newDLQOss(capacity int32) *DLQOss {
	ringBuf := newRingBuffer(capacity)
	dlq := &DLQOss{
		buf:     atomicx.NewPointer(unsafe.Pointer(&ringBuf)),
		m:       &sync.Map{},
		mu:      sync.Mutex{},
		closeCh: make(chan struct{}),
		eventID: atomicx.NewInt64(0),
	}

	return dlq
}

func (d *DLQOss) Push(ctx context.Context, event *DLQEvent) error {
	for {
		select {
		case <-d.closeCh:
			return ErrQueueClosed
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		event.ID = d.eventID.Add(1)
		buf := (*RingBuffer)(d.buf.Load())
		if buf.count.Load() >= buf.capacity.Load() {
			tail := buf.tail.Load()
			capacity := buf.capacity.Load()
			newTailIdx := (tail + 1) % (capacity - 1)
			if newTailIdx == buf.head.Load() {
				// execute expansion
				d.expansion()
			} else {
				// circular wraparound
				buf.tail.Store(newTailIdx)
			}
		}

		pos, err := (*RingBuffer)(d.buf.Load()).push(event)
		if err != nil {
			return err
		}
		d.m.Store(event.ID, pos)

		return nil
	}
}

func (d *DLQOss) Pop(ctx context.Context) (event *DLQEvent, err error) {
	for {
		select {
		case <-d.closeCh:
			return nil, ErrQueueClosed
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		buf := (*RingBuffer)(d.buf.Load())
		if buf.count.Load() == 0 {
			return &DLQEvent{}, nil
		}

		head := buf.head.Load()
		eventPtr := buf.buf[head]
		if eventPtr == nil {
			return &DLQEvent{}, nil
		}

		return (*DLQEvent)(eventPtr), nil
	}
}

func (d *DLQOss) expansion() {
	d.mu.Lock()
	defer d.mu.Unlock()

	buf := (*RingBuffer)(d.buf.Load())
	tail := buf.tail.Load()
	head := buf.head.Load()
	copySlots := buf.count.Load()
	newCapacity := min(maxCapacity, buf.capacity.Load()*growFactor)
	newQ := make([]unsafe.Pointer, newCapacity)

	switch {
	case head <= tail:
		copy(newQ[:copySlots], buf.buf[head:head+copySlots])
	default:
		firstSlots := buf.capacity.Load() - 1 - head
		copy(newQ[:firstSlots], buf.buf[head:])
		if firstSlots < copySlots {
			copy(newQ[firstSlots:], buf.buf[:tail])
		}
	}

	buf.count.Store(copySlots)
	buf.head.Store(0)
	buf.tail.Store(int32(len(newQ) - 1))
	d.buf.Store(unsafe.Pointer(buf))
}

func (d *DLQOss) GetSize() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return int((*RingBuffer)(d.buf.Load()).count.Load())
}

func (d *DLQOss) GetAll(_ context.Context) ([]*DLQEvent, error) {
	return nil, nil
}

func (d *DLQOss) Remove(_ context.Context, _ int64) error {
	return nil
}

func (d *DLQOss) Close() {
	d.once.Do(func() {
		close(d.closeCh)
		d.mu.Lock()
		defer d.mu.Unlock()

		buf := (*RingBuffer)(d.buf.Load())
		if buf.count.Load() > 0 {
			for i := range buf.buf {
				ptr := buf.buf[i]
				if ptr != nil {
					buf.buf[i] = nil
				}
			}
		}
		d.buf = nil
	})
}
