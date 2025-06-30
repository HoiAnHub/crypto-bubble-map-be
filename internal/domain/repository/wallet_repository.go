package repository

import (
	"context"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
)

// WalletRepository defines the interface for wallet data access
type WalletRepository interface {
	// Wallet Network Operations
	GetWalletNetwork(ctx context.Context, input *entity.WalletNetworkInput) (*entity.WalletNetwork, error)
	GetWallet(ctx context.Context, address string) (*entity.Wallet, error)
	GetWalletsByAddresses(ctx context.Context, addresses []string) ([]entity.Wallet, error)

	// Wallet Rankings
	GetWalletRankings(ctx context.Context, category entity.RankingCategory, networkID *string, limit, offset int) (*entity.WalletRankingResult, error)

	// Search Operations
	SearchWallets(ctx context.Context, query string, limit int) ([]entity.WalletSearchResult, error)

	// Risk Operations
	GetRiskScore(ctx context.Context, address string) (*entity.RiskScore, error)
	GetRiskScores(ctx context.Context, addresses []string) ([]entity.RiskScore, error)
	UpdateRiskScore(ctx context.Context, address string, manualFlags []string, whitelistStatus *bool) (*entity.RiskScore, error)

	// Statistics
	GetWalletStats(ctx context.Context, address string) (*entity.WalletStats, error)
}

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Transaction Operations
	GetTransactionsByWallet(ctx context.Context, walletAddress string, limit, offset int64) ([]entity.Transaction, error)
	GetPairwiseTransactions(ctx context.Context, walletA, walletB string, limit, offset int64, filters *entity.TransactionFilters) (*entity.PairwiseTransactionResult, error)
	GetTransaction(ctx context.Context, hash string) (*entity.Transaction, error)

	// Money Flow Operations
	GetMoneyFlowData(ctx context.Context, walletAddress string, filters *entity.MoneyFlowFilters) (*entity.MoneyFlowData, error)

	// Search Operations
	SearchTransactions(ctx context.Context, query string, limit int64) ([]entity.Transaction, error)

	// Statistics
	GetTransactionStats(ctx context.Context, timeRange *entity.TimeRange) (map[string]interface{}, error)
	GetTopTokens(ctx context.Context, limit int64) ([]entity.TokenSummary, error)
	GetRecentActivity(ctx context.Context, limit int64) ([]entity.ActivitySummary, error)
}

// NetworkRepository defines the interface for network data access
type NetworkRepository interface {
	// Network Information
	GetNetworks(ctx context.Context) ([]entity.NetworkInfo, error)
	GetNetwork(ctx context.Context, networkID string) (*entity.NetworkInfo, error)
	GetNetworkStats(ctx context.Context, networkID string) (*entity.NetworkStats, error)
	GetNetworkRankings(ctx context.Context, limit int) ([]entity.NetworkRanking, error)

	// Dashboard Statistics
	GetDashboardStats(ctx context.Context, networkID *string) (*entity.DashboardStats, error)
}

// WatchListRepository defines the interface for watch list data access
type WatchListRepository interface {
	// Watch List Operations
	GetWatchedWallets(ctx context.Context, userID uint) ([]entity.WatchedWallet, error)
	GetWatchedWallet(ctx context.Context, userID uint, walletID uint) (*entity.WatchedWallet, error)
	GetWatchedWalletByAddress(ctx context.Context, userID uint, address string) (*entity.WatchedWallet, error)
	AddWatchedWallet(ctx context.Context, wallet *entity.WatchedWallet) error
	UpdateWatchedWallet(ctx context.Context, wallet *entity.WatchedWallet) error
	RemoveWatchedWallet(ctx context.Context, userID uint, walletID uint) error

	// Watch List Statistics
	GetWatchListStats(ctx context.Context, userID uint) (*entity.WatchListStats, error)

	// Alert Operations
	CreateWalletAlert(ctx context.Context, alert *entity.WalletAlert) error
	GetWalletAlerts(ctx context.Context, userID uint, filters map[string]interface{}) ([]entity.WalletAlert, error)
	AcknowledgeWalletAlert(ctx context.Context, userID uint, alertID uint) error

	// Tag Operations
	GetOrCreateTag(ctx context.Context, name string) (*entity.WatchedWalletTag, error)
	GetAllTags(ctx context.Context) ([]entity.WatchedWalletTag, error)
}

// SecurityRepository defines the interface for security data access
type SecurityRepository interface {
	// Security Alert Operations
	GetSecurityAlerts(ctx context.Context, filters *entity.SecurityAlertFilters, limit, offset int) (*entity.SecurityAlertResult, error)
	GetSecurityAlert(ctx context.Context, alertID string) (*entity.SecurityAlert, error)
	CreateSecurityAlert(ctx context.Context, alert *entity.SecurityAlert) error
	UpdateSecurityAlert(ctx context.Context, alert *entity.SecurityAlert) error
	AcknowledgeSecurityAlert(ctx context.Context, alertID string) error
	ResolveSecurityAlert(ctx context.Context, alertID, resolution, notes string) error

	// Compliance Operations
	GenerateComplianceReport(ctx context.Context, walletAddress string, reportType entity.ComplianceReportType, timeRange entity.TimeRange) (*entity.ComplianceReport, error)
	GetComplianceReport(ctx context.Context, reportID string) (*entity.ComplianceReport, error)
	GetComplianceReports(ctx context.Context, filters map[string]interface{}) ([]entity.ComplianceReport, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// User Operations
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error

	// Authentication
	ValidateUserCredentials(ctx context.Context, email, password string) (*entity.User, error)
	UpdateUserPassword(ctx context.Context, userID uint, hashedPassword string) error

	// Session Management
	CreateSession(ctx context.Context, userID uint, sessionID string, expiresAt time.Time) error
	GetSession(ctx context.Context, sessionID string) (*entity.User, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	// Generic Cache Operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Wallet-specific Cache Operations
	SetWalletNetwork(ctx context.Context, address string, depth int, data interface{}) error
	GetWalletNetwork(ctx context.Context, address string, depth int, dest interface{}) error
	SetWalletRankings(ctx context.Context, category string, data interface{}) error
	GetWalletRankings(ctx context.Context, category string, dest interface{}) error
	SetDashboardStats(ctx context.Context, networkID string, data interface{}) error
	GetDashboardStats(ctx context.Context, networkID string, dest interface{}) error
	SetRiskScore(ctx context.Context, address string, data interface{}) error
	GetRiskScore(ctx context.Context, address string, dest interface{}) error

	// Rate Limiting
	CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)

	// Session Management
	SetSession(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string, dest interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// AIRepository defines the interface for AI operations
type AIRepository interface {
	// AI Assistant Operations
	AskAI(ctx context.Context, question string, context *entity.AIContext, walletAddress *string) (*entity.AIResponse, error)

	// AI Model Management
	GetAvailableModels(ctx context.Context) ([]string, error)
	GetModelInfo(ctx context.Context, modelName string) (map[string]interface{}, error)
}
