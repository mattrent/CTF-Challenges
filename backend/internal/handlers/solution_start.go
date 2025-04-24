package handlers

import (
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TestResponse struct {
	Started  bool `json:"started"`
	Verified bool `json:"verified"`
}

// @Summary Start a test for a challenge
// @Description Starts a test for a challenge
// @Tags solutions
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} handlers.TestResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /solutions/{id}/start [post]
func StartTest(c *gin.Context) {
	userId := auth.GetCurrentUserId(c)
	challengeId := c.Param("id")

	challenge, err := storage.GetChallengeWrapper(challengeId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	if challenge.UserId != userId && !auth.IsAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	runningIdChallenge, err := infrastructure.GetRunningChallengeInstanceId(c, userId, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	runningIdTest, err := infrastructure.GetRunningTestInstanceId(c, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if runningIdTest != "" {
		c.JSON(http.StatusOK, gin.H{
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

	testMode := true
	challengeDomain := getChallengeDomain(runningIdChallenge)
	res, err := createResources(c, userId, &challenge, instanceId, token, challengeDomain, testMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
