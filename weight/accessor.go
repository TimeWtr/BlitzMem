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
	"sort"

	"github.com/TimeWtr/slab/common"
)

type SizeClassAccessor struct {
	common.SizeClassConfig
	details   common.SizeClassDetail
	weightMap map[common.SizeClass]float64
	cached    bool
}

func NewSizeClassAccessor(details common.SizeClassDetail) *SizeClassAccessor {
	return &SizeClassAccessor{
		details:   details,
		weightMap: make(map[common.SizeClass]float64),
	}
}

func (s *SizeClassAccessor) GetWeightMap() map[common.SizeClass]float64 {
	if !s.cached {
		for _, weight := range s.details.Weights {
			s.weightMap[common.SizeClass(weight.Size)] = weight.Weight
		}

		s.cached = true
	}

	return s.weightMap
}

func (s *SizeClassAccessor) GetWeight(size common.SizeClass) (float64, bool) {
	weightMap := s.weightMap
	weight, ok := weightMap[size]
	return weight, ok
}

func (s *SizeClassAccessor) GetSortedWeight() []common.SizeClassWeight {
	sorted := make([]common.SizeClassWeight, len(s.details.Weights))
	copy(sorted, s.details.Weights)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size < sorted[j].Size
	})

	return sorted
}

func (s *SizeClassAccessor) AddWeight(size common.SizeClass, weight float64) {
	if _, exist := s.weightMap[size]; exist {
		for i := 0; i < len(s.details.Weights); i++ {
			if s.details.Weights[i].Size == size.Int() {
				s.details.Weights[i].Weight = weight
				s.weightMap[size] = weight
				return
			}
		}
	}

	s.details.Weights = append(s.details.Weights, common.SizeClassWeight{
		Size:   int(size),
		Weight: weight,
	})
}
