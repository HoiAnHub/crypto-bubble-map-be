package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
	config *config.RedisConfig
	ttl    *config.TTLConfig
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.RedisConfig, ttlCfg *config.TTLConfig, logger *zap.Logger) (*RedisClient, error) {
	// Configure Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	client := &RedisClient{
		client: rdb,
		logger: logger,
		config: cfg,
		ttl:    ttlCfg,
	}

	logger.Info("Redis client initialized successfully",
		zap.String("addr", cfg.GetAddr()),
		zap.Int("db", cfg.DB),
	)

	return client, nil
}

// Close closes the Redis connection
func (c *RedisClient) Close() error {
	return c.client.Close()
}

// GetClient returns the Redis client instance
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// Set stores a value in Redis with TTL
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal value for Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("Failed to set value in Redis",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Get retrieves a value from Redis
func (c *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		c.logger.Error("Failed to get value from Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		c.logger.Error("Failed to unmarshal value from Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Delete removes a key from Redis
func (c *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error("Failed to delete keys from Redis",
			zap.Strings("keys", keys),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Exists checks if a key exists in Redis
func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to check key existence in Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return false, err
	}
	return count > 0, nil
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (c *RedisClient) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal value for Redis SetNX",
			zap.String("key", key),
			zap.Error(err),
		)
		return false, err
	}

	result, err := c.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		c.logger.Error("Failed to execute SetNX in Redis",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return false, err
	}

	return result, nil
}

// Increment increments a numeric value
func (c *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to increment key in Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return 0, err
	}
	return result, nil
}

// IncrementWithExpiry increments a key and sets expiry if it's a new key
func (c *RedisClient) IncrementWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.client.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.logger.Error("Failed to increment with expiry in Redis",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return 0, err
	}

	return incrCmd.Val(), nil
}

// GetMultiple retrieves multiple values from Redis
func (c *RedisClient) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		c.logger.Error("Failed to get multiple values from Redis",
			zap.Strings("keys", keys),
			zap.Error(err),
		)
		return nil, err
	}

	data := make(map[string]interface{})
	for i, result := range results {
		if result != nil {
			var value interface{}
			if err := json.Unmarshal([]byte(result.(string)), &value); err != nil {
				c.logger.Warn("Failed to unmarshal value from Redis",
					zap.String("key", keys[i]),
					zap.Error(err),
				)
				continue
			}
			data[keys[i]] = value
		}
	}

	return data, nil
}

// SetMultiple sets multiple key-value pairs
func (c *RedisClient) SetMultiple(ctx context.Context, data map[string]interface{}, ttl time.Duration) error {
	pipe := c.client.TxPipeline()

	for key, value := range data {
		jsonData, err := json.Marshal(value)
		if err != nil {
			c.logger.Error("Failed to marshal value for Redis SetMultiple",
				zap.String("key", key),
				zap.Error(err),
			)
			continue
		}
		pipe.Set(ctx, key, jsonData, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.logger.Error("Failed to set multiple values in Redis",
			zap.Int("count", len(data)),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Cache-specific methods with predefined TTLs

// SetWalletNetwork caches wallet network data
func (c *RedisClient) SetWalletNetwork(ctx context.Context, address string, depth int, data interface{}) error {
	key := fmt.Sprintf("wallet_network:%s:%d", address, depth)
	return c.Set(ctx, key, data, c.ttl.WalletNetwork)
}

// GetWalletNetwork retrieves cached wallet network data
func (c *RedisClient) GetWalletNetwork(ctx context.Context, address string, depth int, dest interface{}) error {
	key := fmt.Sprintf("wallet_network:%s:%d", address, depth)
	return c.Get(ctx, key, dest)
}

// SetWalletRankings caches wallet rankings
func (c *RedisClient) SetWalletRankings(ctx context.Context, category string, data interface{}) error {
	key := fmt.Sprintf("wallet_rankings:%s", category)
	return c.Set(ctx, key, data, c.ttl.WalletRankings)
}

// GetWalletRankings retrieves cached wallet rankings
func (c *RedisClient) GetWalletRankings(ctx context.Context, category string, dest interface{}) error {
	key := fmt.Sprintf("wallet_rankings:%s", category)
	return c.Get(ctx, key, dest)
}

// SetDashboardStats caches dashboard statistics
func (c *RedisClient) SetDashboardStats(ctx context.Context, networkID string, data interface{}) error {
	key := fmt.Sprintf("dashboard_stats:%s", networkID)
	return c.Set(ctx, key, data, c.ttl.DashboardStats)
}

// GetDashboardStats retrieves cached dashboard statistics
func (c *RedisClient) GetDashboardStats(ctx context.Context, networkID string, dest interface{}) error {
	key := fmt.Sprintf("dashboard_stats:%s", networkID)
	return c.Get(ctx, key, dest)
}

// SetRiskScore caches risk score data
func (c *RedisClient) SetRiskScore(ctx context.Context, address string, data interface{}) error {
	key := fmt.Sprintf("risk_score:%s", address)
	return c.Set(ctx, key, data, c.ttl.RiskScores)
}

// GetRiskScore retrieves cached risk score data
func (c *RedisClient) GetRiskScore(ctx context.Context, address string, dest interface{}) error {
	key := fmt.Sprintf("risk_score:%s", address)
	return c.Get(ctx, key, dest)
}

// SetNetworkStats caches network statistics
func (c *RedisClient) SetNetworkStats(ctx context.Context, networkID string, data interface{}) error {
	key := fmt.Sprintf("network_stats:%s", networkID)
	return c.Set(ctx, key, data, c.ttl.NetworkStats)
}

// GetNetworkStats retrieves cached network statistics
func (c *RedisClient) GetNetworkStats(ctx context.Context, networkID string, dest interface{}) error {
	key := fmt.Sprintf("network_stats:%s", networkID)
	return c.Get(ctx, key, dest)
}

// SetTransactionData caches transaction data
func (c *RedisClient) SetTransactionData(ctx context.Context, key string, data interface{}) error {
	cacheKey := fmt.Sprintf("transaction_data:%s", key)
	return c.Set(ctx, cacheKey, data, c.ttl.TransactionData)
}

// GetTransactionData retrieves cached transaction data
func (c *RedisClient) GetTransactionData(ctx context.Context, key string, dest interface{}) error {
	cacheKey := fmt.Sprintf("transaction_data:%s", key)
	return c.Get(ctx, cacheKey, dest)
}

// Rate limiting methods

// CheckRateLimit checks if a rate limit is exceeded
func (c *RedisClient) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	current, err := c.IncrementWithExpiry(ctx, fmt.Sprintf("rate_limit:%s", key), window)
	if err != nil {
		return false, err
	}

	return current <= limit, nil
}

// Session management

// SetSession stores session data
func (c *RedisClient) SetSession(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.Set(ctx, key, data, ttl)
}

// GetSession retrieves session data
func (c *RedisClient) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.Get(ctx, key, dest)
}

// DeleteSession removes session data
func (c *RedisClient) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return c.Delete(ctx, key)
}

// Health checks the health of the Redis connection
func (c *RedisClient) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// FlushDB flushes the current database (use with caution)
func (c *RedisClient) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// GetStats returns Redis statistics
func (c *RedisClient) GetStats(ctx context.Context) (map[string]string, error) {
	info, err := c.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	// Parse info string into map (simplified)
	stats := make(map[string]string)
	stats["info"] = info

	return stats, nil
}

// Custom errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
