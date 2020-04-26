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
	"time"
)

const (
	lockName     string        = "k8pgkuaylock"
	lockDuration time.Duration = 10 * time.Second
)

type k8sLockable interface {
	GetAnnotations() (map[string]string, error)
	SetAnnotationsIf(map[string]string, map[string]string) error
	Refresh() error
}

func getLockInfo(kl k8sLockable) (string, bool, error) {
	var locked = false
	kl.Refresh()
	annotations, err := kl.GetAnnotations()
	if err != nil {
		return "", locked, err
	}
	lockingPodName, ok := annotations[lockName+"_podname"]
	if !ok {
		return "", locked, nil
	}
	if lockingPodName == "" {
		return "", locked, nil
	}

	strLockTime, ok := annotations[lockName+"_time"]
	if !ok {
		return lockingPodName, true, nil
	}
	if strLockTime == "" {
		return lockingPodName, true, nil
	}
	lockTime, err := time.Parse(time.RFC3339, strLockTime)
	if err != nil {
		return "", false, err
	}
	if time.Now().Before(lockTime) {
		return lockingPodName, true, nil
	}

	return lockingPodName, false, nil
}

// Lock function can do all required to set a Locking annotation on a K8s object
func Lock(kl k8sLockable) (bool, error) {
	lockingPod, locked, err := getLockInfo(kl)
	if err != nil {
		return false, err
	}
	podname, err := FullQualifiedPodName()
	if err != nil {
		return false, err
	}
	if lockingPod != podname && locked {
		return false, fmt.Errorf("k8sLockable.Lock KubeStore.SetClusterState err: There already is a lock from %s", lockingPod)
	}

	var check = map[string]string{}
	check[lockName+"_podname"] = lockingPod
	lockExpires := time.Now().Add(lockDuration).Format(time.RFC3339)
	set := map[string]string{lockName + "_time": lockExpires, lockName + "_podname": podname}

	err = kl.SetAnnotationsIf(check, set)
	if err != nil {
		sugar.Debugf("err: %v", err)
		return false, err
	}
	lockingPod, locked, err = getLockInfo(kl)
	if err != nil {
		return false, err
	}
	if podname == lockingPod && locked {
		return true, nil
	}
	return false, fmt.Errorf("Could not get a lock: me: %v, locking pod: %v, locked: %v", podname, lockingPod, locked)
}

// UnLock function can do all required to remove a Locking annotation from a K8s object
func UnLock(kl k8sLockable) (bool, error) {
	lockingPod, _, err := getLockInfo(kl)
	if err != nil {
		return false, err
	}
	podname, err := FullQualifiedPodName()
	if err != nil {
		return false, err
	}
	if podname != lockingPod {
		return false, fmt.Errorf("I was not locking")
	}
	check := map[string]string{lockName + "_podname": podname}
	set := map[string]string{lockName + "_time": "", lockName + "_podname": ""}
	err = kl.SetAnnotationsIf(check, set)
	if err != nil {
		return false, err
	}
	lockingPod, _, err = getLockInfo(kl)
	if err != nil {
		return false, err
	}
	if podname == "" {
		return true, nil
	}
	return false, fmt.Errorf("Could not release the lock")
}
