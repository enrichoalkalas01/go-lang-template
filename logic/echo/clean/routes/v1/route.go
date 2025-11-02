package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Router struct {
	app    *echo.Echo
	config *viper.Viper
	log    *zap.Logger
}

func NewRouter(
	app *echo.Echo,
	config *viper.Viper,
	log *zap.Logger,
) *Router {
	return &Router{
		app:    app,
		config: config,
		log:    log,
	}
}

func (r *Router) Setup() {
	// Setup Base Routes
	r.setupBaseRoutes()

	// Setup Group Route
	v1 := r.app.Group("/api/v1")

	// Setup Routes Here
	r.setupHealthRoutes(v1)
}
