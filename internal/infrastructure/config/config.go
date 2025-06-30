package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Cache      CacheConfig      `mapstructure:"cache"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	GraphQL    GraphQLConfig    `mapstructure:"graphql"`
	External   ExternalConfig   `mapstructure:"external"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Security   SecurityConfig   `mapstructure:"security"`
	App        AppConfig        `mapstructure:"app"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host                 string        `mapstructure:"host"`
	Port                 int           `mapstructure:"port"`
	Mode                 string        `mapstructure:"mode"`
	ReadTimeout          time.Duration `mapstructure:"read_timeout"`
	WriteTimeout         time.Duration `mapstructure:"write_timeout"`
	IdleTimeout          time.Duration `mapstructure:"idle_timeout"`
	MaxRequestSize       string        `mapstructure:"max_request_size"`
	CORSAllowedOrigins   []string      `mapstructure:"cors_allowed_origins"`
	EnableCORS           bool          `mapstructure:"enable_cors"`
	EnableRequestLogging bool          `mapstructure:"enable_request_logging"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Neo4j      Neo4jConfig      `mapstructure:"neo4j"`
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`
	PostgreSQL PostgreSQLConfig `mapstructure:"postgresql"`
}

// Neo4jConfig holds Neo4j configuration
type Neo4jConfig struct {
	URI                     string        `mapstructure:"uri"`
	Username                string        `mapstructure:"username"`
	Password                string        `mapstructure:"password"`
	Database                string        `mapstructure:"database"`
	MaxConnectionPoolSize   int           `mapstructure:"max_connection_pool_size"`
	ConnectionTimeout       time.Duration `mapstructure:"connection_timeout"`
	MaxTransactionRetryTime time.Duration `mapstructure:"max_transaction_retry_time"`
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI               string        `mapstructure:"uri"`
	Database          string        `mapstructure:"database"`
	MaxPoolSize       uint64        `mapstructure:"max_pool_size"`
	MinPoolSize       uint64        `mapstructure:"min_pool_size"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
	SocketTimeout     time.Duration `mapstructure:"socket_timeout"`
}

// PostgreSQLConfig holds PostgreSQL configuration
type PostgreSQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Redis RedisConfig `mapstructure:"redis"`
	TTL   TTLConfig   `mapstructure:"ttl"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	MaxRetries   int           `mapstructure:"max_retries"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// TTLConfig holds TTL configuration for different cache types
type TTLConfig struct {
	WalletNetwork   time.Duration `mapstructure:"wallet_network"`
	WalletRankings  time.Duration `mapstructure:"wallet_rankings"`
	DashboardStats  time.Duration `mapstructure:"dashboard_stats"`
	RiskScores      time.Duration `mapstructure:"risk_scores"`
	NetworkStats    time.Duration `mapstructure:"network_stats"`
	TransactionData time.Duration `mapstructure:"transaction_data"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	Expiry        time.Duration `mapstructure:"expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
	Issuer        string        `mapstructure:"issuer"`
	Audience      string        `mapstructure:"audience"`
}

// GraphQLConfig holds GraphQL configuration
type GraphQLConfig struct {
	PlaygroundEnabled     bool `mapstructure:"playground_enabled"`
	IntrospectionEnabled  bool `mapstructure:"introspection_enabled"`
	ComplexityLimit       int  `mapstructure:"complexity_limit"`
	DepthLimit            int  `mapstructure:"depth_limit"`
	EnableQueryValidation bool `mapstructure:"enable_query_validation"`
	EnableTracing         bool `mapstructure:"enable_tracing"`
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	EthereumRPCURL  string `mapstructure:"ethereum_rpc_url"`
	CoinGeckoAPIKey string `mapstructure:"coingecko_api_key"`
	InfuraProjectID string `mapstructure:"infura_project_id"`
	AlchemyAPIKey   string `mapstructure:"alchemy_api_key"`
	OpenAIAPIKey    string `mapstructure:"openai_api_key"`
	OpenAIModel     string `mapstructure:"openai_model"`
	OpenAIBaseURL   string `mapstructure:"openai_base_url"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	EnableMetrics   bool   `mapstructure:"enable_metrics"`
	MetricsPort     int    `mapstructure:"metrics_port"`
	EnableTracing   bool   `mapstructure:"enable_tracing"`
	JaegerEndpoint  string `mapstructure:"jaeger_endpoint"`
	EnableProfiling bool   `mapstructure:"enable_profiling"`
	ProfilingPort   int    `mapstructure:"profiling_port"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EnableRateLimiting      bool          `mapstructure:"enable_rate_limiting"`
	RateLimitRequestsPerMin int           `mapstructure:"rate_limit_requests_per_minute"`
	RateLimitBurst          int           `mapstructure:"rate_limit_burst"`
	EnableIPWhitelist       bool          `mapstructure:"enable_ip_whitelist"`
	IPWhitelist             []string      `mapstructure:"ip_whitelist"`
	EnableAPIKeyAuth        bool          `mapstructure:"enable_api_key_auth"`
	PasswordMinLength       int           `mapstructure:"password_min_length"`
	PasswordRequireSpecial  bool          `mapstructure:"password_require_special"`
	SessionTimeout          time.Duration `mapstructure:"session_timeout"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment               string        `mapstructure:"environment"`
	LogLevel                  string        `mapstructure:"log_level"`
	Debug                     bool          `mapstructure:"debug"`
	EnableBackgroundJobs      bool          `mapstructure:"enable_background_jobs"`
	RiskScoreUpdateInterval   time.Duration `mapstructure:"risk_score_update_interval"`
	WalletStatsUpdateInterval time.Duration `mapstructure:"wallet_stats_update_interval"`
	CacheCleanupInterval      time.Duration `mapstructure:"cache_cleanup_interval"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := loadEnvFile(); err != nil {
		// Log warning but don't fail - .env file is optional
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/crypto-bubble-map-be")

	// Set default values
	setDefaults()

	// Enable environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables to config keys
	bindEnvironmentVariables()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	// Try to load .env file from current directory
	if err := loadEnvFromPath(".env"); err == nil {
		return nil
	}

	// Try to load from parent directory (for when running from subdirectories)
	if err := loadEnvFromPath("../.env"); err == nil {
		return nil
	}

	return fmt.Errorf(".env file not found")
}

// loadEnvFromPath loads environment variables from a specific path
func loadEnvFromPath(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// Read the file and set environment variables
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Parse and set environment variables
	envMap, err := parseEnvFile(file)
	if err != nil {
		return err
	}

	for key, value := range envMap {
		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

// parseEnvFile parses a .env file and returns a map of key-value pairs
func parseEnvFile(file *os.File) (map[string]string, error) {
	envMap := make(map[string]string)

	// Read file content
	content := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			content = append(content, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Parse lines
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		envMap[key] = value
	}

	return envMap, nil
}

// bindEnvironmentVariables binds specific environment variables to config keys
func bindEnvironmentVariables() {
	// Server configuration
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.mode", "GIN_MODE")

	// Neo4j configuration
	viper.BindEnv("database.neo4j.uri", "NEO4J_URI")
	viper.BindEnv("database.neo4j.username", "NEO4J_USERNAME")
	viper.BindEnv("database.neo4j.password", "NEO4J_PASSWORD")
	viper.BindEnv("database.neo4j.database", "NEO4J_DATABASE")
	viper.BindEnv("database.neo4j.max_connection_pool_size", "NEO4J_MAX_CONNECTION_POOL_SIZE")
	viper.BindEnv("database.neo4j.connection_timeout", "NEO4J_CONNECTION_TIMEOUT")
	viper.BindEnv("database.neo4j.max_transaction_retry_time", "NEO4J_CONNECTION_ACQUISITION_TIMEOUT")

	// MongoDB configuration
	viper.BindEnv("database.mongodb.uri", "MONGO_URI")
	viper.BindEnv("database.mongodb.database", "MONGO_DATABASE")
	viper.BindEnv("database.mongodb.max_pool_size", "MONGO_MAX_POOL_SIZE")
	viper.BindEnv("database.mongodb.min_pool_size", "MONGO_MIN_POOL_SIZE")
	viper.BindEnv("database.mongodb.connection_timeout", "MONGO_CONNECTION_TIMEOUT")

	// PostgreSQL configuration
	viper.BindEnv("database.postgresql.host", "POSTGRES_HOST")
	viper.BindEnv("database.postgresql.port", "POSTGRES_PORT")
	viper.BindEnv("database.postgresql.user", "POSTGRES_USER")
	viper.BindEnv("database.postgresql.password", "POSTGRES_PASSWORD")
	viper.BindEnv("database.postgresql.database", "POSTGRES_DB")
	viper.BindEnv("database.postgresql.ssl_mode", "POSTGRES_SSL_MODE")
	viper.BindEnv("database.postgresql.max_open_conns", "POSTGRES_MAX_OPEN_CONNS")
	viper.BindEnv("database.postgresql.max_idle_conns", "POSTGRES_MAX_IDLE_CONNS")
	viper.BindEnv("database.postgresql.conn_max_lifetime", "POSTGRES_CONN_MAX_LIFETIME")

	// Redis configuration
	viper.BindEnv("cache.redis.host", "REDIS_HOST")
	viper.BindEnv("cache.redis.port", "REDIS_PORT")
	viper.BindEnv("cache.redis.password", "REDIS_PASSWORD")
	viper.BindEnv("cache.redis.db", "REDIS_DB")
	viper.BindEnv("cache.redis.max_retries", "REDIS_MAX_RETRIES")
	viper.BindEnv("cache.redis.pool_size", "REDIS_POOL_SIZE")
	viper.BindEnv("cache.redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")

	// JWT configuration
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiry", "JWT_EXPIRY")
	viper.BindEnv("jwt.refresh_expiry", "JWT_REFRESH_EXPIRY")

	// Application configuration
	viper.BindEnv("app.environment", "APP_ENV")
	viper.BindEnv("app.log_level", "LOG_LEVEL")
	viper.BindEnv("app.debug", "DEBUG")

	// Security configuration
	viper.BindEnv("security.enable_rate_limiting", "ENABLE_RATE_LIMITING")
	viper.BindEnv("security.rate_limit_requests_per_minute", "RATE_LIMIT_REQUESTS_PER_MINUTE")
	viper.BindEnv("security.rate_limit_burst", "RATE_LIMIT_BURST")

	// Cache TTL configuration
	viper.BindEnv("cache.ttl.wallet_network", "CACHE_TTL_WALLET_NETWORK")
	viper.BindEnv("cache.ttl.wallet_rankings", "CACHE_TTL_WALLET_RANKINGS")
	viper.BindEnv("cache.ttl.dashboard_stats", "CACHE_TTL_DASHBOARD_STATS")
	viper.BindEnv("cache.ttl.risk_scores", "CACHE_TTL_RISK_SCORES")

	// External APIs
	viper.BindEnv("external.ethereum_rpc_url", "ETHEREUM_RPC_URL")
	viper.BindEnv("external.coingecko_api_key", "COINGECKO_API_KEY")
	viper.BindEnv("external.infura_project_id", "INFURA_PROJECT_ID")
	viper.BindEnv("external.alchemy_api_key", "ALCHEMY_API_KEY")
	viper.BindEnv("external.openai_api_key", "OPENAI_API_KEY")
	viper.BindEnv("external.openai_model", "OPENAI_MODEL")
	viper.BindEnv("external.openai_base_url", "OPENAI_BASE_URL")

	// GraphQL configuration
	viper.BindEnv("graphql.playground_enabled", "GRAPHQL_PLAYGROUND_ENABLED")
	viper.BindEnv("graphql.introspection_enabled", "GRAPHQL_INTROSPECTION_ENABLED")
	viper.BindEnv("graphql.complexity_limit", "GRAPHQL_COMPLEXITY_LIMIT")
	viper.BindEnv("graphql.depth_limit", "GRAPHQL_DEPTH_LIMIT")

	// Monitoring configuration
	viper.BindEnv("monitoring.enable_metrics", "ENABLE_METRICS")
	viper.BindEnv("monitoring.metrics_port", "METRICS_PORT")
	viper.BindEnv("monitoring.enable_tracing", "ENABLE_TRACING")
	viper.BindEnv("monitoring.jaeger_endpoint", "JAEGER_ENDPOINT")

	// Background jobs
	viper.BindEnv("app.enable_background_jobs", "ENABLE_BACKGROUND_JOBS")
	viper.BindEnv("app.risk_score_update_interval", "RISK_SCORE_UPDATE_INTERVAL")
	viper.BindEnv("app.wallet_stats_update_interval", "WALLET_STATS_UPDATE_INTERVAL")
	viper.BindEnv("app.cache_cleanup_interval", "CACHE_CLEANUP_INTERVAL")
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.max_request_size", "10MB")
	viper.SetDefault("server.enable_cors", true)
	viper.SetDefault("server.enable_request_logging", true)

	// Neo4j defaults - Updated to match your configuration
	viper.SetDefault("database.neo4j.uri", "neo4j://localhost:7687")
	viper.SetDefault("database.neo4j.username", "neo4j")
	viper.SetDefault("database.neo4j.password", "password")
	viper.SetDefault("database.neo4j.database", "neo4j")
	viper.SetDefault("database.neo4j.max_connection_pool_size", 50)
	viper.SetDefault("database.neo4j.connection_timeout", "10s")
	viper.SetDefault("database.neo4j.max_transaction_retry_time", "60s")

	// MongoDB defaults - Updated to match your configuration
	viper.SetDefault("database.mongodb.uri", "mongodb+srv://haitranwang:eURhdPjFc10NGyDR@cluster0.kzyty5l.mongodb.net/ethereum_raw_data?authSource=admin&maxPoolSize=10&minPoolSize=2&maxIdleTimeMS=60000&serverSelectionTimeoutMS=10000&socketTimeoutMS=60000&connectTimeoutMS=15000&heartbeatFrequencyMS=30000&retryWrites=true&retryReads=true&maxConnecting=3")
	viper.SetDefault("database.mongodb.database", "ethereum_raw_data")
	viper.SetDefault("database.mongodb.max_pool_size", 50)
	viper.SetDefault("database.mongodb.min_pool_size", 2)
	viper.SetDefault("database.mongodb.connection_timeout", "30s")
	viper.SetDefault("database.mongodb.socket_timeout", "60s")

	// PostgreSQL defaults - Updated to match your configuration
	viper.SetDefault("database.postgresql.host", "localhost")
	viper.SetDefault("database.postgresql.port", 5433)
	viper.SetDefault("database.postgresql.user", "hoianhub_user")
	viper.SetDefault("database.postgresql.password", "hoianhub_password")
	viper.SetDefault("database.postgresql.database", "postgres")
	viper.SetDefault("database.postgresql.ssl_mode", "disable")
	viper.SetDefault("database.postgresql.max_open_conns", 25)
	viper.SetDefault("database.postgresql.max_idle_conns", 5)
	viper.SetDefault("database.postgresql.conn_max_lifetime", "5m")

	// Redis defaults
	viper.SetDefault("cache.redis.host", "localhost")
	viper.SetDefault("cache.redis.port", 6379)
	viper.SetDefault("cache.redis.password", "")
	viper.SetDefault("cache.redis.db", 0)
	viper.SetDefault("cache.redis.max_retries", 3)
	viper.SetDefault("cache.redis.pool_size", 10)
	viper.SetDefault("cache.redis.min_idle_conns", 5)
	viper.SetDefault("cache.redis.dial_timeout", "5s")
	viper.SetDefault("cache.redis.read_timeout", "3s")
	viper.SetDefault("cache.redis.write_timeout", "3s")

	// Cache TTL defaults
	viper.SetDefault("cache.ttl.wallet_network", "5m")
	viper.SetDefault("cache.ttl.wallet_rankings", "10m")
	viper.SetDefault("cache.ttl.dashboard_stats", "3m")
	viper.SetDefault("cache.ttl.risk_scores", "15m")
	viper.SetDefault("cache.ttl.network_stats", "5m")
	viper.SetDefault("cache.ttl.transaction_data", "1m")

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiry", "24h")
	viper.SetDefault("jwt.refresh_expiry", "168h")
	viper.SetDefault("jwt.issuer", "crypto-bubble-map-be")
	viper.SetDefault("jwt.audience", "crypto-bubble-map")

	// GraphQL defaults
	viper.SetDefault("graphql.playground_enabled", true)
	viper.SetDefault("graphql.introspection_enabled", true)
	viper.SetDefault("graphql.complexity_limit", 1000)
	viper.SetDefault("graphql.depth_limit", 15)
	viper.SetDefault("graphql.enable_query_validation", true)
	viper.SetDefault("graphql.enable_tracing", false)

	// Monitoring defaults
	viper.SetDefault("monitoring.enable_metrics", true)
	viper.SetDefault("monitoring.metrics_port", 9090)
	viper.SetDefault("monitoring.enable_tracing", false)
	viper.SetDefault("monitoring.enable_profiling", false)
	viper.SetDefault("monitoring.profiling_port", 6060)

	// External API defaults
	viper.SetDefault("external.openai_model", "gpt-3.5-turbo")
	viper.SetDefault("external.openai_base_url", "https://api.openai.com/v1")

	// Security defaults
	viper.SetDefault("security.enable_rate_limiting", true)
	viper.SetDefault("security.rate_limit_requests_per_minute", 100)
	viper.SetDefault("security.rate_limit_burst", 20)
	viper.SetDefault("security.enable_ip_whitelist", false)
	viper.SetDefault("security.enable_api_key_auth", false)
	viper.SetDefault("security.password_min_length", 8)
	viper.SetDefault("security.password_require_special", true)
	viper.SetDefault("security.session_timeout", "24h")

	// App defaults
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.debug", false)
	viper.SetDefault("app.enable_background_jobs", true)
	viper.SetDefault("app.risk_score_update_interval", "1h")
	viper.SetDefault("app.wallet_stats_update_interval", "30m")
	viper.SetDefault("app.cache_cleanup_interval", "6h")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.JWT.Secret == "" || c.JWT.Secret == "your-secret-key" {
		return fmt.Errorf("JWT secret must be set and not be the default value")
	}

	if c.Database.Neo4j.URI == "" {
		return fmt.Errorf("Neo4j URI must be set")
	}

	if c.Database.MongoDB.URI == "" {
		return fmt.Errorf("MongoDB URI must be set")
	}

	if c.Database.PostgreSQL.Host == "" {
		return fmt.Errorf("PostgreSQL host must be set")
	}

	return nil
}

// GetDSN returns the PostgreSQL DSN
func (c *PostgreSQLConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
