package initrouter

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/delivery/router"
	"activity-logbook-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func Initrouter(app *gin.Engine, modules *initregistry.Module, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {
	router.LogbookRouter(app, modules.Logbook.LogbookController, jwks, cfg)
}
