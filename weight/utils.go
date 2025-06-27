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
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/TimeWtr/slab/common"
	"gopkg.in/yaml.v3"
)

func DefaultGlobalWeightConfig() common.GlobalConfig {
	return common.GlobalConfig{
		Large:  0.6,
		Medium: 0.3,
		Small:  0.1,
	}
}

func DefaultSizeClassWeightConfig() common.SizeClassConfig {
	return common.SizeClassConfig{
		Small: newSizeClassBuilder(common.SmallSizeCategory).
			AddWeight(common.SizeClass8B, 0.35).
			AddWeight(common.SizeClass16B, 0.15).
			AddWeight(common.SizeClass32KB, 0.12).
			AddWeight(common.SizeClass64B, 0.10).
			AddWeight(common.SizeClass128B, 0.08).
			AddWeight(common.SizeClass256B, 0.06).
			AddWeight(common.SizeClass512B, 0.05).
			AddWeight(common.SizeClass1KB, 0.04).
			AddWeight(common.SizeClass2KB, 0.03).
			AddWeight(common.SizeClass4KB, 0.02).
			Build(),
		Medium: newSizeClassBuilder(common.MediumSizeCategory).
			AddWeight(common.SizeClass8KB, 0.40).
			AddWeight(common.SizeClass16KB, 0.25).
			AddWeight(common.SizeClass32KB, 0.20).
			AddWeight(common.SizeClass64KB, 0.10).
			AddWeight(common.SizeClass128KB, 0.05).
			Build(),
		Large: newSizeClassBuilder(common.LargeSizeCategory).
			AddWeight(common.SizeClass256KB, 0.35).
			AddWeight(common.SizeClass512KB, 0.30).
			AddWeight(common.SizeClass1MB, 0.20).
			AddWeight(common.SizeClass2MB, 0.10).
			AddWeight(common.SizeClass4MB, 0.05).
			AddWeight(common.SizeClass8MB, 0.04).
			AddWeight(common.SizeClass16MB, 0.03).
			AddWeight(common.SizeClass32MB, 0.02).
			Build(),
	}
}

func parseYaml(data []byte) (common.Config, error) {
	var cfg common.Config
	err := yaml.Unmarshal(data, &cfg)
	return cfg, err
}

func parseJSON(data []byte) (common.Config, error) {
	var cfg common.Config
	err := json.Unmarshal(data, &cfg)
	return cfg, err
}

func parseToml(data []byte) (common.Config, error) {
	var cfg common.Config
	err := toml.Unmarshal(data, &cfg)
	return cfg, err
}
