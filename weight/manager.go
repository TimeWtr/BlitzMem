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
	"sync"
	"sync/atomic"
	"time"

	"github.com/TimeWtr/TurboAlloc/common"
	"github.com/TimeWtr/TurboAlloc/utils/log"
)

type (
	Manager interface {
		// Register registers a new listener with the dispatcher.
		// It returns a channel that receives events for the specified size category.
		Register(tag string, sc common.SizeCategory, bufferSize ...int) <-chan Event

		// Unregister removes a listener from the dispatcher.
		Unregister(tag string)

		// Close shuts down the manager and releases all resources.
		Close()
	}

	ManagerImpl struct {
		// Configuration provider for watching config file changes
		provider Provider
		// Processor for normalizing and transforming config data
		processor Processor
		// Cache of normalized global weights configuration
		global common.GlobalConfig
		// Cache of normalized size class weights by category
		sizeClass map[common.SizeCategory][]float64
		// Mutex to protect concurrent access to configuration data
		mu sync.RWMutex
		// Event hub for dispatching config change events
		eventHub EventHub
		// Channel receiving config updates from provider
		watcher <-chan common.Config
		// Channel to signal manager shutdown
		closeCh chan struct{}
		// Logger instance for recording operational logs
		l log.Logger
		// State indicator (Running/Stopped)
		state atomic.Int32
		// WaitGroup to track background goroutines
		wg sync.WaitGroup
	}
)

// NewManager creates a new Manager instance with the provided dependencies.
//
// Parameters:
//   - provider: Configuration provider for watching config file changes
//   - processor: Processor for normalizing and transforming config data
//   - eventHub: Event hub for dispatching config change events
//   - l: Logger instance for recording operational logs
//
// Returns:
//   - Manager: Initialized manager instance
//   - error: Error if initialization fails
func NewManager(provider Provider, processor Processor, eventHub EventHub, l log.Logger) (Manager, error) {
	// Initialize config watcher to monitor configuration changes
	watcher, err := provider.Watch()
	if err != nil {
		return nil, err
	}

	// Create ManagerImpl instance with initial configuration
	m := &ManagerImpl{
		provider:  provider,
		processor: processor,
		eventHub:  eventHub,
		watcher:   watcher,
		closeCh:   make(chan struct{}),
		mu:        sync.RWMutex{},
		l:         l,
	}

	// Start asyncLoop in background to handle config updates
	m.wg.Add(1)
	go m.asyncLoop()

	return m, nil
}

// asyncLoop runs in a separate goroutine to handle configuration updates and shutdown signals.
// It listens on the watcher channel for new config data, normalizes it, and dispatches events.
// When a close signal is received, it logs the shutdown and exits.
func (m *ManagerImpl) asyncLoop() {
	defer func() {
		m.provider.Close()
		m.wg.Done()
	}()

	for {
		select {
		case rawData, ok := <-m.watcher:
			if !ok {
				return
			}
			// Log that the watcher channel was closed (this typically indicates an unexpected situation)
			m.l.Info("weight provider watcher channel closed")

			// Normalize the raw configuration data
			normalizeConf, err := m.processor.Normalize(rawData)
			if err != nil {
				m.l.Error("the original data normalization failed", log.ErrorField(err))
				continue
			}

			// Build global and size class configurations from normalized data
			global := m.processor.BuildGlobalStruct(normalizeConf)
			sizeClasses := m.processor.BuildSizeClassStruct(normalizeConf)

			// Dispatch configuration change events to notify listeners
			m.dispatchGlobalEvent(global)
			m.dispatchSizeClassEvent(sizeClasses)
		case <-m.closeCh:
			// Handle manager shutdown request
			m.l.Info("receive stop manager signal")
			return
		}
	}
}

// Register adds a new event listener to the manager.
//
// Parameters:
//   - tag: unique identifier for the listener
//   - sc: size category to filter events for this listener
//   - bufferSize: optional buffer size for the event channel (default 0)
//
// Returns:
//   - <-chan Event: channel that will receive events matching the specified size category
func (m *ManagerImpl) Register(tag string, sc common.SizeCategory, bufferSize ...int) <-chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eventHub.Register(tag, sc, bufferSize...)
}

// Unregister removes a listener from the manager.
//
// Parameters:
//   - tag: unique identifier of the listener to remove
func (m *ManagerImpl) Unregister(tag string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHub.Unregister(tag)
}

// dispatchGlobalEvent notifies all registered listeners about a change in global configuration.
//
// Parameters:
//   - global: A map of size categories to their corresponding global weight values.
//     This represents the updated global configuration that needs to be dispatched.
func (m *ManagerImpl) dispatchGlobalEvent(global map[common.SizeCategory]float64) {
	m.eventHub.Dispatch(Event{
		eventType: GlobalConfigChange,
		global:    global,
		timestamp: time.Now().UnixNano(),
	})
}

// dispatchSizeClassEvent notifies all registered listeners about a change in size class configuration.
//
// Parameters:
//   - sizeClasses: A map of size categories to their corresponding array of weight values.
//     Each entry represents the updated size class weights for a specific category.
func (m *ManagerImpl) dispatchSizeClassEvent(sizeClasses map[common.SizeCategory][]float64) {
	for category, weights := range sizeClasses {
		m.eventHub.Dispatch(Event{
			eventType: SizeClassConfigChange,
			category:  category,
			details:   weights,
			timestamp: time.Now().UnixNano(),
		})
	}
}

// Close gracefully shuts down the ManagerImpl instance, ensuring all background processes are terminated
// and resources are released. It performs the following steps:
//  1. Attempts to switch the state from RunningState to StoppedState using atomic CAS.
//     If the state is already StoppedState, it returns immediately.
//  2. Closes the closeCh channel to signal shutdown to all waiting goroutines.
//  3. Waits for all background goroutines to finish via wg.Wait().
//  4. Releases resources held by the processor and eventHub components.
func (m *ManagerImpl) Close() {
	if !m.state.CompareAndSwap(RunningState, StoppedState) {
		return
	}

	close(m.closeCh)
	m.wg.Wait()
	m.processor.Close()
	m.eventHub.Close()
}
