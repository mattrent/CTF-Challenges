package handlers

import (
	"deployer/config"
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Verify a challenge flag
// @Description Verifies the flag for a challenge
// @Tags challenges
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Param flag body handlers.FlagRequest true "Flag request"
// @Success 200 {object} handlers.FlagResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Router /challenges/{id}/verify [post]
func VerifyFlag(c *gin.Context) {
	challengeId := c.Param("id")

	runningIdTest, err := infrastructure.GetRunningTestInstanceId(c, challengeId)
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

	deleteNamespace(c, runningIdTest, infrastructure.GetNamespaceNameTest)
}

// @Summary Start a test for a challenge
// @Description Starts a test for a challenge
// @Tags challenges
// @Security BearerAuth
// @Param id path string true "Challenge ID"
// @Success 200 {object} handlers.TestResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Router /challenges/{id}/verify [get]
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

	nsList, err := infrastructure.TestNameSpacesForUserId(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Only admin can exceed this value
	if !auth.IsAdmin(c) && len(nsList.Items) >= config.Values.MaxConcurrentTests {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many tests running"})
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
