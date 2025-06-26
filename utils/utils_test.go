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

package utils

import "testing"

func TestNormalize_CalculateCores(t *testing.T) {
	testCases := []struct {
		name    string
		smScale float64
		mmScale float64
	}{
		{
			name:    "0.5 - 0.4",
			smScale: 0.5,
			mmScale: 0.4,
		},
		{
			name:    "0.25 - 0.3",
			smScale: 0.25,
			mmScale: 0.3,
		},
		{
			name:    "0.1 - 0.2",
			smScale: 0.1,
			mmScale: 0.2,
		},
		{
			name:    "0.05 - 0.05",
			smScale: 0.05,
			mmScale: 0.05,
		},
		{
			name:    "0.35 - 0.6",
			smScale: 0.35,
			mmScale: 0.6,
		},
		{
			name:    "0.35 - 0.63",
			smScale: 0.35,
			mmScale: 0.63,
		},
		{
			name:    "0.44 - 0.44",
			smScale: 0.44,
			mmScale: 0.44,
		},
	}

	cpuNums := 32

	for _, tc := range testCases {
		smCores, mmCores := CalculateCores(cpuNums, tc.smScale, tc.mmScale)

		t.Logf("small manager cpu counts: %d", smCores)
		t.Logf("medium manager cpu counts: %d", mmCores)
	}
}

func TestNormalize_CalculateCores_MultiCores(t *testing.T) {
	testCases := []struct {
		name    string
		cores   int
		smScale float64
		mmScale float64
	}{
		{
			name:    "32 cores",
			cores:   32,
			smScale: 0.5,
			mmScale: 0.4,
		},
		{
			name:    "2 cores-1",
			cores:   2,
			smScale: 0.25,
			mmScale: 0.3,
		},
		{
			name:    "2 cores-2",
			cores:   2,
			smScale: 0.1,
			mmScale: 0.2,
		},
		{
			name:    "128 cores",
			cores:   128,
			smScale: 0.05,
			mmScale: 0.05,
		},
		{
			name:    "64 cores",
			cores:   64,
			smScale: 0.35,
			mmScale: 0.6,
		},
		{
			name:    "10 cores",
			cores:   10,
			smScale: 0.35,
			mmScale: 0.63,
		},
		{
			name:    "1 core",
			cores:   1,
			smScale: 0.44,
			mmScale: 0.44,
		},
	}

	for _, tc := range testCases {
		smCores, mmCores := CalculateCores(tc.cores, tc.smScale, tc.mmScale)

		t.Logf("small manager cpu counts: %d", smCores)
		t.Logf("medium manager cpu counts: %d", mmCores)
	}
}
