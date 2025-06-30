package entity

import (
	"time"
)

// WalletType represents different types of blockchain entities
type WalletType string

const (
	WalletTypeRegular      WalletType = "REGULAR"
	WalletTypeExchange     WalletType = "EXCHANGE"
	WalletTypeContract     WalletType = "CONTRACT"
	WalletTypeWhale        WalletType = "WHALE"
	WalletTypeMiner        WalletType = "MINER"
	WalletTypeDefi         WalletType = "DEFI"
	WalletTypeBridge       WalletType = "BRIDGE"
	WalletTypeMEVBot       WalletType = "MEV_BOT"
	WalletTypeArbitrageBot WalletType = "ARBITRAGE_BOT"
	WalletTypeMarketMaker  WalletType = "MARKET_MAKER"
	WalletTypeSuspicious   WalletType = "SUSPICIOUS"
	WalletTypeBlacklisted  WalletType = "BLACKLISTED"
)

// RiskLevel represents the risk level of a wallet
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
	RiskLevelUnknown  RiskLevel = "UNKNOWN"
)

// Wallet represents an Ethereum wallet/address
type Wallet struct {
	ID                  string          `json:"id" neo4j:"id"`
	Address             string          `json:"address" neo4j:"address"`
	Label               *string         `json:"label,omitempty" neo4j:"label"`
	Balance             *string         `json:"balance,omitempty" neo4j:"balance"`
	TransactionCount    int64           `json:"transaction_count" neo4j:"total_transactions"`
	WalletType          WalletType      `json:"wallet_type" neo4j:"node_type"`
	Tags                []string        `json:"tags" neo4j:"tags"`
	FirstSeen           time.Time       `json:"first_seen" neo4j:"first_seen"`
	LastSeen            time.Time       `json:"last_seen" neo4j:"last_seen"`
	Network             string          `json:"network" neo4j:"network"`
	IsContract          bool            `json:"is_contract" neo4j:"is_contract"`
	AssociatedExchanges []string        `json:"associated_exchanges" neo4j:"associated_exchanges"`
	AssociatedProtocols []string        `json:"associated_protocols" neo4j:"associated_protocols"`
	SocialProfiles      *SocialProfiles `json:"social_profiles,omitempty"`
	Coordinates         *Coordinates    `json:"coordinates,omitempty"`
	RiskScore           *RiskScore      `json:"risk_score,omitempty"`

	// Enhanced classification fields
	RiskLevel       RiskLevel `json:"risk_level" neo4j:"risk_level"`
	ConfidenceScore float64   `json:"confidence_score" neo4j:"confidence_score"`
	LastClassified  time.Time `json:"last_classified" neo4j:"last_classified"`

	// Computed fields
	TotalSent     *string `json:"total_sent,omitempty" neo4j:"total_sent"`
	TotalReceived *string `json:"total_received,omitempty" neo4j:"total_received"`
}

// WalletConnection represents a connection between two wallets
type WalletConnection struct {
	Source           string    `json:"source" neo4j:"from_address"`
	Target           string    `json:"target" neo4j:"to_address"`
	Value            string    `json:"value" neo4j:"total_value"`
	TransactionCount int64     `json:"transaction_count" neo4j:"tx_count"`
	FirstTransaction time.Time `json:"first_transaction" neo4j:"first_tx"`
	LastTransaction  time.Time `json:"last_transaction" neo4j:"last_tx"`
	RiskLevel        RiskLevel `json:"risk_level"`
}

// WalletNetwork represents a network of connected wallets
type WalletNetwork struct {
	Nodes    []Wallet           `json:"nodes"`
	Links    []WalletConnection `json:"links"`
	Metadata NetworkMetadata    `json:"metadata"`
}

// NetworkMetadata contains metadata about the wallet network
type NetworkMetadata struct {
	TotalNodes   int       `json:"total_nodes"`
	TotalLinks   int       `json:"total_links"`
	MaxDepth     int       `json:"max_depth"`
	CenterWallet string    `json:"center_wallet"`
	GeneratedAt  time.Time `json:"generated_at"`
}

// SocialProfiles represents social media profiles associated with a wallet
type SocialProfiles struct {
	Twitter  *string `json:"twitter,omitempty"`
	Discord  *string `json:"discord,omitempty"`
	Telegram *string `json:"telegram,omitempty"`
	Github   *string `json:"github,omitempty"`
	Website  *string `json:"website,omitempty"`
	LinkedIn *string `json:"linkedin,omitempty"`
}

// Coordinates represents the position of a wallet in the visualization
type Coordinates struct {
	X *float64 `json:"x,omitempty"`
	Y *float64 `json:"y,omitempty"`
}

// RiskScore represents the risk assessment of a wallet
type RiskScore struct {
	Address     string      `json:"address"`
	TotalScore  int         `json:"total_score"`
	RiskLevel   RiskLevel   `json:"risk_level"`
	Factors     RiskFactors `json:"factors"`
	Flags       []string    `json:"flags"`
	LastUpdated time.Time   `json:"last_updated"`
}

// RiskFactors represents different risk factor scores
type RiskFactors struct {
	Phishing   int `json:"phishing"`
	MEV        int `json:"mev"`
	Laundering int `json:"laundering"`
	Sanctions  int `json:"sanctions"`
	Scam       int `json:"scam"`
	Suspicious int `json:"suspicious"`
}

// WalletStats represents statistics for a wallet
type WalletStats struct {
	Address             string `json:"address"`
	IncomingConnections int64  `json:"incoming_connections"`
	OutgoingConnections int64  `json:"outgoing_connections"`
	TotalVolume         string `json:"total_volume"`
	TransactionCount    int64  `json:"transaction_count"`
}

// WalletMetrics represents comprehensive metrics for wallet ranking
type WalletMetrics struct {
	Address                string     `json:"address"`
	Label                  *string    `json:"label,omitempty"`
	QualityScore           float64    `json:"quality_score"`
	RiskScore              float64    `json:"risk_score"`
	ReputationScore        float64    `json:"reputation_score"`
	TransactionCount       int64      `json:"transaction_count"`
	TransactionVolume      string     `json:"transaction_volume"`
	AverageTransactionSize string     `json:"average_transaction_size"`
	ActivityFrequency      float64    `json:"activity_frequency"`
	WalletAge              int        `json:"wallet_age"`
	FirstTransactionDate   time.Time  `json:"first_transaction_date"`
	LastTransactionDate    time.Time  `json:"last_transaction_date"`
	ConnectionCount        int64      `json:"connection_count"`
	UniqueCounterparties   int64      `json:"unique_counterparties"`
	NetworkInfluence       float64    `json:"network_influence"`
	RiskFlags              []string   `json:"risk_flags"`
	IsWhitelisted          bool       `json:"is_whitelisted"`
	IsFlagged              bool       `json:"is_flagged"`
	WalletType             WalletType `json:"wallet_type"`
	ProfitabilityScore     *float64   `json:"profitability_score,omitempty"`
	LiquidityScore         *float64   `json:"liquidity_score,omitempty"`
	HasVerifiedSocials     bool       `json:"has_verified_socials"`
	SocialScore            float64    `json:"social_score"`
}

// WalletRanking represents a wallet's ranking in a specific category
type WalletRanking struct {
	Rank   int           `json:"rank"`
	Wallet WalletMetrics `json:"wallet"`
	Score  float64       `json:"score"`
	Change *int          `json:"change,omitempty"`
}

// RankingCategory represents different ranking categories
type RankingCategory string

const (
	RankingCategoryQuality    RankingCategory = "QUALITY"
	RankingCategoryReputation RankingCategory = "REPUTATION"
	RankingCategoryVolume     RankingCategory = "VOLUME"
	RankingCategoryActivity   RankingCategory = "ACTIVITY"
	RankingCategoryAge        RankingCategory = "AGE"
	RankingCategoryNetwork    RankingCategory = "NETWORK"
	RankingCategorySafety     RankingCategory = "SAFETY"
)

// WalletSearchResult represents a search result for wallets
type WalletSearchResult struct {
	Address          string   `json:"address"`
	Label            *string  `json:"label,omitempty"`
	Tags             []string `json:"tags"`
	RiskScore        *float64 `json:"risk_score,omitempty"`
	TransactionCount int64    `json:"transaction_count"`
	Balance          *string  `json:"balance,omitempty"`
	RelevanceScore   float64  `json:"relevance_score"`
}

// WalletUpdate represents real-time updates for a wallet
type WalletUpdate struct {
	Address          string     `json:"address"`
	Balance          *string    `json:"balance,omitempty"`
	TransactionCount *int64     `json:"transaction_count,omitempty"`
	RiskScore        *RiskScore `json:"risk_score,omitempty"`
	LastActivity     *time.Time `json:"last_activity,omitempty"`
}

// WalletNetworkInput represents input for wallet network queries
type WalletNetworkInput struct {
	Address                   string `json:"address"`
	Depth                     int    `json:"depth"`
	NetworkID                 string `json:"network_id"`
	IncludeRiskAnalysis       bool   `json:"include_risk_analysis"`
	IncludeTransactionVolumes bool   `json:"include_transaction_volumes"`
}

// WalletRankingResult represents paginated wallet ranking results
type WalletRankingResult struct {
	Rankings []WalletRanking `json:"rankings"`
	HasMore  bool            `json:"has_more"`
	Total    int64           `json:"total"`
}

// Helper methods for WalletType
func (wt WalletType) IsHighRisk() bool {
	return wt == WalletTypeSuspicious || wt == WalletTypeBlacklisted
}

func (wt WalletType) IsExchangeRelated() bool {
	return wt == WalletTypeExchange
}

func (wt WalletType) IsContractType() bool {
	return wt == WalletTypeContract || wt == WalletTypeDefi || wt == WalletTypeBridge
}

// Helper methods for RiskLevel
func (rl RiskLevel) ToScore() int {
	switch rl {
	case RiskLevelLow:
		return 25
	case RiskLevelMedium:
		return 50
	case RiskLevelHigh:
		return 75
	case RiskLevelCritical:
		return 100
	default:
		return 0
	}
}

func ScoreToRiskLevel(score int) RiskLevel {
	switch {
	case score >= 80:
		return RiskLevelCritical
	case score >= 60:
		return RiskLevelHigh
	case score >= 30:
		return RiskLevelMedium
	case score > 0:
		return RiskLevelLow
	default:
		return RiskLevelUnknown
	}
}
