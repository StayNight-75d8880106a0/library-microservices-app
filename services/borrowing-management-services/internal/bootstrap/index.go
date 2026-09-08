package bootstrap

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/router/initrouter"
	"borrowing-management-services/internal/infrastructure/database"
	redisdb "borrowing-management-services/internal/infrastructure/redis"
	"borrowing-management-services/internal/registry/initregistry"
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
	"github.com/robfig/cron/v3"
)

func InitApp() {

	gin.SetMode(gin.ReleaseMode)

	appConfig := config.NewAppConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errConnectMysql := database.ConnectMySQL()

	if errConnectMysql != nil {
		log.Fatalf("Failed to connect to Mysql: %v", errConnectMysql)
	}

	errConnectRedisCache := redisdb.ConnectRedis(ctx)

	if errConnectRedisCache != nil {
		log.Fatalf("Failed to connect to Redis Cache: %v", errConnectRedisCache)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	jakartaLoc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}

	scheduler := cron.New(cron.WithSeconds(), cron.WithLocation(jakartaLoc))

	modules := initregistry.NewInitRegistry(redisdb.RDS, appConfig, database.DB, scheduler)
	initrouter.InitRouter(app, modules, jwks, appConfig)

	go modules.KafkaCacheRegistry.AuthConsumer.StartConsuming(
		ctx,
		modules.KafkaCacheRegistry.EventHandler.HandleUserAuthEvent,
	)

	go modules.KafkaCacheRegistry.StatusConsumer.StartConsuming(
		ctx,
		modules.KafkaCacheRegistry.EventHandler.HandleUserStatusUpdateEvent,
	)

	cronScheduler := modules.WaitingListRegistry.WaitingListCronJob
	cronScheduler.Start()

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

	if err := modules.WaitingListRegistry.WaitingListCronJob.Stop(shutdownCtx); err != nil {
		log.Printf("Cron shutdown error: %v", err)
	}

	srv.Shutdown(shutdownCtx)
	modules.KafkaCacheRegistry.AuthConsumer.Close()
	modules.KafkaCacheRegistry.StatusConsumer.Close()

}
