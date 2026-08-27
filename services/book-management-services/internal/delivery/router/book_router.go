package router

import (
	"book-management-services/internal/config"
	"book-management-services/internal/delivery/controller"
	"book-management-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func BookRouter(app *gin.Engine, bookController controller.BookController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	book := app.Group("/api/v1/book", middleware.AuthMiddleware(jwks, cfg))

	book.POST("", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), bookController.Create)
	book.DELETE("/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), bookController.Delete)
	book.PUT("/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), bookController.Update)
	book.GET("", middleware.RequireRole("SUPER_ADMIN", "ADMIN", "USER_PUBLIC"), bookController.GetAll)
	book.GET("/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN", "USER_PUBLIC"), bookController.GetByID)
}
