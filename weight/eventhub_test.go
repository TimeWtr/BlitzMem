package weight

import (
	"testing"
	"time"

	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils/log"
	"go.uber.org/zap"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		category    common.SizeCategory
		bufferSize  []int
		expectedCap int
	}{
		{"default buffer", "tag1", common.SizeCategory(1), nil, 0},
		{"custom buffer", "tag2", common.SizeCategory(2), []int{5}, 5},
	}

	logger := log.NewZapAdapter(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub, _ := newEventHubImpl(logger).(*EventHubImpl)
			ch := hub.Register(tt.tag, tt.category, tt.bufferSize...)
			if cap(ch) != tt.expectedCap {
				t.Errorf("expected channel capacity %d, got %d", tt.expectedCap, cap(ch))
			}
		})
	}
}

func TestUnregister(t *testing.T) {
	tests := []struct {
		name       string
		registered bool
	}{
		{"existing tag", true},
		{"non-existing tag", false},
	}

	logger := log.NewZapAdapter(zap.NewNop())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub, _ := newEventHubImpl(logger).(*EventHubImpl)
			var ch <-chan Event
			if tt.registered {
				ch = hub.Register("tag1", common.SizeCategory(1))
			}
			hub.Unregister("tag1")
			if tt.registered {
				select {
				case _, ok := <-ch:
					if ok {
						t.Error("channel should be closed after unregistering")
					}
				default:
					t.Error("channel should be closed after unregistering")
				}
			}
		})
	}
}

func TestDispatch(t *testing.T) {
	tests := []struct {
		name      string
		listeners []Listener
		event     Event
	}{
		{"matching category", []Listener{{tag: "tag1", category: common.SizeCategory(1), ch: make(chan Event, 1)}}, Event{category: common.SizeCategory(1)}},
		{"all category", []Listener{{tag: "tag2", category: common.AllSizeCategory, ch: make(chan Event, 1)}}, Event{category: common.SizeCategory(2)}},
	}

	logger := log.NewZapAdapter(zap.NewNop())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub, _ := newEventHubImpl(logger).(*EventHubImpl)
			hub.mu.Lock()
			hub.listeners = tt.listeners
			hub.mu.Unlock()

			hub.Dispatch(tt.event)

			for _, listener := range tt.listeners {
				select {
				case ev := <-listener.ch:
					if ev.category != tt.event.category {
						t.Errorf("expected event category %v, got %v", tt.event.category, ev.category)
					}
				case <-time.After(2 * time.Second):
					t.Error("event dispatch timed out")
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	logger := log.NewZapAdapter(zap.NewNop())
	hub, _ := newEventHubImpl(logger).(*EventHubImpl)
	ch := make(chan Event, 1)
	hub.mu.Lock()
	hub.listeners = append(hub.listeners, Listener{tag: "tag1", category: common.SizeCategory(1), ch: ch})
	hub.mu.Unlock()

	hub.Close()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after Close()")
		}
	default:
		t.Error("channel should be closed after Close()")
	}
}
