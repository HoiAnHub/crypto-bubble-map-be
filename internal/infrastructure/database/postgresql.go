package database

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/infrastructure/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// PostgreSQLClient wraps the GORM database connection
type PostgreSQLClient struct {
	db     *gorm.DB
	logger *zap.Logger
	config *config.PostgreSQLConfig
}

// NewPostgreSQLClient creates a new PostgreSQL client
func NewPostgreSQLClient(cfg *config.PostgreSQLConfig, logger *zap.Logger) (*PostgreSQLClient, error) {
	// Configure GORM logger
	var gormLogLevel gormLogger.LogLevel
	switch logger.Level() {
	case zap.DebugLevel:
		gormLogLevel = gormLogger.Info
	case zap.InfoLevel:
		gormLogLevel = gormLogger.Warn
	default:
		gormLogLevel = gormLogger.Error
	}

	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	client := &PostgreSQLClient{
		db:     db,
		logger: logger,
		config: cfg,
	}

	logger.Info("PostgreSQL client initialized successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return client, nil
}

// Close closes the PostgreSQL connection
func (c *PostgreSQLClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the GORM database instance
func (c *PostgreSQLClient) GetDB() *gorm.DB {
	return c.db
}

// AutoMigrate runs database migrations
func (c *PostgreSQLClient) AutoMigrate() error {
	// Import the UserSession type from repository package
	type UserSession struct {
		ID        uint      `gorm:"primaryKey"`
		UserID    uint      `gorm:"not null;index"`
		SessionID string    `gorm:"uniqueIndex;not null"`
		ExpiresAt time.Time `gorm:"not null;index"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	err := c.db.AutoMigrate(
		&entity.User{},
		&entity.WatchedWallet{},
		&entity.WatchedWalletTag{},
		&entity.WalletAlert{},
		&UserSession{},
	)
	if err != nil {
		c.logger.Error("Failed to run auto migration", zap.Error(err))
		return err
	}

	c.logger.Info("Database migration completed successfully")
	return nil
}

// User operations
func (c *PostgreSQLClient) CreateUser(ctx context.Context, user *entity.User) error {
	if err := c.db.WithContext(ctx).Create(user).Error; err != nil {
		c.logger.Error("Failed to create user", zap.Error(err))
		return err
	}
	return nil
}

func (c *PostgreSQLClient) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := c.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		c.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (c *PostgreSQLClient) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	if err := c.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		c.logger.Error("Failed to get user by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (c *PostgreSQLClient) UpdateUser(ctx context.Context, user *entity.User) error {
	if err := c.db.WithContext(ctx).Save(user).Error; err != nil {
		c.logger.Error("Failed to update user", zap.Uint("id", user.ID), zap.Error(err))
		return err
	}
	return nil
}

// Watch list operations
func (c *PostgreSQLClient) GetWatchedWallets(ctx context.Context, userID uint) ([]entity.WatchedWallet, error) {
	var wallets []entity.WatchedWallet
	if err := c.db.WithContext(ctx).
		Preload("Tags").
		Preload("AlertHistory").
		Where("user_id = ?", userID).
		Find(&wallets).Error; err != nil {
		c.logger.Error("Failed to get watched wallets", zap.Uint("userID", userID), zap.Error(err))
		return nil, err
	}
	return wallets, nil
}

func (c *PostgreSQLClient) AddWatchedWallet(ctx context.Context, wallet *entity.WatchedWallet) error {
	if err := c.db.WithContext(ctx).Create(wallet).Error; err != nil {
		c.logger.Error("Failed to add watched wallet",
			zap.Uint("userID", wallet.UserID),
			zap.String("address", wallet.Address),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *PostgreSQLClient) UpdateWatchedWallet(ctx context.Context, wallet *entity.WatchedWallet) error {
	if err := c.db.WithContext(ctx).Save(wallet).Error; err != nil {
		c.logger.Error("Failed to update watched wallet",
			zap.Uint("id", wallet.ID),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *PostgreSQLClient) RemoveWatchedWallet(ctx context.Context, userID uint, walletID uint) error {
	if err := c.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", walletID, userID).
		Delete(&entity.WatchedWallet{}).Error; err != nil {
		c.logger.Error("Failed to remove watched wallet",
			zap.Uint("userID", userID),
			zap.Uint("walletID", walletID),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *PostgreSQLClient) GetWatchedWalletByAddress(ctx context.Context, userID uint, address string) (*entity.WatchedWallet, error) {
	var wallet entity.WatchedWallet
	if err := c.db.WithContext(ctx).
		Preload("Tags").
		Preload("AlertHistory").
		Where("user_id = ? AND address = ?", userID, address).
		First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		c.logger.Error("Failed to get watched wallet by address",
			zap.Uint("userID", userID),
			zap.String("address", address),
			zap.Error(err),
		)
		return nil, err
	}
	return &wallet, nil
}

// Wallet alert operations
func (c *PostgreSQLClient) CreateWalletAlert(ctx context.Context, alert *entity.WalletAlert) error {
	if err := c.db.WithContext(ctx).Create(alert).Error; err != nil {
		c.logger.Error("Failed to create wallet alert",
			zap.Uint("walletID", alert.WalletID),
			zap.String("type", string(alert.Type)),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *PostgreSQLClient) GetWalletAlerts(ctx context.Context, userID uint, filters map[string]interface{}) ([]entity.WalletAlert, error) {
	query := c.db.WithContext(ctx).
		Joins("JOIN watched_wallets ON wallet_alerts.wallet_id = watched_wallets.id").
		Where("watched_wallets.user_id = ?", userID)

	// Apply filters
	if walletID, ok := filters["wallet_id"]; ok {
		query = query.Where("wallet_alerts.wallet_id = ?", walletID)
	}
	if acknowledged, ok := filters["acknowledged"]; ok {
		query = query.Where("wallet_alerts.acknowledged = ?", acknowledged)
	}
	if severity, ok := filters["severity"]; ok {
		query = query.Where("wallet_alerts.severity = ?", severity)
	}
	if limit, ok := filters["limit"]; ok {
		query = query.Limit(limit.(int))
	}

	var alerts []entity.WalletAlert
	if err := query.Order("wallet_alerts.timestamp DESC").Find(&alerts).Error; err != nil {
		c.logger.Error("Failed to get wallet alerts", zap.Uint("userID", userID), zap.Error(err))
		return nil, err
	}
	return alerts, nil
}

func (c *PostgreSQLClient) AcknowledgeWalletAlert(ctx context.Context, userID uint, alertID uint) error {
	result := c.db.WithContext(ctx).
		Model(&entity.WalletAlert{}).
		Joins("JOIN watched_wallets ON wallet_alerts.wallet_id = watched_wallets.id").
		Where("wallet_alerts.id = ? AND watched_wallets.user_id = ?", alertID, userID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_at": time.Now().UTC(),
			"acknowledged_by": userID,
		})

	if result.Error != nil {
		c.logger.Error("Failed to acknowledge wallet alert",
			zap.Uint("userID", userID),
			zap.Uint("alertID", alertID),
			zap.Error(result.Error),
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("alert not found or not owned by user")
	}

	return nil
}

// Tag operations
func (c *PostgreSQLClient) GetOrCreateTag(ctx context.Context, name string) (*entity.WatchedWalletTag, error) {
	var tag entity.WatchedWalletTag
	if err := c.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new tag
			tag = entity.WatchedWalletTag{Name: name}
			if err := c.db.WithContext(ctx).Create(&tag).Error; err != nil {
				c.logger.Error("Failed to create tag", zap.String("name", name), zap.Error(err))
				return nil, err
			}
		} else {
			c.logger.Error("Failed to get tag", zap.String("name", name), zap.Error(err))
			return nil, err
		}
	}
	return &tag, nil
}

func (c *PostgreSQLClient) GetAllTags(ctx context.Context) ([]entity.WatchedWalletTag, error) {
	var tags []entity.WatchedWalletTag
	if err := c.db.WithContext(ctx).Find(&tags).Error; err != nil {
		c.logger.Error("Failed to get all tags", zap.Error(err))
		return nil, err
	}
	return tags, nil
}

// Statistics operations
func (c *PostgreSQLClient) GetWatchListStats(ctx context.Context, userID uint) (*entity.WatchListStats, error) {
	var stats entity.WatchListStats

	// Get total wallets
	if err := c.db.WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalWallets).Error; err != nil {
		return nil, err
	}

	// Get active alerts
	if err := c.db.WithContext(ctx).
		Model(&entity.WalletAlert{}).
		Joins("JOIN watched_wallets ON wallet_alerts.wallet_id = watched_wallets.id").
		Where("watched_wallets.user_id = ? AND wallet_alerts.acknowledged = false", userID).
		Count(&stats.ActiveAlerts).Error; err != nil {
		return nil, err
	}

	// Get high risk wallets (risk score >= 70)
	if err := c.db.WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ? AND risk_score >= 70", userID).
		Count(&stats.HighRiskWallets).Error; err != nil {
		return nil, err
	}

	// Get recent activity (last 24 hours)
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	if err := c.db.WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ? AND last_activity > ?", userID, yesterday).
		Count(&stats.RecentActivity).Error; err != nil {
		return nil, err
	}

	// Total value would need to be calculated from actual wallet balances
	stats.TotalValue = "0" // Placeholder

	return &stats, nil
}

// Health checks the health of the PostgreSQL connection
func (c *PostgreSQLClient) Health(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Transaction executes a function within a database transaction
func (c *PostgreSQLClient) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return c.db.WithContext(ctx).Transaction(fn)
}
