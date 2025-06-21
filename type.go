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

import "time"

const (
	microClass  = "micro"
	smallClass  = "small"
	mediumClass = "medium"
	largeClass  = "large"
)

const (
	defaultSizeLevel       = S64KB
	defaultCompactionRatio = 0.7
	defaultStatsInterval   = 10 * time.Second
)

const numSizeClasses = 19

type SizeClass uint8

const (
	SizeClass8B = SizeClass(iota)
	SizeClass16B
	SizeClass32B
	SizeClass64B
	SizeClass128B
	SizeClass256B
	SizeClass512B
	SizeClass1KB
	SizeClass2KB
	SizeClass4KB
	SizeClass8KB
	SizeClass16KB
	SizeClass32KB
	SizeClass64KB
	SizeClass128KB
	SizeClass256KB
	SizeClass512KB
	SizeClass1MB
	SizeClass2MB
	SizeClassMax
)

func (s SizeClass) uint8() uint8 {
	return uint8(s)
}

func (s SizeClass) int() int {
	return int(s)
}

var sizeClassSizes = [SizeClassMax]int{
	SizeClass8B:    8,
	SizeClass16B:   16,
	SizeClass32B:   32,
	SizeClass64B:   64,
	SizeClass128B:  128,
	SizeClass256B:  256,
	SizeClass512B:  512,
	SizeClass1KB:   1024,
	SizeClass2KB:   1024 * 2,
	SizeClass4KB:   1024 * 4,
	SizeClass8KB:   1024 * 8,
	SizeClass16KB:  1024 * 16,
	SizeClass32KB:  1024 * 32,
	SizeClass64KB:  1024 * 64,
	SizeClass128KB: 1024 * 128,
	SizeClass256KB: 1024 * 256,
	SizeClass512KB: 1024 * 512,
	SizeClass1MB:   1024 * 1024,
	SizeClass2MB:   1024 * 1024 * 2,
}

type SizeLevel int

const (
	S64KB  SizeLevel = 1024 * 64
	S256KB SizeLevel = 1024 * 256
	S512KB SizeLevel = 1024 * 512
	S1MB   SizeLevel = 1024 * 1024
	S2MB   SizeLevel = 1024 * 1024 * 2
)

func (s SizeLevel) String() string {
	switch s {
	case S64KB:
		return "Slab 64KB level"
	case S256KB:
		return "Slab 256KB level"
	case S512KB:
		return "Slab 512KB level"
	case S1MB:
		return "Slab 1MB level"
	case S2MB:
		return "Slab 2MB level"
	default:
		return "Unknown level"
	}
}

func (s SizeLevel) valid() bool {
	switch s {
	case S64KB, S256KB, S512KB, S1MB, S2MB:
		return true
	default:
		return false
	}
}
