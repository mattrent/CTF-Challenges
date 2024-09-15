package infrastructure

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaceName(userId, challengeId string) string {
	return "challenge-" + challengeId[0:13] + "-" + userId[0:13]
}

func BuildNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{},
		},
	}
	return ns
}
