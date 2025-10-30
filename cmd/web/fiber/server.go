package main

import (
	"os"

	"service-golang/configs"
	"service-golang/pkg/logger"
	"service-golang/pkg/server/fiber"

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

	log.Info("Starting application...")

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
	server := fiber.NewFiberServer(configEnv, log.Logger)
	server.SetupMiddlewares()
	server.SetupRoutes()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}

	log.Info("Application started successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
