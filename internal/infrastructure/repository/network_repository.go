package repository

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// NetworkRepository implements NetworkRepository interface
type NetworkRepository struct {
	neo4j  *database.Neo4jClient
	mongo  *database.MongoClient
	logger *zap.Logger
}

// NewNetworkRepository creates a new network repository
func NewNetworkRepository(neo4j *database.Neo4jClient, mongo *database.MongoClient, logger *zap.Logger) repository.NetworkRepository {
	return &NetworkRepository{
		neo4j:  neo4j,
		mongo:  mongo,
		logger: logger,
	}
}

// GetNetworks retrieves all supported networks
func (r *NetworkRepository) GetNetworks(ctx context.Context) ([]entity.NetworkInfo, error) {
	// Return default networks with some dynamic data
	networks := entity.GetDefaultNetworks()

	// TODO: Add real-time data from external APIs
	for i := range networks {
		networks[i].TVL = getRandomFloat64(1000000, 100000000)
		networks[i].DailyTransactions = getRandomInt(10000, 1000000)
		networks[i].TPS = getRandomInt(100, 10000)
		networks[i].GasPrice = getRandomFloat64(0.001, 0.1)
		networks[i].BlockTime = getRandomInt(1, 15)
	}

	return networks, nil
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
	data, err := r.neo4j.GetNetworkStats(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %w", err)
	}

	stats := &entity.NetworkStats{
		NetworkID:         networkID,
		TotalWallets:      getInt64Value(data, "total_wallets"),
		TotalVolume:       "0", // Would be calculated from transaction data
		TotalTransactions: 0,   // Would be calculated from transaction data
		FlaggedWallets:    0,   // Would be calculated from risk data
		WalletTypes: entity.WalletTypeDistribution{
			Regular:  100,
			Exchange: 50,
			Contract: 200,
			Whale:    10,
			Defi:     75,
			Bridge:   25,
			Miner:    15,
		},
		RiskDistribution: entity.RiskDistribution{
			Low:      800,
			Medium:   150,
			High:     40,
			Critical: 10,
		},
		LastUpdate: time.Now(),
	}

	return stats, nil
}

// GetNetworkRankings retrieves network rankings
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

		metrics := entity.NetworkMetrics{
			TVL:               *network.TVL,
			MarketCap:         *network.TVL * 2, // Simplified calculation
			DailyVolume:       float64(*network.DailyTransactions) * 1000,
			ActiveUsers:       int64(*network.DailyTransactions / 10),
			DeveloperActivity: *getRandomFloat64(0.5, 1.0),
			EcosystemGrowth:   *getRandomFloat64(0.8, 1.2),
		}

		// Calculate composite score
		score := (metrics.TVL/1000000)*0.3 +
			(metrics.DailyVolume/1000000)*0.2 +
			float64(metrics.ActiveUsers/1000)*0.2 +
			metrics.DeveloperActivity*100*0.15 +
			metrics.EcosystemGrowth*100*0.15

		ranking := entity.NetworkRanking{
			Rank:    i + 1,
			Network: network,
			Score:   score,
			Change:  getRandomIntPointer(-5, 5),
			Metrics: metrics,
		}

		rankings = append(rankings, ranking)
	}

	return rankings, nil
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
		WalletTypes: entity.WalletTypeDistribution{
			Regular:  800000,
			Exchange: 5000,
			Contract: 150000,
			Whale:    1000,
			Defi:     30000,
			Bridge:   2000,
			Miner:    12000,
		},
		RiskDistribution: entity.RiskDistribution{
			Low:      850000,
			Medium:   120000,
			High:     25000,
			Critical: 5000,
		},
		NetworkActivity: networkActivity,
		TopTokens:       topTokens,
		RecentActivity:  activities,
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
