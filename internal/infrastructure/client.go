package infrastructure

import (
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	scalescheme "k8s.io/client-go/scale/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
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
		infrav1.AddToScheme,
		scalescheme.AddToScheme,
		clusterv1.AddToScheme,
		cdiv1.AddToScheme,
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
