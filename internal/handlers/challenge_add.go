package handlers

import (
	"deployer/config"
	"deployer/internal/storage"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/gin-gonic/gin"
)

func AddChallenge(c *gin.Context) {
	userId, exists := c.Get(userIdValue)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not valid"})
	}

	// Add challenge to DB
	challengeId, err := storage.CreateChallenge(userId.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Store challenge file
	form, _ := c.MultipartForm()
	files := form.File["upload[]"]
	dst := filepath.Join(config.Values.UploadPath, challengeId)
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, file := range files {
		allowedFilenames := []string{"challenge.yml", "challenge.zip", "handout.zip"}
		if !slices.Contains(allowedFilenames, file.Filename) {
			c.JSON(http.StatusBadRequest, "invalid filename")
			return
		}

		log.Printf("Uploaded %s to %s", file.Filename, dst)
		err = c.SaveUploadedFile(file, filepath.Join(dst, file.Filename))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"challengeid": challengeId,
	})
}
