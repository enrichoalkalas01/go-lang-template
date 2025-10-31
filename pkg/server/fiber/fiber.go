package fiber

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type FiberServer struct {
	fiber  *fiber.App
	config *viper.Viper
	log    *zap.Logger
}

func NewFiberServer(cfg *viper.Viper, log *zap.Logger) *FiberServer {
	app := fiber.New(fiber.Config{
		AppName:               cfg.GetString("APP_NAME"),
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		DisableStartupMessage: false, // Show Fiber startup message
	})

	return &FiberServer{
		fiber:  app,
		config: cfg,
		log:    log,
	}
}

func (s *FiberServer) SetupMiddlewares() {
	// Recovery middleware
	s.fiber.Use(recover.New())

	// CORS middleware
	s.fiber.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Request ID middleware
	s.fiber.Use(requestid.New())

	// Fiber native logger middleware (optional - shows in console)
	s.fiber.Use(fiberlogger.New(fiberlogger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	}))

	// Custom Zap logger middleware
	s.fiber.Use(s.loggerMiddleware())
}

func (s *FiberServer) SetupRoutes() {
	// Health check
	s.fiber.Get("/health", s.healthCheck)

	// API routes
	s.setupAPIRoutes()
}

func (s *FiberServer) setupAPIRoutes() {
	// API v1 group
	v1 := s.fiber.Group("/api/v1")

	// User routes
	userGroup := v1.Group("/users")
	userGroup.Get("/", s.getUsers)
	userGroup.Get("/:id", s.getUserByID)
	userGroup.Post("/", s.createUser)
	userGroup.Put("/:id", s.updateUser)
	userGroup.Delete("/:id", s.deleteUser)

	// Product routes
	productGroup := v1.Group("/products")
	productGroup.Get("/", s.getProducts)
	productGroup.Get("/:id", s.getProductByID)
	productGroup.Post("/", s.createProduct)
	productGroup.Put("/:id", s.updateProduct)
	productGroup.Delete("/:id", s.deleteProduct)

	// Auth routes
	authGroup := v1.Group("/auth")
	authGroup.Post("/login", s.login)
	authGroup.Post("/register", s.register)
	authGroup.Post("/logout", s.logout)

	// Admin routes (with auth middleware)
	admin := s.fiber.Group("/admin", s.authMiddleware())
	admin.Get("/dashboard", s.dashboard)
	admin.Get("/users", s.adminGetUsers)
}

func (s *FiberServer) Start() error {
	// Get host and port with fallback
	host := s.config.GetString("FIBER_HOST")
	if host == "" {
		host = s.config.GetString("APP_HOST")
	}
	if host == "" {
		host = "0.0.0.0"
	}

	port := s.config.GetString("FIBER_PORT")
	if port == "" {
		port = s.config.GetString("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("%s:%s", host, port)

	// Zap logger info
	s.log.Info("Starting Fiber server",
		zap.String("framework", "Fiber"),
		zap.String("host", host),
		zap.String("port", port),
		zap.String("address", address),
		zap.String("version", s.config.GetString("APP_VERSION")),
	)

	// Fiber native log
	log.Printf("⚡ Fiber server starting on http://%s", address)

	// ✅ ADD: Manage server lifecycle
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// ✅ ADD: Start server in goroutine
	go func() {
		if err := s.fiber.Listen(address); err != nil {
			s.log.Error("Fiber server startup error", zap.Error(err))
			log.Fatalf("Server startup error: %v", err)
		}
	}()

	s.log.Info("Fiber server started successfully",
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

func (s *FiberServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down Fiber server...")
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.fiber.ShutdownWithContext(shutdownCtx); err != nil {
		s.log.Error("Fiber server forced to shutdown", zap.Error(err))
		log.Fatalf("Server shutdown error: %v", err)
		return err
	}

	s.log.Info("Fiber server exited gracefully")
	log.Println("Server exited gracefully")
	return nil
}

// Custom logger middleware (Zap)
func (s *FiberServer) loggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		// Zap structured logging
		s.log.Info("HTTP Request",
			zap.String("framework", "Fiber"),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
			zap.String("method", c.Method()),
			zap.String("uri", c.OriginalURL()),
			zap.String("remote_ip", c.IP()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		)

		return err
	}
}

// Auth middleware
func (s *FiberServer) authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization token",
			})
		}

		// TODO: Validate token

		return c.Next()
	}
}

// Health check handler
func (s *FiberServer) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"framework": "Fiber",
		"message":   "Service is healthy",
		"version":   s.config.GetString("APP_VERSION"),
		"time":      time.Now().Format(time.RFC3339),
	})
}

// Example handlers
func (s *FiberServer) getUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get all users"})
}

func (s *FiberServer) getUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Get user by ID", "id": id})
}

func (s *FiberServer) createUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created"})
}

func (s *FiberServer) updateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "User updated", "id": id})
}

func (s *FiberServer) deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "User deleted", "id": id})
}

func (s *FiberServer) getProducts(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get all products"})
}

func (s *FiberServer) getProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Get product by ID", "id": id})
}

func (s *FiberServer) createProduct(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Product created"})
}

func (s *FiberServer) updateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Product updated", "id": id})
}

func (s *FiberServer) deleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Product deleted", "id": id})
}

func (s *FiberServer) login(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Login successful", "token": "dummy-token"})
}

func (s *FiberServer) register(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Registration successful"})
}

func (s *FiberServer) logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Logout successful"})
}

func (s *FiberServer) dashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Admin dashboard"})
}

func (s *FiberServer) adminGetUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Admin: Get all users"})
}

func (s *FiberServer) GetFiber() *fiber.App {
	return s.fiber
}
