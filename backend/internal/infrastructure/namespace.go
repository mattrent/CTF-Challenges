package infrastructure

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var challengeNamespacePrefix = "challenge-"
var namespaceLabelChallengeId = "challengeid"
var namespaceLabelInstanceId = "instanceid"
var namespaceLabelPlayerId = "playerid"

func GetNamespaceName(instanceId string) string {
	return challengeNamespacePrefix + instanceId[0:18]
}

func GetRunningInstanceId(c *gin.Context, userId, challengeId string) (string, error) {
	kubeconfig := GetKubeConfigSingleton()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return "", err
	}
	nsList, err := clientset.CoreV1().Namespaces().List(c, metav1.ListOptions{
		LabelSelector: namespaceLabelPlayerId + "=" + userId + "," + namespaceLabelChallengeId + "=" + challengeId,
	})
	if err != nil {
		return "", err
	}
	if len(nsList.Items) > 0 && nsList.Items[0].Status.Phase != corev1.NamespaceTerminating {
		return nsList.Items[0].Labels[namespaceLabelInstanceId], nil
	}
	return "", nil
}

func BuildNamespace(name, challengeId, instanceid, playerId string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				namespaceLabelChallengeId: challengeId,
				namespaceLabelInstanceId:  instanceid,
				namespaceLabelPlayerId:    playerId,
			},
		},
	}
	return ns
}
