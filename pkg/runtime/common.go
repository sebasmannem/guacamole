package runtime

import (
	"fmt"
	"github.com/sebasmannem/k8pgquay/pkg/common"
	"github.com/sebasmannem/k8pgquay/pkg/k8s"
	"github.com/sebasmannem/k8pgquay/pkg/postgresql"
	"github.com/sebasmannem/k8pgquay/pkg/state"
	"go.uber.org/zap"
	"strings"
)

var (
	args         *common.ArgConfig
	sugar        *zap.SugaredLogger
	k8sKubeStore *k8s.KubeStore
	config       *state.Config
	pgManager    *postgresql.Manager
)

const (
	// RunModeUnknown should not occur
	RunModeUnknown uint64 = iota
	// RunModeInit means we are running in a init container
	RunModeInit
	// RunModeAPI means we are running as a (REST?) API
	RunModeAPI
	// RunModeSidecar means we are running as a sidecar, managing Postgres from the side
	RunModeSidecar
	// RunModeParent is meant that we run as a parent of Postgres. It is not implemented...
	RunModeParent
)

// Initialize function is the main entrypoint to run as a init container
func Initialize(arguments *common.ArgConfig, log *zap.SugaredLogger) error {
	args = arguments
	sugar = log

	k8s.InitializeK8s(sugar)
	postgresql.Initialize(sugar)

	err := InitK8s()
	if err != nil {
		return err
	}

	initPostgres()

	return nil
}

// InitK8s function is the main entrypoint to run as a init container
func InitK8s() error {
	store, err := k8s.NewStore(args)
	if err == nil {
		k8sKubeStore = store
	} else {
		return err
	}

	clusterconfig, err := store.GetClusterConfigmapData()
	if err != nil {
		sugar.Errorf("unable to get cluster config from configmap: %v", err)
	}
	sugar.Debugf("runtime.Init clusterconfig %s", clusterconfig)

	stateconfig := state.NewConfig()
	stateconfig.LoadFromHash(clusterconfig)
	config = stateconfig

	clusterstate, err := k8sKubeStore.GetClusterState()
	if err != nil {
		return err
	}

	sugar.Infof("Postgres Port: %s", config.PgPort)
	sugar.Infof("Postgres local hba: %s", config.LocalPGHba)
	sugar.Infof("Logging Level %s", config.LogLevel)
	sugar.Infof("Postgres BIN Directory: %s", config.BinDirectoryPtr)
	sugar.Infof("Postgres Data Directory: %s", config.DataDirectoryPtr)
	sugar.Infof("Clusterstate: %s", clusterstate.String())

	return nil
}

func initPostgres() {
	var initconfig postgresql.InitConfig
	initconfig.Locale = config.Locale
	initconfig.Encoding = config.Encoding
	initconfig.DataChecksums = config.DataChecksums
	pgManager = postgresql.NewManager(config.BinDirectoryPtr, config.DataDirectoryPtr, config.IsEDB, config.AuthMethod, config.SuDBConnection, config.ReplAuthMethod, config.ReplConnection, initconfig)
}

// Run function will pick the right function to run, and run it
func Run() error {
	var err error
	switch strings.ToLower(args.RunMode) {
	case "init":
		err = Init()
	case "sidecar":
		err = Sidecar()
	case "api":
		err = API()
	case "parent":
		err = fmt.Errorf("run as parent pid to postgres currently is not implemented")
	default:
		err = fmt.Errorf("unknown runmode, please set one of init, api, sidecar")
	}

	return err
}

func setReplicas() error {
	// Set pgManager.ClusterNodes
	statefulsetname, err := k8s.StatefulSetName()
	if err != nil {
		return err
	}
	numReplicas, err := k8sKubeStore.GetNumReplicas()
	if err != nil {
		return err
	}
	for i := int32(0); i < *numReplicas; i++ {
		node := fmt.Sprintf("%s-%d", statefulsetname, i)
		pod := k8sKubeStore.GetPod(node)
		ip, err := pod.GetIP()
		if err != nil {
			return err
		}
		if ip == "" {
			continue
		}
		err = pgManager.AddReplica(ip)
		if err != nil {
			return err
		}
	}
	changed, err := pgManager.WritePgHba()
	if err != nil {
		return err
	}
	if changed {
		
	}

	return nil
}
