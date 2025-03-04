package handlers

import (
	"deployer/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifyFlag godoc
// @Summary      Verify Challenge Flag
// @Description  Verifies the flag for the given challenge ID.
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Challenge ID"
// @Param        flag body      string  true  "Flag to verify"
// @Success      200  {object}  map[string]string
// @Router       /challenges/{id}/verify [post]
// @Security     BearerAuth
func VerifyFlag(c *gin.Context) {
	challengeId := c.Param("id")

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
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failure", "expected": expectedFlag})
	}
}
