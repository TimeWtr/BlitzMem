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

package core

import (
	"errors"
	"runtime"

	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils"
)

type Manager struct {
	sm              *SmallManager
	mm              *MediumManager
	lm              *LargeManager
	globalConfig    common.GlobalConfig
	sizeClassConfig common.SizeClassConfig
}

func NewManager(smWeight, mmWeight, _ float64) (*Manager, error) {
	m := &Manager{}

	// Normalization of the percentage of small and medium target managers
	cpuCores := runtime.GOMAXPROCS(0)
	if cpuCores <= 0 {
		return nil, errors.New("invalid cpu core count")
	}

	smCores, mmCores := utils.CalculateCores(cpuCores, smWeight, mmWeight)
	m.sm = newSmallManager(m.calculateShards(cpuCores, smCores))
	m.mm = newMediumManager(m.calculateShards(cpuCores, mmCores))
	return m, nil
}

func (m *Manager) calculateShards(cpuCores, smCores int) int {
	const (
		twice           = 2
		triple          = 3
		fourFold        = 4
		smallThreshold  = 32
		mediumThreshold = 64
	)

	switch {
	case cpuCores <= smallThreshold:
		return smCores * fourFold
	case cpuCores <= mediumThreshold:
		return smCores * triple
	default:
		return smCores * twice
	}
}

func (m *Manager) OnGlobalConfigChange(_, _ common.GlobalConfig) {
}
