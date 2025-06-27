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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var jsonContent = `
{
  "version": "1.0",
    "global": {
      "small": 0.6,
      "medium": 0.3,
      "large": 0.1
    },
    "sizeClass": {
      "small": {
        "description": "Small Size Weight (8B-8KB)",
        "weights": [
          {"size": 8, "weight": 0.35},
          {"size": 16, "weight": 0.15},
          {"size": 32, "weight": 0.12},
          {"size": 64, "weight": 0.10},
          {"size": 128, "weight": 0.08},
          {"size": 256, "weight": 0.06},
          {"size": 512, "weight": 0.05},
          {"size": 1024, "weight": 0.04},
          {"size": 2048, "weight": 0.03},
          {"size": 4096, "weight": 0.02}
        ]
      },
      "medium": {
        "description": "Medium Size Weight (8KB-64KB)",
        "weights": [
          {"size": 8192, "weight": 0.40},
          {"size": 16384, "weight": 0.25},
          {"size": 32768, "weight": 0.20},
          {"size": 49152, "weight": 0.10},
          {"size": 65536, "weight": 0.05}
        ]
      },
      "large": {
        "description": "Large Size Weight (64KB+)",
        "weights": [
          {"size": 131072, "weight": 0.60},
          {"size": 262144, "weight": 0.30},
          {"size": 524288, "weight": 0.05},
          {"size": 1048576, "weight": 0.03},
          {"size": 2097152, "weight": 0.002},
          {"size": 4194304, "weight": 0.01},
          {"size": 8388608, "weight": 0.005},
          {"size": 16777216, "weight": 0.0025}
        ]
      }
    }
}
`

var tomlContent = `
version = "1.0"

[global]
small = 0.6
medium = 0.3
large = 0.1

[sizeClass.small]
description = "Small Size Weight (8B-8KB)"

[[sizeClass.small.weights]]
size = 8
weight = 0.35

[[sizeClass.small.weights]]
size = 16
weight = 0.15

[[sizeClass.small.weights]]
size = 32
weight = 0.12

[[sizeClass.small.weights]]
size = 64
weight = 0.10

[[sizeClass.small.weights]]
size = 128
weight = 0.08

[[sizeClass.small.weights]]
size = 256
weight = 0.06

[[sizeClass.small.weights]]
size = 512
weight = 0.05

[[sizeClass.small.weights]]
size = 1024
weight = 0.04

[[sizeClass.small.weights]]
size = 2048
weight = 0.03

[[sizeClass.small.weights]]
size = 4096
weight = 0.02

[sizeClass.medium]
description = "Medium Size Weight (8KB-64KB)"

[[sizeClass.medium.weights]]
size = 8192
weight = 0.40

[[sizeClass.medium.weights]]
size = 16384
weight = 0.25

[[sizeClass.medium.weights]]
size = 32768
weight = 0.20

[[sizeClass.medium.weights]]
size = 49152
weight = 0.10

[[sizeClass.medium.weights]]
size = 65536
weight = 0.05

[sizeClass.large]
description = "Large Size Weight (64KB+)"

[[sizeClass.large.weights]]
size = 131072
weight = 0.60

[[sizeClass.large.weights]]
size = 262144
weight = 0.30

[[sizeClass.large.weights]]
size = 524288
weight = 0.10
`

var yamlContent = `
version: "1.0"

global:
  small: 0.6
  medium: 0.3
  large: 0.1

sizeClass:
  small:
    description: "Small Size Weight (8B-8KB)"
    weights:
      - size: 8
        weight: 0.35
      - size: 16
        weight: 0.15
      - size: 32
        weight: 0.12
      - size: 64
        weight: 0.10
      - size: 128
        weight: 0.08
      - size: 256
        weight: 0.06
      - size: 512
        weight: 0.05
      - size: 1024
        weight: 0.04
      - size: 2048
        weight: 0.03
      - size: 4096
        weight: 0.02
  medium:
    description: "Medium Size Weight (8KB-64KB)"
    weights:
      - size: 8192
        weight: 0.40
      - size: 16384
        weight: 0.25
      - size: 32768
        weight: 0.20
      - size: 49152
        weight: 0.10
      - size: 65536
        weight: 0.05
  large:
    description: "Large Size Weight (64KB+)"
    weights:
      - size: 131072
        weight: 0.60
      - size: 262144
        weight: 0.30
      - size: 524288
        weight: 0.10
`

func TestFileProvider_Basic(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "1.0", cfg.Version)
	case <-time.After(1 * time.Second):
		t.Fatal("initial configuration not received")
	}
	const version = "v2.0"
	cfg.Version = version
	jsonWriteConfig(t, cfgPath, cfg)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "v2.0", cfg.Version)
	case <-time.After(2 * time.Second):
		t.Fatal("update configuration not received")
	}

	provider.Close()

	select {
	case _, ok := <-configCh:
		assert.False(t, ok, "channel not closed")
	case <-time.After(1 * time.Second):
		t.Fatal("closed timeout")
	}
}

func TestFileProvider_Basic_Toml(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.toml")
	cfg, err := tomlUnmarshal()
	assert.NoError(t, err)
	tomlWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeTOML, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "1.0", cfg.Version)
	case <-time.After(1 * time.Second):
		t.Fatal("initial configuration not received")
	}
	cfg.Version = "v2.0"
	tomlWriteConfig(t, cfgPath, cfg)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "v2.0", cfg.Version)
	case <-time.After(2 * time.Second):
		t.Fatal("update configuration not received")
	}

	provider.Close()

	select {
	case _, ok := <-configCh:
		assert.False(t, ok, "channel not closed")
	case <-time.After(1 * time.Second):
		t.Fatal("closed timeout")
	}
}

func TestFileProvider_Basic_Yaml(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.toml")
	cfg, err := yamlUnmarshal()
	assert.NoError(t, err)
	yamlWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeYAML, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "1.0", cfg.Version)
	case <-time.After(1 * time.Second):
		t.Fatal("initial configuration not received")
	}
	cfg.Version = "v2.0"
	yamlWriteConfig(t, cfgPath, cfg)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "v2.0", cfg.Version)
	case <-time.After(2 * time.Second):
		t.Fatal("update configuration not received")
	}

	provider.Close()

	select {
	case _, ok := <-configCh:
		assert.False(t, ok, "channel not closed")
	case <-time.After(1 * time.Second):
		t.Fatal("closed timeout")
	}
}

func TestFileProvider_FileRemoval(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	<-configCh
	require.NoError(t, os.Remove(cfgPath))

	time.Sleep(100 * time.Millisecond)

	cfg.Version = "v3.0"
	jsonWriteConfig(t, cfgPath, cfg)

	select {
	case newCfg := <-configCh:
		assert.Equal(t, "v3.0", newCfg.Version)
	case <-time.After(2 * time.Second):
		t.Fatal("the restored configuration was not received")
	}

	provider.Close()
}

func TestFileProvider_RapidUpdates(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	<-configCh

	for i := 0; i < 10; i++ {
		cfg.Version = "update-" + string(rune('A'+i))
		jsonWriteConfig(t, cfgPath, cfg)
		time.Sleep(100 * time.Millisecond)
	}

	finalCfg := common.Config{Version: "final"}
	jsonWriteConfig(t, cfgPath, finalCfg)

	select {
	case cfg := <-configCh:
		assert.Equal(t, "final", cfg.Version)
	case <-time.After(2 * time.Second):
		t.Fatal("final configuration not received")
	}

	select {
	case cfg := <-configCh:
		t.Fatalf("received unexpected configuration update: %+v", cfg)
	case <-time.After(500 * time.Millisecond):
	}

	provider.Close()
}

func TestFileProvider_CloseSafety(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	<-configCh

	cfg.Version = "1.0.1"
	jsonWriteConfig(t, cfgPath, cfg)

	provider.Close()

	select {
	case _, ok := <-configCh:
		assert.False(t, ok, "configuration channel not closed properly")
	case <-time.After(1 * time.Second):
		t.Fatal("closed timeout")
	}
}

func TestFileProvider_MultiClose(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	provider, err := NewFileProvider(ParseTypeJSON, cfgPath, logger)
	require.NoError(t, err)

	configCh, err := provider.Watch()
	require.NoError(t, err)

	<-configCh

	cfg.Version = "1.0.1"
	jsonWriteConfig(t, cfgPath, cfg)

	provider.Close()
	provider.Close()

	select {
	case _, ok := <-configCh:
		assert.False(t, ok, "configuration channel not closed properly")
	case <-time.After(1 * time.Second):
		t.Fatal("closed timeout")
	}
	provider.Close()
}

func TestFileProvider_InitialFailure(t *testing.T) {
	cfgPath := filepath.Join("weight_test", "weight.json")
	cfg, err := jsonUnmarshal()
	assert.NoError(t, err)
	jsonWriteConfig(t, cfgPath, cfg)
	l, err := zap.NewDevelopment()
	assert.NoError(t, err)
	logger := log.NewZapAdapter(l)
	require.NoError(t, os.Remove(cfgPath))

	_, err = NewFileProvider(ParseTypeJSON, cfgPath, logger)
	assert.Error(t, err, "expected error but did not occur")
}

func jsonWriteConfig(t *testing.T, path string, cfg common.Config) {
	t.Helper()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	err = ioutil.WriteFile(path, data, 0o600)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
}

func tomlWriteConfig(t *testing.T, path string, cfg common.Config) {
	t.Helper()
	data, err := toml.Marshal(cfg)
	require.NoError(t, err)

	err = ioutil.WriteFile(path, data, 0o600)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
}

func yamlWriteConfig(t *testing.T, path string, cfg common.Config) {
	t.Helper()
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	err = ioutil.WriteFile(path, data, 0o600)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
}

func jsonUnmarshal() (common.Config, error) {
	var cfg common.Config
	err := json.Unmarshal([]byte(jsonContent), &cfg)
	return cfg, err
}

func tomlUnmarshal() (common.Config, error) {
	var cfg common.Config
	err := toml.Unmarshal([]byte(tomlContent), &cfg)
	return cfg, err
}

func yamlUnmarshal() (common.Config, error) {
	var cfg common.Config
	err := yaml.Unmarshal([]byte(yamlContent), &cfg)
	return cfg, err
}
