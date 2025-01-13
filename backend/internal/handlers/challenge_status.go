package handlers

import (
	"deployer/config"
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ChallengeStatus godoc
// @Summary      Challenge Status
// @Tags         challenges
// @Param        id	path		string				true	"Challenge ID"
// @Accept       json
// @Produce      json
// @Router       /challenges/{id}/status [get]
// @Security BearerAuth
func GetChallengeStatus(c *gin.Context) {
	challengeId := c.Param("id")
	userId := auth.GetCurrentUserId(c)

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

	instanceId, err := infrastructure.GetRunningInstanceId(c, userId, challenge.Id)
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

	namespace := infrastructure.GetNamespaceName(instanceId)
	pods, err := clientset.CoreV1().Pods(namespace).List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ns, err := clientset.CoreV1().Namespaces().Get(c, namespace, metav1.GetOptions{})
	if err != nil {
		log.Println("Could not get namespace: " + namespace + " " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"started": false,
		})
		return
	}

	for _, status := range pods.Items[0].Status.ContainerStatuses {
		if status.Name == "compute" {
			c.JSON(http.StatusOK, gin.H{
				"url":         getChallengeDomain(instanceId),
				"ready":       status.Ready,
				"secondsleft": int(((time.Minute * time.Duration(config.Values.ChallengeLifetimeMinutes)) - time.Since(ns.CreationTimestamp.Time)).Seconds()),
				"started":     true,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Challenge not found"})
}
