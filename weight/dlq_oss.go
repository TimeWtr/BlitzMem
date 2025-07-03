package weight

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"github.com/TimeWtr/TurboAlloc/utils/atomicx"
)

const (
	minCapacity  = 32
	maxCapacity  = 1 << 20
	shrinkFactor = 4
	shrinkStep   = 2
	growFactor   = 2
)

var (
	ErrQueueClosed = errors.New("queue closed")
	ErrBufFull     = errors.New("buf full")
	ErrBufClosed   = errors.New("buf closed")
)

type (
	DLQOss struct {
		buf        *atomicx.Pointer
		m          *sync.Map
		mu         sync.Mutex
		closeCh    chan struct{}
		eventID    *atomicx.Int64
		once       sync.Once
		popCounter *atomicx.Int64
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
	var (
		capacity     int32
		currentCount int32
		pos          int32
	)

	for {
		if !r.state.Load() {
			return 0, ErrBufClosed
		}

		capacity = r.capacity.Load()
		currentCount = r.count.Load()
		if currentCount >= capacity {
			return 0, ErrBufFull
		}

		// Reserve a slot
		if !r.count.CompareAndSwap(currentCount, currentCount+1) {
			continue
		}

		pos = (r.tail.Add(1) - 1) % capacity
		r.buf[pos] = unsafe.Pointer(event)
		break
	}

	return pos, nil
}

func (r *RingBuffer) pop(ctx context.Context) (*DLQEvent, error) {
	for {
		if !r.state.Load() {
			return nil, ErrQueueClosed
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if r.count.Load() <= 0 {
			return &DLQEvent{}, nil
		}

		// Reserve a slot
		currentCount := r.count.Load()
		if !r.count.CompareAndSwap(currentCount, currentCount-1) {
			continue
		}

		head := r.head.Add(1) - 1
		pos := head % r.capacity.Load()
		event := (*DLQEvent)(r.buf[pos])
		r.buf[pos] = nil
		return event, nil
	}
}

func (r *RingBuffer) close() {
	if !r.state.CompareAndSwap(false, true) {
		return
	}
}

func newDLQOss(capacity int32) *DLQOss {
	if capacity < minCapacity {
		capacity = minCapacity
	}

	ringBuf := newRingBuffer(capacity)
	dlq := &DLQOss{
		buf:        atomicx.NewPointer(unsafe.Pointer(ringBuf)),
		m:          &sync.Map{},
		closeCh:    make(chan struct{}),
		eventID:    atomicx.NewInt64(0),
		popCounter: atomicx.NewInt64(0),
	}

	return dlq
}

func (d *DLQOss) Push(ctx context.Context, event *DLQEvent) error {
	select {
	case <-d.closeCh:
		return ErrQueueClosed
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	event.ID = d.eventID.Add(1)

	for {
		buf := (*RingBuffer)(d.buf.Load())
		pos, err := buf.push(event)
		switch {
		case err == nil:
			d.m.Store(event.ID, pos)
			return nil
		case errors.Is(err, ErrBufFull):
			d.expansion()

			select {
			case <-d.closeCh:
				return ErrQueueClosed
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Avoid busy waiting
				runtime.Gosched()
			}
		default:
			return err
		}
	}
}

func (d *DLQOss) Pop(ctx context.Context) (event *DLQEvent, err error) {
	const batch = 10
	select {
	case <-d.closeCh:
		return nil, ErrQueueClosed
	case <-ctx.Done():
		return &DLQEvent{}, ctx.Err()
	default:
	}

	if d.popCounter.Add(1)%batch == 0 {
		if d.shouldShrink() {
			d.shrink()
		}
	}

	buf := (*RingBuffer)(d.buf.Load())
	event, err = buf.pop(ctx)
	switch {
	case err == nil:
		if event != nil {
			d.m.Delete(event.ID)
		}
		return event, nil
	default:
		return nil, err
	}
}

func (d *DLQOss) shouldShrink() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	buf := (*RingBuffer)(d.buf.Load())
	count := buf.count.Load()
	if count < buf.capacity.Load()/shrinkFactor && count > minCapacity {
		return true
	}
	return false
}

func (d *DLQOss) shrink() {
	d.mu.Lock()
	defer d.mu.Unlock()

	buf := (*RingBuffer)(d.buf.Load())
	tail := buf.tail.Load()
	head := buf.head.Load()
	capacity := buf.capacity.Load()
	newCapacity := max(minCapacity, capacity/shrinkStep)
	newQ := make([]unsafe.Pointer, newCapacity)
	switch {
	case tail >= head:
		copy(newQ, buf.buf[head:tail])
	default:
		freeSlots := capacity - buf.count.Load()
		firstSlots := capacity - 1 - head
		copy(newQ, buf.buf[head:])
		if firstSlots < freeSlots {
			copy(newQ[firstSlots:], buf.buf[:tail])
		}
	}
	buf.head.Store(0)
	buf.tail.Store(int32(len(newQ) - 1))
	d.buf.Store(unsafe.Pointer(buf))

	// adjust the relationship between event and position in the buffer
	d.adjustMap(buf)
}

// expansion is responsible for expanding the ring buffer used by the DLQOss.
// This function ensures that the buffer can accommodate more data by reallocating
// the internal array with a larger capacity, while preserving the existing data and
// updating the relevant pointers and counters.
func (d *DLQOss) expansion() {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Load the current ring buffer and its tail, head, and count information.
	buf := (*RingBuffer)(d.buf.Load())
	// double check
	count := buf.count.Load()
	capacity := buf.capacity.Load()
	if count < capacity {
		return
	}

	tail := buf.tail.Load()
	head := buf.head.Load()
	copySlots := buf.count.Load()

	// Calculate the new capacity, ensuring it does not exceed the maximum capacity.
	newCapacity := min(maxCapacity, copySlots*growFactor)
	newQ := make([]unsafe.Pointer, newCapacity)
	switch {
	case head <= tail:
		// If head is less than or equal to tail, it means the data is continuous,
		// and a direct copy can be performed.
		copy(newQ[:copySlots], buf.buf[head:head+copySlots])
	default:
		// If head is greater than tail, it means the data is split into two segments.
		firstSlots := buf.capacity.Load() - 1 - head
		copy(newQ[:firstSlots], buf.buf[head:])
		if firstSlots < copySlots {
			// If the first segment is not enough, copy the second segment of data.
			copy(newQ[firstSlots:copySlots], buf.buf[:tail])
		}
	}

	buf.count.Store(copySlots)
	buf.head.Store(0)
	buf.tail.Store(copySlots - 1)
	d.buf.Store(unsafe.Pointer(buf))

	// adjust the relationship between event and position in the buffer
	d.adjustMap(buf)
}

// adjustMap updates the mapping from event IDs to their positions in the buffer.
// This is necessary after resizing the ring buffer to ensure that the stored positions
// accurately reflect the new layout of events in the buffer.
//
// Parameters:
//   - buf: The RingBuffer whose buffer's event mappings need adjustment.
func (d *DLQOss) adjustMap(buf *RingBuffer) {
	for i := range buf.buf {
		ptr := (*DLQEvent)(buf.buf[i])
		if ptr != nil {
			d.m.Store(ptr.ID, i)
		}
	}
}

// GetSize returns the current number of events stored in the ring buffer.
// This method acquires a lock to ensure thread-safe access to the buffer's count.
//
// Returns:
//   - int: The current size of the ring buffer, representing the number of stored events.
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

// Close gracefully shuts down the DLQOss instance, ensuring it is executed only once.
// It closes the closeCh channel to notify all waiting goroutines, then acquires the mutex lock
// to perform cleanup operations on the internal buffer. If there are elements in the buffer,
// it clears each non-nil pointer to help garbage collection and sets the buffer reference to nil.
func (d *DLQOss) Close() {
	d.once.Do(func() {
		close(d.closeCh)
		d.mu.Lock()
		defer d.mu.Unlock()

		buf := (*RingBuffer)(d.buf.Load())
		buf.close()
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
