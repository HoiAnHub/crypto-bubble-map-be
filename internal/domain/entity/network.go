package entity

import "time"

// NetworkCategory represents different categories of blockchain networks
type NetworkCategory string

const (
	NetworkCategoryLayer1   NetworkCategory = "LAYER1"
	NetworkCategoryLayer2   NetworkCategory = "LAYER2"
	NetworkCategorySidechain NetworkCategory = "SIDECHAIN"
)

// NetworkInfo represents information about a blockchain network
type NetworkInfo struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Symbol       string          `json:"symbol"`
	Category     NetworkCategory `json:"category"`
	TVL          *float64        `json:"tvl,omitempty"`
	DailyTransactions *int       `json:"daily_transactions,omitempty"`
	TPS          *int            `json:"tps,omitempty"`
	GasPrice     *float64        `json:"gas_price,omitempty"`
	BlockTime    *int            `json:"block_time,omitempty"`
	IsActive     bool            `json:"is_active"`
	GradientFrom string          `json:"gradient_from"`
	GradientTo   string          `json:"gradient_to"`
	Icon         string          `json:"icon"`
	Description  *string         `json:"description,omitempty"`
	Website      *string         `json:"website,omitempty"`
	Explorer     *string         `json:"explorer,omitempty"`
}

// NetworkStats represents statistics for a specific network
type NetworkStats struct {
	NetworkID       string                  `json:"network_id"`
	TotalWallets    int64                   `json:"total_wallets"`
	TotalVolume     string                  `json:"total_volume"`
	TotalTransactions int64                 `json:"total_transactions"`
	FlaggedWallets  int64                   `json:"flagged_wallets"`
	WalletTypes     WalletTypeDistribution  `json:"wallet_types"`
	RiskDistribution RiskDistribution       `json:"risk_distribution"`
	LastUpdate      time.Time               `json:"last_update"`
}

// WalletTypeDistribution represents the distribution of wallet types
type WalletTypeDistribution struct {
	Regular  int64 `json:"regular"`
	Exchange int64 `json:"exchange"`
	Contract int64 `json:"contract"`
	Whale    int64 `json:"whale"`
	Defi     int64 `json:"defi"`
	Bridge   int64 `json:"bridge"`
	Miner    int64 `json:"miner"`
}

// NetworkRanking represents a network's ranking
type NetworkRanking struct {
	Rank    int            `json:"rank"`
	Network NetworkInfo    `json:"network"`
	Score   float64        `json:"score"`
	Change  *int           `json:"change,omitempty"`
	Metrics NetworkMetrics `json:"metrics"`
}

// NetworkMetrics represents comprehensive metrics for network ranking
type NetworkMetrics struct {
	TVL                float64 `json:"tvl"`
	MarketCap          float64 `json:"market_cap"`
	DailyVolume        float64 `json:"daily_volume"`
	ActiveUsers        int64   `json:"active_users"`
	DeveloperActivity  float64 `json:"developer_activity"`
	EcosystemGrowth    float64 `json:"ecosystem_growth"`
}

// NetworkActivity represents network activity over time
type NetworkActivity struct {
	Timestamp        time.Time `json:"timestamp"`
	TransactionCount int64     `json:"transaction_count"`
	Volume           string    `json:"volume"`
	UniqueWallets    int64     `json:"unique_wallets"`
}

// NetworkActivityUpdate represents real-time network activity updates
type NetworkActivityUpdate struct {
	NetworkID        string    `json:"network_id"`
	Timestamp        time.Time `json:"timestamp"`
	TransactionCount int64     `json:"transaction_count"`
	Volume           string    `json:"volume"`
	UniqueWallets    int64     `json:"unique_wallets"`
}

// DashboardStats represents overall dashboard statistics
type DashboardStats struct {
	TotalWallets         int64                   `json:"total_wallets"`
	TotalVolume          string                  `json:"total_volume"`
	TotalTransactions    int64                   `json:"total_transactions"`
	FlaggedWallets       int64                   `json:"flagged_wallets"`
	WhitelistedWallets   int64                   `json:"whitelisted_wallets"`
	AverageQualityScore  float64                 `json:"average_quality_score"`
	WalletTypes          WalletTypeDistribution  `json:"wallet_types"`
	RiskDistribution     RiskDistribution        `json:"risk_distribution"`
	NetworkActivity      []NetworkActivity       `json:"network_activity"`
	TopTokens            []string                `json:"top_tokens"`
	RecentActivity       []ActivitySummary       `json:"recent_activity"`
}

// ActivitySummary represents a summary of recent activity
type ActivitySummary struct {
	Type      string    `json:"type"`
	Count     int64     `json:"count"`
	Volume    string    `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// GetDefaultNetworks returns a list of default supported networks
func GetDefaultNetworks() []NetworkInfo {
	return []NetworkInfo{
		{
			ID:           "ethereum",
			Name:         "Ethereum",
			Symbol:       "ETH",
			Category:     NetworkCategoryLayer1,
			IsActive:     true,
			GradientFrom: "#627EEA",
			GradientTo:   "#8A92B2",
			Icon:         "FaEthereum",
			Description:  stringPtr("The world's programmable blockchain"),
			Website:      stringPtr("https://ethereum.org"),
			Explorer:     stringPtr("https://etherscan.io"),
		},
		{
			ID:           "binance-smart-chain",
			Name:         "BNB Smart Chain",
			Symbol:       "BNB",
			Category:     NetworkCategorySidechain,
			IsActive:     true,
			GradientFrom: "#F3BA2F",
			GradientTo:   "#FCD535",
			Icon:         "SiBinance",
			Description:  stringPtr("High-performance blockchain for DeFi"),
			Website:      stringPtr("https://www.bnbchain.org"),
			Explorer:     stringPtr("https://bscscan.com"),
		},
		{
			ID:           "polygon",
			Name:         "Polygon",
			Symbol:       "MATIC",
			Category:     NetworkCategoryLayer2,
			IsActive:     true,
			GradientFrom: "#8247E5",
			GradientTo:   "#A855F7",
			Icon:         "SiPolygon",
			Description:  stringPtr("Ethereum's Internet of Blockchains"),
			Website:      stringPtr("https://polygon.technology"),
			Explorer:     stringPtr("https://polygonscan.com"),
		},
		{
			ID:           "arbitrum",
			Name:         "Arbitrum",
			Symbol:       "ARB",
			Category:     NetworkCategoryLayer2,
			IsActive:     true,
			GradientFrom: "#28A0F0",
			GradientTo:   "#4FC3F7",
			Icon:         "SiArbitrum",
			Description:  stringPtr("Optimistic rollup for Ethereum"),
			Website:      stringPtr("https://arbitrum.io"),
			Explorer:     stringPtr("https://arbiscan.io"),
		},
		{
			ID:           "optimism",
			Name:         "Optimism",
			Symbol:       "OP",
			Category:     NetworkCategoryLayer2,
			IsActive:     true,
			GradientFrom: "#FF0420",
			GradientTo:   "#FF6B6B",
			Icon:         "SiOptimism",
			Description:  stringPtr("Optimistic rollup scaling solution"),
			Website:      stringPtr("https://optimism.io"),
			Explorer:     stringPtr("https://optimistic.etherscan.io"),
		},
		{
			ID:           "solana",
			Name:         "Solana",
			Symbol:       "SOL",
			Category:     NetworkCategoryLayer1,
			IsActive:     true,
			GradientFrom: "#9945FF",
			GradientTo:   "#14F195",
			Icon:         "SiSolana",
			Description:  stringPtr("High-performance blockchain"),
			Website:      stringPtr("https://solana.com"),
			Explorer:     stringPtr("https://explorer.solana.com"),
		},
		{
			ID:           "near",
			Name:         "NEAR Protocol",
			Symbol:       "NEAR",
			Category:     NetworkCategoryLayer1,
			IsActive:     true,
			GradientFrom: "#00C08B",
			GradientTo:   "#58D68D",
			Icon:         "SiNear",
			Description:  stringPtr("Sharded proof-of-stake blockchain"),
			Website:      stringPtr("https://near.org"),
			Explorer:     stringPtr("https://explorer.near.org"),
		},
		{
			ID:           "aptos",
			Name:         "Aptos",
			Symbol:       "APT",
			Category:     NetworkCategoryLayer1,
			IsActive:     true,
			GradientFrom: "#00D4AA",
			GradientTo:   "#4ECDC4",
			Icon:         "SiAptos",
			Description:  stringPtr("Scalable and safe blockchain"),
			Website:      stringPtr("https://aptoslabs.com"),
			Explorer:     stringPtr("https://explorer.aptoslabs.com"),
		},
		{
			ID:           "sui",
			Name:         "Sui",
			Symbol:       "SUI",
			Category:     NetworkCategoryLayer1,
			IsActive:     true,
			GradientFrom: "#4DA2FF",
			GradientTo:   "#6BCFFF",
			Icon:         "SiSui",
			Description:  stringPtr("Next-generation smart contract platform"),
			Website:      stringPtr("https://sui.io"),
			Explorer:     stringPtr("https://explorer.sui.io"),
		},
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper methods for NetworkInfo
func (ni *NetworkInfo) IsLayer1() bool {
	return ni.Category == NetworkCategoryLayer1
}

func (ni *NetworkInfo) IsLayer2() bool {
	return ni.Category == NetworkCategoryLayer2
}

func (ni *NetworkInfo) IsSidechain() bool {
	return ni.Category == NetworkCategorySidechain
}

// Helper methods for NetworkStats
func (ns *NetworkStats) GetWalletTypePercentage(walletType string) float64 {
	if ns.TotalWallets == 0 {
		return 0
	}
	
	var count int64
	switch walletType {
	case "regular":
		count = ns.WalletTypes.Regular
	case "exchange":
		count = ns.WalletTypes.Exchange
	case "contract":
		count = ns.WalletTypes.Contract
	case "whale":
		count = ns.WalletTypes.Whale
	case "defi":
		count = ns.WalletTypes.Defi
	case "bridge":
		count = ns.WalletTypes.Bridge
	case "miner":
		count = ns.WalletTypes.Miner
	}
	
	return float64(count) / float64(ns.TotalWallets) * 100
}

func (ns *NetworkStats) GetRiskPercentage(riskLevel string) float64 {
	total := ns.RiskDistribution.Low + ns.RiskDistribution.Medium + 
			ns.RiskDistribution.High + ns.RiskDistribution.Critical
	
	if total == 0 {
		return 0
	}
	
	var count int64
	switch riskLevel {
	case "low":
		count = ns.RiskDistribution.Low
	case "medium":
		count = ns.RiskDistribution.Medium
	case "high":
		count = ns.RiskDistribution.High
	case "critical":
		count = ns.RiskDistribution.Critical
	}
	
	return float64(count) / float64(total) * 100
}
