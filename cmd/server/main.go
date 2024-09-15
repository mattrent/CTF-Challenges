package main

import (
	"deployer/internal/auth"
	"deployer/internal/handlers"
	"deployer/internal/infrastructure"
	"deployer/internal/storage"
	"log"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gin-gonic/gin"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	log.Println("Starting...")
	logf.SetLogger(zap.New())

	storage.InitDb()

	go infrastructure.StartCleaner()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Healthy")
	})

	router.POST("/users/login", handlers.Login)

	router.POST("/challenges", auth.RequireDeveloper, handlers.AddChallenge)

	router.POST("/challenges/:id/start", auth.RequireAuth, handlers.StartChallenge)

	router.POST("/challenges/:id/stop", auth.RequireAuth, handlers.StopChallenge)

	router.GET("/challenges/:id/status", auth.RequireAuth, handlers.GetChallengeStatus)

	router.GET("/challenges/:id/logs", auth.RequireDeveloper, handlers.GetChallengeLogs)

	router.GET("/challenges/:id/download", handlers.DownloadChallenge)

	router.POST("/challenges/:id/publish", auth.RequireDeveloper, handlers.PublishChallenge)

	err := router.SetTrustedProxies(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatal(err.Error())
	}
}
