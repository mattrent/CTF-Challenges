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

// @Summary Download solution
// @Description Downloads a solution
// @Tags solutions
// @Param id path string true "Challenge ID"
// @Param token query string true "Token"
// @Success 200 {file} file "Challenge file"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Failure 404 {object} handlers.ErrorResponse "File not found"
// @Router /solutions/{id}/download [get]
func DownloadSolution(c *gin.Context) {
	challengeId := c.Param("id")
	token := c.Query("token")

	instance, err := storage.GetInstance(challengeId, token)
	if err != nil || instance.CreatedAt.Add(time.Minute*20).Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	file := filepath.Join(config.Values.UploadPath, challengeId, "solution.zip")
	_, err = os.Stat(file)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.FileAttachment(file, "solution.zip")
}
