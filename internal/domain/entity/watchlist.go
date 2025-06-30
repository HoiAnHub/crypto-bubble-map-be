package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// WatchedWallet represents a wallet in the user's watch list
type WatchedWallet struct {
	ID               uint               `json:"id" gorm:"primaryKey"`
	UserID           uint               `json:"user_id" gorm:"not null;index"`
	Address          string             `json:"address" gorm:"not null;index"`
	Label            *string            `json:"label,omitempty"`
	Tags             []WatchedWalletTag `json:"tags" gorm:"many2many:watched_wallet_tags;"`
	AddedAt          time.Time          `json:"added_at" gorm:"autoCreateTime"`
	LastActivity     *time.Time         `json:"last_activity,omitempty"`
	Balance          *string            `json:"balance,omitempty"`
	TransactionCount *int64             `json:"transaction_count,omitempty"`
	RiskScore        *float64           `json:"risk_score,omitempty"`
	AlertsEnabled    bool               `json:"alerts_enabled" gorm:"default:true"`
	CustomThresholds *CustomThresholds  `json:"custom_thresholds,omitempty" gorm:"embedded"`
	Notes            *string            `json:"notes,omitempty"`
	LastChecked      *time.Time         `json:"last_checked,omitempty"`
	AlertHistory     []WalletAlert      `json:"alert_history" gorm:"foreignKey:WalletID"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	DeletedAt        gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`
}

// WatchedWalletTag represents tags for watched wallets
type WatchedWalletTag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"`
	Color     *string        `json:"color,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// CustomThresholds represents custom alert thresholds for a watched wallet
type CustomThresholds struct {
	BalanceChange     float64 `json:"balance_change" gorm:"column:balance_change_threshold"`           // percentage
	TransactionVolume string  `json:"transaction_volume" gorm:"column:transaction_volume_threshold"`   // ETH/token amount
	RiskScoreIncrease float64 `json:"risk_score_increase" gorm:"column:risk_score_increase_threshold"` // points
}

// WalletAlertType represents different types of wallet alerts
type WalletAlertType string

const (
	WalletAlertTypeBalanceChange      WalletAlertType = "BALANCE_CHANGE"
	WalletAlertTypeHighVolume         WalletAlertType = "HIGH_VOLUME"
	WalletAlertTypeRiskIncrease       WalletAlertType = "RISK_INCREASE"
	WalletAlertTypeSuspiciousActivity WalletAlertType = "SUSPICIOUS_ACTIVITY"
	WalletAlertTypeNewTransaction     WalletAlertType = "NEW_TRANSACTION"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "LOW"
	AlertSeverityMedium   AlertSeverity = "MEDIUM"
	AlertSeverityHigh     AlertSeverity = "HIGH"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

// WalletAlert represents an alert for a watched wallet
type WalletAlert struct {
	ID             uint            `json:"id" gorm:"primaryKey"`
	WalletID       uint            `json:"wallet_id" gorm:"not null;index"`
	WatchedWallet  WatchedWallet   `json:"watched_wallet" gorm:"foreignKey:WalletID"`
	Type           WalletAlertType `json:"type" gorm:"not null"`
	Severity       AlertSeverity   `json:"severity" gorm:"not null"`
	Message        string          `json:"message" gorm:"not null"`
	Details        *string         `json:"details,omitempty"` // JSON string
	Timestamp      time.Time       `json:"timestamp" gorm:"autoCreateTime"`
	Acknowledged   bool            `json:"acknowledged" gorm:"default:false"`
	AcknowledgedAt *time.Time      `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uint           `json:"acknowledged_by,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
}

// WatchListStats represents statistics for the watch list
type WatchListStats struct {
	TotalWallets    int64  `json:"total_wallets"`
	ActiveAlerts    int64  `json:"active_alerts"`
	HighRiskWallets int64  `json:"high_risk_wallets"`
	TotalValue      string `json:"total_value"`     // in ETH
	RecentActivity  int64  `json:"recent_activity"` // wallets with activity in last 24h
}

// User represents a user in the system
type User struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	Email           string          `json:"email" gorm:"uniqueIndex;not null"`
	Username        *string         `json:"username,omitempty" gorm:"uniqueIndex"`
	PasswordHash    string          `json:"-" gorm:"not null"`
	FirstName       *string         `json:"first_name,omitempty"`
	LastName        *string         `json:"last_name,omitempty"`
	Role            UserRole        `json:"role" gorm:"default:'USER'"`
	IsActive        bool            `json:"is_active" gorm:"default:true"`
	EmailVerified   bool            `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time      `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time      `json:"last_login_at,omitempty"`
	WatchedWallets  []WatchedWallet `json:"watched_wallets" gorm:"foreignKey:UserID"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
}

// UserRole represents different user roles
type UserRole string

const (
	UserRoleUser      UserRole = "USER"
	UserRoleAdmin     UserRole = "ADMIN"
	UserRoleModerator UserRole = "MODERATOR"
	UserRoleAnalyst   UserRole = "ANALYST"
)

// WatchedWalletInput represents input for adding a wallet to watch list
type WatchedWalletInput struct {
	Address          string            `json:"address" validate:"required,eth_addr"`
	Label            *string           `json:"label,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	AlertsEnabled    bool              `json:"alerts_enabled"`
	CustomThresholds *CustomThresholds `json:"custom_thresholds,omitempty"`
	Notes            *string           `json:"notes,omitempty"`
}

// WatchedWalletUpdateInput represents input for updating a watched wallet
type WatchedWalletUpdateInput struct {
	Label            *string           `json:"label,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	AlertsEnabled    *bool             `json:"alerts_enabled,omitempty"`
	CustomThresholds *CustomThresholds `json:"custom_thresholds,omitempty"`
	Notes            *string           `json:"notes,omitempty"`
}

// TableName returns the table name for WatchedWallet
func (WatchedWallet) TableName() string {
	return "watched_wallets"
}

// TableName returns the table name for WatchedWalletTag
func (WatchedWalletTag) TableName() string {
	return "watched_wallet_tags"
}

// TableName returns the table name for WalletAlert
func (WalletAlert) TableName() string {
	return "wallet_alerts"
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}

// Helper methods for WatchedWallet
func (ww *WatchedWallet) IsHighRisk() bool {
	if ww.RiskScore == nil {
		return false
	}
	return *ww.RiskScore >= 70.0
}

func (ww *WatchedWallet) HasRecentActivity() bool {
	if ww.LastActivity == nil {
		return false
	}
	return time.Since(*ww.LastActivity) <= 24*time.Hour
}

func (ww *WatchedWallet) GetTagNames() []string {
	names := make([]string, len(ww.Tags))
	for i, tag := range ww.Tags {
		names[i] = tag.Name
	}
	return names
}

func (ww *WatchedWallet) GetUnacknowledgedAlerts() []WalletAlert {
	var unacknowledged []WalletAlert
	for _, alert := range ww.AlertHistory {
		if !alert.Acknowledged {
			unacknowledged = append(unacknowledged, alert)
		}
	}
	return unacknowledged
}

// Helper methods for WalletAlert
func (wa *WalletAlert) Acknowledge(userID uint) {
	wa.Acknowledged = true
	now := time.Now()
	wa.AcknowledgedAt = &now
	wa.AcknowledgedBy = &userID
}

func (wa *WalletAlert) IsRecent() bool {
	return time.Since(wa.Timestamp) <= 24*time.Hour
}

func (wa *WalletAlert) IsCritical() bool {
	return wa.Severity == AlertSeverityCritical
}

// Helper methods for User
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

func (u *User) IsModerator() bool {
	return u.Role == UserRoleModerator || u.Role == UserRoleAdmin
}

func (u *User) IsAnalyst() bool {
	return u.Role == UserRoleAnalyst || u.Role == UserRoleModerator || u.Role == UserRoleAdmin
}

func (u *User) CanManageWatchList() bool {
	return u.IsActive && u.EmailVerified
}

func (u *User) GetWatchListStats() WatchListStats {
	totalWallets := int64(len(u.WatchedWallets))
	var activeAlerts, highRiskWallets, recentActivity int64

	for _, wallet := range u.WatchedWallets {
		// Count unacknowledged alerts
		for _, alert := range wallet.AlertHistory {
			if !alert.Acknowledged {
				activeAlerts++
			}
		}

		// Count high risk wallets
		if wallet.IsHighRisk() {
			highRiskWallets++
		}

		// Count recent activity
		if wallet.HasRecentActivity() {
			recentActivity++
		}
	}

	return WatchListStats{
		TotalWallets:    totalWallets,
		ActiveAlerts:    activeAlerts,
		HighRiskWallets: highRiskWallets,
		TotalValue:      "0", // This would be calculated from actual balances
		RecentActivity:  recentActivity,
	}
}

// Validation methods
func (input *WatchedWalletInput) Validate() error {
	// This would contain validation logic
	// For now, just check if address is provided
	if input.Address == "" {
		return ErrInvalidAddress
	}
	return nil
}

// Custom errors
var (
	ErrInvalidAddress      = fmt.Errorf("invalid wallet address")
	ErrWalletNotFound      = fmt.Errorf("wallet not found in watch list")
	ErrWalletAlreadyExists = fmt.Errorf("wallet already exists in watch list")
	ErrUnauthorized        = fmt.Errorf("unauthorized access")
)
