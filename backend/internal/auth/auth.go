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

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserId string
	Role   string
	jwt.RegisteredClaims
}

type KeycloakClaims struct {
	Roles []string `json:"roles"`
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
	parts := strings.Split(authHeader, "Bearer ")

	if len(parts) != 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if len(config.Values.JwksUrl) > 0 {
		k, err := keyfunc.NewDefaultCtx(c, []string{config.Values.JwksUrl})
		if err != nil {
			log.Println("Failed to create keyfunc" + err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := &KeycloakClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, k.Keyfunc)
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

		var intersect []string
		for _, role := range claims.Roles {
			if slices.Contains(allowedRoles, role) {
				intersect = append(intersect, role)
			}
		}

		if len(allowedRoles) != 0 && len(intersect) == 0 {
			log.Println("Missing role")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userid", claims.Subject)
		log.Println("Setting context for userid for: " + claims.Subject)
		c.Next()

	} else {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
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
}
