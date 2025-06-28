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
	"github.com/TimeWtr/TurboAlloc/common"
	"gopkg.in/yaml.v3"
)

func DefaultGlobalWeightConfig() common.GlobalConfig {
	return common.GlobalConfig{
		Large:  0.6,
		Medium: 0.3,
		Small:  0.1,
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
