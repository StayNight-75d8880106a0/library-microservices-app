package bootstrap

import (
	"auth-services/internal/config"
	"auth-services/internal/delivery/router/initrouter"
	redisdb "auth-services/internal/infrastructure/redis"
	"auth-services/internal/registry/initregistry"
	"context"
	"fmt"
	"log"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitApp() {

	gin.SetMode(gin.ReleaseMode)

	appConfig := config.NewAppConfig()

	log.Printf("Connecting to Redis on %s", appConfig.Redis.RedisHost)

	if err := redisdb.ConnectRedis(context.Background()); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	modules := initregistry.NewInitRegistry(appConfig, redisdb.RDS)
	initrouter.InitRouter(app, modules, jwks, appConfig)

	app.Run(":" + appConfig.Port.PORT)

}
