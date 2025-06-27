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
	"fmt"
	"math"
	"sort"

	"github.com/TimeWtr/slab/common"
)

type (
	Processor interface {
		Normalize(cfg common.Config) (common.Config, error)
		BuildGlobalStruct(normalizeConf common.Config) map[common.SizeCategory]float64
		BuildSizeClassStruct(normalizeConf common.Config) map[common.SizeCategory][]float64
		Close()
	}

	ProcessorImpl struct{}
)

func newProcessorImpl() Processor {
	return &ProcessorImpl{}
}

func (p *ProcessorImpl) Normalize(cfg common.Config) (common.Config, error) {
	const (
		tolerance    = 0.001
		requireTotal = 1.0
	)
	globalTotal := cfg.Global.Large + cfg.Global.Medium + cfg.Global.Small
	if math.Abs(globalTotal-requireTotal) >= tolerance {
		return common.Config{}, fmt.Errorf("invalid global weights: %f", globalTotal)
	}

	smallTotal := 0.0
	for _, w := range cfg.SizeClass.Small.Weights {
		smallTotal += w.Weight
	}
	if math.Abs(smallTotal-requireTotal) >= tolerance {
		return common.Config{}, fmt.Errorf("invalid small weights: %f", smallTotal)
	}

	mediumTotal := 0.0
	for _, w := range cfg.SizeClass.Medium.Weights {
		mediumTotal += w.Weight
	}
	if math.Abs(mediumTotal-requireTotal) >= tolerance {
		return common.Config{}, fmt.Errorf("invalid medium weights: %f", mediumTotal)
	}

	largeTotal := 0.0
	for _, w := range cfg.SizeClass.Large.Weights {
		largeTotal += w.Weight
	}
	if math.Abs(largeTotal-requireTotal) >= tolerance {
		return common.Config{}, fmt.Errorf("invalid large weights: %f", largeTotal)
	}

	return cfg, nil
}

func (p *ProcessorImpl) BuildGlobalStruct(normalizeConf common.Config) map[common.SizeCategory]float64 {
	return map[common.SizeCategory]float64{
		common.SmallSizeCategory:  normalizeConf.Global.Small,
		common.MediumSizeCategory: normalizeConf.Global.Medium,
		common.LargeSizeCategory:  normalizeConf.Global.Large,
	}
}

func (p *ProcessorImpl) BuildSizeClassStruct(normalizeConf common.Config) map[common.SizeCategory][]float64 {
	res := map[common.SizeCategory][]float64{}
	smallConf := p.buildSmall(normalizeConf)
	res[common.SmallSizeCategory] = smallConf

	mediumConf := p.buildMedium(normalizeConf)
	res[common.MediumSizeCategory] = mediumConf

	largeConf := p.buildLarge(normalizeConf)
	res[common.LargeSizeCategory] = largeConf

	return res
}

func (p *ProcessorImpl) buildSmall(normalizeConf common.Config) []float64 {
	smallConf := make([]float64, 0, len(normalizeConf.SizeClass.Small.Weights))
	for _, weight := range normalizeConf.SizeClass.Small.Weights {
		smallConf = append(smallConf, weight.Weight)
	}
	sort.Slice(smallConf, func(i, j int) bool {
		return smallConf[i] < smallConf[j]
	})

	return smallConf
}

func (p *ProcessorImpl) buildMedium(normalizeConf common.Config) []float64 {
	mediumConf := make([]float64, 0, len(normalizeConf.SizeClass.Medium.Weights))
	for _, weight := range normalizeConf.SizeClass.Medium.Weights {
		mediumConf = append(mediumConf, weight.Weight)
	}
	sort.Slice(mediumConf, func(i, j int) bool {
		return mediumConf[i] < mediumConf[j]
	})
	return mediumConf
}

func (p *ProcessorImpl) buildLarge(normalizeConf common.Config) []float64 {
	largeConf := make([]float64, 0, len(normalizeConf.SizeClass.Large.Weights))
	for _, weight := range normalizeConf.SizeClass.Large.Weights {
		largeConf = append(largeConf, weight.Weight)
	}
	sort.Slice(largeConf, func(i, j int) bool {
		return largeConf[i] < largeConf[j]
	})

	return largeConf
}

func (p *ProcessorImpl) Close() {}
