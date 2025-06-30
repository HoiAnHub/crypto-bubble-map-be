package graph

import (
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/cache"
	"crypto-bubble-map-be/internal/infrastructure/logger"
)

// Resolver is the root GraphQL resolver
type Resolver struct {
	// Repositories
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	networkRepo     repository.NetworkRepository
	watchListRepo   repository.WatchListRepository
	securityRepo    repository.SecurityRepository
	userRepo        repository.UserRepository
	cacheRepo       repository.CacheRepository
	aiRepo          repository.AIRepository

	// Infrastructure
	cache  *cache.RedisClient
	logger *logger.Logger
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	networkRepo repository.NetworkRepository,
	watchListRepo repository.WatchListRepository,
	securityRepo repository.SecurityRepository,
	userRepo repository.UserRepository,
	cacheRepo repository.CacheRepository,
	aiRepo repository.AIRepository,
	cache *cache.RedisClient,
	logger *logger.Logger,
) *Resolver {
	return &Resolver{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		networkRepo:     networkRepo,
		watchListRepo:   watchListRepo,
		securityRepo:    securityRepo,
		userRepo:        userRepo,
		cacheRepo:       cacheRepo,
		aiRepo:          aiRepo,
		cache:           cache,
		logger:          logger,
	}
}
