package initrouter

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/router"
	"borrowing-management-services/internal/registry/initregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitRouter(app *gin.Engine, modules *initregistry.Module, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {
	router.BorrowingRouter(app, modules.BorrowingRegistry.BorrowingController, jwks, cfg)
}
