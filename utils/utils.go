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

func CalculateCores(cores int, smScale, mmScale float64) (smCores, mmCores int) {
	effectiveScale := smScale + mmScale
	if effectiveScale < 0.01 {
		effectiveScale = 1.0
		smScale = 0.5
		mmScale = 0.5
	}

	normalizedSmScale := smScale / effectiveScale
	normalizedMmScale := mmScale / effectiveScale
	effectiveCores := int(float64(cores)*effectiveScale + 0.5)
	if effectiveCores < 2 {
		effectiveCores = min(2, cores)
	}

	const filling = 0.5
	smCores = int(float64(effectiveCores)*normalizedSmScale + filling)
	mmCores = int(float64(effectiveCores)*normalizedMmScale + filling)

	return
}
