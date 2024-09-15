package handlers

import (
	"context"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"github.com/gin-gonic/gin"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

func GetChallengeLogs(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)

	challenge, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge.UserId != userId {
		c.JSON(http.StatusUnauthorized, gin.H{})
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

	req := clientset.CoreV1().Pods(namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{Container: "guest-console-log"})
	logReader, err := req.Stream(context.Background())
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
