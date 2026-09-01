package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"user-management-services/internal/config"
	"user-management-services/internal/database"
	"user-management-services/internal/delivery/grpc"
	"user-management-services/internal/delivery/router/initrouter"
	"user-management-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitApp() {

	gin.SetMode(gin.ReleaseMode)

	ctx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	appConfig := config.NewAppConfig()

	errConnectDatabase := database.Connect()

	if errConnectDatabase != nil {
		log.Fatalf("Failed to connect to Postgres: %v", errConnectDatabase)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	modules := initregistry.NewInitRegistry(appConfig, database.DB)
	initrouter.InitRouter(app, modules, jwks, appConfig)

	modules.User.UserConsumer.StartConsuming(ctx)
	defer modules.User.UserConsumer.Close()

	grpcServer, errGrpcServer := grpc.NewGrpcServer(appConfig.GrpcConfig.PORT, modules.User.UserUsecase)

	if errGrpcServer != nil {
		log.Fatalf("Failed to initialize gRPC Server: %v", err)
	}
	grpcServer.Start()
	defer grpcServer.Stop()

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
	modules.User.UserConsumer.Close()
}
