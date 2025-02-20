package handlers

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v4"

	"deployer/config"
)

func init() {
	// Initialize Unleash client assynchronously
	unleash.Initialize(
		unleash.WithListener(&unleash.DebugListener{}),
		unleash.WithAppName("deployer-service"),
		unleash.WithUrl(config.Values.Unleash.Url),
		unleash.WithEnvironment(config.Values.Unleash.Environment),
		unleash.WithCustomHeaders(
			http.Header{"Authorization": {
				config.Values.Unleash.ApiKey,
			}},
		),
	)
}
