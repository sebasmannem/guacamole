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

	apicorev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Pod represents a Kubernetes Pod as configured in K8s. It has a recrd of the K8s store, so it can update itself.
type Pod struct {
	client    *kubernetes.Clientset
	iface     typedcorev1.PodInterface
	object    *apicorev1.Pod
	Namespace string
	Name      string
	valid     bool
}

// NewK8sPod returns a freshly instantiated K8sPod object.
func NewK8sPod(client *kubernetes.Clientset, namespace string, podname string) *Pod {
	var err error
	if podname == "" {
		podname, err = PodName()
		if err != nil {
			panic(fmt.Sprintf("Cannot detect podname, %v", err))
		}
	}
	obj := Pod{
		client:    client,
		Namespace: namespace,
		Name:      podname,
		valid:     false,
	}
	obj.connect()
	return &obj
}

func (kp *Pod) connect() error {
	sugar.Debugf("Pod.connect Namespace %s", kp.Namespace)
	sugar.Debugf("Pod.connect Name %s", kp.Name)
	if kp.valid {
		return nil
	}
	iface := kp.client.CoreV1().Pods(kp.Namespace)
	object, err := iface.Get(kp.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get latest version of Pod %s: %v", kp.Name, err)
	}
	if err != nil {
		return err
	}
	kp.iface = iface
	kp.object = object
	return nil
}

// Refresh method will re-read the latest info on this Pod object
func (kp *Pod) Refresh() error {
	kp.valid = false
	return kp.connect()
}

func (kp *Pod) update() error {
	object, err := kp.iface.Update(kp.object)
	if err != nil {
		return err
	}
	kp.object = object
	return nil
}

// GetIP method will return the IP address for this Pod
func (kp *Pod) GetIP() (string, error) {
	err := kp.connect()
	if err != nil {
		return "", err
	}
	return kp.object.Status.PodIP, nil
}

// GetAnnotations method will return the annotations set on this Pod
func (kp *Pod) GetAnnotations() (map[string]string, error) {
	err := kp.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version of pod %s: %v", kp.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		// configmap exists
		return kp.object.Annotations, nil
	}
	// configmap does not exist
	return map[string]string{}, err
}

// SetAnnotationsIf method will update the annotations on this Pod
func (kp *Pod) SetAnnotationsIf(Check map[string]string, Set map[string]string) error {
	err := kp.connect()
	if err != nil {
		return fmt.Errorf("failed to get latest version of pod %s: %v", kp.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		for key, value := range Check {
			x, found := kp.object.Annotations[key]
			if !found {
				return fmt.Errorf("Pod %s is missing key (%s) in annotations", kp.Name, key)
			}
			if x != value {
				return fmt.Errorf("Pod %s has unexpected value (%s=%s!=%s) in annotations: %v", kp.Name, key, value, x, kp.object.Annotations)
			}
		}

		for key, value := range Set {
			kp.object.Annotations[key] = value
		}
		kp.update()

		//KubeClusterDataKey is not set in configmap.Data. Return empty string
		return nil
	}
	// configmap does not exist
	return fmt.Errorf("Pod %s does not exist: %v", kp.Name, err)
}

// Lock method will Lock this configmap so others will not update it
func (kp *Pod) Lock() (bool, error) {
	return Lock(kp)
}

// UnLock method will UnLock this configmap
func (kp *Pod) UnLock() (bool, error) {
	return UnLock(kp)
}
