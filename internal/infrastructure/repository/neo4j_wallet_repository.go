package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/database"

	"go.uber.org/zap"
)

// Neo4jWalletRepository implements WalletRepository using Neo4j
type Neo4jWalletRepository struct {
	neo4j  *database.Neo4jClient
	logger *zap.Logger
}

// NewNeo4jWalletRepository creates a new Neo4j wallet repository
func NewNeo4jWalletRepository(neo4j *database.Neo4jClient, logger *zap.Logger) repository.WalletRepository {
	return &Neo4jWalletRepository{
		neo4j:  neo4j,
		logger: logger,
	}
}

// GetWalletNetwork retrieves wallet network data
func (r *Neo4jWalletRepository) GetWalletNetwork(ctx context.Context, input *entity.WalletNetworkInput) (*entity.WalletNetwork, error) {
	data, err := r.neo4j.GetWalletNetwork(ctx, input.Address, input.Depth)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet network: %w", err)
	}

	// Convert Neo4j data to domain entities
	network := &entity.WalletNetwork{
		Nodes: []entity.Wallet{},
		Links: []entity.WalletConnection{},
		Metadata: entity.NetworkMetadata{
			CenterWallet: input.Address,
			MaxDepth:     input.Depth,
			GeneratedAt:  time.Now(),
		},
	}

	// Process nodes and links
	nodeMap := make(map[string]*entity.Wallet)

	for _, record := range data {
		// Process center wallet
		centerAddr := getStringValue(record, "center_address")
		if centerAddr != "" && nodeMap[centerAddr] == nil {
			centerWallet := &entity.Wallet{
				ID:               centerAddr,
				Address:          centerAddr,
				WalletType:       entity.WalletType(getStringValue(record, "center_type")),
				RiskLevel:        entity.RiskLevel(getStringValue(record, "center_risk")),
				TransactionCount: getInt64Value(record, "center_tx_count"),
				Balance:          getStringPointer(record, "center_balance"),
				Network:          input.NetworkID,
			}
			nodeMap[centerAddr] = centerWallet
			network.Nodes = append(network.Nodes, *centerWallet)
		}

		// Process connected wallet
		connectedAddr := getStringValue(record, "connected_address")
		if connectedAddr != "" && nodeMap[connectedAddr] == nil {
			connectedWallet := &entity.Wallet{
				ID:               connectedAddr,
				Address:          connectedAddr,
				WalletType:       entity.WalletType(getStringValue(record, "connected_type")),
				RiskLevel:        entity.RiskLevel(getStringValue(record, "connected_risk")),
				TransactionCount: getInt64Value(record, "connected_tx_count"),
				Balance:          getStringPointer(record, "connected_balance"),
				Network:          input.NetworkID,
			}
			nodeMap[connectedAddr] = connectedWallet
			network.Nodes = append(network.Nodes, *connectedWallet)
		}

		// Process connection
		if centerAddr != "" && connectedAddr != "" {
			connection := entity.WalletConnection{
				Source:           centerAddr,
				Target:           connectedAddr,
				Value:            fmt.Sprintf("%.0f", getFloat64Value(record, "connection_value")),
				TransactionCount: getInt64Value(record, "connection_tx_count"),
				RiskLevel:        entity.RiskLevelLow, // Default, would be calculated
			}
			network.Links = append(network.Links, connection)
		}
	}

	network.Metadata.TotalNodes = len(network.Nodes)
	network.Metadata.TotalLinks = len(network.Links)

	return network, nil
}

// GetWallet retrieves a single wallet by address
func (r *Neo4jWalletRepository) GetWallet(ctx context.Context, address string) (*entity.Wallet, error) {
	data, err := r.neo4j.GetWalletInfo(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet info: %w", err)
	}

	wallet := &entity.Wallet{
		ID:                  address,
		Address:             address,
		WalletType:          entity.WalletType(getStringValue(data, "wallet_type")),
		RiskLevel:           entity.RiskLevel(getStringValue(data, "risk_level")),
		TransactionCount:    getInt64Value(data, "transaction_count"),
		Balance:             getStringPointer(data, "balance"),
		FirstSeen:           getTimeValue(data, "first_seen"),
		LastSeen:            getTimeValue(data, "last_seen"),
		Tags:                getStringSliceValue(data, "tags"),
		AssociatedExchanges: getStringSliceValue(data, "associated_exchanges"),
		AssociatedProtocols: getStringSliceValue(data, "associated_protocols"),
		ConfidenceScore:     getFloat64Value(data, "confidence_score"),
	}

	return wallet, nil
}

// GetWalletsByAddresses retrieves multiple wallets by addresses
func (r *Neo4jWalletRepository) GetWalletsByAddresses(ctx context.Context, addresses []string) ([]entity.Wallet, error) {
	var wallets []entity.Wallet

	for _, address := range addresses {
		wallet, err := r.GetWallet(ctx, address)
		if err != nil {
			r.logger.Warn("Failed to get wallet", zap.String("address", address), zap.Error(err))
			continue
		}
		wallets = append(wallets, *wallet)
	}

	return wallets, nil
}

// GetWalletRankings retrieves wallet rankings
func (r *Neo4jWalletRepository) GetWalletRankings(ctx context.Context, category entity.RankingCategory, networkID *string, limit, offset int) (*entity.WalletRankingResult, error) {
	data, err := r.neo4j.GetWalletRankings(ctx, string(category), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet rankings: %w", err)
	}

	var rankings []entity.WalletRanking
	for _, record := range data {
		metrics := entity.WalletMetrics{
			Address:              getStringValue(record, "address"),
			WalletType:           entity.WalletType(getStringValue(record, "wallet_type")),
			QualityScore:         getFloat64Value(record, "confidence_score"),
			TransactionCount:     getInt64Value(record, "transaction_count"),
			TransactionVolume:    getStringValue(record, "total_sent"),
			FirstTransactionDate: getTimeValue(record, "first_seen"),
			LastTransactionDate:  getTimeValue(record, "last_seen"),
			ConnectionCount:      getInt64Value(record, "connection_count"),
		}

		ranking := entity.WalletRanking{
			Rank:   getIntValue(record, "rank"),
			Wallet: metrics,
			Score:  getFloat64Value(record, "confidence_score"),
		}
		rankings = append(rankings, ranking)
	}

	result := &entity.WalletRankingResult{
		Rankings:   rankings,
		TotalCount: int64(len(rankings)), // This would need a separate count query
		Category:   entity.RankingCategory(category),
	}

	return result, nil
}

// SearchWallets searches for wallets
func (r *Neo4jWalletRepository) SearchWallets(ctx context.Context, query string, limit int) ([]entity.WalletSearchResult, error) {
	data, err := r.neo4j.SearchWallets(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search wallets: %w", err)
	}

	var results []entity.WalletSearchResult
	for _, record := range data {
		result := entity.WalletSearchResult{
			Address:          getStringValue(record, "address"),
			Label:            getStringPointer(record, "label"),
			Tags:             getStringSliceValue(record, "tags"),
			TransactionCount: getInt64Value(record, "transaction_count"),
			Balance:          getStringPointer(record, "balance"),
			RelevanceScore:   getFloat64Value(record, "confidence_score"),
		}
		results = append(results, result)
	}

	return results, nil
}

// GetRiskScore retrieves risk score for a wallet
func (r *Neo4jWalletRepository) GetRiskScore(ctx context.Context, address string) (*entity.RiskScore, error) {
	// This would typically come from a separate risk scoring service
	// For now, return a placeholder
	return &entity.RiskScore{
		Address:     address,
		TotalScore:  50,
		RiskLevel:   entity.RiskLevelMedium,
		Factors:     entity.RiskFactors{},
		Flags:       []string{},
		LastUpdated: time.Now(),
	}, nil
}

// GetRiskScores retrieves risk scores for multiple wallets
func (r *Neo4jWalletRepository) GetRiskScores(ctx context.Context, addresses []string) ([]entity.RiskScore, error) {
	var scores []entity.RiskScore
	for _, address := range addresses {
		score, err := r.GetRiskScore(ctx, address)
		if err != nil {
			r.logger.Warn("Failed to get risk score", zap.String("address", address), zap.Error(err))
			continue
		}
		scores = append(scores, *score)
	}
	return scores, nil
}

// UpdateRiskScore updates risk score for a wallet
func (r *Neo4jWalletRepository) UpdateRiskScore(ctx context.Context, address string, manualFlags []string, whitelistStatus *bool) (*entity.RiskScore, error) {
	// This would update the risk score in the database
	// For now, return the current score
	return r.GetRiskScore(ctx, address)
}

// GetWalletStats retrieves wallet statistics
func (r *Neo4jWalletRepository) GetWalletStats(ctx context.Context, address string) (*entity.WalletStats, error) {
	data, err := r.neo4j.GetWalletInfo(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet stats: %w", err)
	}

	stats := &entity.WalletStats{
		Address:          address,
		TransactionCount: getInt64Value(data, "transaction_count"),
		TotalVolume:      getStringValue(data, "total_volume"),
	}

	return stats, nil
}

// Helper functions to safely extract values from Neo4j records
func getStringValue(record map[string]interface{}, key string) string {
	if val, ok := record[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringPointer(record map[string]interface{}, key string) *string {
	if val := getStringValue(record, key); val != "" {
		return &val
	}
	return nil
}

func getInt64Value(record map[string]interface{}, key string) int64 {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return 0
}

func getIntValue(record map[string]interface{}, key string) int {
	return int(getInt64Value(record, key))
}

func getFloat64Value(record map[string]interface{}, key string) float64 {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int64:
			return float64(v)
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func getTimeValue(record map[string]interface{}, key string) time.Time {
	if val, ok := record[key]; ok && val != nil {
		if t, ok := val.(time.Time); ok {
			return t
		}
		if str, ok := val.(string); ok {
			if t, err := time.Parse(time.RFC3339, str); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func getStringSliceValue(record map[string]interface{}, key string) []string {
	if val, ok := record[key]; ok && val != nil {
		if slice, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return []string{}
}
