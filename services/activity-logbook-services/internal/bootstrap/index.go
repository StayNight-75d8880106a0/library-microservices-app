package bootstrap

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/delivery/router/initrouter"
	"activity-logbook-services/internal/infrastructure/database"
	redisdb "activity-logbook-services/internal/infrastructure/redis"
	"activity-logbook-services/internal/registry/initregistry"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitApp() {

	gin.SetMode(gin.ReleaseMode)

	appConfig := config.NewAppConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errDatabase := database.Connect()

	if errDatabase != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", errDatabase)
	}

	errRedis := redisdb.ConnectRedis(ctx)

	if errRedis != nil {
		log.Fatalf("Failed to connect to Redis: %v", errRedis)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	modules := initregistry.NewInitRegistry(database.DB, redisdb.RDS, appConfig)
	initrouter.Initrouter(app, modules, jwks, appConfig)

	modules.Logbook.LogbookConsumer.StartConsuming(ctx)
	defer modules.Logbook.LogbookConsumer.Close()

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
	modules.Logbook.LogbookConsumer.Close()
}
