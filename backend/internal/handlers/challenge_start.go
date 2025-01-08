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

// ChallengeStart godoc
// @Summary      Challenge Start
// @Tags         challenges
// @Param        id	path		string				true	"Challenge ID"
// @Accept       json
// @Produce      json
// @Router       /challenges/{id}/start [post]
// @Security BearerAuth
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

	runningId, err := infrastructure.GetRunningInstanceId(c, userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if runningId != "" {
		c.JSON(http.StatusOK, gin.H{
			"url":     getChallengeDomain(runningId),
			"started": true,
		})
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

func getChallengeDomain(instanceId string) string {
	return instanceId[0:18] + config.Values.ChallengeDomain
}

func createResources(ctx context.Context, userId, challengeId, instanceId, token string) (*StartChallengeResponse, error) {
	challengeDomain := getChallengeDomain(instanceId)

	kubeClient, err := infrastructure.CreateClient()
	if err != nil {
		return nil, err
	}

	// Create resources
	name := infrastructure.GetNamespaceName(instanceId)
	ns := infrastructure.BuildNamespace(name, challengeId, instanceId, userId)
	resources := []client.Object{
		ns,
		infrastructure.BuildVm(challengeId, token, ns.Name, challengeDomain),
		infrastructure.BuildHttpService(ns.Name),
		infrastructure.BuildHttpsService(ns.Name),
		infrastructure.BuildSshService(ns.Name),
		infrastructure.BuildHttpIngress(ns.Name, challengeDomain),
		infrastructure.BuildNetworkPolicy(ns),
		//infrastructure.BuildHttpsIngress(ns.Name, challengeDomain),
		//infrastructure.BuildHttpsIngressRoute(ns.Name, challengeDomain),
		//infrastructure.BuildHttpIngressRoute(ns.Name, challengeDomain),
	}

	for _, val := range resources {
		err = kubeClient.Create(ctx, val)
		if err != nil {
			log.Println(err.Error())
			_ = kubeClient.Delete(ctx, ns)
			return nil, err
		}
	}

	log.Println("Started: " + challengeDomain)
	return &StartChallengeResponse{
		Url:         challengeDomain,
		SecondsLeft: int((time.Minute * time.Duration(config.Values.ChallengeLifetimeMinutes)).Seconds()),
		Started:     true,
	}, nil
}
