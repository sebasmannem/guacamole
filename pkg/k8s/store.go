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

	"github.com/sebasmannem/k8pgquay/pkg/common"
	"github.com/sebasmannem/k8pgquay/pkg/state"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var sugar *zap.SugaredLogger

// InitializeK8s function will initialize the K8s module
func InitializeK8s(log *zap.SugaredLogger) {
	sugar = log
}

// KubeStore struct creates a K8s connection on which we can acquire and update k8s object definitions.
type KubeStore struct {
	client    *kubernetes.Clientset
	namespace string
	configmap string
}

// NewKubeStore will return a freshly initiat'ed K8s Store object.
func newKubeStore(kubecli *kubernetes.Clientset, namespace string, configmap string) (*KubeStore, error) {
	return &KubeStore{
		client:    kubecli,
		namespace: namespace,
		configmap: configmap,
	}, nil
}

// NewStore function will do all things required to connect to K8s
func NewStore(cfg *common.ArgConfig) (*KubeStore, error) {
	var s *KubeStore
	namespace, err := NameSpace()
	if err != nil {
		return nil, err
	}

	kubeClientConfig := NewKubeClientConfig(cfg.KubeConfigPath, cfg.KubeContext, namespace)
	kubecfg, err := kubeClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	kubecfg.Timeout = cfg.Timeout
	kubecli, err := kubernetes.NewForConfig(kubecfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create kubernetes client: %v", err)
	}

	s, err = newKubeStore(kubecli, namespace, cfg.ConfigMap)
	if err != nil {
		return nil, fmt.Errorf("cannot create store: %v", err)
	}

	return s, nil
}

// NewKubeClientConfig returns a kube client config that will by default use an
// in cluster client config or, if not available or overriden an external client
// config using the default client behavior used also by kubectl.
func NewKubeClientConfig(kubeconfigPath, context, namespace string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	if kubeconfigPath != "" {
		rules.ExplicitPath = kubeconfigPath
	}

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if context != "" {
		overrides.CurrentContext = context
	}

	if namespace != "" {
		overrides.Context.Namespace = namespace
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}

// GetPodData method retrieves and returns a pod object
func (s *KubeStore) GetPodData() (map[string]string, error) {
	pod := NewK8sPod(s.client, s.namespace, "")
	poddata, err := pod.GetAnnotations()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving pod data %v", err)
	}

	return poddata, nil
}

// GetPod method retrieves and returns a pod object
func (s *KubeStore) GetPod(name string) *Pod {
	pod := NewK8sPod(s.client, s.namespace, name)
	return pod
}

// GetClusterConfigmapData method retrieves and returns a Configmap object
func (s *KubeStore) GetClusterConfigmapData() (map[string]string, error) {
	ConfigMap := NewK8sConfigMap(s.client, s.namespace, s.configmap)
	ConfigMapData, err := ConfigMap.GetData()
	if err != nil {
		return map[string]string{}, err
	}

	//KubeClusterDataKey is not set in configmap.Data. Return empty string
	return ConfigMapData, nil
}

// GetClusterConfigmap method retrieves and returns a Configmap object
func (s *KubeStore) GetClusterConfigmap() *Configmap {
	return NewK8sConfigMap(s.client, s.namespace, s.configmap)
}

// SetClusterState method will set the cluster to the specified state, tetsing if it is feasible
func (s *KubeStore) SetClusterState(newState state.ClusterState) error {
	Statefulset := NewK8sStatefulSet(s.client, s.namespace)
	annotations, err := Statefulset.GetAnnotations()
	if err != nil {
		sugar.Debugf("KubeStore.SetClusterState err: %v", err)
		return err
	}
	clusterstate, ok := annotations[ClusterStateName]
	if !ok {
		clusterstate = ""
	}
	curState := state.ClusterStateFromString(clusterstate)
	if curState == newState {
		return nil
	}
	if !curState.ValidateNextState(newState) {
		return fmt.Errorf("newState %v is invalid as next state from %v", newState, curState)
	}
	err = Statefulset.SetAnnotationsIf(map[string]string{}, map[string]string{ClusterStateName: newState.String()})
	if err != nil {
		sugar.Debugf("KubeStore.SetClusterState err: %v", err)
		return err
	}
	return nil
}

// GetClusterState method will get the current clusterstate
func (s *KubeStore) GetClusterState() (state.ClusterState, error) {
	Statefulset := NewK8sStatefulSet(s.client, s.namespace)
	annotations, err := Statefulset.GetAnnotations()
	if err != nil {
		return state.ClusterStateUnknown, err
	}
	clusterstate, ok := annotations[ClusterStateName]
	if ok {
		return state.ClusterStateFromString(clusterstate), nil
	}
	return state.ClusterStateEmpty, nil
}

// GetNumReplicas method will get the number of replicas in this statefulset
func (s *KubeStore) GetNumReplicas() (*int32, error) {
	Statefulset := NewK8sStatefulSet(s.client, s.namespace)
	return Statefulset.GetNumReplicas()
}
