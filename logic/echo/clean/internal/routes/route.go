package routes

import (
	"net/http"
	v1 "service-golang/logic/echo/clean/internal/routes/v1"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Router struct {
	app    *echo.Echo
	config *viper.Viper
	log    *zap.Logger

	v1Router *v1.Router
}

func NewRouter(
	app *echo.Echo,
	config *viper.Viper,
	log *zap.Logger,
) *Router {
	v1Router := v1.NewRouter(app, config, log)

	return &Router{
		app:      app,
		config:   config,
		log:      log,
		v1Router: v1Router,
	}
}

func (r *Router) SetupCleanEchoRoutes() {
	r.app.GET("/", r.rootHandler)
	// r.v1Router.Setup()
}

func (r *Router) rootHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Server is running...",
		"version": r.config.GetString("APP_VERSION"),
		"time":    time.Now().Format(time.RFC3339),
	})
}
