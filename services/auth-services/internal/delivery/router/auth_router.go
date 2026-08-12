package router

import (
	"auth-services/internal/config"
	"auth-services/internal/delivery/controller"
	"auth-services/internal/delivery/middleware"
	"auth-services/internal/helper"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func AuthRouter(app *gin.Engine, authController *controller.AuthController, deps *helper.Depedencies, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	auth := app.Group("/api/v1")

	authMiddleware := middleware.AuthMiddleware(jwks, deps.IsblacklistToken, cfg)

	auth.POST("/login", authController.Login)
	auth.POST("/register", authController.Register)
	auth.POST("/logout", authMiddleware, authController.Logout)
	auth.GET("/me", authMiddleware, authController.Me)
	auth.POST("/refresh", authController.RefreshToken)

}
