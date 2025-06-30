package repository

import (
	"context"
	"time"

	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/cache"

	"go.uber.org/zap"
)

// RedisCacheRepository implements CacheRepository using Redis
type RedisCacheRepository struct {
	redis  *cache.RedisClient
	logger *zap.Logger
}

// NewRedisCacheRepository creates a new Redis cache repository
func NewRedisCacheRepository(redis *cache.RedisClient, logger *zap.Logger) repository.CacheRepository {
	return &RedisCacheRepository{
		redis:  redis,
		logger: logger,
	}
}

// Set stores a value in cache with TTL
func (r *RedisCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.redis.Set(ctx, key, value, ttl)
}

// Get retrieves a value from cache
func (r *RedisCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	return r.redis.Get(ctx, key, dest)
}

// Delete removes keys from cache
func (r *RedisCacheRepository) Delete(ctx context.Context, keys ...string) error {
	return r.redis.Delete(ctx, keys...)
}

// Exists checks if a key exists in cache
func (r *RedisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	return r.redis.Exists(ctx, key)
}

// SetWalletNetwork caches wallet network data
func (r *RedisCacheRepository) SetWalletNetwork(ctx context.Context, address string, depth int, data interface{}) error {
	return r.redis.SetWalletNetwork(ctx, address, depth, data)
}

// GetWalletNetwork retrieves cached wallet network data
func (r *RedisCacheRepository) GetWalletNetwork(ctx context.Context, address string, depth int, dest interface{}) error {
	return r.redis.GetWalletNetwork(ctx, address, depth, dest)
}

// SetWalletRankings caches wallet rankings
func (r *RedisCacheRepository) SetWalletRankings(ctx context.Context, category string, data interface{}) error {
	return r.redis.SetWalletRankings(ctx, category, data)
}

// GetWalletRankings retrieves cached wallet rankings
func (r *RedisCacheRepository) GetWalletRankings(ctx context.Context, category string, dest interface{}) error {
	return r.redis.GetWalletRankings(ctx, category, dest)
}

// SetDashboardStats caches dashboard statistics
func (r *RedisCacheRepository) SetDashboardStats(ctx context.Context, networkID string, data interface{}) error {
	return r.redis.SetDashboardStats(ctx, networkID, data)
}

// GetDashboardStats retrieves cached dashboard statistics
func (r *RedisCacheRepository) GetDashboardStats(ctx context.Context, networkID string, dest interface{}) error {
	return r.redis.GetDashboardStats(ctx, networkID, dest)
}

// SetRiskScore caches risk score data
func (r *RedisCacheRepository) SetRiskScore(ctx context.Context, address string, data interface{}) error {
	return r.redis.SetRiskScore(ctx, address, data)
}

// GetRiskScore retrieves cached risk score data
func (r *RedisCacheRepository) GetRiskScore(ctx context.Context, address string, dest interface{}) error {
	return r.redis.GetRiskScore(ctx, address, dest)
}

// CheckRateLimit checks if a rate limit is exceeded
func (r *RedisCacheRepository) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	return r.redis.CheckRateLimit(ctx, key, limit, window)
}

// SetSession stores session data
func (r *RedisCacheRepository) SetSession(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error {
	return r.redis.SetSession(ctx, sessionID, data, ttl)
}

// GetSession retrieves session data
func (r *RedisCacheRepository) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	return r.redis.GetSession(ctx, sessionID, dest)
}

// DeleteSession removes session data
func (r *RedisCacheRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.redis.DeleteSession(ctx, sessionID)
}
