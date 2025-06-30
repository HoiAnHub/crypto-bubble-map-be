package container

import (
	"context"
	"fmt"

	"crypto-bubble-map-be/graph"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/cache"
	"crypto-bubble-map-be/internal/infrastructure/config"
	"crypto-bubble-map-be/internal/infrastructure/database"
	"crypto-bubble-map-be/internal/infrastructure/logger"
	repoImpl "crypto-bubble-map-be/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Container holds all application dependencies
type Container struct {
	Config     *config.Config
	Logger     *logger.Logger
	Neo4j      *database.Neo4jClient
	MongoDB    *database.MongoClient
	PostgreSQL *database.PostgreSQLClient
	Redis      *cache.RedisClient
	Resolver   *graph.Resolver
}

// NewContainer creates a new dependency injection container
func NewContainer() *fx.App {
	return fx.New(
		// Configuration
		fx.Provide(config.Load),

		// Logger
		fx.Provide(NewLogger),

		// Databases
		fx.Provide(NewNeo4jClient),
		fx.Provide(NewMongoClient),
		fx.Provide(NewPostgreSQLClient),
		fx.Provide(NewRedisClient),

		// Repositories
		fx.Provide(NewWalletRepository),
		fx.Provide(NewTransactionRepository),
		fx.Provide(NewNetworkRepository),
		fx.Provide(NewWatchListRepository),
		fx.Provide(NewSecurityRepository),
		fx.Provide(NewUserRepository),
		fx.Provide(NewCacheRepository),
		fx.Provide(NewAIRepository),

		// GraphQL Resolver
		fx.Provide(NewGraphQLResolver),

		// Container
		fx.Provide(NewContainerStruct),
	)
}

// Provider functions

func NewLogger(cfg *config.Config) (*logger.Logger, error) {
	loggerCfg := &logger.Config{
		Level:       cfg.App.LogLevel,
		Environment: cfg.App.Environment,
		Debug:       cfg.App.Debug,
	}
	return logger.NewLogger(loggerCfg)
}

func NewNeo4jClient(cfg *config.Config, logger *logger.Logger) (*database.Neo4jClient, error) {
	return database.NewNeo4jClient(&cfg.Database.Neo4j, logger.Logger)
}

func NewMongoClient(cfg *config.Config, logger *logger.Logger) (*database.MongoClient, error) {
	return database.NewMongoClient(&cfg.Database.MongoDB, logger.Logger)
}

func NewPostgreSQLClient(cfg *config.Config, logger *logger.Logger) (*database.PostgreSQLClient, error) {
	return database.NewPostgreSQLClient(&cfg.Database.PostgreSQL, logger.Logger)
}

func NewRedisClient(cfg *config.Config, logger *logger.Logger) (*cache.RedisClient, error) {
	return cache.NewRedisClient(&cfg.Cache.Redis, &cfg.Cache.TTL, logger.Logger)
}

// Repository providers

func NewWalletRepository(neo4j *database.Neo4jClient, logger *logger.Logger) repository.WalletRepository {
	return repoImpl.NewNeo4jWalletRepository(neo4j, logger.Logger)
}

func NewTransactionRepository(mongo *database.MongoClient, logger *logger.Logger) repository.TransactionRepository {
	return repoImpl.NewMongoTransactionRepository(mongo, logger.Logger)
}

func NewNetworkRepository(neo4j *database.Neo4jClient, mongo *database.MongoClient, logger *logger.Logger) repository.NetworkRepository {
	return repoImpl.NewNetworkRepository(neo4j, mongo, logger.Logger)
}

func NewWatchListRepository(postgres *database.PostgreSQLClient, logger *logger.Logger) repository.WatchListRepository {
	// This would be implemented similar to other repositories
	// For now, return nil to avoid compilation errors
	return nil
}

func NewSecurityRepository(logger *logger.Logger) repository.SecurityRepository {
	// This would be implemented similar to other repositories
	// For now, return nil to avoid compilation errors
	return nil
}

func NewUserRepository(postgres *database.PostgreSQLClient, logger *logger.Logger) repository.UserRepository {
	// This would be implemented similar to other repositories
	// For now, return nil to avoid compilation errors
	return nil
}

func NewCacheRepository(redis *cache.RedisClient, logger *logger.Logger) repository.CacheRepository {
	return repoImpl.NewRedisCacheRepository(redis, logger.Logger)
}

func NewAIRepository(logger *logger.Logger) repository.AIRepository {
	return repoImpl.NewMockAIRepository(logger.Logger)
}

// GraphQL resolver provider

func NewGraphQLResolver(
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	networkRepo repository.NetworkRepository,
	watchListRepo repository.WatchListRepository,
	securityRepo repository.SecurityRepository,
	userRepo repository.UserRepository,
	cacheRepo repository.CacheRepository,
	aiRepo repository.AIRepository,
	redis *cache.RedisClient,
	logger *logger.Logger,
) *graph.Resolver {
	return graph.NewResolver(
		walletRepo,
		transactionRepo,
		networkRepo,
		watchListRepo,
		securityRepo,
		userRepo,
		cacheRepo,
		aiRepo,
		redis,
		logger,
	)
}

// Container struct provider

func NewContainerStruct(
	cfg *config.Config,
	logger *logger.Logger,
	neo4j *database.Neo4jClient,
	mongo *database.MongoClient,
	postgres *database.PostgreSQLClient,
	redis *cache.RedisClient,
	resolver *graph.Resolver,
) *Container {
	return &Container{
		Config:     cfg,
		Logger:     logger,
		Neo4j:      neo4j,
		MongoDB:    mongo,
		PostgreSQL: postgres,
		Redis:      redis,
		Resolver:   resolver,
	}
}

// Lifecycle hooks

func RegisterHooks(
	lc fx.Lifecycle,
	container *Container,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			container.Logger.Info("Starting application dependencies")

			// Run database migrations
			if err := container.PostgreSQL.AutoMigrate(); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			// Create MongoDB indexes
			if err := container.MongoDB.CreateIndexes(ctx); err != nil {
				container.Logger.Warn("Failed to create MongoDB indexes", zap.Error(err))
			}

			container.Logger.Info("Application dependencies started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			container.Logger.Info("Stopping application dependencies")

			// Close database connections
			if err := container.Neo4j.Close(ctx); err != nil {
				container.Logger.Error("Failed to close Neo4j connection", zap.Error(err))
			}

			if err := container.MongoDB.Close(ctx); err != nil {
				container.Logger.Error("Failed to close MongoDB connection", zap.Error(err))
			}

			if err := container.PostgreSQL.Close(); err != nil {
				container.Logger.Error("Failed to close PostgreSQL connection", zap.Error(err))
			}

			if err := container.Redis.Close(); err != nil {
				container.Logger.Error("Failed to close Redis connection", zap.Error(err))
			}

			// Close logger
			if err := container.Logger.Close(); err != nil {
				// Can't log this error since logger is closing
			}

			return nil
		},
	})
}

// Helper function to get container from fx.App
func GetContainer(app *fx.App) (*Container, error) {
	var container *Container
	if err := app.Err(); err != nil {
		return nil, err
	}

	// This is a simplified way to get the container
	// In a real application, you might want to use fx.Populate or similar
	return container, nil
}
