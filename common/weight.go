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

package common

type SizeClassWeight struct {
	Size   int     `json:"size" yaml:"size" toml:"size"`
	Weight float64 `json:"weight" yaml:"weight" toml:"weight"`
}

type SizeClassDetail struct {
	Description string            `json:"description" yaml:"description" toml:"description"`
	Weights     []SizeClassWeight `json:"weights" yaml:"weights" toml:"weights"`
}

type SizeClassConfig struct {
	Small  SizeClassDetail `json:"small" yaml:"small" toml:"small"`
	Medium SizeClassDetail `json:"medium" yaml:"medium" toml:"medium"`
	Large  SizeClassDetail `json:"large" yaml:"large" toml:"large"`
}

type GlobalConfig struct {
	Small  float64 `json:"small" yaml:"small" toml:"small"`
	Medium float64 `json:"medium" yaml:"medium" toml:"medium"`
	Large  float64 `json:"large" yaml:"large" toml:"large"`
}

type Config struct {
	Version   string          `json:"version" yaml:"version" toml:"version"`
	Global    GlobalConfig    `json:"global" yaml:"global" toml:"global"`
	SizeClass SizeClassConfig `json:"sizeClass" yaml:"sizeClass" toml:"sizeClass"`
}
