package handlers

import (
	"context"
	"deployer/config"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StartChallengeResponse struct {
	Url string
}

func StartChallenge(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)

	challenge, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Challenge not found",
		})
		return
	}
	if challenge.UserId != userId && !challenge.Published {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	instanceId, token, err := storage.CreateInstance(userId, challengeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := createResources(userId, challengeId, instanceId, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func createResources(userId, challengeId, instanceId, token string) (*StartChallengeResponse, error) {
	challengeUrl := instanceId + config.Values.ChallengeDomain

	kubeClient, err := infrastructure.CreateClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create resources
	name := infrastructure.GetNamespaceName(userId, challengeId)
	ns := infrastructure.BuildNamespace(name)
	resources := []client.Object{
		ns,
		infrastructure.BuildVm(challengeId, token, ns.Name),
		infrastructure.BuildHttpService(ns.Name),
		infrastructure.BuildHttpsService(ns.Name),
		infrastructure.BuildHttpsIngressRoute(ns.Name, challengeUrl),
		infrastructure.BuildHttpIngressRoute(ns.Name, challengeUrl),
	}

	for _, val := range resources {
		err = kubeClient.Create(ctx, val)
		if err != nil {
			log.Println(err.Error())
			_ = kubeClient.Delete(ctx, ns)
			return nil, err
		}
	}

	log.Println("Started: " + challengeUrl)
	return &StartChallengeResponse{
		Url: challengeUrl,
	}, nil
}
