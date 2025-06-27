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

	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils/log"
)

type (
	Manager interface {
		Close()
	}

	ManagerImpl struct {
		provider  Provider
		processor Processor
		global    common.GlobalConfig
		sizeClass map[common.SizeCategory][]float64
		mu        sync.RWMutex
		eventHub  EventHub
		watcher   <-chan common.Config
		closeCh   chan struct{}
		l         log.Logger
		state     atomic.Int32
		wg        sync.WaitGroup
	}
)

func NewManager(provider Provider, processor Processor, eventHub EventHub, l log.Logger) (Manager, error) {
	watcher, err := provider.Watch()
	if err != nil {
		return nil, err
	}
	m := &ManagerImpl{
		provider:  provider,
		processor: processor,
		eventHub:  eventHub,
		watcher:   watcher,
		closeCh:   make(chan struct{}),
		mu:        sync.RWMutex{},
		l:         l,
	}

	m.wg.Add(1)
	go m.asyncLoop()

	return m, nil
}

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
			m.l.Info("weight provider watcher channel closed")
			normalizeConf, err := m.processor.Normalize(rawData)
			if err != nil {
				m.l.Error("the original data normalization failed", log.ErrorField(err))
				continue
			}

			global := m.processor.BuildGlobalStruct(normalizeConf)
			sizeClasses := m.processor.BuildSizeClassStruct(normalizeConf)
			m.dispatchGlobalEvent(global)
			m.dispatchSizeClassEvent(sizeClasses)
		case <-m.closeCh:
			m.l.Info("receive stop manager signal")
			return
		}
	}
}

func (m *ManagerImpl) dispatchGlobalEvent(global map[common.SizeCategory]float64) {
	m.eventHub.Dispatch(Event{
		eventType: GlobalConfigChange,
		global:    global,
		timestamp: time.Now().UnixNano(),
	})
}

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

func (m *ManagerImpl) Close() {
	if !m.state.CompareAndSwap(RunningState, StoppedState) {
		return
	}

	close(m.closeCh)
	m.wg.Wait()
	m.processor.Close()
	m.processor.Close()
	m.eventHub.Close()
}
