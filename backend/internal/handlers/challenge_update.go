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

// ChallengeUpdate godoc
// @Summary      Challenge Update
// @Tags         challenges
// @Accept       mpfd
// @Produce      json
// @Param        id	path		string				true	"Challenge ID"
// @Param upload[] formData []file true "files" collectionFormat(multi)
// @Router       /challenges/{id} [put]
// @Security BearerAuth
func UpdateChallenge(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)

	challenge, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Challenge not found",
		})
		return
	}
	if challenge.UserId != userId {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	// Replace challenge files
	dst := filepath.Join(config.Values.UploadPath, challengeId)
	err = os.RemoveAll(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = os.MkdirAll(dst, 0750)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	form, _ := c.MultipartForm()
	files := form.File["upload[]"]

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
