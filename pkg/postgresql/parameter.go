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
	"reflect"
)

// Parameters is a map of parameters (key=values)
type Parameters map[string]string

// Equals returns true if both parameter maps are completely equal
func (s Parameters) Equals(is Parameters) bool {
	return reflect.DeepEqual(s, is)
}

// Diff returns the list of pgParameters changed(newly added, existing deleted and value changed)
func (s Parameters) Diff(newParams Parameters) []string {
	var changedParams []string
	for k, v := range newParams {
		if val, ok := s[k]; !ok || v != val {
			changedParams = append(changedParams, k)
		}
	}

	for k := range s {
		if _, ok := newParams[k]; !ok {
			changedParams = append(changedParams, k)
		}
	}
	return changedParams
}
