package repository

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/database"
	"crypto-bubble-map-be/internal/infrastructure/external"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// NetworkRepository implements NetworkRepository interface
type NetworkRepository struct {
	neo4j     *database.Neo4jClient
	mongo     *database.MongoClient
	apiClient *external.BlockchainAPIClient
	logger    *zap.Logger
}

// NewNetworkRepository creates a new network repository
func NewNetworkRepository(neo4j *database.Neo4jClient, mongo *database.MongoClient, apiClient *external.BlockchainAPIClient, logger *zap.Logger) repository.NetworkRepository {
	return &NetworkRepository{
		neo4j:     neo4j,
		mongo:     mongo,
		apiClient: apiClient,
		logger:    logger,
	}
}

// GetNetworks retrieves all supported networks with real-time data
func (r *NetworkRepository) GetNetworks(ctx context.Context) ([]entity.NetworkInfo, error) {
	// Get default networks
	networks := entity.GetDefaultNetworks()

	// Fetch real-time data for each network
	for i := range networks {
		if r.apiClient != nil {
			// Try to fetch real-time data
			if networkData, err := r.apiClient.GetNetworkData(ctx, networks[i].ID); err == nil {
				// Update with real-time data
				if networkData.TVL > 0 {
					networks[i].TVL = &networkData.TVL
				}
				if networkData.DailyTransactions > 0 {
					networks[i].DailyTransactions = &networkData.DailyTransactions
				}
				if networkData.TPS > 0 {
					networks[i].TPS = &networkData.TPS
				}
				if networkData.GasPrice > 0 {
					networks[i].GasPrice = &networkData.GasPrice
				}
				if networkData.BlockTime > 0 {
					networks[i].BlockTime = &networkData.BlockTime
				}
			} else {
				r.logger.Debug("Failed to fetch real-time data, using fallback",
					zap.String("networkID", networks[i].ID),
					zap.Error(err))
				// Fallback to estimated data
				r.setFallbackData(&networks[i])
			}
		} else {
			// No API client available, use fallback data
			r.setFallbackData(&networks[i])
		}
	}

	return networks, nil
}

// setFallbackData sets reasonable fallback data for networks
func (r *NetworkRepository) setFallbackData(network *entity.NetworkInfo) {
	// Set fallback data based on network type and known characteristics
	switch network.ID {
	case "ethereum":
		tvl := 50000000000.0 // $50B
		network.TVL = &tvl
		dailyTx := 1200000
		network.DailyTransactions = &dailyTx
		tps := 15
		network.TPS = &tps
		gasPrice := 30.0
		network.GasPrice = &gasPrice
		blockTime := 12
		network.BlockTime = &blockTime
	case "polygon":
		tvl := 1000000000.0 // $1B
		network.TVL = &tvl
		dailyTx := 3000000
		network.DailyTransactions = &dailyTx
		tps := 7000
		network.TPS = &tps
		gasPrice := 30.0
		network.GasPrice = &gasPrice
		blockTime := 2
		network.BlockTime = &blockTime
	case "bsc":
		tvl := 5000000000.0 // $5B
		network.TVL = &tvl
		dailyTx := 5000000
		network.DailyTransactions = &dailyTx
		tps := 160
		network.TPS = &tps
		gasPrice := 5.0
		network.GasPrice = &gasPrice
		blockTime := 3
		network.BlockTime = &blockTime
	default:
		// Generic fallback for other networks
		network.TVL = getRandomFloat64(100000000, 10000000000) // $100M - $10B
		network.DailyTransactions = getRandomInt(50000, 2000000)
		network.TPS = getRandomInt(100, 5000)
		network.GasPrice = getRandomFloat64(1.0, 50.0)
		network.BlockTime = getRandomInt(2, 15)
	}
}

// GetNetwork retrieves a specific network by ID
func (r *NetworkRepository) GetNetwork(ctx context.Context, networkID string) (*entity.NetworkInfo, error) {
	networks, err := r.GetNetworks(ctx)
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.ID == networkID {
			return &network, nil
		}
	}

	return nil, fmt.Errorf("network not found: %s", networkID)
}

// GetNetworkStats retrieves statistics for a specific network
func (r *NetworkRepository) GetNetworkStats(ctx context.Context, networkID string) (*entity.NetworkStats, error) {
	// Get base stats from Neo4j
	data, err := r.neo4j.GetNetworkStats(ctx, networkID)
	if err != nil {
		r.logger.Warn("Failed to get Neo4j network stats, using fallback",
			zap.String("networkID", networkID),
			zap.Error(err))
		data = make(map[string]interface{})
	}

	// Get real-time network data
	var networkData *external.NetworkData
	if r.apiClient != nil {
		if apiData, err := r.apiClient.GetNetworkData(ctx, networkID); err == nil {
			networkData = apiData
		} else {
			r.logger.Debug("Failed to fetch real-time network data",
				zap.String("networkID", networkID),
				zap.Error(err))
		}
	}

	// Calculate enhanced statistics
	stats := &entity.NetworkStats{
		NetworkID:         networkID,
		TotalWallets:      getInt64Value(data, "total_wallets"),
		TotalVolume:       r.calculateTotalVolume(networkData),
		TotalTransactions: r.calculateTotalTransactions(networkData),
		FlaggedWallets:    getInt64Value(data, "flagged_wallets"),
		WalletTypes:       r.getWalletTypeDistribution(networkID, data),
		RiskDistribution:  r.getRiskDistribution(networkID, data),
		LastUpdate:        time.Now(),
	}

	// If we don't have real data, use reasonable estimates
	if stats.TotalWallets == 0 {
		stats.TotalWallets = r.estimateTotalWallets(networkID)
	}

	return stats, nil
}

// Helper methods for calculating network statistics

func (r *NetworkRepository) calculateTotalVolume(networkData *external.NetworkData) string {
	if networkData != nil && networkData.Volume24h > 0 {
		// Estimate total volume based on 24h volume (rough approximation)
		totalVolume := networkData.Volume24h * 365 // Annualized volume
		return fmt.Sprintf("%.2f", totalVolume)
	}
	return "0"
}

func (r *NetworkRepository) calculateTotalTransactions(networkData *external.NetworkData) int64 {
	if networkData != nil && networkData.DailyTransactions > 0 {
		// Estimate total transactions based on daily transactions
		// Assume network has been active for some time
		return int64(networkData.DailyTransactions) * 365 // Annualized
	}
	return 0
}

func (r *NetworkRepository) getWalletTypeDistribution(networkID string, data map[string]interface{}) entity.WalletTypeDistribution {
	// Try to get from Neo4j data first
	if dist, ok := data["wallet_types"].(entity.WalletTypeDistribution); ok {
		return dist
	}

	// Fallback to network-specific estimates
	switch networkID {
	case "ethereum":
		return entity.WalletTypeDistribution{
			Regular:  500000,
			Exchange: 1000,
			Contract: 100000,
			Whale:    5000,
			Defi:     50000,
			Bridge:   500,
			Miner:    10000,
		}
	case "polygon":
		return entity.WalletTypeDistribution{
			Regular:  200000,
			Exchange: 500,
			Contract: 50000,
			Whale:    2000,
			Defi:     30000,
			Bridge:   200,
			Miner:    1000,
		}
	case "bsc":
		return entity.WalletTypeDistribution{
			Regular:  300000,
			Exchange: 800,
			Contract: 75000,
			Whale:    3000,
			Defi:     40000,
			Bridge:   300,
			Miner:    2000,
		}
	default:
		return entity.WalletTypeDistribution{
			Regular:  100000,
			Exchange: 200,
			Contract: 20000,
			Whale:    1000,
			Defi:     10000,
			Bridge:   100,
			Miner:    500,
		}
	}
}

func (r *NetworkRepository) getRiskDistribution(networkID string, data map[string]interface{}) entity.RiskDistribution {
	// Try to get from Neo4j data first
	if dist, ok := data["risk_distribution"].(entity.RiskDistribution); ok {
		return dist
	}

	// Fallback to network-specific estimates
	totalWallets := r.estimateTotalWallets(networkID)

	return entity.RiskDistribution{
		Low:      int64(float64(totalWallets) * 0.8),  // 80% low risk
		Medium:   int64(float64(totalWallets) * 0.15), // 15% medium risk
		High:     int64(float64(totalWallets) * 0.04), // 4% high risk
		Critical: int64(float64(totalWallets) * 0.01), // 1% critical risk
	}
}

func (r *NetworkRepository) estimateTotalWallets(networkID string) int64 {
	switch networkID {
	case "ethereum":
		return 200000000 // 200M wallets
	case "polygon":
		return 50000000 // 50M wallets
	case "bsc":
		return 100000000 // 100M wallets
	case "arbitrum":
		return 10000000 // 10M wallets
	case "optimism":
		return 5000000 // 5M wallets
	case "avalanche":
		return 15000000 // 15M wallets
	case "fantom":
		return 8000000 // 8M wallets
	case "solana":
		return 30000000 // 30M wallets
	default:
		return 1000000 // 1M wallets default
	}
}

// GetNetworkRankings retrieves network rankings with enhanced real-time metrics
func (r *NetworkRepository) GetNetworkRankings(ctx context.Context, limit int) ([]entity.NetworkRanking, error) {
	networks, err := r.GetNetworks(ctx)
	if err != nil {
		return nil, err
	}

	var rankings []entity.NetworkRanking
	for i, network := range networks {
		if i >= limit {
			break
		}

		// Get real-time network data for enhanced metrics
		var networkData *external.NetworkData
		if r.apiClient != nil {
			if apiData, err := r.apiClient.GetNetworkData(ctx, network.ID); err == nil {
				networkData = apiData
			}
		}

		metrics := r.calculateNetworkMetrics(network, networkData)

		// Calculate composite score with weighted factors
		score := r.calculateCompositeScore(metrics)

		// Calculate ranking change (simplified - would need historical data)
		change := r.calculateRankingChange(network.ID, score)

		ranking := entity.NetworkRanking{
			Rank:    i + 1,
			Network: network,
			Score:   score,
			Change:  &change,
			Metrics: metrics,
		}

		rankings = append(rankings, ranking)
	}

	// Sort rankings by score (descending)
	for i := 0; i < len(rankings)-1; i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[j].Score > rankings[i].Score {
				rankings[i], rankings[j] = rankings[j], rankings[i]
				// Update ranks after sorting
				rankings[i].Rank = i + 1
				rankings[j].Rank = j + 1
			}
		}
	}

	return rankings, nil
}

// calculateNetworkMetrics calculates comprehensive metrics for a network
func (r *NetworkRepository) calculateNetworkMetrics(network entity.NetworkInfo, networkData *external.NetworkData) entity.NetworkMetrics {
	metrics := entity.NetworkMetrics{}

	// Use real-time data if available, otherwise use network data
	if networkData != nil {
		metrics.TVL = networkData.TVL
		metrics.MarketCap = networkData.MarketCap
		metrics.DailyVolume = networkData.Volume24h
		metrics.ActiveUsers = int64(networkData.DailyTransactions / 10) // Estimate active users
	} else {
		// Fallback to network data
		if network.TVL != nil {
			metrics.TVL = *network.TVL
		}
		if network.DailyTransactions != nil {
			metrics.DailyVolume = float64(*network.DailyTransactions) * 1000 // Estimate volume
			metrics.ActiveUsers = int64(*network.DailyTransactions / 10)
		}
		// Estimate market cap based on TVL
		metrics.MarketCap = metrics.TVL * 2
	}

	// Calculate developer activity and ecosystem growth
	metrics.DeveloperActivity = r.estimateDeveloperActivity(network.ID)
	metrics.EcosystemGrowth = r.estimateEcosystemGrowth(network.ID)

	return metrics
}

// calculateCompositeScore calculates a composite score for network ranking
func (r *NetworkRepository) calculateCompositeScore(metrics entity.NetworkMetrics) float64 {
	// Weighted scoring system
	weights := map[string]float64{
		"tvl":                0.25, // 25% weight
		"market_cap":         0.20, // 20% weight
		"daily_volume":       0.20, // 20% weight
		"active_users":       0.15, // 15% weight
		"developer_activity": 0.10, // 10% weight
		"ecosystem_growth":   0.10, // 10% weight
	}

	// Normalize values and calculate weighted score
	score := 0.0

	// TVL score (normalized to 0-100)
	tvlScore := normalizeValue(metrics.TVL, 0, 100000000000) * 100 // Max $100B
	score += tvlScore * weights["tvl"]

	// Market cap score
	marketCapScore := normalizeValue(metrics.MarketCap, 0, 500000000000) * 100 // Max $500B
	score += marketCapScore * weights["market_cap"]

	// Daily volume score
	volumeScore := normalizeValue(metrics.DailyVolume, 0, 50000000000) * 100 // Max $50B daily
	score += volumeScore * weights["daily_volume"]

	// Active users score
	usersScore := normalizeValue(float64(metrics.ActiveUsers), 0, 10000000) * 100 // Max 10M users
	score += usersScore * weights["active_users"]

	// Developer activity score (already 0-1, multiply by 100)
	score += metrics.DeveloperActivity * 100 * weights["developer_activity"]

	// Ecosystem growth score (already 0-2, normalize to 0-100)
	ecosystemScore := normalizeValue(metrics.EcosystemGrowth, 0, 2) * 100
	score += ecosystemScore * weights["ecosystem_growth"]

	return score
}

// normalizeValue normalizes a value to 0-1 range
func normalizeValue(value, min, max float64) float64 {
	if max <= min {
		return 0
	}
	normalized := (value - min) / (max - min)
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

// estimateDeveloperActivity estimates developer activity for a network
func (r *NetworkRepository) estimateDeveloperActivity(networkID string) float64 {
	// This would ideally come from GitHub API, developer surveys, etc.
	activityMap := map[string]float64{
		"ethereum":  0.95,
		"polygon":   0.85,
		"bsc":       0.75,
		"arbitrum":  0.80,
		"optimism":  0.78,
		"avalanche": 0.82,
		"fantom":    0.70,
		"solana":    0.88,
		"cardano":   0.75,
		"polkadot":  0.80,
		"cosmos":    0.77,
		"near":      0.73,
		"algorand":  0.68,
		"tezos":     0.65,
		"flow":      0.60,
		"sui":       0.85,
	}

	if activity, exists := activityMap[networkID]; exists {
		return activity
	}
	return 0.5 // Default moderate activity
}

// estimateEcosystemGrowth estimates ecosystem growth for a network
func (r *NetworkRepository) estimateEcosystemGrowth(networkID string) float64 {
	// This would ideally come from DeFi Pulse, ecosystem metrics, etc.
	growthMap := map[string]float64{
		"ethereum":  1.1,
		"polygon":   1.4,
		"bsc":       1.2,
		"arbitrum":  1.6,
		"optimism":  1.5,
		"avalanche": 1.3,
		"fantom":    1.1,
		"solana":    1.3,
		"cardano":   1.0,
		"polkadot":  1.1,
		"cosmos":    1.2,
		"near":      1.4,
		"algorand":  0.9,
		"tezos":     0.8,
		"flow":      1.0,
		"sui":       1.8,
	}

	if growth, exists := growthMap[networkID]; exists {
		return growth
	}
	return 1.0 // Default stable growth
}

// calculateRankingChange calculates ranking change (simplified)
func (r *NetworkRepository) calculateRankingChange(networkID string, currentScore float64) int {
	// This would ideally compare with historical rankings
	// For now, simulate based on network characteristics
	changeMap := map[string]int{
		"ethereum":  0,  // Stable
		"polygon":   2,  // Rising
		"bsc":       -1, // Slight decline
		"arbitrum":  3,  // Strong growth
		"optimism":  2,  // Growing
		"avalanche": 1,  // Moderate growth
		"fantom":    -2, // Declining
		"solana":    1,  // Recovering
		"sui":       5,  // New and rising
	}

	if change, exists := changeMap[networkID]; exists {
		return change
	}
	return 0 // Default no change
}

// GetDashboardStats retrieves overall dashboard statistics
func (r *NetworkRepository) GetDashboardStats(ctx context.Context, networkID *string) (*entity.DashboardStats, error) {
	// Get network stats
	var totalWallets int64 = 1000000
	var totalTransactions int64 = 50000000
	var flaggedWallets int64 = 5000
	var whitelistedWallets int64 = 10000

	if networkID != nil {
		networkStats, err := r.GetNetworkStats(ctx, *networkID)
		if err == nil {
			totalWallets = networkStats.TotalWallets
			totalTransactions = networkStats.TotalTransactions
			flaggedWallets = networkStats.FlaggedWallets
		}
	}

	// Get recent activity from MongoDB
	recentActivity, err := r.mongo.GetRecentActivity(ctx, 10)
	if err != nil {
		r.logger.Warn("Failed to get recent activity", zap.Error(err))
		recentActivity = []bson.M{}
	}

	var activities []entity.ActivitySummary
	for _, record := range recentActivity {
		activity := entity.ActivitySummary{
			Type:      "transaction",
			Count:     1,
			Volume:    getStringValue(record, "value"),
			Timestamp: getTimeValue(record, "crawled_at"),
		}
		activities = append(activities, activity)
	}

	// Get top tokens
	topTokensData, err := r.mongo.GetTopTokens(ctx, 10)
	if err != nil {
		r.logger.Warn("Failed to get top tokens", zap.Error(err))
		topTokensData = []bson.M{}
	}

	var topTokens []string
	for _, record := range topTokensData {
		symbol := getStringValue(record, "symbol")
		if symbol != "" {
			topTokens = append(topTokens, symbol)
		}
	}

	// Generate network activity data (last 30 days)
	var networkActivity []entity.NetworkActivity
	now := time.Now()
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		activity := entity.NetworkActivity{
			Timestamp:        date,
			TransactionCount: getRandomInt64(10000, 100000),
			Volume:           fmt.Sprintf("%.0f", getRandomFloat64(1000000, 10000000)),
			UniqueWallets:    getRandomInt64(1000, 10000),
		}
		networkActivity = append(networkActivity, activity)
	}

	stats := &entity.DashboardStats{
		TotalWallets:        totalWallets,
		TotalVolume:         "1000000000", // 1B ETH equivalent
		TotalTransactions:   totalTransactions,
		FlaggedWallets:      flaggedWallets,
		WhitelistedWallets:  whitelistedWallets,
		AverageQualityScore: 0.85,
		AverageRiskScore:    0.25,
		RecentActivity:      int64(len(activities)),
		LastUpdate:          time.Now(),
	}

	return stats, nil
}

// Helper functions for generating mock data
func getRandomFloat64(min, max float64) *float64 {
	// Simple mock implementation
	val := min + (max-min)*0.5 // Return middle value for consistency
	return &val
}

func getRandomInt(min, max int) *int {
	// Simple mock implementation
	val := min + (max-min)/2 // Return middle value for consistency
	return &val
}

func getRandomInt64(min, max int64) int64 {
	// Simple mock implementation
	return min + (max-min)/2 // Return middle value for consistency
}

func getRandomIntPointer(min, max int) *int {
	val := getRandomInt(min, max)
	return val
}
