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

func StopChallenge(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)

	_, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Challenge not found",
		})
		return
	}

	kubeconfig := infrastructure.GetKubeConfigSingleton()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete namespace
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = clientset.CoreV1().Namespaces().Delete(ctx, infrastructure.GetNamespaceName(userId, challengeId), metav1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stopping challenge",
	})
}
