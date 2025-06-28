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

//go:build e2e

package weight

import (
	"path/filepath"
	"testing"

	"github.com/TimeWtr/TurboAlloc/common"
	"github.com/TimeWtr/TurboAlloc/utils/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestManagerImpl_Register(t *testing.T) {
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	assert.NoError(t, err)

	eventHub := newEventHubImpl(logger)
	manager, err := NewManager(provider, newProcessorImpl(), eventHub, logger)
	assert.NoError(t, err)
	defer manager.Close()

	testCases := []struct {
		name    string
		tag     string
		sc      common.SizeCategory
		bufSize int
	}{
		{
			name:    "register small category",
			tag:     "name1",
			sc:      common.SmallSizeCategory,
			bufSize: 10,
		},
		{
			name:    "register medium category",
			tag:     "name2",
			sc:      common.MediumSizeCategory,
			bufSize: 10,
		},
		{
			name:    "register large category",
			tag:     "name3",
			sc:      common.LargeSizeCategory,
			bufSize: 10,
		},
		{
			name: "register global config",
			tag:  "name4",
			sc:   common.AllSizeCategory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ch := manager.Register(tc.tag, tc.sc, tc.bufSize)
			assert.NotNil(t, ch)
		})
	}
}

func TestManagerImpl_Unregister(t *testing.T) {
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	assert.NoError(t, err)

	eventHub := newEventHubImpl(logger)
	manager, err := NewManager(provider, newProcessorImpl(), eventHub, logger)
	assert.NoError(t, err)
	defer manager.Close()

	testCases := []struct {
		name    string
		tag     string
		sc      common.SizeCategory
		bufSize int
	}{
		{
			name:    "register small category",
			tag:     "name1",
			sc:      common.SmallSizeCategory,
			bufSize: 10,
		},
		{
			name:    "register medium category",
			tag:     "name2",
			sc:      common.MediumSizeCategory,
			bufSize: 10,
		},
		{
			name:    "register large category",
			tag:     "name3",
			sc:      common.LargeSizeCategory,
			bufSize: 10,
		},
		{
			name: "register global config",
			tag:  "name4",
			sc:   common.AllSizeCategory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ch := manager.Register(tc.tag, tc.sc, tc.bufSize)
			assert.NotNil(t, ch)

			manager.Unregister(tc.tag)
		})
	}
}
