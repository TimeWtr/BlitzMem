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
	"math"

	"github.com/TimeWtr/slab/common"
)

var defaultCategoryDescription = map[common.SizeCategory]string{
	common.SmallSizeCategory:  "Small size category",
	common.MediumSizeCategory: "Medium size category",
	common.LargeSizeCategory:  "Large size category",
}

type SizeClassBuilder struct {
	category common.SizeCategory
	details  common.SizeClassDetail
}

func newSizeClassBuilder(category common.SizeCategory) *SizeClassBuilder {
	return &SizeClassBuilder{
		category: category,
		details: common.SizeClassDetail{
			Description: defaultCategoryDescription[category],
		},
	}
}

func (s *SizeClassBuilder) AddWeight(size common.SizeClass, weight float64) *SizeClassBuilder {
	s.details.Weights = append(s.details.Weights, common.SizeClassWeight{
		Size:   size.Int(),
		Weight: weight,
	})

	return s
}

func (s *SizeClassBuilder) Build() common.SizeClassDetail {
	total := 0.0
	for _, w := range s.details.Weights {
		total += w.Weight
	}

	const (
		standard  = 1.0
		tolerance = 0.001
	)

	// Normalize weights
	if math.Abs(total-standard) > tolerance {
		scale := standard / total
		for i := range s.details.Weights {
			s.details.Weights[i].Weight *= scale
		}
	}

	return s.details
}
