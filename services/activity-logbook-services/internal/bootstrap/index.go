package bootstrap

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/infrastructure/database"
	redisdb "activity-logbook-services/internal/infrastructure/redis"
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

	errDatabase := database.Connect()

	if errDatabase != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", errDatabase)
	}

	errRedis := redisdb.ConnectRedis(ctx)

	if errRedis != nil {
		log.Fatalf("Failed to connect to Redis: %v", errRedis)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	_, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

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

}
