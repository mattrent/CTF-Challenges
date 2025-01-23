package infrastructure

import (
	"deployer/config"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func BuildNetworkPolicy(ns *corev1.Namespace) *networkingv1.NetworkPolicy {
	policy := &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "policy",
			Namespace: ns.Name,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"vm.kubevirt.io/name": "challenge",
				},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "0.0.0.0/0",
						},
					}},
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: &intstr.IntOrString{IntVal: 22}},
						{Port: &intstr.IntOrString{IntVal: 8022}},
						{Port: &intstr.IntOrString{IntVal: 8080}},
						{Port: &intstr.IntOrString{IntVal: 8443}},
					},
				}},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"k8s-app": "kube-dns",
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port:     &intstr.IntOrString{IntVal: 53},
							Protocol: ptr.To(corev1.ProtocolUDP),
						},
					},
				},
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": config.Values.Namespace,
								},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "deployer",
								},
							},
						},
					},
				},
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							IPBlock: &networkingv1.IPBlock{
								CIDR: "0.0.0.0/0",
								Except: []string{
									"10.0.0.0/8",
									"192.168.0.0/16",
									"172.16.0.0/20",
								},
							}},
					},
				}},
		},
	}
	return policy
}
