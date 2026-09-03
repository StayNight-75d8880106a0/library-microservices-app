package router

import (
	"user-management-services/internal/config"
	"user-management-services/internal/delivery/controller"
	"user-management-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func UserRouter(app *gin.Engine, UserController *controller.UserController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	user := app.Group("/api/v1/user/my", middleware.AuthMiddleware(jwks, cfg), middleware.RequireRole("USER_PUBLIC"))

	user.PUT("/password", UserController.UpdateMyPassword)
	user.PUT("/profile", UserController.UpdateMyProfile)
	user.GET("", UserController.GetMyProfile)

}
