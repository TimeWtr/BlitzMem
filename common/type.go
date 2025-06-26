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

type SizeClass uint8

const (
	SizeClass8B SizeClass = iota
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
	SizeClass4MB
	SizeClass8MB
	SizeClass16MB
	SizeClass32MB
	SizeClassMax = SizeClass32MB
)

const Unknown = "unknown"

func (s SizeClass) Uint8() uint8 {
	return uint8(s)
}

func (s SizeClass) Int() int {
	return int(s)
}

func (s SizeClass) String() string {
	switch {
	case s.Uint8() < SizeClass8KB.Uint8():
		return s.smallString()
	case s.Uint8() < SizeClass128KB.Uint8():
		return s.mediumString()
	default:
		return s.largeString()
	}
}

func (s SizeClass) smallString() string {
	switch s {
	case SizeClass128B:
		return "128B"
	case SizeClass256B:
		return "256B"
	case SizeClass512B:
		return "512B"
	case SizeClass1KB:
		return "1KB"
	case SizeClass2KB:
		return "2KB"
	case SizeClass4KB:
		return "4KB"
	default:
		return Unknown
	}
}

func (s SizeClass) mediumString() string {
	switch s {
	case SizeClass8B:
		return "8B"
	case SizeClass16B:
		return "16B"
	case SizeClass32B:
		return "32B"
	case SizeClass64B:
		return "64B"
	default:
		return Unknown
	}
}

func (s SizeClass) largeString() string {
	switch s {
	case SizeClass8KB:
		return "8KB"
	case SizeClass16KB:
		return "16KB"
	case SizeClass32KB:
		return "32KB"
	case SizeClass64KB:
		return "64KB"
	case SizeClass128KB:
		return "128KB"
	case SizeClass256KB:
		return "256KB"
	case SizeClass512KB:
		return "512KB"
	case SizeClass1MB:
		return "1MB"
	case SizeClass2MB:
		return "2MB"
	case SizeClass4MB:
		return "4MB"
	case SizeClass8MB:
		return "8MB"
	case SizeClass16MB:
		return "16MB"
	case SizeClass32MB:
		return "32MB"
	default:
		return Unknown
	}
}

//nolint:unused  // will be used
func (s SizeClass) valid() bool {
	return s <= SizeClassMax
}

const (
	B8   = 8
	B16  = 16
	B32  = 32
	B64  = 64
	B128 = 128
	B256 = 256
	B512 = 512
	KB   = 1024
	MB   = 1024 * 1024
)

const (
	SmallSizeClassNums  = 10
	mediumSizeClassNums = 4
	largeSizeClassNums  = 9
)

var SizeClassSizes = map[SizeCategory]map[SizeClass]int{
	SmallSizeCategory: {
		SizeClass8B:   B8,
		SizeClass16B:  B16,
		SizeClass32B:  B32,
		SizeClass64B:  B64,
		SizeClass128B: B128,
		SizeClass256B: B256,
		SizeClass512B: B512,
		SizeClass1KB:  KB,
		SizeClass2KB:  KB * 2,
		SizeClass4KB:  KB * 4,
	},
	MediumSizeCategory: {
		SizeClass8KB:  KB * 8,
		SizeClass16KB: KB * 16,
		SizeClass32KB: KB * 32,
		SizeClass64KB: KB * 64,
	},
	LargeSizeCategory: {
		SizeClass128KB: KB * 128,
		SizeClass256KB: KB * 256,
		SizeClass512KB: KB * 512,
		SizeClass1MB:   MB,
		SizeClass2MB:   MB * 2,
		SizeClass4MB:   MB * 4,
		SizeClass8MB:   MB * 8,
		SizeClass16MB:  MB * 16,
		SizeClass32MB:  MB * 32,
	},
}

type SizeCategory uint8

const (
	SmallSizeCategory SizeCategory = iota
	MediumSizeCategory
	LargeSizeCategory
	AllSizeCategory
)
