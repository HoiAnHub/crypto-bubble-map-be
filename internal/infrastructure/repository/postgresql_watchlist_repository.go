package repository

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgreSQLWatchListRepository implements WatchListRepository using PostgreSQL
type PostgreSQLWatchListRepository struct {
	db     *database.PostgreSQLClient
	logger *zap.Logger
}

// NewPostgreSQLWatchListRepository creates a new PostgreSQL watch list repository
func NewPostgreSQLWatchListRepository(db *database.PostgreSQLClient, logger *zap.Logger) repository.WatchListRepository {
	return &PostgreSQLWatchListRepository{
		db:     db,
		logger: logger,
	}
}

// GetWatchedWallets retrieves all watched wallets for a user
func (r *PostgreSQLWatchListRepository) GetWatchedWallets(ctx context.Context, userID uint) ([]entity.WatchedWallet, error) {
	var wallets []entity.WatchedWallet

	err := r.db.GetDB().WithContext(ctx).
		Preload("Tags").
		Preload("AlertHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp DESC").Limit(10) // Limit recent alerts
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&wallets).Error

	if err != nil {
		r.logger.Error("Failed to get watched wallets",
			zap.Uint("userID", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get watched wallets: %w", err)
	}

	r.logger.Debug("Retrieved watched wallets",
		zap.Uint("userID", userID),
		zap.Int("count", len(wallets)))

	return wallets, nil
}

// GetWatchedWallet retrieves a specific watched wallet by ID
func (r *PostgreSQLWatchListRepository) GetWatchedWallet(ctx context.Context, userID uint, walletID uint) (*entity.WatchedWallet, error) {
	var wallet entity.WatchedWallet

	err := r.db.GetDB().WithContext(ctx).
		Preload("Tags").
		Preload("AlertHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp DESC")
		}).
		Where("id = ? AND user_id = ?", walletID, userID).
		First(&wallet).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get watched wallet",
			zap.Uint("userID", userID),
			zap.Uint("walletID", walletID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get watched wallet: %w", err)
	}

	return &wallet, nil
}

// GetWatchedWalletByAddress retrieves a watched wallet by address
func (r *PostgreSQLWatchListRepository) GetWatchedWalletByAddress(ctx context.Context, userID uint, address string) (*entity.WatchedWallet, error) {
	var wallet entity.WatchedWallet

	err := r.db.GetDB().WithContext(ctx).
		Preload("Tags").
		Preload("AlertHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp DESC")
		}).
		Where("user_id = ? AND address = ?", userID, address).
		First(&wallet).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get watched wallet by address",
			zap.Uint("userID", userID),
			zap.String("address", address),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get watched wallet by address: %w", err)
	}

	return &wallet, nil
}

// AddWatchedWallet adds a new wallet to the watch list
func (r *PostgreSQLWatchListRepository) AddWatchedWallet(ctx context.Context, wallet *entity.WatchedWallet) error {
	// Check if wallet already exists for this user
	existing, err := r.GetWatchedWalletByAddress(ctx, wallet.UserID, wallet.Address)
	if err != nil {
		return fmt.Errorf("failed to check existing wallet: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("wallet already exists in watch list")
	}

	// Start transaction
	tx := r.db.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Handle tags
	if len(wallet.Tags) > 0 {
		var tags []entity.WatchedWalletTag
		for _, tag := range wallet.Tags {
			existingTag, err := r.getOrCreateTagTx(tx, tag.Name)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to handle tag: %w", err)
			}
			tags = append(tags, *existingTag)
		}
		wallet.Tags = tags
	}

	// Create the watched wallet
	if err := tx.Create(wallet).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Failed to add watched wallet",
			zap.Uint("userID", wallet.UserID),
			zap.String("address", wallet.Address),
			zap.Error(err))
		return fmt.Errorf("failed to add watched wallet: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Added watched wallet",
		zap.Uint("userID", wallet.UserID),
		zap.String("address", wallet.Address),
		zap.Uint("walletID", wallet.ID))

	return nil
}

// UpdateWatchedWallet updates an existing watched wallet
func (r *PostgreSQLWatchListRepository) UpdateWatchedWallet(ctx context.Context, wallet *entity.WatchedWallet) error {
	// Start transaction
	tx := r.db.GetDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Handle tags if they are being updated
	if len(wallet.Tags) > 0 {
		// Clear existing associations
		if err := tx.Model(wallet).Association("Tags").Clear(); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear existing tags: %w", err)
		}

		// Add new tags
		var tags []entity.WatchedWalletTag
		for _, tag := range wallet.Tags {
			existingTag, err := r.getOrCreateTagTx(tx, tag.Name)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to handle tag: %w", err)
			}
			tags = append(tags, *existingTag)
		}
		wallet.Tags = tags
	}

	// Update the wallet
	if err := tx.Save(wallet).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Failed to update watched wallet",
			zap.Uint("walletID", wallet.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update watched wallet: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Updated watched wallet",
		zap.Uint("walletID", wallet.ID))

	return nil
}

// RemoveWatchedWallet removes a wallet from the watch list
func (r *PostgreSQLWatchListRepository) RemoveWatchedWallet(ctx context.Context, userID uint, walletID uint) error {
	result := r.db.GetDB().WithContext(ctx).
		Where("id = ? AND user_id = ?", walletID, userID).
		Delete(&entity.WatchedWallet{})

	if result.Error != nil {
		r.logger.Error("Failed to remove watched wallet",
			zap.Uint("userID", userID),
			zap.Uint("walletID", walletID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to remove watched wallet: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("watched wallet not found")
	}

	r.logger.Info("Removed watched wallet",
		zap.Uint("userID", userID),
		zap.Uint("walletID", walletID))

	return nil
}

// GetWatchListStats retrieves statistics for the watch list
func (r *PostgreSQLWatchListRepository) GetWatchListStats(ctx context.Context, userID uint) (*entity.WatchListStats, error) {
	var stats entity.WatchListStats

	// Get total wallets
	if err := r.db.GetDB().WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalWallets).Error; err != nil {
		return nil, fmt.Errorf("failed to count total wallets: %w", err)
	}

	// Get active alerts (unacknowledged)
	if err := r.db.GetDB().WithContext(ctx).
		Model(&entity.WalletAlert{}).
		Joins("JOIN watched_wallets ON wallet_alerts.wallet_id = watched_wallets.id").
		Where("watched_wallets.user_id = ? AND wallet_alerts.acknowledged = false", userID).
		Count(&stats.ActiveAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count active alerts: %w", err)
	}

	// Get high risk wallets (risk score >= 70)
	if err := r.db.GetDB().WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ? AND risk_score >= ?", userID, 70.0).
		Count(&stats.HighRiskWallets).Error; err != nil {
		return nil, fmt.Errorf("failed to count high risk wallets: %w", err)
	}

	// Get wallets with recent activity (last 24 hours)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	if err := r.db.GetDB().WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ? AND last_activity > ?", userID, twentyFourHoursAgo).
		Count(&stats.RecentActivity).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent activity: %w", err)
	}

	// Calculate total value (sum of all balances)
	var totalValue float64
	if err := r.db.GetDB().WithContext(ctx).
		Model(&entity.WatchedWallet{}).
		Where("user_id = ? AND balance IS NOT NULL", userID).
		Select("COALESCE(SUM(CAST(balance AS DECIMAL)), 0)").
		Scan(&totalValue).Error; err != nil {
		r.logger.Warn("Failed to calculate total value, setting to 0", zap.Error(err))
		totalValue = 0
	}
	stats.TotalValue = fmt.Sprintf("%.6f", totalValue)

	r.logger.Debug("Retrieved watch list stats",
		zap.Uint("userID", userID),
		zap.Int64("totalWallets", stats.TotalWallets),
		zap.Int64("activeAlerts", stats.ActiveAlerts))

	return &stats, nil
}

// CreateWalletAlert creates a new wallet alert
func (r *PostgreSQLWatchListRepository) CreateWalletAlert(ctx context.Context, alert *entity.WalletAlert) error {
	if err := r.db.GetDB().WithContext(ctx).Create(alert).Error; err != nil {
		r.logger.Error("Failed to create wallet alert",
			zap.Uint("walletID", alert.WalletID),
			zap.String("type", string(alert.Type)),
			zap.Error(err))
		return fmt.Errorf("failed to create wallet alert: %w", err)
	}

	r.logger.Info("Created wallet alert",
		zap.Uint("alertID", alert.ID),
		zap.Uint("walletID", alert.WalletID),
		zap.String("type", string(alert.Type)))

	return nil
}

// GetWalletAlerts retrieves wallet alerts with optional filters
func (r *PostgreSQLWatchListRepository) GetWalletAlerts(ctx context.Context, userID uint, filters map[string]interface{}) ([]entity.WalletAlert, error) {
	var alerts []entity.WalletAlert

	query := r.db.GetDB().WithContext(ctx).
		Joins("JOIN watched_wallets ON wallet_alerts.wallet_id = watched_wallets.id").
		Where("watched_wallets.user_id = ?", userID).
		Preload("WatchedWallet")

	// Apply filters
	if acknowledged, ok := filters["acknowledged"].(bool); ok {
		query = query.Where("wallet_alerts.acknowledged = ?", acknowledged)
	}

	if severity, ok := filters["severity"].(string); ok {
		query = query.Where("wallet_alerts.severity = ?", severity)
	}

	if alertType, ok := filters["type"].(string); ok {
		query = query.Where("wallet_alerts.type = ?", alertType)
	}

	if walletID, ok := filters["wallet_id"].(uint); ok {
		query = query.Where("wallet_alerts.wallet_id = ?", walletID)
	}

	// Apply time range filter if provided
	if since, ok := filters["since"].(time.Time); ok {
		query = query.Where("wallet_alerts.timestamp >= ?", since)
	}

	err := query.Order("wallet_alerts.timestamp DESC").
		Limit(100). // Limit to prevent large result sets
		Find(&alerts).Error

	if err != nil {
		r.logger.Error("Failed to get wallet alerts",
			zap.Uint("userID", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get wallet alerts: %w", err)
	}

	r.logger.Debug("Retrieved wallet alerts",
		zap.Uint("userID", userID),
		zap.Int("count", len(alerts)))

	return alerts, nil
}

// AcknowledgeWalletAlert acknowledges a wallet alert
func (r *PostgreSQLWatchListRepository) AcknowledgeWalletAlert(ctx context.Context, userID uint, alertID uint) error {
	now := time.Now()

	result := r.db.GetDB().WithContext(ctx).
		Model(&entity.WalletAlert{}).
		Joins("JOIN watched_wallets ON wallet_alerts.wallet_id = watched_wallets.id").
		Where("wallet_alerts.id = ? AND watched_wallets.user_id = ?", alertID, userID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_at": now,
			"acknowledged_by": userID,
		})

	if result.Error != nil {
		r.logger.Error("Failed to acknowledge wallet alert",
			zap.Uint("userID", userID),
			zap.Uint("alertID", alertID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to acknowledge wallet alert: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("wallet alert not found or not owned by user")
	}

	r.logger.Info("Acknowledged wallet alert",
		zap.Uint("userID", userID),
		zap.Uint("alertID", alertID))

	return nil
}

// GetOrCreateTag gets an existing tag or creates a new one
func (r *PostgreSQLWatchListRepository) GetOrCreateTag(ctx context.Context, name string) (*entity.WatchedWalletTag, error) {
	return r.getOrCreateTagTx(r.db.GetDB().WithContext(ctx), name)
}

// GetAllTags retrieves all available tags
func (r *PostgreSQLWatchListRepository) GetAllTags(ctx context.Context) ([]entity.WatchedWalletTag, error) {
	var tags []entity.WatchedWalletTag

	err := r.db.GetDB().WithContext(ctx).
		Order("name ASC").
		Find(&tags).Error

	if err != nil {
		r.logger.Error("Failed to get all tags", zap.Error(err))
		return nil, fmt.Errorf("failed to get all tags: %w", err)
	}

	r.logger.Debug("Retrieved all tags", zap.Int("count", len(tags)))

	return tags, nil
}

// Helper method to get or create a tag within a transaction
func (r *PostgreSQLWatchListRepository) getOrCreateTagTx(tx *gorm.DB, name string) (*entity.WatchedWalletTag, error) {
	var tag entity.WatchedWalletTag

	// Try to find existing tag
	err := tx.Where("name = ?", name).First(&tag).Error
	if err == nil {
		return &tag, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query tag: %w", err)
	}

	// Create new tag
	tag = entity.WatchedWalletTag{
		Name: name,
	}

	if err := tx.Create(&tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return &tag, nil
}
