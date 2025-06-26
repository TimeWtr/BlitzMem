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
	"sync"
	"time"

	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils/log"
)

type EventHub interface {
	Register(tag string, sc common.SizeCategory, bufferSize ...int) <-chan Event
	Unregister(tag string)
	Dispatch(ev Event)
	Close()
}

type Listener struct {
	tag      string
	category common.SizeCategory
	ch       chan Event
}

// EventHubImpl manages event listeners and broadcasts events to them
type EventHubImpl struct {
	listeners []Listener    // Slice of listeners identified by tags
	closeCh   chan struct{} // Channel for closing signal
	once      sync.Once     // Ensures close operation executes only once
	l         log.Logger    // Logger instance
	mu        sync.RWMutex  // RWMutex for concurrent access protection
}

// newEventHubImpl creates a new EventHubImpl instance
func newEventHubImpl(l log.Logger) EventHub {
	return &EventHubImpl{
		listeners: []Listener{},
		l:         l,
		closeCh:   make(chan struct{}),
	}
}

// Register adds an event listener to the dispatcher
// Parameters:
//
//	tag - unique identifier for the listener
//	ch - event channel
func (d *EventHubImpl) Register(tag string, sc common.SizeCategory, bufferSize ...int) <-chan Event {
	d.mu.Lock()
	defer d.mu.Unlock()
	size := 0
	if len(bufferSize) > 0 {
		size = bufferSize[0]
	}

	ch := make(chan Event, size)
	d.listeners = append(d.listeners, Listener{
		tag:      tag,
		category: sc,
		ch:       ch,
	})
	return ch
}

// Unregister removes a listener from the dispatcher
// Parameters:
//
//	tag - unique identifier of the listener to remove
func (d *EventHubImpl) Unregister(tag string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, l := range d.listeners {
		if l.tag == tag {
			close(l.ch)
			d.listeners = append(d.listeners[:i], d.listeners[i+1:]...)
			break
		}
	}
}

// Dispatch broadcasts an event to all registered listeners
// Parameters:
//
//	ev - event to broadcast
func (d *EventHubImpl) Dispatch(ev Event) {
	d.mu.RLock()
	listeners := d.listeners
	d.mu.RUnlock()

	for _, listener := range listeners {
		if listener.category == ev.category || listener.category == common.AllSizeCategory {
			select {
			case listener.ch <- ev:
				// Event successfully sent
			case <-time.After(time.Second):
				d.l.Error("dispatch event error",
					log.StringField("listener", listener.tag),
					log.ErrorField(context.DeadlineExceeded))
			case <-d.closeCh:
				// Received close signal, stop dispatching
				d.l.Info("dispatcher loop receive stop signal")
			}
		}
	}
}

// Close shuts down the dispatcher
func (d *EventHubImpl) Close() {
	d.once.Do(func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		close(d.closeCh)
		for _, listener := range d.listeners {
			close(listener.ch)
		}
	})
}
