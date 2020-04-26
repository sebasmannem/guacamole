// Copyright 2019 Oscar M Herrera(KnowledgeSource Solutions Inc}
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package postgresql

import (
	"sort"
	"testing"
)

// compareStringSliceNoOrder compares two slices of strings regardless of their order, a nil slice is considered an empty one
func compareStringSliceNoOrder(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// This isn't the faster way but it's cleaner and enough for us

	// Take a copy of the original slice
	a = append([]string(nil), a...)
	b = append([]string(nil), b...)

	sort.Sort(sort.StringSlice(a))
	sort.Sort(sort.StringSlice(b))

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestDiffReturnsChangedParams(t *testing.T) {
	var curParams Parameters = map[string]string{
		"max_connections": "100",
		"shared_buffers":  "10MB",
		"huge":            "off",
	}

	var newParams Parameters = map[string]string{
		"max_connections": "200",
		"shared_buffers":  "10MB",
		"work_mem":        "4MB",
	}

	expectedDiff := []string{"max_connections", "huge", "work_mem"}

	diff := curParams.Diff(newParams)

	if !compareStringSliceNoOrder(expectedDiff, diff) {
		t.Errorf("Expected diff is %v, but got %v", expectedDiff, diff)
	}
}
