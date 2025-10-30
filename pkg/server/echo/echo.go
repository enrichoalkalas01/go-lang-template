package echo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type EchoServer struct {
	echo   *echo.Echo
	config *viper.Viper
	log    *zap.Logger
}

func NewEchoServer(cfg *viper.Viper, log *zap.Logger) *EchoServer {
	e := echo.New()

	// Enable Echo logger (not hide banner)
	e.HideBanner = false // Show Echo banner
	e.HidePort = false   // Show port info

	// Configure Echo's built-in logger
	e.Logger.SetLevel(1) // Set to INFO level

	// // Print custom banner
	// fmt.Println(`
	// 	 ____                  _            ____       _
	// 	/ ___|  ___ _ ____   _(_) ___ ___  / ___| ___ | | __ _ _ __   __ _
	// 	\___ \ / _ \ '__\ \ / / |/ __/ _ \| |  _ / _ \| |/ _' | '_ \ / _' |
	// 	___) |  __/ |   \ V /| | (_|  __/| |_| | (_) | | (_| | | | | (_| |
	// 	|____/ \___|_|    \_/ |_|\___\___| \____|\___/|_|\__,_|_| |_|\__, |
	// 																  |___/
	// `)

	return &EchoServer{
		echo:   e,
		config: cfg,
		log:    log,
	}
}

func (s *EchoServer) SetupMiddlewares() {
	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// CORS middleware
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Request ID middleware
	s.echo.Use(middleware.RequestID())

	// Custom logger middleware (Zap)
	s.echo.Use(s.loggerMiddleware())

	// Timeout middleware
	s.echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
}

func (s *EchoServer) SetupRoutes() {
	// Health check
	s.echo.GET("/health", s.healthCheck)

	// API routes
	s.setupAPIRoutes()
}

func (s *EchoServer) setupAPIRoutes() {
	// API v1 group
	v1 := s.echo.Group("/api/v1")
	{
		// Example: User routes
		userGroup := v1.Group("/users")
		{
			userGroup.GET("", s.getUsers)
			userGroup.GET("/:id", s.getUserByID)
			userGroup.POST("", s.createUser)
			userGroup.PUT("/:id", s.updateUser)
			userGroup.DELETE("/:id", s.deleteUser)
		}

		// Example: Product routes
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

	// Admin routes (with auth middleware example)
	admin := s.echo.Group("/admin", s.authMiddleware())
	{
		admin.GET("/dashboard", s.dashboard)
		admin.GET("/users", s.adminGetUsers)
	}
}

func (s *EchoServer) Start() error {
	// Get host and port with fallback
	host := s.config.GetString("ECHO_HOST")
	if host == "" {
		host = s.config.GetString("APP_HOST")
	}
	if host == "" {
		host = "0.0.0.0"
	}

	port := s.config.GetString("ECHO_PORT")
	if port == "" {
		port = s.config.GetString("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("%s:%s", host, port)

	// Zap logger info
	s.log.Info("Starting Echo server",
		zap.String("framework", "Echo"),
		zap.String("host", host),
		zap.String("port", port),
		zap.String("address", address),
		zap.String("version", s.config.GetString("APP_VERSION")),
	)

	// Echo native log
	s.echo.Logger.Infof("⇨ http server started on %s", address)

	// Manage server lifecycle
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server in goroutine
	go func() {
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.log.Error("Failed to start Echo server", zap.Error(err))
			s.echo.Logger.Fatalf("Server startup error: %v", err)
		}
	}()

	// Zap log success
	s.log.Info("Echo server started successfully",
		zap.String("address", address),
		zap.String("url", fmt.Sprintf("http://%s:%s", host, port)),
	)

	// Wait for shutdown signal
	<-ctx.Done()

	return s.Shutdown(context.Background())
}

func (s *EchoServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down Echo server...")
	s.echo.Logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		s.log.Error("Echo server forced to shutdown", zap.Error(err))
		s.echo.Logger.Fatalf("Server shutdown error: %v", err)
		return err
	}

	s.log.Info("Echo server exited gracefully")
	s.echo.Logger.Info("Server exited gracefully")
	return nil
}

// Custom logger middleware
func (s *EchoServer) loggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()

			// Zap structured logging
			s.log.Info("HTTP Request",
				zap.String("framework", "Echo"),
				zap.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("remote_ip", c.RealIP()),
				zap.Int("status", res.Status),
				zap.Int64("latency_ms", time.Since(start).Milliseconds()),
			)

			// Echo native logging (optional, for visibility)
			s.echo.Logger.Infof("%s %s %d %dms",
				req.Method,
				req.RequestURI,
				res.Status,
				time.Since(start).Milliseconds(),
			)

			return err
		}
	}
}

// Auth middleware example
func (s *EchoServer) authMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authorization token",
				})
			}

			// TODO: Validate token
			// For now, just pass through

			return next(c)
		}
	}
}

// Health check handler
func (s *EchoServer) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"framework": "Echo",
		"message":   "Service is healthy",
		"version":   s.config.GetString("APP_VERSION"),
		"time":      time.Now().Format(time.RFC3339),
	})
}

// Example handlers (placeholder)
func (s *EchoServer) getUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Get all users"})
}

func (s *EchoServer) getUserByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"message": "Get user by ID", "id": id})
}

func (s *EchoServer) createUser(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}

func (s *EchoServer) updateUser(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"message": "User updated", "id": id})
}

func (s *EchoServer) deleteUser(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted", "id": id})
}

func (s *EchoServer) getProducts(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Get all products"})
}

func (s *EchoServer) getProductByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"message": "Get product by ID", "id": id})
}

func (s *EchoServer) createProduct(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{"message": "Product created"})
}

func (s *EchoServer) updateProduct(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"message": "Product updated", "id": id})
}

func (s *EchoServer) deleteProduct(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted", "id": id})
}

func (s *EchoServer) login(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Login successful", "token": "dummy-token"})
}

func (s *EchoServer) register(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{"message": "Registration successful"})
}

func (s *EchoServer) logout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Logout successful"})
}

func (s *EchoServer) dashboard(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Admin dashboard"})
}

func (s *EchoServer) adminGetUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Admin: Get all users"})
}

func (s *EchoServer) GetEcho() *echo.Echo {
	return s.echo
}
