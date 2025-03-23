package handlers

import (
	"deployer/internal/auth"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SolutionStop godoc
// @Summary      Solution Stop
// @Tags         solutions
// @Param        id	path		string				true	"Challenge ID"
// @Accept       json
// @Produce      json
// @Router       /solutions/{id}/stop [post]
// @Security BearerAuth
func StopTest(c *gin.Context) {
	challengeId := c.Param("id")
	userId := auth.GetCurrentUserId(c)

	challenge, err := storage.GetChallengeWrapper(challengeId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	if challenge.UserId != userId && !auth.IsAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	instanceIdTest, err := infrastructure.GetRunningTestInstanceId(c, challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if instanceIdTest == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "Test instance not running"})
		return
	}

	err = deleteNamespace(c, instanceIdTest, infrastructure.GetNamespaceNameTest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stopping test instance",
	})
}
