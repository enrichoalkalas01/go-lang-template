package main

import (
	"net/http"
	"os"
	"time"

	"service-golang/configs"
	"service-golang/pkg/logger"

	"service-golang/pkg/server/echo"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Step 1 - Init Logger
	log, err := logger.NewLogger(logger.Config{
		Environtment: getEnv("APP_ENV", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		OutputPath:   "stdout",
	})

	if err != nil {
		panic("Failed to init logger: " + err.Error())
	}

	defer log.Sync()

	log.Info("Starting Echo application...")

	// Step 2 - Load env configuration
	configEnv, err := configs.NewViper(".env", "env", ".", "../../")
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	log.Info("ENV configuration loaded successfully",
		zap.String("app_name", configEnv.GetString("APP_NAME")),
		zap.String("app_env", configEnv.GetString("APP_ENV")),
		zap.String("app_version", configEnv.GetString("APP_VERSION")),
	)

	// Init Echo Server
	server := echo.NewEchoServer(configEnv, log.Logger)
	server.SetupMiddlewares()
	// server.SetupRoutes()

	// Setup routes - create instance router first
	// Setup routes - create instance router first
	echoApp := server.GetEcho()
	router := NewRouter(echoApp, configEnv, log.Logger)
	router.SetupCleanEchoRoutes()

	// Start server
	// Start server
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}

	log.Info("Application started successfully")
}

// Simple local router implementation used instead of the missing external package.
type Router struct {
	app    *echo.EchoServer
	env    *viper.Viper
	logger *zap.Logger
}

func NewRouter(app *echo.EchoServer, env *viper.Viper, logger *zap.Logger) *Router {
	return &Router{app: app, env: env, logger: logger}
}

func (r *Router) SetupCleanEchoRoutes() {
	// v1 := r.app.GetEcho().Group("/v1")

	r.app.GetEcho().GET("/", r.rootHandler)
	// Placeholder: register Echo routes here using r.app (echo instance), r.cfg and r.logger.
	// This is a no-op implementation to satisfy the build when the external package is not present.
}

func (r *Router) rootHandler(c echo.EchoServer) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Server is running...",
		"version": r.config.GetString("APP_VERSION"),
		"time":    time.Now().Format(time.RFC3339),
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
