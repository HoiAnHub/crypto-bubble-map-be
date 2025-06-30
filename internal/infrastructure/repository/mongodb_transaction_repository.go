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

// MongoTransactionRepository implements TransactionRepository using MongoDB
type MongoTransactionRepository struct {
	mongo  *database.MongoClient
	logger *zap.Logger
}

// NewMongoTransactionRepository creates a new MongoDB transaction repository
func NewMongoTransactionRepository(mongo *database.MongoClient, logger *zap.Logger) repository.TransactionRepository {
	return &MongoTransactionRepository{
		mongo:  mongo,
		logger: logger,
	}
}

// GetTransactionsByWallet retrieves transactions for a specific wallet
func (r *MongoTransactionRepository) GetTransactionsByWallet(ctx context.Context, walletAddress string, limit, offset int64) ([]entity.Transaction, error) {
	data, err := r.mongo.GetTransactionsByWallet(ctx, walletAddress, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by wallet: %w", err)
	}

	var transactions []entity.Transaction
	for _, record := range data {
		tx := r.convertToTransaction(record)
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetPairwiseTransactions retrieves transactions between two specific wallets
func (r *MongoTransactionRepository) GetPairwiseTransactions(ctx context.Context, walletA, walletB string, limit, offset int64, filters *entity.TransactionFilters) (*entity.PairwiseTransactionResult, error) {
	data, err := r.mongo.GetPairwiseTransactions(ctx, walletA, walletB, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pairwise transactions: %w", err)
	}

	var transactions []entity.PairwiseTransaction
	for _, record := range data {
		tx := r.convertToPairwiseTransaction(record, walletA, walletB)
		transactions = append(transactions, tx)
	}

	// Create summary
	summary := entity.PairwiseTransactionSummary{
		WalletA:           walletA,
		WalletB:           walletB,
		TotalTransactions: int64(len(transactions)),
		TopTokens:         []entity.TopToken{},
		RiskDistribution:  entity.RiskDistribution{},
		TransactionTypes:  make(map[string]int64),
	}

	if len(transactions) > 0 {
		summary.FirstTransaction = transactions[len(transactions)-1].Timestamp
		summary.LastTransaction = transactions[0].Timestamp
	}

	// Convert PairwiseTransaction to Transaction
	convertedTransactions := make([]entity.Transaction, len(transactions))
	for i, tx := range transactions {
		convertedTransactions[i] = entity.Transaction{
			Hash:            tx.Hash,
			From:            tx.From,
			To:              &tx.To,
			Value:           tx.Value,
			Timestamp:       tx.Timestamp,
			BlockNumber:     tx.BlockNumber,
			GasUsed:         uint64(0), // Convert from string if needed
			GasPrice:        tx.GasPrice,
			GasFee:          tx.GasFee,
			TransactionType: tx.TransactionType,
			TxStatus:        entity.TransactionStatus(tx.Status),
			RiskLevel:       tx.RiskLevel,
		}
	}

	result := &entity.PairwiseTransactionResult{
		Transactions: convertedTransactions,
		Summary:      summary,
		TotalCount:   int64(len(transactions)),
	}

	return result, nil
}

// GetTransaction retrieves a single transaction by hash
func (r *MongoTransactionRepository) GetTransaction(ctx context.Context, hash string) (*entity.Transaction, error) {
	filter := bson.M{"hash": hash}
	data, err := r.mongo.GetTransactions(ctx, filter, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("transaction not found")
	}

	tx := r.convertToTransaction(data[0])
	return &tx, nil
}

// GetMoneyFlowData retrieves money flow analysis data
func (r *MongoTransactionRepository) GetMoneyFlowData(ctx context.Context, walletAddress string, filters *entity.MoneyFlowFilters) (*entity.MoneyFlowData, error) {
	flowType := "BOTH"
	if filters != nil {
		flowType = string(filters.FlowType)
	}

	data, err := r.mongo.GetMoneyFlowData(ctx, walletAddress, flowType, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get money flow data: %w", err)
	}

	// Convert to money flow data structure
	centerAccount := entity.MoneyFlowAccount{
		Address: walletAddress,
	}

	var inboundAccounts []entity.MoneyFlowAccount
	var outboundAccounts []entity.MoneyFlowAccount
	var transactions []entity.MoneyFlowTransaction

	for _, record := range data {
		counterparty := getStringValue(record, "_id")
		if counterparty == "" {
			continue
		}

		account := entity.MoneyFlowAccount{
			Address:          counterparty,
			TotalValue:       fmt.Sprintf("%.0f", getFloat64Value(record, "total_value")),
			TotalUsdValue:    fmt.Sprintf("%.2f", getFloat64Value(record, "total_value")), // Would need price conversion
			TransactionCount: getInt64Value(record, "transaction_count"),
			FirstSeen:        getTimeValue(record, "first_transaction"),
			LastSeen:         getTimeValue(record, "last_transaction"),
		}

		// Determine if inbound or outbound based on transaction direction
		// This is simplified - would need more logic to determine direction
		if getFloat64Value(record, "total_value") > 0 {
			outboundAccounts = append(outboundAccounts, account)
		} else {
			inboundAccounts = append(inboundAccounts, account)
		}
	}

	summary := entity.MoneyFlowSummary{
		TotalInbound:         "0",
		TotalOutbound:        "0",
		TotalInboundUsd:      "0",
		TotalOutboundUsd:     "0",
		UniqueCounterparties: len(inboundAccounts) + len(outboundAccounts),
		TimeRange: entity.TimeRange{
			Start: time.Now().AddDate(0, -1, 0),
			End:   time.Now(),
		},
		TopTokens: []entity.TopToken{},
	}

	sankeyData := entity.SankeyData{
		Nodes: []entity.SankeyNode{},
		Links: []entity.SankeyLink{},
	}

	result := &entity.MoneyFlowData{
		CenterAccount:    centerAccount,
		InboundAccounts:  inboundAccounts,
		OutboundAccounts: outboundAccounts,
		Transactions:     transactions,
		Summary:          summary,
		SankeyData:       sankeyData,
	}

	return result, nil
}

// SearchTransactions searches transactions by hash or address
func (r *MongoTransactionRepository) SearchTransactions(ctx context.Context, query string, limit int64) ([]entity.Transaction, error) {
	data, err := r.mongo.SearchTransactions(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search transactions: %w", err)
	}

	var transactions []entity.Transaction
	for _, record := range data {
		tx := r.convertToTransaction(record)
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetTransactionStats retrieves transaction statistics
func (r *MongoTransactionRepository) GetTransactionStats(ctx context.Context, timeRange *entity.TimeRange) (map[string]interface{}, error) {
	var mongoTimeRange *database.TimeRange
	if timeRange != nil {
		mongoTimeRange = &database.TimeRange{
			Start: timeRange.Start,
			End:   timeRange.End,
		}
	}

	stats, err := r.mongo.GetTransactionStats(ctx, mongoTimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %w", err)
	}

	return stats, nil
}

// GetTopTokens retrieves the most transacted tokens
func (r *MongoTransactionRepository) GetTopTokens(ctx context.Context, limit int64) ([]entity.TokenSummary, error) {
	data, err := r.mongo.GetTopTokens(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top tokens: %w", err)
	}

	var tokens []entity.TokenSummary
	for _, record := range data {
		token := entity.TokenSummary{
			Symbol:           getStringValue(record, "symbol"),
			Volume:           fmt.Sprintf("%.0f", getFloat64Value(record, "total_volume")),
			VolumeUSD:        getFloat64Value(record, "total_volume"), // Would need price conversion
			TransactionCount: getInt64Value(record, "transaction_count"),
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// GetRecentActivity retrieves recent blockchain activity
func (r *MongoTransactionRepository) GetRecentActivity(ctx context.Context, limit int64) ([]entity.ActivitySummary, error) {
	data, err := r.mongo.GetRecentActivity(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	var activities []entity.ActivitySummary
	for _, record := range data {
		activity := entity.ActivitySummary{
			Type:      "transaction",
			Count:     1,
			Volume:    getStringValue(record, "value"),
			Timestamp: getTimeValue(record, "crawled_at"),
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// Helper methods to convert MongoDB records to domain entities

func (r *MongoTransactionRepository) convertToTransaction(record bson.M) entity.Transaction {
	tx := entity.Transaction{
		Hash:        getStringValue(record, "hash"),
		From:        getStringValue(record, "from"),
		To:          getStringPointer(record, "to"),
		Value:       getStringValue(record, "value"),
		Timestamp:   getTimeValue(record, "crawled_at"),
		BlockNumber: getStringValue(record, "block_number"),
		GasUsed:     uint64(getInt64Value(record, "gas_used")),
		GasPrice:    getStringValue(record, "gas_price"),
		Network:     getStringValue(record, "network"),
		TxStatus:    entity.TransactionStatusSuccess, // Default
		RiskLevel:   entity.RiskLevelLow,             // Default
	}

	// Calculate gas fee
	gasUsed := getInt64Value(record, "gas_used")
	gasPrice := getInt64Value(record, "gas_price")
	tx.GasFee = fmt.Sprintf("%d", gasUsed*gasPrice)

	// Determine transaction type
	if tx.To == nil {
		tx.TransactionType = entity.TransactionTypeContractCall
	} else {
		tx.TransactionType = entity.TransactionTypeTransfer
	}

	return tx
}

func (r *MongoTransactionRepository) convertToPairwiseTransaction(record bson.M, walletA, walletB string) entity.PairwiseTransaction {
	from := getStringValue(record, "from")
	to := getStringValue(record, "to")

	// Determine direction
	direction := entity.TransactionDirectionOutgoing
	if from == walletB || to == walletA {
		direction = entity.TransactionDirectionIncoming
	}

	tx := entity.PairwiseTransaction{
		ID:              getStringValue(record, "hash"),
		Hash:            getStringValue(record, "hash"),
		From:            from,
		To:              to,
		Value:           getStringValue(record, "value"),
		Token:           "ETH", // Default
		TokenSymbol:     "ETH",
		Timestamp:       getTimeValue(record, "crawled_at"),
		BlockNumber:     getStringValue(record, "block_number"),
		GasUsed:         fmt.Sprintf("%d", getInt64Value(record, "gas_used")),
		GasPrice:        getStringValue(record, "gas_price"),
		TransactionType: entity.TransactionTypeTransfer,
		RiskLevel:       entity.RiskLevelLow,
		Status:          entity.TransactionStatusSuccess,
		Direction:       direction,
	}

	// Calculate gas fee
	gasUsed := getInt64Value(record, "gas_used")
	gasPrice := getInt64Value(record, "gas_price")
	tx.GasFee = fmt.Sprintf("%d", gasUsed*gasPrice)

	return tx
}
