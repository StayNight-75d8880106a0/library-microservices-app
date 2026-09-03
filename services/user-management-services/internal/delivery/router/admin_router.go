package router

import (
	"user-management-services/internal/config"
	"user-management-services/internal/delivery/controller"
	"user-management-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func AdminRouter(app *gin.Engine, AdminController *controller.AdminController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	admin := app.Group("/api/v1/user", middleware.AuthMiddleware(jwks, cfg), middleware.RequireRole("SUPER_ADMIN"))

	admin.GET("", AdminController.GetAllUsers)
	admin.POST("", AdminController.CreateAdmin)
	admin.PUT("/:adminID", AdminController.UpdateAdmin)
	admin.DELETE("/:adminID", AdminController.DeleteAdmin)

}
