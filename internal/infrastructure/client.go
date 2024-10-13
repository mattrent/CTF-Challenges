package infrastructure

import (
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateClient() (client.Client, error) {
	kubeconfig := GetKubeConfigSingleton()
	schemes, err := registerSchemes()
	if err != nil {
		return nil, err
	}

	kubeClient, err := client.New(kubeconfig, client.Options{
		Scheme: schemes,
	})
	return kubeClient, err
}

func registerSchemes() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	for _, funcScheme := range []func(*runtime.Scheme) error{
		kubevirtv1.AddToScheme,
		kubescheme.AddToScheme,
		traefik.AddToScheme,
	} {
		if err := funcScheme(scheme); err != nil {
			return nil, err
		}
	}
	return scheme, nil
}
