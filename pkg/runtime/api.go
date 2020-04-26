package runtime

import (
	"github.com/gin-gonic/gin"
	"github.com/sebasmannem/k8pgquay/pkg/services"
)

// API function is the main entrypoint to run as a init container
func API() error {
	services.RequestInit(sugar, k8sKubeStore, pgManager)

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	routerAPI := gin.Default()

	sugar.Info("Defining Routes")

	v1 := routerAPI.Group("/api/v1")
	{
		manage := v1.Group("/manage")
		{
			manage.GET("/start", services.ManageStart)
			manage.GET("/initdb", services.ManageInitDb)
			manage.GET("/clonedb", services.ManageCloneDb)
			manage.GET("/stop", services.ManageStop)
			manage.GET("/reload", services.ManageReload)
			manage.GET("/restart", services.ManageRestart)
			manage.PUT("/promote", services.ManagePromote)
			manage.GET("/get_pghba", services.ManageGetHba)
			manage.POST("/set_pghba", services.ManageSetHba)
			manage.PATCH("/apply_pghba", services.ManageApplyHba)
			manage.GET("/get_pgconf", services.ManageGetPostgresConf)
			manage.GET("/get_clusterdata", services.ManageGetK8sClusterData)
			manage.GET("/get_clusterstate", services.ManageGetK8sClusterState)
			manage.GET("/set_clusterstate/:clusterstate", services.ManageSetK8sClusterState)
			manage.GET("/get_poddata", services.ManageGetK8sPodData)
			manage.GET("/get_replication_slots", services.ManageGetReplicationSlots)
			manage.PUT("/create_replication_slot/:slotname", services.ManageCreateReplicationSlot)
			manage.DELETE("/drop_replication_slot/:slotname", services.ManageDropReplicationSlot)

		}

		status := v1.Group("/status")
		{
			status.GET("/is_initialized", services.StatusIsInitialized)
			status.GET("/is_started", services.StatusIsStarted)
			status.GET("/ping", services.StatusPing)
			status.GET("/binary_version", services.StatusGetBinaryVersion)
			status.GET("/data_version", services.StatusGetDataVersion)
		}

	}

	sugar.Info("Starting Server")
	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	err := routerAPI.Run()
	if err != nil {
		sugar.Errorf("Error running API server: %v", err)
		// router.Run(":3000") for a hard coded port
	}
	return err
}
