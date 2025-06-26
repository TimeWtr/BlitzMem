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

	"github.com/TimeWtr/slab/common"
)

type Extension interface {
	Name() string
	Init(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type ConfigExtension interface {
	Extension
	BeforeConfigLoad(ctx context.Context) error
	AfterConfigLoad(ctx context.Context, cfg *common.Config) error
	BeforeConfigReload(ctx context.Context) error
	AfterConfigReload(ctx context.Context, cfg *common.Config) error
}

type RuntimeExtension interface {
	Extension
	BeforeWeightGet(ctx context.Context, sizeClass common.SizeClass) (context.Context, error)
	AfterWeightGet(ctx context.Context, sizeClass common.SizeClass, weight float64) (float64, error)
	WeightCalculationHook(ctx context.Context, sizeClass common.SizeClass, baseWeight float64) (float64, error)
}

type LifeCycleExtension interface {
	Extension
	OnInit(ctx context.Context) error
	OnStart(ctx context.Context) error
	OnShutdown(ctx context.Context) error
}

type Monitor interface {
	Extension
	Monitor(ctx context.Context) error
}
