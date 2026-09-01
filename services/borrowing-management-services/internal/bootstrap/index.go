package bootstrap

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/infrastructure/database"
	redisdb "borrowing-management-services/internal/infrastructure/redis"
	"borrowing-management-services/internal/registry/initregistry"
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

	ctx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	errConnectMysql := database.ConnectMySQL()

	if errConnectMysql != nil {
		log.Fatalf("Failed to connect to Mysql: %v", errConnectMysql)
	}

	errConnectRedisCache := redisdb.ConnectRedis(ctx)

	if errConnectRedisCache != nil {
		log.Fatalf("Failed to connect to Redis Cache: %v", errConnectRedisCache)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	_, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	modules := initregistry.NewInitRegistry(redisdb.RDS, appConfig)

	go modules.KafkaCacheRegistry.AuthConsumer.StartConsuming(
		ctx,
		modules.KafkaCacheRegistry.EventHandler.HandleUserAuthEvent,
	)

	go modules.KafkaCacheRegistry.StatusConsumer.StartConsuming(
		ctx,
		modules.KafkaCacheRegistry.EventHandler.HandleUserStatusUpdateEvent,
	)

	srv := &http.Server{Addr: ":" + appConfig.PortConfig.PORT, Handler: app}

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
	modules.KafkaCacheRegistry.AuthConsumer.Close()
	modules.KafkaCacheRegistry.StatusConsumer.Close()

}
