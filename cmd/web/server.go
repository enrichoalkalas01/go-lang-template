package main

import (
	"os"

	"service-golang/configs"
	"service-golang/pkg/database"
	"service-golang/pkg/logger"

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

	// Step 3 - Init PostgreSQL
	postgresDB, err := database.NewPostgres(configEnv, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}

	sqlDB, _ := postgresDB.DB()
	defer sqlDB.Close()

	log.Info("Application started successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
