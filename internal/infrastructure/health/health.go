package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/cache"
	"crypto-bubble-map-be/internal/infrastructure/config"
	"crypto-bubble-map-be/internal/infrastructure/database"

	"go.uber.org/zap"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
	StatusUnknown   Status = "unknown"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      Status                 `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     Status                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     time.Duration              `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker interface for health check implementations
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// HealthManager manages health checks for all system components
type HealthManager struct {
	checkers  []HealthChecker
	config    *config.Config
	logger    *zap.Logger
	startTime time.Time
	mu        sync.RWMutex
	lastCheck map[string]ComponentHealth
}

// NewHealthManager creates a new health manager
func NewHealthManager(cfg *config.Config, logger *zap.Logger) *HealthManager {
	return &HealthManager{
		checkers:  make([]HealthChecker, 0),
		config:    cfg,
		logger:    logger,
		startTime: time.Now(),
		lastCheck: make(map[string]ComponentHealth),
	}
}

// RegisterChecker registers a health checker
func (hm *HealthManager) RegisterChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers = append(hm.checkers, checker)
}

// CheckHealth performs health checks on all registered components
func (hm *HealthManager) CheckHealth(ctx context.Context) SystemHealth {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	components := make(map[string]ComponentHealth)
	overallStatus := StatusHealthy

	// Run health checks concurrently
	type result struct {
		name   string
		health ComponentHealth
	}

	resultChan := make(chan result, len(hm.checkers))

	for _, checker := range hm.checkers {
		go func(c HealthChecker) {
			health := c.Check(ctx)
			resultChan <- result{name: c.Name(), health: health}
		}(checker)
	}

	// Collect results
	for i := 0; i < len(hm.checkers); i++ {
		select {
		case res := <-resultChan:
			components[res.name] = res.health
			hm.lastCheck[res.name] = res.health

			// Determine overall status
			switch res.health.Status {
			case StatusUnhealthy:
				overallStatus = StatusUnhealthy
			case StatusDegraded:
				if overallStatus == StatusHealthy {
					overallStatus = StatusDegraded
				}
			}

		case <-ctx.Done():
			hm.logger.Warn("Health check timeout", zap.Error(ctx.Err()))
			overallStatus = StatusUnknown
		}
	}

	return SystemHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    "1.0.0", // TODO: Get from build info or config
		Uptime:     time.Since(hm.startTime),
		Components: components,
	}
}

// GetLastCheck returns the last health check result for a component
func (hm *HealthManager) GetLastCheck(componentName string) (ComponentHealth, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	health, exists := hm.lastCheck[componentName]
	return health, exists
}

// PostgreSQLHealthChecker checks PostgreSQL database health
type PostgreSQLHealthChecker struct {
	client *database.PostgreSQLClient
	logger *zap.Logger
}

// NewPostgreSQLHealthChecker creates a new PostgreSQL health checker
func NewPostgreSQLHealthChecker(client *database.PostgreSQLClient, logger *zap.Logger) *PostgreSQLHealthChecker {
	return &PostgreSQLHealthChecker{
		client: client,
		logger: logger,
	}
}

// Name returns the checker name
func (c *PostgreSQLHealthChecker) Name() string {
	return "postgresql"
}

// Check performs the health check
func (c *PostgreSQLHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        c.Name(),
		LastChecked: start,
		Metadata:    make(map[string]interface{}),
	}

	// Check database connection
	db := c.client.GetDB()
	if db == nil {
		health.Status = StatusUnhealthy
		health.Message = "Database connection is nil"
		health.Duration = time.Since(start)
		return health
	}

	// Ping database
	sqlDB, err := db.DB()
	if err != nil {
		health.Status = StatusUnhealthy
		health.Message = "Failed to get underlying sql.DB"
		health.Error = err.Error()
		health.Duration = time.Since(start)
		return health
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = "Database ping failed"
		health.Error = err.Error()
		health.Duration = time.Since(start)
		return health
	}

	// Get database stats
	stats := sqlDB.Stats()
	health.Metadata["open_connections"] = stats.OpenConnections
	health.Metadata["in_use"] = stats.InUse
	health.Metadata["idle"] = stats.Idle

	// Check if connection pool is healthy
	if stats.OpenConnections > 0 {
		health.Status = StatusHealthy
		health.Message = "Database is healthy"
	} else {
		health.Status = StatusDegraded
		health.Message = "No open database connections"
	}

	health.Duration = time.Since(start)
	return health
}

// MongoHealthChecker checks MongoDB health
type MongoHealthChecker struct {
	client *database.MongoClient
	logger *zap.Logger
}

// NewMongoHealthChecker creates a new MongoDB health checker
func NewMongoHealthChecker(client *database.MongoClient, logger *zap.Logger) *MongoHealthChecker {
	return &MongoHealthChecker{
		client: client,
		logger: logger,
	}
}

// Name returns the checker name
func (c *MongoHealthChecker) Name() string {
	return "mongodb"
}

// Check performs the health check
func (c *MongoHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        c.Name(),
		LastChecked: start,
		Metadata:    make(map[string]interface{}),
	}

	// Check MongoDB health
	if err := c.client.Health(ctx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = "MongoDB health check failed"
		health.Error = err.Error()
		health.Duration = time.Since(start)
		return health
	}

	health.Status = StatusHealthy
	health.Message = "MongoDB is healthy"
	health.Duration = time.Since(start)
	return health
}

// Neo4jHealthChecker checks Neo4j health
type Neo4jHealthChecker struct {
	client *database.Neo4jClient
	logger *zap.Logger
}

// NewNeo4jHealthChecker creates a new Neo4j health checker
func NewNeo4jHealthChecker(client *database.Neo4jClient, logger *zap.Logger) *Neo4jHealthChecker {
	return &Neo4jHealthChecker{
		client: client,
		logger: logger,
	}
}

// Name returns the checker name
func (c *Neo4jHealthChecker) Name() string {
	return "neo4j"
}

// Check performs the health check
func (c *Neo4jHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        c.Name(),
		LastChecked: start,
		Metadata:    make(map[string]interface{}),
	}

	// Check Neo4j health
	if err := c.client.Health(ctx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = "Neo4j health check failed"
		health.Error = err.Error()
		health.Duration = time.Since(start)
		return health
	}

	health.Status = StatusHealthy
	health.Message = "Neo4j is healthy"
	health.Duration = time.Since(start)
	return health
}

// RedisHealthChecker checks Redis health
type RedisHealthChecker struct {
	client *cache.RedisClient
	logger *zap.Logger
}

// NewRedisHealthChecker creates a new Redis health checker
func NewRedisHealthChecker(client *cache.RedisClient, logger *zap.Logger) *RedisHealthChecker {
	return &RedisHealthChecker{
		client: client,
		logger: logger,
	}
}

// Name returns the checker name
func (c *RedisHealthChecker) Name() string {
	return "redis"
}

// Check performs the health check
func (c *RedisHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        c.Name(),
		LastChecked: start,
		Metadata:    make(map[string]interface{}),
	}

	// Check Redis health
	if err := c.client.Health(ctx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = "Redis health check failed"
		health.Error = err.Error()
		health.Duration = time.Since(start)
		return health
	}

	health.Status = StatusHealthy
	health.Message = "Redis is healthy"
	health.Duration = time.Since(start)
	return health
}

// ExternalAPIHealthChecker checks external API health
type ExternalAPIHealthChecker struct {
	config *config.ExternalConfig
	logger *zap.Logger
}

// NewExternalAPIHealthChecker creates a new external API health checker
func NewExternalAPIHealthChecker(config *config.ExternalConfig, logger *zap.Logger) *ExternalAPIHealthChecker {
	return &ExternalAPIHealthChecker{
		config: config,
		logger: logger,
	}
}

// Name returns the checker name
func (c *ExternalAPIHealthChecker) Name() string {
	return "external_apis"
}

// Check performs the health check
func (c *ExternalAPIHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        c.Name(),
		LastChecked: start,
		Metadata:    make(map[string]interface{}),
	}

	// Check if API keys are configured
	apiCount := 0
	configuredAPIs := make([]string, 0)

	if c.config.CoinGeckoAPIKey != "" {
		apiCount++
		configuredAPIs = append(configuredAPIs, "CoinGecko")
	}

	if c.config.OpenAIAPIKey != "" {
		apiCount++
		configuredAPIs = append(configuredAPIs, "OpenAI")
	}

	if c.config.EthereumRPCURL != "" {
		apiCount++
		configuredAPIs = append(configuredAPIs, "Ethereum RPC")
	}

	health.Metadata["configured_apis"] = configuredAPIs
	health.Metadata["api_count"] = apiCount

	if apiCount > 0 {
		health.Status = StatusHealthy
		health.Message = fmt.Sprintf("%d external APIs configured", apiCount)
	} else {
		health.Status = StatusDegraded
		health.Message = "No external APIs configured"
	}

	health.Duration = time.Since(start)
	return health
}

// SetupHealthCheckers sets up all health checkers
func SetupHealthCheckers(
	hm *HealthManager,
	postgres *database.PostgreSQLClient,
	mongo *database.MongoClient,
	neo4j *database.Neo4jClient,
	redis *cache.RedisClient,
	config *config.Config,
	logger *zap.Logger,
) {
	// Register database health checkers
	if postgres != nil {
		hm.RegisterChecker(NewPostgreSQLHealthChecker(postgres, logger))
	}

	if mongo != nil {
		hm.RegisterChecker(NewMongoHealthChecker(mongo, logger))
	}

	if neo4j != nil {
		hm.RegisterChecker(NewNeo4jHealthChecker(neo4j, logger))
	}

	if redis != nil {
		hm.RegisterChecker(NewRedisHealthChecker(redis, logger))
	}

	// Register external API health checker
	hm.RegisterChecker(NewExternalAPIHealthChecker(&config.External, logger))
}
