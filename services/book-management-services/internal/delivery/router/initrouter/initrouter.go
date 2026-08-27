package initrouter

import (
	"book-management-services/internal/config"
	"book-management-services/internal/delivery/router"
	"book-management-services/internal/registry/initialregistry"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func InitRouter(app *gin.Engine, modules *initialregistry.Module, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {
	router.BookRouter(app, *modules.Book.BookController, jwks, cfg)
}
