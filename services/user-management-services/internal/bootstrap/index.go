package bootstrap

import (
	"fmt"
	"log"
	"user-management-services/internal/config"
	"user-management-services/internal/delivery/router/initrouter"
	"user-management-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitApp() {

	gin.SetMode(gin.ReleaseMode)

	appConfig := config.NewAppConfig()

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", appConfig.Keycloak.KeycloakURL, appConfig.Keycloak.Realm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})

	if err != nil {
		log.Fatalf("Failed to fetch JWKS from Keycloak: %v", err)
	}

	app := gin.Default()

	modules := initregistry.NewInitRegistry(appConfig)
	initrouter.InitRouter(app, modules, jwks, appConfig)

	app.Run(":" + appConfig.Port.PORT)

}
