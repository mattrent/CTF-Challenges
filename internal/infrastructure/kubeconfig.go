package infrastructure

import (
	"flag"
	"path/filepath"
	"sync"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var instance *rest.Config
var once sync.Once

func GetKubeConfigSingleton() *rest.Config {
	once.Do(func() {
		instance = getKubeConfig()
	})
	return instance
}

func getKubeConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}
