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

// MemberState is an ENUM of all states the cluster could have
type MemberState int

// This is all the states the ENUM could have
const (
	MemberStateUnknown MemberState = iota
	MemberStateInit
	MemberStateStandby
	MemberStateReinstate
	MemberStateReconfigS
	MemberStateRestart
	MemberStateMaster
	MemberStateReconfigM
	MemberStateDemote
	MemberStateStop
)

//MemberStates needs to fit to above defined ENUM
var MemberStates = []string{"UNKNOWN", "INIT", "STANDBY", "REINSTATE", "RECONFIGS", "RESTART", "MASTER", "RECONFIGM", "DEMOTE", "STOP"}

var (
	validTransitions = map[MemberState][]MemberState{
		MemberStateInit:      {MemberStateStandby},
		MemberStateStandby:   {MemberStateReinstate, MemberStateReconfigS, MemberStateMaster},
		MemberStateReinstate: {MemberStateStop},
		MemberStateReconfigS: {MemberStateRestart, MemberStateStandby},
		MemberStateRestart:   {MemberStateStop, MemberStateReconfigM},
		MemberStateMaster:    {MemberStateReconfigM, MemberStateDemote},
		MemberStateReconfigM: {MemberStateRestart, MemberStateDemote, MemberStateMaster},
		MemberStateDemote:    {MemberStateStop},
	}
)

// MemberStateFromString function creates a new state from a string that should hold the state
func MemberStateFromString(sClusterstate string) MemberState {
	if sClusterstate == "" {
		return MemberStateInit
	}
	sClusterstate = strings.ToUpper(sClusterstate)
	for iState, sState := range MemberStates {
		if sState == sClusterstate {

			return MemberState(iState)
		}
	}
	return MemberStateUnknown
}

func (kcs *MemberState) String() string {
	return MemberStates[*kcs]
}

// ValidateNextState method checks if the nextstate would be a valid next state
func (kcs *MemberState) ValidateNextState(nextState MemberState) bool {
	// If this state is the next state, that is always ok
	if *kcs == nextState {
		return true
	}
	if *kcs == MemberStateUnknown {
		return true
	}
	//Derive the valid states for this state from validTransitions hash
	nextStatesList, ok := validTransitions[*kcs]
	if !ok {
		// This state is not in the validTransitions hash. That is a bug. COuld return false, but lets fail hard instead.
		panic(fmt.Sprintf("State %v is not known in validTransitions map", kcs))
	}
	for _, state := range nextStatesList {
		if state == nextState {
			// Found nextState in list from validTransitions. nextState is valid.
			return true
		}
	}
	// nextState not found in list from validTransitions. nextState is not valid.
	return false
}
