package gin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type GinServer struct {
	gin    *gin.Engine
	config *viper.Viper
	log    *zap.Logger
	server *http.Server
}

func NewGinServer(cfg *viper.Viper, log *zap.Logger) *GinServer {
	// Set Gin mode
	if cfg.GetString("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode) // Show routes and info
	}

	r := gin.New()

	return &GinServer{
		gin:    r,
		config: cfg,
		log:    log,
	}
}

func (s *GinServer) SetupMiddlewares() {
	// Recovery middleware
	s.gin.Use(gin.Recovery())

	// CORS middleware
	s.gin.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Request ID middleware
	s.gin.Use(s.requestIDMiddleware())

	// Gin native logger middleware
	s.gin.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s\n",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
			)
		},
	}))

	// Custom Zap logger middleware
	s.gin.Use(s.loggerMiddleware())
}

func (s *GinServer) SetupRoutes() {
	// Health check
	s.gin.GET("/health", s.healthCheck)

	// API routes
	s.setupAPIRoutes()
}

func (s *GinServer) setupAPIRoutes() {
	// API v1 group
	v1 := s.gin.Group("/api/v1")
	{
		// User routes
		userGroup := v1.Group("/users")
		{
			userGroup.GET("", s.getUsers)
			userGroup.GET("/:id", s.getUserByID)
			userGroup.POST("", s.createUser)
			userGroup.PUT("/:id", s.updateUser)
			userGroup.DELETE("/:id", s.deleteUser)
		}

		// Product routes
		productGroup := v1.Group("/products")
		{
			productGroup.GET("", s.getProducts)
			productGroup.GET("/:id", s.getProductByID)
			productGroup.POST("", s.createProduct)
			productGroup.PUT("/:id", s.updateProduct)
			productGroup.DELETE("/:id", s.deleteProduct)
		}

		// Auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", s.login)
			authGroup.POST("/register", s.register)
			authGroup.POST("/logout", s.logout)
		}
	}

	// Admin routes (with auth middleware)
	admin := s.gin.Group("/admin", s.authMiddleware())
	{
		admin.GET("/dashboard", s.dashboard)
		admin.GET("/users", s.adminGetUsers)
	}
}

func (s *GinServer) Start() error {
	// Get host and port with fallback
	host := s.config.GetString("GIN_HOST")
	if host == "" {
		host = s.config.GetString("APP_HOST")
	}
	if host == "" {
		host = "0.0.0.0"
	}

	port := s.config.GetString("GIN_PORT")
	if port == "" {
		port = s.config.GetString("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("%s:%s", host, port)

	// Zap logger info
	s.log.Info("Starting Gin server",
		zap.String("framework", "Gin"),
		zap.String("host", host),
		zap.String("port", port),
		zap.String("address", address),
		zap.String("version", s.config.GetString("APP_VERSION")),
	)

	// Gin native log
	log.Printf("[GIN-debug] Listening and serving HTTP on %s", address)

	s.server = &http.Server{
		Addr:    address,
		Handler: s.gin,
	}

	// ✅ ADD: Manage server lifecycle
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// ✅ ADD: Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("Gin server startup error", zap.Error(err))
			log.Fatalf("Server startup error: %v", err)
		}
	}()

	s.log.Info("Gin server started successfully",
		zap.String("address", address),
		zap.String("url", fmt.Sprintf("http://%s:%s", host, port)),
	)

	// ✅ ADD: Wait for shutdown signal
	<-ctx.Done()

	// ✅ ADD: Graceful shutdown with configurable timeout
	shutdownTimeout := s.config.GetDuration("TIMEOUT_GRACEFUL_SHUTDOWN")
	if shutdownTimeout == 0 {
		shutdownTimeout = 10 // default 10 seconds
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}

func (s *GinServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down Gin server...")
	log.Println("[GIN] Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.log.Error("Gin server forced to shutdown", zap.Error(err))
		log.Fatalf("Server shutdown error: %v", err)
		return err
	}

	s.log.Info("Gin server exited gracefully")
	log.Println("[GIN] Server exited gracefully")
	return nil
}

// Request ID middleware
func (s *GinServer) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Custom logger middleware (Zap)
func (s *GinServer) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		requestID, _ := c.Get("RequestID")

		// Zap structured logging
		s.log.Info("HTTP Request",
			zap.String("framework", "Gin"),
			zap.Any("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("remote_ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	}
}

// Auth middleware
func (s *GinServer) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization token",
			})
			c.Abort()
			return
		}

		// TODO: Validate token

		c.Next()
	}
}

// Health check handler
func (s *GinServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"framework": "Gin",
		"message":   "Service is healthy",
		"version":   s.config.GetString("APP_VERSION"),
		"time":      time.Now().Format(time.RFC3339),
	})
}

// Example handlers
func (s *GinServer) getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get all users"})
}

func (s *GinServer) getUserByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Get user by ID", "id": id})
}

func (s *GinServer) createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

func (s *GinServer) updateUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "User updated", "id": id})
}

func (s *GinServer) deleteUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "User deleted", "id": id})
}

func (s *GinServer) getProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get all products"})
}

func (s *GinServer) getProductByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Get product by ID", "id": id})
}

func (s *GinServer) createProduct(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Product created"})
}

func (s *GinServer) updateProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Product updated", "id": id})
}

func (s *GinServer) deleteProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted", "id": id})
}

func (s *GinServer) login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": "dummy-token"})
}

func (s *GinServer) register(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

func (s *GinServer) logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (s *GinServer) dashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Admin dashboard"})
}

func (s *GinServer) adminGetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Admin: Get all users"})
}

func (s *GinServer) GetGin() *gin.Engine {
	return s.gin
}
