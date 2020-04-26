package common

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"time"
)

// ArgConfig holds all config required to connect to K8s and read a configmap
type ArgConfig struct {
	RunMode        string
	KubeConfigPath string
	KubeContext    string
	ConfigMap      string
	Timeout        time.Duration
	PrintVersion   bool
}

var (
	runModeEnum uint64
	runMode     string
)

// NewArgConfig function to create a new (initialized) ArgConfig
func NewArgConfig() (*ArgConfig, error) {
	kingpin.CommandLine.HelpFlag.Short('h')
	versionVar := kingpin.Flag("version", "print version").Short('v').Bool()
	runMode := kingpin.Flag("runmode", "The runmode (init, rest, sidecar)").Default("sidecar").String()
	configmap := kingpin.Flag("configmap", "The name of the config map to use").Default("configdata").String()
	timeout := kingpin.Flag("timeout", "Timeout for connecting to K8s").Default("5").Int()
	kubeconfigpath := kingpin.Flag("kubeconfigpath", "The path to Kubeconfig").String()
	kubecontext := kingpin.Flag("kubecontext", "The kube context").String()
	kingpin.Parse()

	return &ArgConfig{
		RunMode:        *runMode,
		KubeConfigPath: *kubeconfigpath,
		KubeContext:    *kubecontext,
		ConfigMap:      *configmap,
		Timeout:        time.Duration(*timeout) * time.Second,
		PrintVersion:   *versionVar,
	}, nil
}
