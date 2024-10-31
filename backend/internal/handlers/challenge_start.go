package handlers

import (
	"context"
	"deployer/config"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StartChallengeResponse struct {
	Url         string `json:"url"`
	SecondsLeft int    `json:"secondslseft"`
	Started     bool   `json:"started"`
}

func StartChallenge(c *gin.Context) {
	userId := c.GetString(userIdValue)
	challengeId := c.Param("id")

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

	if challenge.UserId != userId && !challenge.Published {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	instanceId, token, err := storage.CreateInstance(userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := createResources(c, userId, challenge.Id, instanceId, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func getChallengeUrl(instanceId string) string {
	return instanceId + config.Values.ChallengeDomain
}

func createResources(ctx context.Context, userId, challengeId, instanceId, token string) (*StartChallengeResponse, error) {
	challengeUrl := getChallengeUrl(instanceId)

	kubeClient, err := infrastructure.CreateClient()
	if err != nil {
		return nil, err
	}

	// Create resources
	name := infrastructure.GetNamespaceName(userId, challengeId)
	ns := infrastructure.BuildNamespace(name)
	resources := []client.Object{
		ns,
		infrastructure.BuildNetworkPolicy(ns),
		infrastructure.BuildVm(challengeId, token, ns.Name, challengeUrl),
		infrastructure.BuildHttpService(ns.Name),
		infrastructure.BuildHttpsService(ns.Name),
		infrastructure.BuildHttpIngress(ns.Name, challengeUrl),
		infrastructure.BuildHttpsIngress(ns.Name, challengeUrl),
		//infrastructure.BuildHttpsIngressRoute(ns.Name, challengeUrl),
		//infrastructure.BuildHttpIngressRoute(ns.Name, challengeUrl),
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
		Url:         challengeUrl,
		SecondsLeft: int((time.Minute * time.Duration(config.Values.ChallengeLifetimeMinutes)).Seconds()),
		Started:     true,
	}, nil
}
