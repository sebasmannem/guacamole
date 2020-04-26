package runtime

import (
	"fmt"
	"github.com/sebasmannem/k8pgquay/pkg/k8s"
	"github.com/sebasmannem/k8pgquay/pkg/state"
	"os"
)

// Init function is the main entrypoint to run as a init container
func Init() error {
	exists, err := pgManager.IsInitialized()
	if err != nil {
		return err
	}

	if exists {
		// $PGDATA/PG_VERSION exists
		sugar.Infof("Datadir %s seems inited", config.DataDirectoryPtr)
	} else {
		sugar.Infof("Datadir %s does not seem inited", config.DataDirectoryPtr)
	}

	clusterstate, err := k8sKubeStore.GetClusterState()
	if clusterstate == state.ClusterStateEmpty && !exists {
		podID, err := k8s.PodID()
		if err != nil {
			return err
		}
		if podID > 0 {
			return fmt.Errorf("How we be running in pod>0 with a clusterstate of %v?", clusterstate.String())
		}
		err = pgManager.Init()
		if err != nil {
			return err
		}
		err = k8sKubeStore.SetClusterState(state.ClusterStateDown)
		if err != nil {
			os.RemoveAll(config.DataDirectoryPtr)
			return err
		}
	} else if !exists {
		err = pgManager.Clone()
		if err != nil {
			return err
		}
	} else {
		pgManager.ReadConfs()
	}

	setReplicas()
	if err != nil {
		return err
	}

	// In all OK situationsm ReadConfs is read, and we now write, which is odd.
	// We should have an intermediate step that updates configs according to Configmap, POD state, etc.
	// We should probably add that to WriteConfs...
	pgManager.WriteConfs()

	return nil
}
