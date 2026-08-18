package initrouter

import (
	"user-management-services/internal/config"
	"user-management-services/internal/delivery/router"
	"user-management-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitRouter(app *gin.Engine, modules *initregistry.Modules, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {
	router.AdminRouter(app, modules.User.AdminController, jwks, cfg)
	router.UserRouter(app, modules.User.UserController, jwks, cfg)
}
