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

type Manager interface{}

type ManagerImpl struct {
	provider       Provider
	Loader         Loader
	_              common.GlobalConfig
	_              common.SizeClassConfig
	mu             sync.RWMutex
	eventSig       chan struct{}
	eventNotifiers map[string]chan Event
	closeCh        chan struct{}
	l              log.Logger
	once           sync.Once
}

func NewManager(provider Provider, loader Loader, l log.Logger) Manager {
	return &ManagerImpl{
		provider:       provider,
		Loader:         loader,
		eventSig:       make(chan struct{}),
		eventNotifiers: map[string]chan Event{},
		closeCh:        make(chan struct{}),
		mu:             sync.RWMutex{},
		l:              l,
	}
}

func (m *ManagerImpl) RegisterEventNotify(tag string, ch chan Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventNotifiers[tag] = ch
}

func (m *ManagerImpl) UnregisterEventNotify(tag string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.eventNotifiers, tag)
}

//nolint:unused // will be used in new version
func (m *ManagerImpl) globalEventNotify(gc common.GlobalConfig) {
	m.mu.RLock()
	notifiers := m.eventNotifiers
	m.mu.RUnlock()
	if len(notifiers) == 0 {
		return
	}

	event := Event{
		eventType: GlobalConfigChange,
		global:    gc,
	}
	for _, ch := range notifiers {
		select {
		case ch <- event:
		case <-time.After(time.Millisecond * 100):
			m.l.Error("failed to notify global config",
				log.ErrorField(context.DeadlineExceeded))
		case <-m.closeCh:
			m.l.Info("global notifier loop receive stop signal")
			return
		}
	}
}

func (m *ManagerImpl) sizeClassEventNotify(details common.SizeClassDetail) {
	m.mu.RLock()
	notifiers := m.eventNotifiers
	m.mu.RUnlock()
	if len(notifiers) == 0 {
		return
	}

	event := Event{
		details:   details,
		eventType: SizeClassConfigChange,
	}

	for _, ch := range notifiers {
		select {
		case ch <- event:
		case <-time.After(time.Millisecond * 100):
			m.l.Error("failed to notify size class config",
				log.StringField("description", event.details.Description),
				log.ErrorField(context.DeadlineExceeded))
		case <-m.closeCh:
			m.l.Info("size class notifier loop receive stop signal")
		}
	}
}

func (m *ManagerImpl) monitor() {
	for {
		select {
		case <-m.eventSig:
			m.l.Debug("receive weight config event!")
			gc, err := m.Loader.LoadGlobalConfig()
			if err != nil {
				m.l.Error("failed to load global config", log.ErrorField(err))
				continue
			}
			m.globalEventNotify(gc)

			sc, err := m.Loader.LoadSizeClassConfig()
			if err != nil {
				m.l.Error("failed to load size class config", log.ErrorField(err))
				continue
			}
			m.sizeClassEventNotify(sc)
		case <-m.closeCh:
			return
		}
	}
}

func (m *ManagerImpl) Close() {
	m.once.Do(func() {
		close(m.closeCh)
		for _, ch := range m.eventNotifiers {
			close(ch)
		}
	})
}
