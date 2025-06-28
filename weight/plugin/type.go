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

package plugin

import (
	"context"

	"github.com/TimeWtr/TurboAlloc/common"
)

type TypePlugin uint8

// TypePlugin represents the type category of an extension, used to classify different functionalities
// within the system. Each type serves a distinct role in the lifecycle and operation of components.
const (
	TypePluginAny       TypePlugin = iota // Generic or unspecified extension type
	TypePluginConfig                      // Configuration-related extensions
	TypePluginRuntime                     // Runtime behavior-affecting extensions
	TypePluginMonitor                     // Monitoring and observability extensions
	TypePluginResource                    // Resource management extensions
	TypePluginAlgorithm                   // Algorithm implementation extensions
	TypePluginLifecycle                   // Component lifecycle management extensions
)

// Plugin represents a pluggable component in the system that can be initialized and shut down.
// It serves as the base interface for all extension types, providing fundamental lifecycle operations.
type Plugin interface {
	// Name returns the unique identifier of the extension.
	Name() string

	// Init performs initialization tasks for the extension.
	// It takes a context for cancellation signals and configuration purposes.
	// Returns an error if initialization fails.
	Init(ctx context.Context) error

	// Shutdown gracefully terminates the extension.
	// It takes a context for cancellation signals during shutdown procedures.
	// Returns an error if shutdown fails.
	Shutdown(ctx context.Context) error
}

type ConfigPlugin interface {
	Plugin
	BeforeConfigLoad(ctx context.Context) error
	AfterConfigLoad(ctx context.Context, cfg *common.Config) error
	BeforeConfigReload(ctx context.Context) error
	AfterConfigReload(ctx context.Context, cfg *common.Config) error
}

type RuntimePlugin interface {
	Plugin
	BeforeWeightGet(ctx context.Context, sizeClass common.SizeClass) (context.Context, error)
	AfterWeightGet(ctx context.Context, sizeClass common.SizeClass, weight float64) (float64, error)
	WeightCalculationHook(ctx context.Context, sizeClass common.SizeClass, baseWeight float64) (float64, error)
}

type LifeCyclePlugin interface {
	Plugin
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type MonitorPlugin interface {
	Plugin
	Monitor(ctx context.Context) error
}
