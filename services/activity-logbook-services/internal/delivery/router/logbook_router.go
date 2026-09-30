package router

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/delivery/controller"
	"activity-logbook-services/internal/delivery/middleware"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

func LogbookRouter(app *gin.Engine, logbookController *controller.LogbookController, jwks keyfunc.Keyfunc, cfg *config.AppConfig) {

	logbook := app.Group("/api/v1/logbook", middleware.AuthMiddleware(jwks, cfg), middleware.RequireRole("SUPER_ADMIN"))

	logbook.GET("", logbookController.GetAll)
	logbook.GET("/:traceID", logbookController.GetByTraceID)
	logbook.GET("/:traceID/:id", logbookController.GetByID)

}
