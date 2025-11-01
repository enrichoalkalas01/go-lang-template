package routes

import (
	v1 "service-golang/logic/echo/clean/internal/routes/v1"

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
	r.v1Router.Setup()
}
