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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Configmap resembles a Configmap Object. It has a record of the K8s Store, so it can update itself.
type Configmap struct {
	client    *kubernetes.Clientset
	iface     typedcorev1.ConfigMapInterface
	object    *apicorev1.ConfigMap
	Namespace string
	Name      string
	valid     bool
	stopWatch chan struct{}
}

// NewK8sConfigMap function returns a newly init'ed NewConfigMap object
func NewK8sConfigMap(client *kubernetes.Clientset, namespace string, name string) *Configmap {
	obj := Configmap{
		client:    client,
		Namespace: namespace,
		Name:      name,
		valid:     false,
	}
	obj.connect()
	return &obj
}

func (kc *Configmap) connect() error {
	sugar.Debugf("Configmap.connect Namespace %s", kc.Namespace)
	sugar.Debugf("Configmap.connect Name %s", kc.Name)
	if kc.valid {
		return nil
	}
	iface := kc.client.CoreV1().ConfigMaps(kc.Namespace)
	object, err := iface.Get(kc.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get latest version of Configmap %s: %v", kc.Name, err)
	}
	if err != nil {
		return err
	}
	kc.iface = iface
	kc.object = object
	return nil
}

// Refresh method will re-read the latest info on this ConfigMap object
func (kc *Configmap) Refresh() error {
	kc.valid = false
	return kc.connect()
}

func (kc *Configmap) update() error {
	object, err := kc.iface.Update(kc.object)
	if err != nil {
		return err
	}
	kc.object = object
	return nil
}

// GetData method will return the data in the configmap
func (kc *Configmap) GetData() (map[string]string, error) {
	err := kc.connect()
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to get latest version of Configmap %s: %v", kc.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		// configmap exists
		return kc.object.Data, nil
	}
	// configmap does not exist
	return map[string]string{}, err
}

// SetData method will update the data in the configmap
func (kc *Configmap) SetData(Check map[string]string, Set map[string]string) error {
	err := kc.connect()
	if err != nil {
		return fmt.Errorf("failed to get latest version of Configmap %s: %v", kc.Name, err)
	}
	for key, value := range Check {
		x, found := kc.object.Annotations[key]
		if !found {
			return fmt.Errorf("Configmap %s is missing key (%s) in annotations", kc.Name, key)
		}
		if x != value {
			return fmt.Errorf("Configmap %s has unexpected value (%s=%s!=%s) in annotations: %v", kc.Name, key, value, x, kc.object.Annotations)
		}
	}

	for key, value := range Set {
		kc.object.Data[key] = value
	}
	kc.update()

	//KubeClusterDataKey is not set in configmap.Data. Return empty string
	return nil
}

// GetAnnotations method will return the annotations set on this configmap
func (kc *Configmap) GetAnnotations() (map[string]string, error) {
	err := kc.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version of Configmap %s: %v", kc.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		// configmap exists
		return kc.object.Annotations, nil
	}
	// configmap does not exist
	return map[string]string{}, err
}

// SetAnnotationsIf method will update the annotations on this configmap
func (kc *Configmap) SetAnnotationsIf(Check map[string]string, Set map[string]string) error {
	err := kc.connect()
	if err != nil {
		return fmt.Errorf("failed to get latest version of Configmap %s: %v", kc.Name, err)
	}
	if !apierrors.IsNotFound(err) {
		for key, value := range Check {
			x, found := kc.object.Annotations[key]
			if !found {
				return fmt.Errorf("Configmap %s is missing key (%s) in annotations", kc.Name, key)
			}
			if x != value {
				return fmt.Errorf("Configmap %s has unexpected value (%s=%s!=%s) in annotations: %v", kc.Name, key, value, x, kc.object.Annotations)
			}
		}

		for key, value := range Set {
			kc.object.Annotations[key] = value
		}
		kc.update()

		//KubeClusterDataKey is not set in configmap.Data. Return empty string
		return nil
	}
	// configmap does not exist
	return fmt.Errorf("Configmap %s does not exist: %v", kc.Name, err)
}

// Lock method will Lock this configmap so others will not update it
func (kc *Configmap) Lock() (bool, error) {
	return Lock(kc)
}

// UnLock method will UnLock this configmap
func (kc *Configmap) UnLock() (bool, error) {
	return UnLock(kc)
}

// Watch keeps track of changes on this object and taes action accordingly.
// This will probably run as a seperate thread (next to other checks).
// Examples:
// - Change of numstandys: reload hba, go to UP/DEGRADED state
// - Change clusterstate=SWITCHOVER: trigger switchover process
// Below comes from example at: https://stackoverflow.com/questions/40975307/how-to-watch-events-on-a-kubernetes-service-using-its-go-client

// StartWatch method can be used to start tracking changes on this configmap
func (kc *Configmap) StartWatch() {
	fieldSelector := fields.OneTermEqualSelector("metadata.name", kc.Name)

	watchlist := cache.NewListWatchFromClient(kc.client.CoreV1().RESTClient(), "configmaps", kc.Namespace, fieldSelector)

	_, controller := cache.NewInformer(
		watchlist,
		&apicorev1.ConfigMap{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc:    nil,
			DeleteFunc: nil,
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldConfigMap := oldObj.(*apicorev1.ConfigMap)
				newConfigMap := newObj.(*apicorev1.ConfigMap)
				if oldConfigMap.ObjectMeta.ResourceVersion != newConfigMap.ObjectMeta.ResourceVersion {
					kc.object = newConfigMap
					fmt.Printf("Configmap changed:\n%v\n%v\n\n", oldConfigMap.ObjectMeta.ResourceVersion, newConfigMap.ObjectMeta.ResourceVersion)
				}
			},
		},
	)
	// I found it in k8s scheduler module. Maybe it's help if you interested in.
	// serviceInformer := cache.NewSharedIndexInformer(watchlist, &v1.Service{}, 0, cache.Indexers{
	//     cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	// })
	// go serviceInformer.Run(stop)
	kc.stopWatch = make(chan struct{})
	go controller.Run(kc.stopWatch)
}

// StopWatch method can be used to stop tracking the configmap changes
func (kc *Configmap) StopWatch() {
	if kc.stopWatch != nil {
		close(kc.stopWatch)
		kc.stopWatch = nil
	}
}

// Example: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubeletconfig/watch.go #newSharedNodeInformer
// Used by https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/vendor/k8s.io/kubernetes/pkg/kubelet/kubeletconfig/controller.go #StartSync

// func (kc *Configmap) Watch() cache.SharedInformer {
// 	// select nodes by name
// 	fieldSelector := fields.OneTermEqualSelector("metadata.name", kc.Name)
//
// 	// add some randomness to resync period, which can help avoid controllers falling into lock-step
// 	minResyncPeriod := 15 * time.Minute
// 	factor := rand.Float64() + 1
// 	resyncPeriod := time.Duration(float64(minResyncPeriod.Nanoseconds()) * factor)
//
// 	lw := cache.NewListWatchFromClient(kc.client.CoreV1().RESTClient(), "configmaps", kc.Namespace, fieldSelector)
//
// 	handler := cache.ResourceEventHandlerFuncs{
// 		AddFunc:    addFunc,
// 		UpdateFunc: updateFunc,
// 		DeleteFunc: deleteFunc,
// 	}
//
// 	informer := cache.NewSharedInformer(lw, &apicorev1.ConfigMap{}, resyncPeriod)
// 	informer.AddEventHandler(handler)
//
// 	return informer
// }

// Another example: https://medium.com/programming-kubernetes/building-stuff-with-the-kubernetes-api-part-4-using-go-b1d0e3c1c899
