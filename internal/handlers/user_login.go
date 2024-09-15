package handlers

import (
	"deployer/internal/auth"
	"deployer/internal/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type LoginRequest struct {
	Username string
	Password string
}

func Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := storage.GetUser(request.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	token := auth.CreateToken(user.Id, user.Role)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
