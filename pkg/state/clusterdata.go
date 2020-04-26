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

package state

import (
	"github.com/mitchellh/mapstructure"
	"time"
)

const currentCDFormatVersion uint64 = 1

// ClusterData will hold all informaton on a cluster, like state
type ClusterData struct {
	// ClusterData format version. Used to detect incompatible
	// version and do upgrade. Needs to be bumped when a non
	// backward compatible change is done to the other struct
	// members.
	FormatVersion uint64    `json:"formatVersion"`
	ChangeTime    time.Time `json:"changeTime"`
	ClusterState  string    `json:"ClusterState"`
	Candidates    string    `json:"candidates"`
	Master        string    `json:"master"`
}

// NewClusterData returns a newly init'ed ClusterData
func NewClusterData() *ClusterData {
	return &ClusterData{
		FormatVersion: currentCDFormatVersion,
		ClusterState:  "",
		Candidates:    "[]",
		Master:        "",
	}
}

// LoadFromAnnotations method can load a new CLusterdata from annotations
func (cd *ClusterData) LoadFromAnnotations(Annotations map[string]string) {
	mapstructure.Decode(Annotations, cd)
	state, ok := Annotations["clusterState"]
	if ok {
		cd.ClusterState = state
	} else {
		cd.ClusterState = ""
	}
}
