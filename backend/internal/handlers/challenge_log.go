package handlers

import (
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// @Summary Get challenge logs
// @Description Returns the logs of a challenge
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {string} string "Logs"
// @Failure 401 {object} handlers.ErrorResponse
// @Router /challenges/{id}/logs [get]
func GetChallengeLogs(c *gin.Context) {
	challengeId := c.Param("id")
	userId := auth.GetCurrentUserId(c)

	challenge, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge.UserId != userId && !auth.IsAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	instanceId, err := infrastructure.GetRunningChallengeInstanceId(c, userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if instanceId == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "Challenge instance not running"})
		return
	}

	kubeconfig := infrastructure.GetKubeConfigSingleton()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	namespace := infrastructure.GetNamespaceNameChallenge(instanceId)
	pods, err := clientset.CoreV1().Pods(namespace).List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO check if this works with containers
	req := clientset.CoreV1().Pods(namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{Container: "guest-console-log"})
	logReader, err := req.Stream(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer logReader.Close()
	b, err := io.ReadAll(logReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, string(b))
}
