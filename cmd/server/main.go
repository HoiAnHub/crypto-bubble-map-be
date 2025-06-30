package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/cache"
	"crypto-bubble-map-be/internal/infrastructure/config"
	"crypto-bubble-map-be/internal/infrastructure/database"
	"crypto-bubble-map-be/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	logger     *logger.Logger
	neo4j      *database.Neo4jClient
	mongodb    *database.MongoClient
	postgresql *database.PostgreSQLClient
	redis      *cache.RedisClient
	httpServer *http.Server
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	loggerCfg := &logger.Config{
		Level:       cfg.App.LogLevel,
		Environment: cfg.App.Environment,
		Debug:       cfg.App.Debug,
	}

	log, err := logger.NewLogger(loggerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize global logger
	if err := logger.InitGlobalLogger(loggerCfg); err != nil {
		return nil, fmt.Errorf("failed to initialize global logger: %w", err)
	}

	log.Info("Starting Crypto Bubble Map Backend",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.App.Environment),
		zap.String("log_level", cfg.App.LogLevel),
	)

	// Initialize databases
	neo4jClient, err := database.NewNeo4jClient(&cfg.Database.Neo4j, log.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Neo4j: %w", err)
	}

	mongoClient, err := database.NewMongoClient(&cfg.Database.MongoDB, log.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB: %w", err)
	}

	postgresClient, err := database.NewPostgreSQLClient(&cfg.Database.PostgreSQL, log.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Run database migrations
	if err := postgresClient.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Initialize Redis cache
	redisClient, err := cache.NewRedisClient(&cfg.Cache.Redis, &cfg.Cache.TTL, log.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Create MongoDB indexes
	if err := mongoClient.CreateIndexes(context.Background()); err != nil {
		log.Warn("Failed to create MongoDB indexes", zap.Error(err))
	}

	server := &Server{
		config:     cfg,
		logger:     log,
		neo4j:      neo4jClient,
		mongodb:    mongoClient,
		postgresql: postgresClient,
		redis:      redisClient,
	}

	// Setup HTTP server
	if err := server.setupHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	return server, nil
}

// setupHTTPServer configures the HTTP server
func (s *Server) setupHTTPServer() error {
	// Set Gin mode
	gin.SetMode(s.config.Server.Mode)

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(s.corsMiddleware())
	router.Use(s.loggingMiddleware())
	router.Use(s.recoveryMiddleware())

	if s.config.Security.EnableRateLimiting {
		router.Use(s.rateLimitMiddleware())
	}

	// Health check endpoint
	router.GET("/health", s.healthHandler)
	router.GET("/ready", s.readinessHandler)

	// GraphQL endpoint (placeholder - will be implemented later)
	router.POST("/graphql", s.graphqlHandler)

	if s.config.GraphQL.PlaygroundEnabled {
		router.GET("/playground", s.playgroundHandler)
	}

	// Metrics endpoint
	if s.config.Monitoring.EnableMetrics {
		router.GET("/metrics", s.metricsHandler)
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:      router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	return nil
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server",
		zap.String("addr", s.httpServer.Addr),
		zap.String("mode", s.config.Server.Mode),
	)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	s.logger.Info("Server started successfully", zap.String("addr", s.httpServer.Addr))
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		return err
	}

	// Close database connections
	if err := s.neo4j.Close(ctx); err != nil {
		s.logger.Error("Failed to close Neo4j connection", zap.Error(err))
	}

	if err := s.mongodb.Close(ctx); err != nil {
		s.logger.Error("Failed to close MongoDB connection", zap.Error(err))
	}

	if err := s.postgresql.Close(); err != nil {
		s.logger.Error("Failed to close PostgreSQL connection", zap.Error(err))
	}

	if err := s.redis.Close(); err != nil {
		s.logger.Error("Failed to close Redis connection", zap.Error(err))
	}

	// Close logger
	if err := s.logger.Close(); err != nil {
		fmt.Printf("Failed to close logger: %v\n", err)
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

// HTTP Handlers

// healthHandler handles health check requests
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// readinessHandler handles readiness check requests
func (s *Server) readinessHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := map[string]string{
		"neo4j":      "healthy",
		"mongodb":    "healthy",
		"postgresql": "healthy",
		"redis":      "healthy",
	}

	// Check Neo4j
	if err := s.neo4j.Health(ctx); err != nil {
		checks["neo4j"] = "unhealthy"
		s.logger.Error("Neo4j health check failed", zap.Error(err))
	}

	// Check MongoDB
	if err := s.mongodb.Health(ctx); err != nil {
		checks["mongodb"] = "unhealthy"
		s.logger.Error("MongoDB health check failed", zap.Error(err))
	}

	// Check PostgreSQL
	if err := s.postgresql.Health(ctx); err != nil {
		checks["postgresql"] = "unhealthy"
		s.logger.Error("PostgreSQL health check failed", zap.Error(err))
	}

	// Check Redis
	if err := s.redis.Health(ctx); err != nil {
		checks["redis"] = "unhealthy"
		s.logger.Error("Redis health check failed", zap.Error(err))
	}

	// Determine overall status
	status := "ready"
	statusCode := http.StatusOK
	for _, check := range checks {
		if check == "unhealthy" {
			status = "not ready"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"checks":    checks,
		"timestamp": time.Now().UTC(),
	})
}

// graphqlHandler handles GraphQL requests (placeholder)
func (s *Server) graphqlHandler(c *gin.Context) {
	// Simple GraphQL response for testing
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "GraphQL endpoint is working",
			"version": "1.0.0",
		},
	})
}

// playgroundHandler serves the GraphQL playground
func (s *Server) playgroundHandler(c *gin.Context) {
	// This will serve the GraphQL playground HTML
	c.JSON(http.StatusOK, gin.H{
		"message": "GraphQL Playground - coming soon",
	})
}

// metricsHandler handles metrics requests
func (s *Server) metricsHandler(c *gin.Context) {
	// This will be implemented with Prometheus metrics
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics endpoint - coming soon",
	})
}

// Middleware

// corsMiddleware handles CORS
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.Server.EnableCORS {
			origin := c.Request.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range s.config.Server.CORSAllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		c.Next()
	}
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		s.logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// recoveryMiddleware handles panics
func (s *Server) recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		s.logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// rateLimitMiddleware implements rate limiting
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		allowed, err := s.redis.CheckRateLimit(
			c.Request.Context(),
			key,
			int64(s.config.Security.RateLimitRequestsPerMin),
			time.Minute,
		)

		if err != nil {
			s.logger.Error("Rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if !allowed {
			s.logger.Warn("Rate limit exceeded", zap.String("client_ip", clientIP))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// main function
func main() {
	// Create server
	server, err := NewServer()
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Start server
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server exited")
}
