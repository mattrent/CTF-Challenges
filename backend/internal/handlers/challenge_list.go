package handlers

import (
	"deployer/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChallengeList godoc
// @Summary      Callenges List
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Router       /challenges [get]
// @Security BearerAuth
func ListChallenges(c *gin.Context) {
	res, err := storage.ListChallenges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"challenges": res,
	})
}
