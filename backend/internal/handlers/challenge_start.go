package handlers

import (
	"context"
	"deployer/config"
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Unleash/unleash-client-go/v4"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StartChallengeResponse struct {
	Url         string `json:"url"`
	SecondsLeft int    `json:"secondslseft"`
	Started     bool   `json:"started"`
	Verified    bool   `json:"verified"`
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
	userId := auth.GetCurrentUserId(c)
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

	if challenge.UserId != userId && !challenge.Published && !auth.IsAdmin(c) {
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
			"url":      getChallengeDomain(runningId),
			"started":  true,
			"verified": challenge.Verified,
		})
		return
	}

	instanceId, token, err := storage.CreateInstance(userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	testMode := false
	challengeDomain := getChallengeDomain(instanceId)
	res, err := createResources(c, userId, &challenge, instanceId, token, challengeDomain, testMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func getChallengeDomain(instanceId string) string {
	return instanceId[0:18] + config.Values.ChallengeDomain
}

func createResources(ctx context.Context, userId string, challenge *storage.Challenge, instanceId, token string, challengeDomain string, testMode bool) (*StartChallengeResponse, error) {
	kubeClient, err := infrastructure.CreateClient()
	if err != nil {
		return nil, err
	}

	// Create resources
	name := infrastructure.GetNamespaceName(instanceId)

	var challengeIdLabel string
	if testMode {
		challengeIdLabel = "test-" + challenge.Id
	} else {
		challengeIdLabel = challenge.Id
	}

	ns := infrastructure.BuildNamespace(name, challengeIdLabel, instanceId, userId)

	var mainResource client.Object
	if useVm := unleash.IsEnabled("use-virtual-machine"); useVm {
		mainResource = infrastructure.BuildVm(challenge.Id, userId, token, ns.Name, challengeDomain, testMode)
	} else {
		mainResource = infrastructure.BuildContainer(challenge.Id, userId, token, ns.Name, challengeDomain, testMode)
	}

	resources := []client.Object{
		ns,
		mainResource,
		infrastructure.BuildNetworkPolicy(ns),
	}

	if !testMode {
		resources = append(resources,
			infrastructure.BuildHttpService(ns.Name),
			infrastructure.BuildSshService(ns.Name),
			infrastructure.BuildHttpIngress(ns.Name, challengeDomain),
		)
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
		Verified:    challenge.Verified,
	}, nil
}
