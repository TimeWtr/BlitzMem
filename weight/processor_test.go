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
	"testing"

	"github.com/TimeWtr/slab/common"
	"github.com/stretchr/testify/assert"
)

func getRawData(t *testing.T) (common.Config, error) {
	t.Helper()
	var cfg common.Config
	err := json.Unmarshal([]byte(jsonContent), &cfg)
	return cfg, err
}

func Test_Normalize_Global_Invalid(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	cfg.Global.Small = 0.7
	process := newProcessorImpl()
	_, err = process.Normalize(cfg)
	assert.Error(t, err)
}

func Test_Normalize_Small_Size_Invalid(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	cfg.SizeClass.Small.Weights[0].Weight = 0.3
	cfg.SizeClass.Small.Weights[1].Weight = 0.3
	process := newProcessorImpl()
	_, err = process.Normalize(cfg)
	assert.Error(t, err)
}

func Test_Normalize_Medium_Size_Invalid(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	cfg.SizeClass.Medium.Weights[2].Weight = 0.3
	cfg.SizeClass.Medium.Weights[3].Weight = 0.3
	process := newProcessorImpl()
	_, err = process.Normalize(cfg)
	assert.Error(t, err)
}

func Test_Normalize_Large_Size_Invalid(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	cfg.SizeClass.Large.Weights[0].Weight = 0.3
	cfg.SizeClass.Large.Weights[1].Weight = 0.8
	process := newProcessorImpl()
	_, err = process.Normalize(cfg)
	assert.Error(t, err)
}

func Test_Normalize_Weights(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	process := newProcessorImpl()
	_, err = process.Normalize(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, cfg.Global.Large+cfg.Global.Medium+cfg.Global.Small)
}

func Test_BuildGlobalStruct(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	process := newProcessorImpl()
	res := process.BuildGlobalStruct(cfg)
	requireMap := map[common.SizeCategory]float64{
		common.SmallSizeCategory:  cfg.Global.Small,
		common.MediumSizeCategory: cfg.Global.Medium,
		common.LargeSizeCategory:  cfg.Global.Large,
	}
	assert.Equal(t, requireMap, res)
}

func TestProcessorImpl_BuildSizeClassStruct(t *testing.T) {
	cfg, err := getRawData(t)
	assert.NoError(t, err)
	process := newProcessorImpl()
	res := process.BuildSizeClassStruct(cfg)
	t.Logf("%+v", res)
}
