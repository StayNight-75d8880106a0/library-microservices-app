package router

import (
	"notification-services/internal/config"
	"notification-services/internal/delivery/controller"
	"notification-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func EmailLogRouter(app *gin.Engine, emailController controller.EmailController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	emailLog := app.Group("/api/v1/notifications/log-email", middleware.AuthMiddleware(jwks, cfg))

	emailLog.GET("", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), emailController.GetAllEmailLogs)
	emailLog.GET("/:id", middleware.RequireRole("SUPER_ADMIN", "ADMIN"), emailController.GetEmailLogByID)

}
