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
	"context"
	"sync"

	"github.com/TimeWtr/TurboAlloc/utils/atomicx"
	"github.com/TimeWtr/TurboAlloc/weight/plugin"
)

//go:generate mockgen -source=plugins_hub.go -destination=plugins_hub_mock.go -package=weight
type (
	// PluginsHub defines a comprehensive interface combining multiple plugin-related capabilities.
	// It embeds several sub-interfaces to manage plugins' lifecycle, registration, discovery,
	// and scheduling functionalities. This interface serves as a central hub for interacting
	// with various types of plugins within the system.
	PluginsHub interface {
		// PluginsManager defines the interface for managing plugin registration and unregistration operations.
		// It provides methods to register single or batch plugins, and to unregister plugins either by instance
		// or by name.
		PluginsManager
		// PluginsDiscover Provides methods for discovering plugins managed by the PluginsHub.
		// It allows querying registered plugin types and retrieving plugins by type.
		PluginsDiscover
		// PluginsLifecycle Embeds lifecycle management methods (InitAll, StartAll, ShutdownAll, HealthCheck)
		PluginsLifecycle
		// Scheduler provides the capability to invoke plugins at specific hook points.
		// It includes a CallPlugin method that executes applicable plugins under a given hook name,
		// and collects results from each plugin's execution. This interface enables the system
		// to dynamically extend behavior at different stages.
		Scheduler
	}

	PluginsManager interface {
		// Register adds a single plugin to the manager.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//   tp: Type of the plugin being registered.
		//   plugin: The plugin instance to register.
		//
		// Returns:
		//   error: If registration fails, an error is returned; otherwise, nil.
		Register(ctx context.Context, tp plugin.TypePlugin, plugin plugin.Plugin) error

		// RegisterBatch adds multiple plugins of the same type to the manager.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//   tp: Type of the plugins being registered.
		//   plugins: Slice of plugin instances to register.
		//
		// Returns:
		//   error: If batch registration fails, an error is returned; otherwise, nil.
		RegisterBatch(ctx context.Context, tp plugin.TypePlugin, plugins []plugin.Plugin) error

		// Unregister removes a specific plugin from the manager.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//   tp: Type of the plugin being unregistered.
		//   plugin: The plugin instance to remove.
		//
		// Returns:
		//   error: If unregistration fails, an error is returned; otherwise, nil.
		Unregister(ctx context.Context, tp plugin.TypePlugin, plugin plugin.Plugin) error

		// UnregisterByName removes a plugin by its name rather than its instance.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//   tp: Type of the plugin being unregistered.
		//   name: Name of the plugin to remove.
		//
		// Returns:
		//   error: If unregistration by name fails, an error is returned; otherwise, nil.
		UnregisterByName(ctx context.Context, tp plugin.TypePlugin, name string) error
	}

	// PluginsDiscover provides methods for discovering plugins managed by the PluginsHub.
	// It allows querying registered plugin types and retrieving plugins by type.
	PluginsDiscover interface {
		// GetPluginTypes returns a slice of all plugin types currently registered in the system.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//
		// Returns:
		//   []plugin.TypePlugin: A slice containing all registered plugin types.
		GetPluginTypes(ctx context.Context) []plugin.TypePlugin

		// GetPluginsByType returns all plugins of the specified type.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//   tp: The type of plugins to retrieve.
		//
		// Returns:
		//   []plugin.Plugin: A slice containing all plugins of the specified type.
		GetPluginsByType(ctx context.Context, tp plugin.TypePlugin) []plugin.Plugin
	}

	// PluginsLifecycle defines the lifecycle management interface for plugins.
	// It includes methods for initializing, starting, shutting down, and health checking all plugins.
	PluginsLifecycle interface {
		// InitAll initializes all plugins across the system.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//
		// Returns:
		//   error: If initialization fails, an error is returned; otherwise, nil.
		InitAll(ctx context.Context) error

		// StartAll starts all plugins across the system.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//
		// Returns:
		//   error: If starting plugins fails, an error is returned; otherwise, nil.
		StartAll(ctx context.Context) error

		// ShutdownAll gracefully shuts down all plugins across the system.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//
		// Returns:
		//   error: If shutting down plugins fails, an error is returned; otherwise, nil.
		ShutdownAll(ctx context.Context) error

		// HealthCheck performs a health check on all plugins and returns their statuses.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//
		// Returns:
		//   map[string]PluginStatus: A map containing plugin names as keys and their corresponding status.
		HealthCheck(ctx context.Context) map[string]PluginStatus
	}

	// Scheduler defines an interface for invoking plugins during specific hooks.
	// It provides a method to call a function across all relevant plugins and collect results.
	Scheduler interface {
		// CallPlugin executes the provided hook function across applicable plugins.
		//
		// Parameters:
		//   ctx: Context for managing request-scoped data and cancellation.
		//   hookName: Name of the hook being executed, used for identifying the context of plugin execution.
		//   hookFunc: Function to apply to each plugin; takes a plugin.Plugin and returns (result any, error).
		//
		// Returns:
		//   []HookResult: A slice containing results from each plugin's execution.
		//   error: If execution fails during the hook process, an error is returned; otherwise, nil.
		CallPlugin(ctx context.Context,
			hookName string,
			hookFunc func(p plugin.Plugin) (any, error)) ([]HookResult, error)
	}

	PluginStatus struct {
		Name string
	}

	HookResult struct{}

	PluginsHubImpl struct {
		plugins map[plugin.TypePlugin][]plugin.Plugin
		mu      sync.RWMutex
		state   atomicx.Bool
	}
)
