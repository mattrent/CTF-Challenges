package handlers

import (
	"deployer/config"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetChallengeStatus(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)

	var challenge storage.Challenge
	if id, err := strconv.Atoi(challengeId); err == nil {
		challenge, err = storage.GetChallengeByCtfdId(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "CTFd challenge not found"})
			return
		}
	} else {
		challenge, err = storage.GetChallenge(challengeId)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Challenge not found"})
			return
		}
	}

	instance, err := storage.GetInstanceByPlayer(challenge.Id, userId)
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

	namespace := infrastructure.GetNamespaceName(userId, challenge.Id)
	pods, err := clientset.CoreV1().Pods(namespace).List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ns, err := clientset.CoreV1().Namespaces().Get(c, namespace, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"started": false,
		})
		return
	}

	for _, status := range pods.Items[0].Status.ContainerStatuses {
		if status.Name == "compute" {
			c.JSON(http.StatusOK, gin.H{
				"url":         getChallengeUrl(instance.Id),
				"ready":       status.Ready,
				"secondsleft": int(((time.Minute * time.Duration(config.Values.ChallengeLifetimeMinutes)) - time.Since(ns.CreationTimestamp.Time)).Seconds()),
				"started":     true,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Challenge not found"})
}
