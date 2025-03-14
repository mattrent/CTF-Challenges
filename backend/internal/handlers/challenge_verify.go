package handlers

import (
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"fmt"
	"net/http"
	"strconv"

	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
)

// VerifyFlag godoc
// @Summary      Verify Challenge Flag
// @Description  Verifies the flag for the given challenge ID.
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Challenge ID"
// @Param        flag    body      string  true  "Flag to verify"
// @Param        userid  query     string  true  "User ID"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /challenges/{id}/verify [post]
// @Security     BearerAuth
func VerifyFlag(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.Query(auth.ContextUserIdKey)

	runningIdTest, err := infrastructure.GetRunningInstanceId(c, userId, "test-"+challengeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if runningIdTest == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Test is not running",
		})
		return
	}

	var json struct {
		Flag string `json:"flag" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flag := json.Flag
	expectedFlag, err := storage.GetChallengeFlag(challengeId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if flag == expectedFlag {
		if err := storage.MarkChallengeVerified(challengeId); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	} else {
		quotedFlag := fmt.Sprintf("\"%s\"", flag)
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failure", "submitted-flag": quotedFlag})
	}

	kubeconfig := infrastructure.GetKubeConfigSingleton()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	deleteNamespace(c, clientset, runningIdTest)
}

// @Summary Start a test for a challenge
// @Description Start a test for a specific challenge
// @Tags challenges
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Success 200 {object} handlers.StartTestResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /challenges/{id}/verify [get]
func StartTest(c *gin.Context) {
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

	if challenge.UserId != userId && !auth.IsAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	runningIdChallenge, err := infrastructure.GetRunningInstanceId(c, userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	runningIdTest, err := infrastructure.GetRunningInstanceId(c, userId, "test-"+challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if runningIdChallenge == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Challenge not started",
		})
		return
	}

	if runningIdTest != "" {
		c.JSON(http.StatusOK, gin.H{
			"started": true,
		})
		return
	}

	instanceId, token, err := storage.CreateInstance(userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	testMode := true
	challengeDomain := getChallengeDomain(runningIdChallenge)
	res, err := createResources(c, userId, &challenge, instanceId, token, challengeDomain, testMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
