package handlers

import (
	"deployer/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
