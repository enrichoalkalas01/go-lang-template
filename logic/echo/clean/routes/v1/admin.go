package v1

import (
	handler "service-golang/logic/echo/clean/internal/handler/admin"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type AdminRoutes struct {
	app    *echo.Echo
	config *viper.Viper
	log    *zap.Logger
	
	adminHandler handler.AdminHandler
}

func (r *AdminRoutes) setupAdminRoutes(v1 *echo.Group) {
	admin := v1.Group("/admin")

	admin.GET("/", handler.AdminHandler)
}
