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
	ResourceAccess struct {
		Deployer struct {
			Roles []string `json:"roles"`
		} `json:"deployer"`
	} `json:"resource_access"`
	jwt.RegisteredClaims
}

const (
	AdminRoleKey     = "admin"
	DeveloperRoleKey = "developer"
)

const (
	ContextUserIdKey = "userid"
	ContextRoleKey   = "role"
)

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
	RequireRole(c, []string{AdminRoleKey})
}

func RequireDeveloper(c *gin.Context) {
	RequireRole(c, []string{AdminRoleKey, DeveloperRoleKey})
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
		_, err = jwt.ParseWithClaims(parts[1], claims, k.Keyfunc)

		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				log.Println("Signature error")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			log.Println("Expired token error")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var intersect []string
		for _, role := range claims.ResourceAccess.Deployer.Roles {
			if slices.Contains(allowedRoles, role) {
				intersect = append(intersect, role)
			}
		}

		if len(allowedRoles) != 0 && len(intersect) == 0 {
			log.Println("Missing role")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(ContextUserIdKey, claims.Subject)
		if slices.Contains(claims.ResourceAccess.Deployer.Roles, AdminRoleKey) {
			c.Set(ContextRoleKey, AdminRoleKey)
		} else if slices.Contains(claims.ResourceAccess.Deployer.Roles, DeveloperRoleKey) {
			c.Set(ContextRoleKey, DeveloperRoleKey)
		}

		log.Println("Setting context for userid for: " + claims.Subject)
	} else {
		claims := &Claims{}
		_, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
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
			log.Println("Token expired error")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if len(allowedRoles) != 0 && !slices.Contains(allowedRoles, claims.Role) {
			log.Println("Missing role")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(ContextUserIdKey, claims.UserId)
		c.Set(ContextRoleKey, claims.Role)
		log.Println("Setting context for userid for: " + claims.UserId)
	}
	c.Next()
}

func GetCurrentUserId(c *gin.Context) string {
	return c.GetString(ContextUserIdKey)
}

func IsAdmin(c *gin.Context) bool {
	return c.GetString(ContextRoleKey) == AdminRoleKey
}
