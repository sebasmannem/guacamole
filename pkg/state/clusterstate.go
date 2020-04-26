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
	"fmt"
	"strings"
)

// ClusterState is an ENUM of all states the cluster could have
type ClusterState int

// This is all the states the ENUM could have
const (
	ClusterStateUnknown ClusterState = iota
	ClusterStateEmpty
	ClusterStateDown
	ClusterStateElection
	ClusterStateDegraded
	ClusterStateRestart
	ClusterStateUp
	ClusterStateSwitchOver
	ClusterStateMasterUnavailable
	ClusterStateFailOver
)

//clusterStates needs to fit to above defined ENUM
var clusterStates = []string{"UNKNOWN", "EMPTY", "DOWN", "ELECTION", "DEGRADED", "RESTART", "UP", "SWITCHOVER", "MASTERUNAVAILABLE", "FAILOVER"}

var (
	validStates = map[ClusterState][]ClusterState{
		ClusterStateEmpty:             {ClusterStateDown},
		ClusterStateDown:              {ClusterStateElection},
		ClusterStateElection:          {ClusterStateDegraded},
		ClusterStateDegraded:          {ClusterStateRestart, ClusterStateUp},
		ClusterStateRestart:           {ClusterStateSwitchOver, ClusterStateDegraded},
		ClusterStateUp:                {ClusterStateDegraded, ClusterStateSwitchOver, ClusterStateMasterUnavailable},
		ClusterStateSwitchOver:        {ClusterStateElection},
		ClusterStateMasterUnavailable: {ClusterStateFailOver, ClusterStateDegraded},
		ClusterStateFailOver:          {ClusterStateElection},
	}
)

// ClusterStateFromString function creates a new state from a string that should hold the state
func ClusterStateFromString(sClusterstate string) ClusterState {
	if sClusterstate == "" {
		return ClusterStateEmpty
	}
	sClusterstate = strings.ToUpper(sClusterstate)
	for iState, sState := range clusterStates {
		if sState == sClusterstate {

			return ClusterState(iState)
		}
	}
	return ClusterStateUnknown
}

func (cs *ClusterState) String() string {
	return clusterStates[*cs]
}

// ValidateNextState method checks if the nextstate would be a valid next state
func (cs *ClusterState) ValidateNextState(nextState ClusterState) bool {
	// If this state is the next state, that is always ok
	if *cs == nextState {
		return true
	}
	if *cs == ClusterStateUnknown {
		return true
	}
	//Derive the valid states for this state from validStates hash
	nextStatesList, ok := validStates[*cs]
	if !ok {
		// This state is not in the validStates hash. That is a bug. COuld return false, but lets fail hard instead.
		panic(fmt.Sprintf("ClusterState %v is not known in validStates map", cs))
	}
	for _, state := range nextStatesList {
		if state == nextState {
			// Found nextState in list from validStates. nextState is valid.
			return true
		}
	}
	// nextState not found in list from validStates. nextState is not valid.
	return false
}
