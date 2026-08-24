package bootstrap

import (
	"auth-services/internal/config"
	"auth-services/internal/delivery/router/initrouter"
	redisdb "auth-services/internal/infrastructure/redis"
	"auth-services/internal/registry/initregistry"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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

	ctx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	srv := &http.Server{Addr: ":" + appConfig.Port.PORT, Handler: app}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)
}
