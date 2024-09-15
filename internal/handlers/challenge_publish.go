package handlers

import (
	"deployer/internal/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PublishChallenge(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)
	challenge, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge.UserId != userId {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	err = storage.PublishChallenge(challenge.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("Challenge status changed to published")

	c.IndentedJSON(http.StatusCreated, gin.H{})
}
