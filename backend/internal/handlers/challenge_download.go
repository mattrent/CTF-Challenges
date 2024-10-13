package handlers

import (
	"deployer/config"
	"deployer/internal/storage"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func DownloadChallenge(c *gin.Context) {
	challengeId := c.Param("id")
	token := c.Query("token")

	instance, err := storage.GetInstance(challengeId, token)
	if err != nil || instance.CreatedAt.Add(time.Minute*20).Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	file := filepath.Join(config.Values.UploadPath, challengeId, "challenge.zip")
	_, err = os.Stat(file)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.FileAttachment(file, "challenge.zip")
}
