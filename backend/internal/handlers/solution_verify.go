package handlers

import (
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

type FlagResponse struct {
	Status string `json:"status"`
}

// @Summary Verify a challenge flag
// @Description Verifies the flag for a challenge
// @Tags solutions
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Param flag body handlers.FlagRequest true "Flag request"
// @Success 200 {object} handlers.FlagResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Router /solutions/{id}/verify [post]
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

	err = deleteNamespace(c, runningIdTest, infrastructure.GetNamespaceNameTest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
}
