package router

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/controller"
	"borrowing-management-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func BorrowingRouter(app *gin.Engine, borrowingController *controller.BorrowingController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	borrowing := app.Group("/api/v1/borrowing", middleware.AuthMiddleware(jwks, cfg))

	borrowing.POST("", middleware.RequireRole("USER_PUBLIC"), borrowingController.Create)
	borrowing.GET("", middleware.RequireRole("SUPER_ADMIN", "ADMIN", "USER_PUBLIC"), borrowingController.GetALL)
	borrowing.GET("/:id/my", middleware.RequireRole("USER_PUBLIC"), borrowingController.GetMyByID)
	borrowing.GET("/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), borrowingController.GetByID)
	borrowing.PATCH("/:id/status", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), borrowingController.UpdateStatus)

}
