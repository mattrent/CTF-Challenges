package main

import (
	"crypto/tls"
	"crypto/x509"
	"deployer/config"
	"deployer/internal/auth"
	"deployer/internal/handlers"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"log"
	"net/http"
	"os"

	swaggerFiles "github.com/swaggo/files"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	_ "deployer/docs"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <jwt-token>"
func main() {
	log.Println("Starting...")
	logf.SetLogger(zap.New())

	if config.Values.RootCert != "" {
		rootCertPool := x509.NewCertPool()
		certs, errCert := os.ReadFile(config.Values.RootCert)
		if errCert != nil {
			log.Fatalf("Failed to read root certificate: %v", errCert)
		}

		if ok := rootCertPool.AppendCertsFromPEM(certs); !ok {
			log.Fatalf("Failed to append root certificate to pool")
		}

		// Create a custom TLS configuration
		tlsConfig := &tls.Config{
			RootCAs: rootCertPool,
		}

		// Setup the HTTP client with the custom transport
		http.DefaultTransport.(*http.Transport).TLSClientConfig = tlsConfig
	}

	storage.InitDb()

	go infrastructure.StartCleaner()

	router := gin.Default()
	router.Use(ErrorHandler)

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Healthy")
	})

	router.POST("/users/login", handlers.Login)

	router.GET("/challenges", auth.RequireAdmin, handlers.ListChallenges)

	router.POST("/challenges", auth.RequireDeveloper, handlers.AddChallenge)

	router.PUT("/challenges/:id", auth.RequireDeveloper, handlers.UpdateChallenge)

	router.DELETE("/challenges/:id", auth.RequireAuth, handlers.DeleteChallenge)

	router.POST("/challenges/:id/start", auth.RequireAuth, handlers.StartChallenge)

	router.POST("/challenges/:id/stop", auth.RequireAuth, handlers.StopChallenge)

	router.GET("/challenges/:id/status", auth.RequireAuth, handlers.GetChallengeStatus)

	router.GET("/challenges/:id/logs", auth.RequireDeveloper, handlers.GetChallengeLogs)

	router.GET("/challenges/:id/download", handlers.DownloadChallenge)

	router.POST("/challenges/:id/publish", auth.RequireDeveloper, handlers.PublishChallenge)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := router.SetTrustedProxies(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors {
		log.Println(err.Error())
	}

	if len(c.Errors) > 0 {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
