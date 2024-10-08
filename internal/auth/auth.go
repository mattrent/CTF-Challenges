package auth

import (
	"deployer/config"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserId string
	Role   string
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CreateToken(userId, role string) string {
	expirationTime := time.Now().Add(time.Hour * 24)
	claims := &Claims{
		UserId: userId,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.Values.JwtSecret)
	return token
}

func RequireAuth(c *gin.Context) {
	RequireRole(c, []string{})
}

func RequireAdmin(c *gin.Context) {
	RequireRole(c, []string{"admin"})
}

func RequireDeveloper(c *gin.Context) {
	RequireRole(c, []string{"admin", "developer"})
}

func RequireRole(c *gin.Context, allowedRoles []string) {
	authHeader := c.GetHeader("Authorization")
	reqTokenParts := strings.Split(authHeader, "Bearer ")

	if len(reqTokenParts) != 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	reqToken := reqTokenParts[1]
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method invalid: %v", token.Header["alg"])
		}
		return config.Values.JwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.Println("Signature error")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Println("Signature error")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if !token.Valid {
		log.Println("Token invalid")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		log.Println("Token exp error")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if exp.Before(time.Now()) {
		log.Println("Token expired")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if len(allowedRoles) != 0 && !slices.Contains(allowedRoles, claims.Role) {
		log.Println("Missing role")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("userid", claims.UserId)
	log.Println("Setting context for userid for: " + claims.UserId)
	c.Next()
}
