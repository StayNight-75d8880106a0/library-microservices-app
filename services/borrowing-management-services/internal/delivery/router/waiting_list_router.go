package router

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/controller"
	"borrowing-management-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func WaitingListRouter(app *gin.Engine, borrowingController *controller.WaitingListController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	waiting := app.Group("/api/v1/borrowing/waiting-list", middleware.AuthMiddleware(jwks, cfg))

	waiting.POST("", middleware.RequireRole("USER_PUBLIC"), borrowingController.Create)
	waiting.GET("", middleware.RequireRole("USER_PUBLIC", "SUPER_ADMIN", "ADMIN"), borrowingController.GetAll)
	waiting.GET("/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), borrowingController.GetByID)
	waiting.PATCH("/:id/cancel", middleware.RequireRole("USER_PUBLIC"), borrowingController.Cancel)

}
