package runtime

// check cluster state
// check pod state
// get config
// check datadir
// loop:
//   check Master
//   check me
//   switch state
//   state specific stuff
//   update cluster state, pod state, config
// exit 0

import (
	"time"
)

const (
	connectnumtries  = 10
	connectwaitsec   = 1
	heartbeatwaitsec = 10
)

// Sidecar function is the main entrypoint to run as a init container
func Sidecar() error {
	clusterconfig, err := k8sKubeStore.GetClusterConfigmapData()
	if err != nil {
		sugar.Errorf("unable to get cluster config from configmap: %v", err)
	}
	sugar.Debugf("runtime.Init clusterconfig %s", clusterconfig)

	for i := 1; i <= connectnumtries; i++ {
		err = pgManager.Connect()
		if err != nil {
			if i < connectnumtries {
				sugar.Debugf("Unable to initialize database connection: %v", err)
			} else {
				sugar.Errorf("Unable to initialize database connection: %v", err)
			}
			//return
		} else {
			sugar.Info("Database Management module initialized")
			break
		}
		time.Sleep(connectwaitsec * time.Second)
	}

	sugar.Debugf("Get Config Map")
	clusterConfigMap := k8sKubeStore.GetClusterConfigmap()
	sugar.Debugf("Watch Config Map")
	clusterConfigMap.StartWatch()
	defer clusterConfigMap.StopWatch()

	for {
		err = pgManager.Ping()
		if err != nil {
			sugar.Debugf("Heartbeat checkup on Postgres failed: %v", err)
		}
		time.Sleep(heartbeatwaitsec * time.Second)
	}

	return nil
}
