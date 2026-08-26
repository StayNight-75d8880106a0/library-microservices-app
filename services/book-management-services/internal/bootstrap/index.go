package bootstrap

import (
	"book-management-services/internal/config"
	mysql "book-management-services/internal/infrastructure/database"
	redisdb "book-management-services/internal/infrastructure/redis"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func InitApp() {
	gin.SetMode(gin.ReleaseMode)

	appConfig := config.NewAppConfig()

	ctx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	errConnectDatabase := mysql.ConnectMySQL()

	if errConnectDatabase != nil {
		log.Fatalf("Failed to connect to Mysql: %v", errConnectDatabase)
	}

	errConnectRedis := redisdb.ConnectRedis(ctx)

	if errConnectRedis != nil {
		log.Fatalf("Failed to connect to Redis: %v", errConnectRedis)
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
