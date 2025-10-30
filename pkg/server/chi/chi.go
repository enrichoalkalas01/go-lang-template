package chi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ChiServer struct {
	chi    *chi.Mux
	config *viper.Viper
	log    *zap.Logger
	server *http.Server
}

func NewChiServer(cfg *viper.Viper, log *zap.Logger) *ChiServer {
	r := chi.NewRouter()

	return &ChiServer{
		chi:    r,
		config: cfg,
		log:    log,
	}
}

func (s *ChiServer) SetupMiddlewares() {
	// Recovery middleware
	s.chi.Use(middleware.Recoverer)

	// CORS middleware
	s.chi.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Request ID middleware
	s.chi.Use(middleware.RequestID)

	// Chi native logger middleware (compact)
	s.chi.Use(middleware.Logger)

	// Real IP middleware
	s.chi.Use(middleware.RealIP)

	// Timeout middleware
	s.chi.Use(middleware.Timeout(30 * time.Second))

	// Custom Zap logger middleware
	s.chi.Use(s.loggerMiddleware())
}

func (s *ChiServer) SetupRoutes() {
	// Health check
	s.chi.Get("/health", s.healthCheck)

	// API routes
	s.setupAPIRoutes()
}

func (s *ChiServer) setupAPIRoutes() {
	// API v1 routes
	s.chi.Route("/api/v1", func(r chi.Router) {
		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", s.getUsers)
			r.Get("/{id}", s.getUserByID)
			r.Post("/", s.createUser)
			r.Put("/{id}", s.updateUser)
			r.Delete("/{id}", s.deleteUser)
		})

		// Product routes
		r.Route("/products", func(r chi.Router) {
			r.Get("/", s.getProducts)
			r.Get("/{id}", s.getProductByID)
			r.Post("/", s.createProduct)
			r.Put("/{id}", s.updateProduct)
			r.Delete("/{id}", s.deleteProduct)
		})

		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", s.login)
			r.Post("/register", s.register)
			r.Post("/logout", s.logout)
		})
	})

	// Admin routes (with auth middleware)
	s.chi.Route("/admin", func(r chi.Router) {
		r.Use(s.authMiddleware())
		r.Get("/dashboard", s.dashboard)
		r.Get("/users", s.adminGetUsers)
	})
}

func (s *ChiServer) Start() error {
	// Get host and port with fallback
	host := s.config.GetString("CHI_HOST")
	if host == "" {
		host = s.config.GetString("APP_HOST")
	}
	if host == "" {
		host = "0.0.0.0"
	}

	port := s.config.GetString("CHI_PORT")
	if port == "" {
		port = s.config.GetString("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("%s:%s", host, port)

	// Zap logger info
	s.log.Info("Starting Chi server",
		zap.String("framework", "Chi"),
		zap.String("host", host),
		zap.String("port", port),
		zap.String("address", address),
		zap.String("version", s.config.GetString("APP_VERSION")),
	)

	// Chi native log
	log.Printf("[CHI] Starting server on http://%s", address)
	log.Printf("[CHI] Routes:")
	s.printRoutes()

	s.server = &http.Server{
		Addr:         address,
		Handler:      s.chi,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Manage server lifecycle
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("Failed to start Chi server", zap.Error(err))
			log.Fatalf("Server startup error: %v", err)
		}
	}()

	// Zap log success
	s.log.Info("Chi server started successfully",
		zap.String("address", address),
		zap.String("url", fmt.Sprintf("http://%s:%s", host, port)),
	)

	// Wait for shutdown signal
	<-ctx.Done()

	return s.Shutdown(context.Background())
}

func (s *ChiServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down Chi server...")
	log.Println("[CHI] Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.log.Error("Chi server forced to shutdown", zap.Error(err))
		log.Fatalf("Server shutdown error: %v", err)
		return err
	}

	s.log.Info("Chi server exited gracefully")
	log.Println("[CHI] Server exited gracefully")
	return nil
}

// Print routes (Chi feature)
func (s *ChiServer) printRoutes() {
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("[CHI]   %-7s %s", method, route)
		return nil
	}
	if err := chi.Walk(s.chi, walkFunc); err != nil {
		s.log.Error("Failed to walk routes", zap.Error(err))
	}
}

// Custom logger middleware (Zap)
func (s *ChiServer) loggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			// Get request ID from context
			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Zap structured logging
			s.log.Info("HTTP Request",
				zap.String("framework", "Chi"),
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.String("remote_ip", r.RemoteAddr),
				zap.Int("status", ww.Status()),
				zap.Int64("latency_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

// Auth middleware
func (s *ChiServer) authMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "missing authorization token"}`))
				return
			}

			// TODO: Validate token

			next.ServeHTTP(w, r)
		})
	}
}

// Health check handler
func (s *ChiServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{
		"status": "ok",
		"framework": "Chi",
		"message": "Service is healthy",
		"version": "%s",
		"time": "%s"
	}`, s.config.GetString("APP_VERSION"), time.Now().Format(time.RFC3339))
	w.Write([]byte(response))
}

// Example handlers
func (s *ChiServer) getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Get all users"}`))
}

func (s *ChiServer) getUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Get user by ID", "id": "%s"}`, id)))
}

func (s *ChiServer) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "User created"}`))
}

func (s *ChiServer) updateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "User updated", "id": "%s"}`, id)))
}

func (s *ChiServer) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "User deleted", "id": "%s"}`, id)))
}

func (s *ChiServer) getProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Get all products"}`))
}

func (s *ChiServer) getProductByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Get product by ID", "id": "%s"}`, id)))
}

func (s *ChiServer) createProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Product created"}`))
}

func (s *ChiServer) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Product updated", "id": "%s"}`, id)))
}

func (s *ChiServer) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Product deleted", "id": "%s"}`, id)))
}

func (s *ChiServer) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Login successful", "token": "dummy-token"}`))
}

func (s *ChiServer) register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Registration successful"}`))
}

func (s *ChiServer) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logout successful"}`))
}

func (s *ChiServer) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Admin dashboard"}`))
}

func (s *ChiServer) adminGetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Admin: Get all users"}`))
}

func (s *ChiServer) GetChi() *chi.Mux {
	return s.chi
}
