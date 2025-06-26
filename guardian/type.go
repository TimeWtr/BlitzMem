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

package guardian

import "unsafe"

type Guardian interface {
	Alloc(size int) (unsafe.Pointer, error)
	Access(ptr unsafe.Pointer, accessFunc func([]byte) error) error
	Free(ptr unsafe.Pointer, size int) error
	SetSecurityLevel(level SecurityLevel) error
	Destroy()
}

type AlgorithmInfo struct {
	Name      string
	KeySize   int
	BlockSize int
	IsGuoMi   bool
	IsHWAccel bool
}

type SecurityLevel int

const (
	SecurityNone SecurityLevel = iota
	SecurityBasic
	SecurityAdvanced
	SecurityGuoMiCompliance
)
