package infrastructure

import (
	"fmt"

	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildHttpsIngressRoute(namespace string, challengeUrl string) *v1alpha1.IngressRouteTCP {
	// Traefik TCP Ingress route for port 443
	return &v1alpha1.IngressRouteTCP{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IngressRouteTCP",
			APIVersion: "traefik.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpsingress",
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: v1alpha1.IngressRouteTCPSpec{
			EntryPoints: []string{"websecure"},
			Routes: []v1alpha1.RouteTCP{{
				Match: fmt.Sprintf("HostSNI(`%s`) || HostSNIRegexp(`{anydomain:.*}.%s`)", challengeUrl, challengeUrl),
				Services: []v1alpha1.ServiceTCP{
					{
						Name: "webs",
						Port: intstr.IntOrString{IntVal: 443},
					},
				},
			}},
			TLS: &v1alpha1.TLSTCP{
				Passthrough: true,
			},
		},
	}
}

func BuildHttpIngressRoute(namespace string, challengeUrl string) *v1alpha1.IngressRoute {
	// Traefik ingress route for port 80
	return &v1alpha1.IngressRoute{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IngressRoute",
			APIVersion: "traefik.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httpingress",
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: v1alpha1.IngressRouteSpec{
			EntryPoints: []string{"web"},
			Routes: []v1alpha1.Route{{
				Kind:  "Rule",
				Match: fmt.Sprintf("Host(`%s`) || HostRegexp(`{anydomain:.*}.%s`)", challengeUrl, challengeUrl),
				Services: []v1alpha1.Service{
					{
						LoadBalancerSpec: v1alpha1.LoadBalancerSpec{
							Name: "web",
							Kind: "Service",
							Port: intstr.IntOrString{IntVal: 80},
						},
					},
				},
			}},
		},
	}
}
