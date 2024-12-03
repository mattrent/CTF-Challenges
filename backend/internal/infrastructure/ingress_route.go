package infrastructure

import (
	"deployer/config"
	"fmt"

	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func BuildHttpIngress(namespace string, challengeDomain string) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "challenge-http-ingress",
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &config.Values.IngressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: challengeDomain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: ptr.To(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "web",
											Port: networkingv1.ServiceBackendPort{
												Number: 8080,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: "*." + challengeDomain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: ptr.To(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "web",
											Port: networkingv1.ServiceBackendPort{
												Number: 8080,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func BuildHttpsIngress(namespace string, challengeDomain string) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "challenge-https-ingress",
			Namespace: namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/backend-protocol":   "HTTPS",
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"cert-manager.io/issuer":                         "step-issuer",
				"cert-manager.io/issuer-kind":                    "StepIssuer",
				"cert-manager.io/issuer-group":                   "certmanager.step.sm",
				//&config.Values.IngressAnnotations,
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &config.Values.IngressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: challengeDomain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: ptr.To(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "webs",
											Port: networkingv1.ServiceBackendPort{
												Number: 8443,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: "*." + challengeDomain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: ptr.To(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "webs",
											Port: networkingv1.ServiceBackendPort{
												Number: 8443,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: []string{
						challengeDomain,
					},
					SecretName: config.Values.IngressTlsSecretName,
				},
			},
		},
	}
}

func BuildHttpsIngressRoute(namespace string, challengeDomain string) *v1alpha1.IngressRouteTCP {
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
				Match: fmt.Sprintf("HostSNI(`%s`) || HostSNIRegexp(`{anydomain:.*}.%s`)", challengeDomain, challengeDomain),
				Services: []v1alpha1.ServiceTCP{
					{
						Name: "webs",
						Port: intstr.IntOrString{IntVal: 8443},
					},
				},
			}},
			TLS: &v1alpha1.TLSTCP{
				Passthrough: true,
			},
		},
	}
}

func BuildHttpIngressRoute(namespace string, challengeDomain string) *v1alpha1.IngressRoute {
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
				Match: fmt.Sprintf("Host(`%s`) || HostRegexp(`{anydomain:.*}.%s`)", challengeDomain, challengeDomain),
				Services: []v1alpha1.Service{
					{
						LoadBalancerSpec: v1alpha1.LoadBalancerSpec{
							Name: "web",
							Kind: "Service",
							Port: intstr.IntOrString{IntVal: 8080},
						},
					},
				},
			}},
		},
	}
}
