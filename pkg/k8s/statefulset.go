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

	apicorev1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedappv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

var (
	requiresLocks = []string{ClusterStateName}
	isLocked      bool
)

// Statefulset struct represents a statefulset as was read from k8s. It has a record of the K8s Store, so it can update itself.
type Statefulset struct {
	client    *kubernetes.Clientset
	iface     typedappv1.StatefulSetInterface
	object    *apicorev1.StatefulSet
	Namespace string
	Name      string
	valid     bool
}

// NewK8sStatefulSet function returns a newly init'ed NewStatefulset object
func NewK8sStatefulSet(client *kubernetes.Clientset, namespace string) *Statefulset {
	statefulsetname, err := StatefulSetName()
	if err != nil {
		panic(fmt.Sprintf("Cannot detect statefulset name, %v", err))
	}
	obj := Statefulset{
		client:    client,
		Namespace: namespace,
		Name:      statefulsetname,
		valid:     false,
	}
	//obj.connect()
	return &obj
}

func (ks *Statefulset) connect() error {
	if ks.valid {
		return nil
	}
	sugar.Debugf("Statefulset.connect Name %s", ks.Name)
	iface := ks.client.AppsV1().StatefulSets(ks.Namespace)
	object, err := iface.Get(ks.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get latest version of Statefulset %s: %v", ks.Name, err)
	}
	if err != nil {
		return err
	}
	ks.iface = iface
	ks.object = object
	ks.valid = true
	return nil
}

// Refresh method will re-read the latest info on this Statefulset object
func (ks *Statefulset) Refresh() error {
	ks.valid = false
	return ks.connect()
}

func (ks *Statefulset) update() error {
	object, err := ks.iface.Update(ks.object)
	if err != nil {
		return err
	}
	ks.object = object
	return nil
}

// GetAnnotations method will return the annotations set on this Statefulset
func (ks *Statefulset) GetAnnotations() (map[string]string, error) {
	err := ks.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version of Statefulset %s: %v", ks.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		// Statefulset exists
		return ks.object.Annotations, nil
	}
	// Statefulset does not exist
	return map[string]string{}, err
}

// GetNumReplicas method will return the number of replicas that is configured in this Statefulset
func (ks *Statefulset) GetNumReplicas() (*int32, error) {
	err := ks.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version of Statefulset %s: %v", ks.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		// Statefulset exists
		return ks.object.Spec.Replicas, nil
	}
	// Statefulset does not exist
	return nil, err
}

// SetAnnotationsIf method will update the annotations on this Statefulset
func (ks *Statefulset) SetAnnotationsIf(Check map[string]string, Set map[string]string) error {
	var haveLocked = false
	err := ks.connect()
	if err != nil {
		return fmt.Errorf("failed to get latest version of Statefulset %s: %v", ks.Name, err)
	}
	for key, value := range Check {
		x, found := ks.object.Annotations[key]
		if !found {
			//return fmt.Errorf("Statefulset %s is missing key (%s) in annotations", ks.Name, key)
			continue
		}
		if x != value {
			return fmt.Errorf("Statefulset %s has unexpected value (%s=%s!=%s) in annotations: %v", ks.Name, key, value, x, ks.object.Annotations)
		}
	}

	for setKey := range Set {
		for _, requireLockKey := range requiresLocks {
			if setKey == requireLockKey {
				sugar.Debugf("Statefulset.SetAnnotationsIf locking")
				haveLocked, err = ks.Lock()
				if err != nil {
					return err
				}
				break
			}
		}
	}

	if len(ks.object.Annotations) == 0 {
		ks.object.Annotations = make(map[string]string)
	}
	for key, value := range Set {
		ks.object.Annotations[key] = value
	}
	ks.update()

	if haveLocked {
		sugar.Debugf("Statefulset.SetAnnotationsIf unlocking")
		ks.UnLock()
	}

	//KubeClusterDataKey is not set in Statefulset.Data. Return empty string
	return nil
}

// Lock method will Lock this Statefulset so others will not update it
func (ks *Statefulset) Lock() (bool, error) {
	locked, err := Lock(ks)
	if err == nil && locked {
		isLocked = true
	}
	return locked, err
}

// UnLock method will UnLock this Statefulset
func (ks *Statefulset) UnLock() (bool, error) {
	unlocked, err := UnLock(ks)
	if err == nil && unlocked {
		isLocked = false
	}
	return unlocked, err
}
