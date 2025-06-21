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

package slab

import "sync/atomic"

var stats = struct {
	microAllocs, microFrees   uint64
	smallAllocs, smallFrees   uint64
	mediumAllocs, mediumFrees uint64
	largeAllocs, largeFrees   uint64
	compactions               uint64
	releases                  uint64
	cacheHits, cacheMisses    uint64
	hitRate                   float64
}{}

type Stats struct {
	MicroAllocs, MicroFrees   uint64
	SmallAllocs, SmallFrees   uint64
	MediumAllocs, MediumFrees uint64
	LargeAllocs, LargeFrees   uint64
	Compactions               uint64
	Releases                  uint64
	CacheHits, CacheMisses    uint64
	HitRate                   float64
}

func getStats() Stats {
	totalAllocs := atomic.LoadUint64(&stats.microAllocs) +
		atomic.LoadUint64(&stats.smallAllocs) +
		atomic.LoadUint64(&stats.mediumAllocs) +
		atomic.LoadUint64(&stats.largeAllocs)

	cacheHits := atomic.LoadUint64(&stats.cacheHits)
	hitRate := 0.0
	if totalAllocs > 0 {
		hitRate = float64(cacheHits) / float64(totalAllocs) * 100
	}

	return Stats{
		MicroAllocs:  atomic.LoadUint64(&stats.microAllocs),
		MicroFrees:   atomic.LoadUint64(&stats.microFrees),
		SmallAllocs:  atomic.LoadUint64(&stats.smallAllocs),
		SmallFrees:   atomic.LoadUint64(&stats.smallFrees),
		MediumAllocs: atomic.LoadUint64(&stats.mediumAllocs),
		MediumFrees:  atomic.LoadUint64(&stats.mediumFrees),
		LargeAllocs:  atomic.LoadUint64(&stats.largeAllocs),
		LargeFrees:   atomic.LoadUint64(&stats.largeFrees),
		Compactions:  atomic.LoadUint64(&stats.compactions),
		Releases:     atomic.LoadUint64(&stats.releases),
		CacheHits:    atomic.LoadUint64(&stats.cacheHits),
		CacheMisses:  atomic.LoadUint64(&stats.cacheMisses),
		HitRate:      hitRate,
	}
}

func recordAlloc(class string) {
	switch class {
	case microClass:
		atomic.AddUint64(&stats.microAllocs, 1)
	case smallClass:
		atomic.AddUint64(&stats.smallAllocs, 1)
	case mediumClass:
		atomic.AddUint64(&stats.mediumAllocs, 1)
	case largeClass:
		atomic.AddUint64(&stats.largeAllocs, 1)
	}
}

func recordFree(class string) {
	switch class {
	case microClass:
		atomic.AddUint64(&stats.microFrees, 1)
	case smallClass:
		atomic.AddUint64(&stats.smallFrees, 1)
	case mediumClass:
		atomic.AddUint64(&stats.mediumFrees, 1)
	case largeClass:
		atomic.AddUint64(&stats.largeFrees, 1)
	}
}

func recordCacheHits() {
	atomic.AddUint64(&stats.cacheHits, 1)
}

func recordCacheMisses() {
	atomic.AddUint64(&stats.cacheMisses, 1)
}

func recordRelease(count uint64) {
	atomic.AddUint64(&stats.releases, count)
}

func recordCompaction(count uint64) {
	atomic.AddUint64(&stats.compactions, count)
}
