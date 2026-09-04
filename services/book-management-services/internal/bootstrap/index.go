package bootstrap

import (
	"book-management-services/internal/config"
	"book-management-services/internal/delivery/router/initrouter"
	mysql "book-management-services/internal/infrastructure/database"
	"book-management-services/internal/infrastructure/elasticsearch"
	redisdb "book-management-services/internal/infrastructure/redis"
	"book-management-services/internal/registry/initialregistry"
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

	errConnectDatabase := mysql.ConnectMySQL()

	if errConnectDatabase != nil {
		log.Fatalf("Failed to connect to Mysql: %v", errConnectDatabase)
	}

	errConnectRedis := redisdb.ConnectRedis(ctx)

	if errConnectRedis != nil {
		log.Fatalf("Failed to connect to Redis: %v", errConnectRedis)
	}

	errConnectElasticsearch := elasticsearch.InitElasticsearch(appConfig.Elasticsearch.ElasticsearchURL)

	if errConnectElasticsearch != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", errConnectElasticsearch)
	}

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	modules := initialregistry.NewInitRegistry(mysql.DB, redisdb.RDS, elasticsearch.ElasticseearchClient, appConfig)
	initrouter.InitRouter(app, modules, jwks, appConfig)

	modules.Book.BookConsumer.StartConsuming(ctx)
	defer modules.Book.BookConsumer.Close()

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
	modules.Book.BookConsumer.Close()
}
