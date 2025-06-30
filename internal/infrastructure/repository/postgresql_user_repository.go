package repository

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/database"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// PostgreSQLUserRepository implements UserRepository using PostgreSQL
type PostgreSQLUserRepository struct {
	db     *database.PostgreSQLClient
	logger *zap.Logger
}

// NewPostgreSQLUserRepository creates a new PostgreSQL user repository
func NewPostgreSQLUserRepository(db *database.PostgreSQLClient, logger *zap.Logger) repository.UserRepository {
	return &PostgreSQLUserRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user
func (r *PostgreSQLUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	// Check if user already exists
	existing, err := r.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Create the user
	if err := r.db.GetDB().WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("Failed to create user",
			zap.String("email", user.Email),
			zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("Created user",
		zap.Uint("userID", user.ID),
		zap.String("email", user.Email))

	return nil
}

// GetUserByID retrieves a user by ID
func (r *PostgreSQLUserRepository) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User

	err := r.db.GetDB().WithContext(ctx).
		Preload("WatchedWallets").
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get user by ID",
			zap.Uint("userID", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgreSQLUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	err := r.db.GetDB().WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get user by email",
			zap.String("email", email),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *PostgreSQLUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	if err := r.db.GetDB().WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error("Failed to update user",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.Info("Updated user",
		zap.Uint("userID", user.ID))

	return nil
}

// DeleteUser soft deletes a user
func (r *PostgreSQLUserRepository) DeleteUser(ctx context.Context, id uint) error {
	result := r.db.GetDB().WithContext(ctx).
		Delete(&entity.User{}, id)

	if result.Error != nil {
		r.logger.Error("Failed to delete user",
			zap.Uint("userID", id),
			zap.Error(result.Error))
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("Deleted user",
		zap.Uint("userID", id))

	return nil
}

// ValidateUserCredentials validates user credentials and returns the user if valid
func (r *PostgreSQLUserRepository) ValidateUserCredentials(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		r.logger.Warn("Invalid password attempt",
			zap.String("email", email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := r.UpdateUser(ctx, user); err != nil {
		r.logger.Warn("Failed to update last login time",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		// Don't fail authentication for this
	}

	r.logger.Info("User authenticated successfully",
		zap.Uint("userID", user.ID),
		zap.String("email", email))

	return user, nil
}

// UpdateUserPassword updates a user's password
func (r *PostgreSQLUserRepository) UpdateUserPassword(ctx context.Context, userID uint, hashedPassword string) error {
	result := r.db.GetDB().WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword)

	if result.Error != nil {
		r.logger.Error("Failed to update user password",
			zap.Uint("userID", userID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to update user password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("Updated user password",
		zap.Uint("userID", userID))

	return nil
}

// UserSession represents a user session stored in database
type UserSession struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	SessionID string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      entity.User `gorm:"foreignKey:UserID"`
}

// TableName returns the table name for UserSession
func (UserSession) TableName() string {
	return "user_sessions"
}

// CreateSession creates a new user session
func (r *PostgreSQLUserRepository) CreateSession(ctx context.Context, userID uint, sessionID string, expiresAt time.Time) error {
	session := UserSession{
		UserID:    userID,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}

	if err := r.db.GetDB().WithContext(ctx).Create(&session).Error; err != nil {
		r.logger.Error("Failed to create session",
			zap.Uint("userID", userID),
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.Debug("Created session",
		zap.Uint("userID", userID),
		zap.String("sessionID", sessionID))

	return nil
}

// GetSession retrieves a user by session ID
func (r *PostgreSQLUserRepository) GetSession(ctx context.Context, sessionID string) (*entity.User, error) {
	var session UserSession

	err := r.db.GetDB().WithContext(ctx).
		Preload("User").
		Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get session",
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if user is still active
	if !session.User.IsActive {
		// Delete the session for inactive user
		r.DeleteSession(ctx, sessionID)
		return nil, nil
	}

	return &session.User, nil
}

// DeleteSession deletes a user session
func (r *PostgreSQLUserRepository) DeleteSession(ctx context.Context, sessionID string) error {
	result := r.db.GetDB().WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&UserSession{})

	if result.Error != nil {
		r.logger.Error("Failed to delete session",
			zap.String("sessionID", sessionID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	r.logger.Debug("Deleted session",
		zap.String("sessionID", sessionID))

	return nil
}

// CleanupExpiredSessions removes expired sessions (utility method)
func (r *PostgreSQLUserRepository) CleanupExpiredSessions(ctx context.Context) error {
	result := r.db.GetDB().WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&UserSession{})

	if result.Error != nil {
		r.logger.Error("Failed to cleanup expired sessions", zap.Error(result.Error))
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		r.logger.Info("Cleaned up expired sessions",
			zap.Int64("count", result.RowsAffected))
	}

	return nil
}
