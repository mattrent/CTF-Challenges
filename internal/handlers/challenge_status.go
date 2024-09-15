package handlers

import (
	"context"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetChallengeStatus(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)

	_, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Challenge not found"})
		return
	}

	kubeconfig := infrastructure.GetKubeConfigSingleton()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	namespace := infrastructure.GetNamespaceName(userId, challengeId)
	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, status := range pods.Items[0].Status.ContainerStatuses {
		if status.Name == "compute" {
			c.JSON(http.StatusOK, gin.H{"ready": status.Ready})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Challenge not found"})
}
