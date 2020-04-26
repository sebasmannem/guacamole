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

package k8s

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

const (
	kubePodName   string = "HOSTNAME"
	kubeNameSpace string = "POD_NAMESPACE"
)

const (
	// KubeClusterDataKey holds the name of the key in the annotations that hold the cluster config
	KubeClusterDataKey string = "clusterconfig"
	// ClusterStateName holds the name of the key in the annotations that hold the cluster state
	ClusterStateName string = "ClusterState"
)

var (
	namespace       string
	podname         string
	statefulsetname string
	podID           int
)

// PodName derives and returns the PodName from the environment variables
func PodName() (string, error) {
	if podname != "" {
		return podname, nil
	}
	podname = os.Getenv(kubePodName)
	if len(podname) == 0 {
		return "", fmt.Errorf("missing required env variable %q", kubePodName)
	}
	return podname, nil
}

// NameSpace derives and returns the NameSpace from the environment variables
func NameSpace() (string, error) {
	if namespace != "" {
		return namespace, nil
	}
	namespace = os.Getenv(kubeNameSpace)
	if len(namespace) == 0 {
		return "", fmt.Errorf("missing required env variable %q", kubeNameSpace)
	}
	return namespace, nil
}

// StatefulSetName derives and returns the StatefulSetName from the PodName
func StatefulSetName() (string, error) {
	if statefulsetname != "" {
		return statefulsetname, nil
	}

	podName, err := PodName()
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile(`-[0-9]+$`)
	statefulsetname = reg.ReplaceAllString(podName, "")
	if statefulsetname == podname {
		return "", fmt.Errorf("pod name same as statefulset name. Is this a statefulset?")
	}
	return statefulsetname, nil
}

// PodID derives the ID from a POD name that is part of a statefulset.
func PodID() (int, error) {
	if podID >= 0 {
		return podID, nil
	}

	podName, err := PodName()
	if err != nil {
		return -1, err
	}
	reg := regexp.MustCompile(`-[0-9]+$`)
	strPodID := reg.FindString(podName)
	if strPodID == "" {
		return -1, fmt.Errorf("pod name does not end in dashed numeric string. Is this a statefulset?")
	}
	podID, err := strconv.Atoi(strPodID[1:])
	if err != nil {
		return -1, fmt.Errorf("The podid could not be cast to an integer")
	}
	return podID, nil
}

// FullQualifiedPodName formats a identifier from the NameSpace and PodName
func FullQualifiedPodName() (string, error) {
	podName, err := PodName()
	if err != nil {
		return "", err
	}
	namespace, err := NameSpace()
	if err != nil {
		return "", err
	}
	return namespace + "." + podName, nil
}
