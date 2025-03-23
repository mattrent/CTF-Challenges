package infrastructure

import (
	"strconv"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var challengeNamespacePrefix = "challenge-"
var testNamespacePrefix = "test-"
var namespaceLabelChallengeId = "challengeid"
var namespaceLabelInstanceId = "instanceid"
var namespaceLabelPlayerId = "playerid"
var testLabel = "testmode"

func GetNamespaceNameChallenge(instanceId string) string {
	return challengeNamespacePrefix + instanceId[0:18]
}

func GetNamespaceNameTest(instanceId string) string {
	return testNamespacePrefix + instanceId[0:18]
}

func getNameSpaces(c *gin.Context, selector string) (*corev1.NamespaceList, error) {
	kubeconfig := GetKubeConfigSingleton()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	nsList, err := clientset.CoreV1().Namespaces().List(c, metav1.ListOptions{
		LabelSelector: selector,
	})
	return nsList, err
}

func GetRunningChallengeInstanceId(c *gin.Context, userId, challengeId string) (string, error) {
	selector := namespaceLabelPlayerId + "=" + userId + "," + namespaceLabelChallengeId + "=" + challengeId + "," + testLabel + "=false"
	nsList, err := getNameSpaces(c, selector)
	if err != nil {
		return "", err
	}
	if len(nsList.Items) > 0 && nsList.Items[0].Status.Phase != corev1.NamespaceTerminating {
		return nsList.Items[0].Labels[namespaceLabelInstanceId], nil
	}
	return "", nil
}

func GetRunningTestInstanceId(c *gin.Context, challengeId string) (string, error) {
	selector := namespaceLabelChallengeId + "=" + challengeId + "," + testLabel + "=true"
	nsList, err := getNameSpaces(c, selector)
	if err != nil {
		return "", err
	}
	if len(nsList.Items) > 0 && nsList.Items[0].Status.Phase != corev1.NamespaceTerminating {
		return nsList.Items[0].Labels[namespaceLabelInstanceId], nil
	}
	return "", nil
}

func BuildNamespace(challengeId, instanceid, playerId string, testMode bool) *corev1.Namespace {

	var name string
	if testMode {
		name = GetNamespaceNameTest(instanceid)
	} else {
		name = GetNamespaceNameChallenge(instanceid)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				namespaceLabelChallengeId: challengeId,
				namespaceLabelInstanceId:  instanceid,
				namespaceLabelPlayerId:    playerId,
				testLabel:                 strconv.FormatBool(testMode),
			},
		},
	}
	return ns
}
