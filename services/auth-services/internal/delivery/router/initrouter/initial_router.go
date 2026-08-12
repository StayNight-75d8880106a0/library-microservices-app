package initrouter

import (
	"auth-services/internal/config"
	"auth-services/internal/delivery/router"
	"auth-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitRouter(app *gin.Engine, modules *initregistry.Modules, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {
	router.AuthRouter(app, modules.Auth.AuthController, modules.Dependencies, jwks, cfg)
}
