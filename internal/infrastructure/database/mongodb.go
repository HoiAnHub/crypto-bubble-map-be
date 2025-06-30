package database

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoClient wraps the MongoDB client
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	logger   *zap.Logger
	config   *config.MongoDBConfig
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(cfg *config.MongoDBConfig, logger *zap.Logger) (*MongoClient, error) {
	// Configure client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetConnectTimeout(cfg.ConnectionTimeout).
		SetSocketTimeout(cfg.SocketTimeout)

	// Create client
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.Database)

	mongoClient := &MongoClient{
		client:   client,
		database: database,
		logger:   logger,
		config:   cfg,
	}

	logger.Info("MongoDB client initialized successfully",
		zap.String("uri", cfg.URI),
		zap.String("database", cfg.Database),
	)

	return mongoClient, nil
}

// Close closes the MongoDB connection
func (c *MongoClient) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// GetDatabase returns the database instance
func (c *MongoClient) GetDatabase() *mongo.Database {
	return c.database
}

// GetCollection returns a collection instance
func (c *MongoClient) GetCollection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// GetTransactions retrieves transactions from MongoDB
func (c *MongoClient) GetTransactions(ctx context.Context, filter bson.M, limit int64, skip int64) ([]bson.M, error) {
	collection := c.GetCollection("transactions")

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.D{{Key: "crawled_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.logger.Error("Failed to find transactions", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []bson.M
	if err := cursor.All(ctx, &transactions); err != nil {
		c.logger.Error("Failed to decode transactions", zap.Error(err))
		return nil, err
	}

	return transactions, nil
}

// GetPairwiseTransactions retrieves transactions between two specific wallets
func (c *MongoClient) GetPairwiseTransactions(ctx context.Context, walletA, walletB string, limit int64, skip int64) ([]bson.M, error) {
	collection := c.GetCollection("transactions")

	filter := bson.M{
		"$or": []bson.M{
			{
				"from": walletA,
				"to":   walletB,
			},
			{
				"from": walletB,
				"to":   walletA,
			},
		},
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.D{{Key: "crawled_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.logger.Error("Failed to find pairwise transactions",
			zap.String("walletA", walletA),
			zap.String("walletB", walletB),
			zap.Error(err),
		)
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []bson.M
	if err := cursor.All(ctx, &transactions); err != nil {
		c.logger.Error("Failed to decode pairwise transactions", zap.Error(err))
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsByWallet retrieves transactions for a specific wallet
func (c *MongoClient) GetTransactionsByWallet(ctx context.Context, walletAddress string, limit int64, skip int64) ([]bson.M, error) {
	collection := c.GetCollection("transactions")

	filter := bson.M{
		"$or": []bson.M{
			{"from": walletAddress},
			{"to": walletAddress},
		},
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.D{{Key: "crawled_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.logger.Error("Failed to find wallet transactions",
			zap.String("wallet", walletAddress),
			zap.Error(err),
		)
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []bson.M
	if err := cursor.All(ctx, &transactions); err != nil {
		c.logger.Error("Failed to decode wallet transactions", zap.Error(err))
		return nil, err
	}

	return transactions, nil
}

// GetMoneyFlowData retrieves money flow data for a wallet
func (c *MongoClient) GetMoneyFlowData(ctx context.Context, walletAddress string, flowType string, limit int64) ([]bson.M, error) {
	collection := c.GetCollection("transactions")

	var filter bson.M
	switch flowType {
	case "INBOUND":
		filter = bson.M{"to": walletAddress}
	case "OUTBOUND":
		filter = bson.M{"from": walletAddress}
	case "BOTH":
		filter = bson.M{
			"$or": []bson.M{
				{"from": walletAddress},
				{"to": walletAddress},
			},
		}
	default:
		filter = bson.M{
			"$or": []bson.M{
				{"from": walletAddress},
				{"to": walletAddress},
			},
		}
	}

	// Aggregate pipeline to group by counterparty and sum values
	pipeline := []bson.M{
		{"$match": filter},
		{
			"$addFields": bson.M{
				"counterparty": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": []interface{}{"$from", walletAddress}},
						"then": "$to",
						"else": "$from",
					},
				},
				"direction": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": []interface{}{"$from", walletAddress}},
						"then": "OUTBOUND",
						"else": "INBOUND",
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":               "$counterparty",
				"total_value":       bson.M{"$sum": bson.M{"$toDouble": "$value"}},
				"transaction_count": bson.M{"$sum": 1},
				"first_transaction": bson.M{"$min": "$crawled_at"},
				"last_transaction":  bson.M{"$max": "$crawled_at"},
				"transactions":      bson.M{"$push": "$$ROOT"},
			},
		},
		{"$sort": bson.M{"total_value": -1}},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.logger.Error("Failed to aggregate money flow data",
			zap.String("wallet", walletAddress),
			zap.String("flowType", flowType),
			zap.Error(err),
		)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.logger.Error("Failed to decode money flow data", zap.Error(err))
		return nil, err
	}

	return results, nil
}

// GetTransactionStats retrieves transaction statistics
func (c *MongoClient) GetTransactionStats(ctx context.Context, timeRange *TimeRange) (bson.M, error) {
	collection := c.GetCollection("transactions")

	matchStage := bson.M{}
	if timeRange != nil {
		matchStage["crawled_at"] = bson.M{
			"$gte": timeRange.Start,
			"$lte": timeRange.End,
		}
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{
			"$group": bson.M{
				"_id":                nil,
				"total_transactions": bson.M{"$sum": 1},
				"total_volume":       bson.M{"$sum": bson.M{"$toDouble": "$value"}},
				"unique_wallets": bson.M{
					"$addToSet": []interface{}{"$from", "$to"},
				},
				"avg_gas_price": bson.M{"$avg": bson.M{"$toDouble": "$gas_price"}},
				"avg_gas_used":  bson.M{"$avg": "$gas_used"},
			},
		},
		{
			"$addFields": bson.M{
				"unique_wallet_count": bson.M{"$size": "$unique_wallets"},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.logger.Error("Failed to get transaction stats", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var result bson.M
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			c.logger.Error("Failed to decode transaction stats", zap.Error(err))
			return nil, err
		}
	}

	return result, nil
}

// GetTopTokens retrieves the most transacted tokens
func (c *MongoClient) GetTopTokens(ctx context.Context, limit int64) ([]bson.M, error) {
	collection := c.GetCollection("token_transfers")

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":               "$token_address",
				"symbol":            bson.M{"$first": "$token_symbol"},
				"name":              bson.M{"$first": "$token_name"},
				"transaction_count": bson.M{"$sum": 1},
				"total_volume":      bson.M{"$sum": bson.M{"$toDouble": "$value"}},
				"unique_holders": bson.M{
					"$addToSet": []interface{}{"$from", "$to"},
				},
			},
		},
		{
			"$addFields": bson.M{
				"unique_holder_count": bson.M{"$size": "$unique_holders"},
			},
		},
		{"$sort": bson.M{"transaction_count": -1}},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.logger.Error("Failed to get top tokens", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []bson.M
	if err := cursor.All(ctx, &tokens); err != nil {
		c.logger.Error("Failed to decode top tokens", zap.Error(err))
		return nil, err
	}

	return tokens, nil
}

// GetRecentActivity retrieves recent blockchain activity
func (c *MongoClient) GetRecentActivity(ctx context.Context, limit int64) ([]bson.M, error) {
	collection := c.GetCollection("transactions")

	pipeline := []bson.M{
		{"$sort": bson.M{"crawled_at": -1}},
		{"$limit": limit},
		{
			"$project": bson.M{
				"hash":         1,
				"from":         1,
				"to":           1,
				"value":        1,
				"gas_used":     1,
				"gas_price":    1,
				"crawled_at":   1,
				"block_number": 1,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.logger.Error("Failed to get recent activity", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []bson.M
	if err := cursor.All(ctx, &activities); err != nil {
		c.logger.Error("Failed to decode recent activity", zap.Error(err))
		return nil, err
	}

	return activities, nil
}

// SearchTransactions searches transactions by hash or address
func (c *MongoClient) SearchTransactions(ctx context.Context, query string, limit int64) ([]bson.M, error) {
	collection := c.GetCollection("transactions")

	filter := bson.M{
		"$or": []bson.M{
			{"hash": bson.M{"$regex": query, "$options": "i"}},
			{"from": bson.M{"$regex": query, "$options": "i"}},
			{"to": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "crawled_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.logger.Error("Failed to search transactions",
			zap.String("query", query),
			zap.Error(err),
		)
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []bson.M
	if err := cursor.All(ctx, &transactions); err != nil {
		c.logger.Error("Failed to decode search results", zap.Error(err))
		return nil, err
	}

	return transactions, nil
}

// Health checks the health of the MongoDB connection
func (c *MongoClient) Health(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

// CreateIndexes creates necessary indexes for optimal performance
func (c *MongoClient) CreateIndexes(ctx context.Context) error {
	// Transaction indexes
	transactionCollection := c.GetCollection("transactions")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "from", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "to", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "crawled_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "block_number", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "from", Value: 1},
				{Key: "to", Value: 1},
				{Key: "crawled_at", Value: -1},
			},
		},
	}

	_, err := transactionCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		c.logger.Error("Failed to create transaction indexes", zap.Error(err))
		return err
	}

	c.logger.Info("MongoDB indexes created successfully")
	return nil
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
