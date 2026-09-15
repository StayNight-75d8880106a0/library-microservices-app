package initrouter

import (
	"notification-services/internal/config"
	"notification-services/internal/delivery/router"
	"notification-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitRouter(app *gin.Engine, modules *initregistry.Module, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {
	router.EmailLogRouter(app, *modules.EmailLog.EmailLogController, jwks, cfg)
}
