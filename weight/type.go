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
	"github.com/TimeWtr/slab/common"
)

type Loader interface {
	LoadGlobalConfig() (common.GlobalConfig, error)
	LoadSizeClassConfig() (common.SizeClassDetail, error)
	LoadSizeClassDescription() (string, error)
}

type Event struct {
	eventType EventType
	timestamp int64
	category  common.SizeCategory
	global    common.GlobalConfig
	details   common.SizeClassDetail
}

type EventType int

const (
	GlobalConfigChange EventType = iota
	SizeClassConfigChange
)

type ParseType string

const (
	ParseTypeYAML ParseType = "YAML"
	ParseTypeJSON ParseType = "JSON"
	ParseTypeTOML ParseType = "TOML"
)

func (p ParseType) String() string {
	return string(p)
}

func (p ParseType) valid() bool {
	switch p {
	case ParseTypeYAML, ParseTypeJSON, ParseTypeTOML:
		return true
	default:
		return false
	}
}
