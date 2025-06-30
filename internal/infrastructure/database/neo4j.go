package database

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/config"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

// Neo4jClient wraps the Neo4j driver
type Neo4jClient struct {
	driver neo4j.DriverWithContext
	logger *zap.Logger
	config *config.Neo4jConfig
}

// NewNeo4jClient creates a new Neo4j client
func NewNeo4jClient(cfg *config.Neo4jConfig, logger *zap.Logger) (*Neo4jClient, error) {
	// Configure authentication
	auth := neo4j.BasicAuth(cfg.Username, cfg.Password, "")

	// Configure driver
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		auth,
		func(config *neo4j.Config) {
			config.MaxConnectionPoolSize = cfg.MaxConnectionPoolSize
			config.ConnectionAcquisitionTimeout = cfg.ConnectionTimeout
			config.MaxTransactionRetryTime = cfg.MaxTransactionRetryTime
			config.Log = neo4j.ConsoleLogger(neo4j.INFO)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	client := &Neo4jClient{
		driver: driver,
		logger: logger,
		config: cfg,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	logger.Info("Neo4j client initialized successfully",
		zap.String("uri", cfg.URI),
		zap.String("database", cfg.Database),
	)

	return client, nil
}

// VerifyConnectivity verifies the connection to Neo4j
func (c *Neo4jClient) VerifyConnectivity(ctx context.Context) error {
	return c.driver.VerifyConnectivity(ctx)
}

// Close closes the Neo4j driver
func (c *Neo4jClient) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

// ExecuteRead executes a read transaction
func (c *Neo4jClient) ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.config.Database,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, work, configurers...)
}

// ExecuteWrite executes a write transaction
func (c *Neo4jClient) ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.config.Database,
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work, configurers...)
}

// NewSession creates a new Neo4j session
func (c *Neo4jClient) NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext {
	if config.DatabaseName == "" {
		config.DatabaseName = c.config.Database
	}
	return c.driver.NewSession(ctx, config)
}

// GetWalletNetwork retrieves wallet network data from Neo4j
func (c *Neo4jClient) GetWalletNetwork(ctx context.Context, address string, depth int) ([]map[string]interface{}, error) {
	query := `
		MATCH path = (center:Wallet {address: $address})-[r:TRANSACTED_WITH*1..$depth]-(connected:Wallet)
		WITH center, connected, r,
			 reduce(totalValue = 0, rel in r | totalValue + rel.total_value) as pathValue,
			 reduce(totalTxs = 0, rel in r | totalTxs + rel.tx_count) as pathTxCount
		RETURN DISTINCT
			center.address as center_address,
			center.node_type as center_type,
			center.risk_level as center_risk,
			center.total_transactions as center_tx_count,
			center.balance as center_balance,
			connected.address as connected_address,
			connected.node_type as connected_type,
			connected.risk_level as connected_risk,
			connected.total_transactions as connected_tx_count,
			connected.balance as connected_balance,
			pathValue as connection_value,
			pathTxCount as connection_tx_count
		ORDER BY pathValue DESC
		LIMIT 1000
	`

	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
			"depth":   depth,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}

		var networkData []map[string]interface{}
		for _, record := range records {
			networkData = append(networkData, record.AsMap())
		}

		return networkData, nil
	})

	if err != nil {
		c.logger.Error("Failed to get wallet network",
			zap.String("address", address),
			zap.Int("depth", depth),
			zap.Error(err),
		)
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}

// GetWalletInfo retrieves detailed information about a wallet
func (c *Neo4jClient) GetWalletInfo(ctx context.Context, address string) (map[string]interface{}, error) {
	query := `
		MATCH (w:Wallet {address: $address})
		OPTIONAL MATCH (w)-[r:TRANSACTED_WITH]-(connected:Wallet)
		WITH w, count(connected) as connection_count,
			 sum(r.total_value) as total_volume,
			 sum(r.tx_count) as total_transactions
		RETURN w.address as address,
			   w.node_type as wallet_type,
			   w.risk_level as risk_level,
			   w.confidence_score as confidence_score,
			   w.total_transactions as transaction_count,
			   w.balance as balance,
			   w.first_seen as first_seen,
			   w.last_seen as last_seen,
			   w.tags as tags,
			   w.associated_exchanges as associated_exchanges,
			   w.associated_protocols as associated_protocols,
			   connection_count,
			   total_volume,
			   total_transactions
	`

	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		return record.AsMap(), nil
	})

	if err != nil {
		c.logger.Error("Failed to get wallet info",
			zap.String("address", address),
			zap.Error(err),
		)
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

// GetWalletRankings retrieves wallet rankings from Neo4j
func (c *Neo4jClient) GetWalletRankings(ctx context.Context, category string, limit int, offset int) ([]map[string]interface{}, error) {
	var orderBy string
	switch category {
	case "QUALITY":
		orderBy = "w.confidence_score DESC"
	case "VOLUME":
		orderBy = "w.total_sent + w.total_received DESC"
	case "ACTIVITY":
		orderBy = "w.total_transactions DESC"
	case "NETWORK":
		orderBy = "connection_count DESC"
	case "SAFETY":
		orderBy = "w.risk_level ASC, w.confidence_score DESC"
	default:
		orderBy = "w.confidence_score DESC"
	}

	query := fmt.Sprintf(`
		MATCH (w:Wallet)
		WHERE w.node_type <> 'BLACKLISTED'
		OPTIONAL MATCH (w)-[:TRANSACTED_WITH]-(connected:Wallet)
		WITH w, count(connected) as connection_count
		RETURN w.address as address,
			   w.node_type as wallet_type,
			   w.risk_level as risk_level,
			   w.confidence_score as confidence_score,
			   w.total_transactions as transaction_count,
			   w.balance as balance,
			   w.first_seen as first_seen,
			   w.last_seen as last_seen,
			   w.total_sent as total_sent,
			   w.total_received as total_received,
			   connection_count
		ORDER BY %s
		SKIP $offset
		LIMIT $limit
	`, orderBy)

	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}

		var rankings []map[string]interface{}
		for i, record := range records {
			data := record.AsMap()
			data["rank"] = offset + i + 1
			rankings = append(rankings, data)
		}

		return rankings, nil
	})

	if err != nil {
		c.logger.Error("Failed to get wallet rankings",
			zap.String("category", category),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
			zap.Error(err),
		)
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}

// GetNetworkStats retrieves network statistics from Neo4j
func (c *Neo4jClient) GetNetworkStats(ctx context.Context, networkID string) (map[string]interface{}, error) {
	query := `
		MATCH (w:Wallet)
		WHERE w.network = $networkId OR $networkId IS NULL
		WITH w.node_type as wallet_type, w.risk_level as risk_level, count(*) as count
		RETURN collect({type: wallet_type, risk: risk_level, count: count}) as stats,
			   sum(count) as total_wallets
	`

	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"networkId": networkID,
		})
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		return record.AsMap(), nil
	})

	if err != nil {
		c.logger.Error("Failed to get network stats",
			zap.String("networkId", networkID),
			zap.Error(err),
		)
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

// SearchWallets searches for wallets by address or label
func (c *Neo4jClient) SearchWallets(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	cypherQuery := `
		MATCH (w:Wallet)
		WHERE w.address CONTAINS $query
		   OR any(tag in w.tags WHERE tag CONTAINS $query)
		   OR w.label CONTAINS $query
		RETURN w.address as address,
			   w.label as label,
			   w.node_type as wallet_type,
			   w.risk_level as risk_level,
			   w.confidence_score as confidence_score,
			   w.total_transactions as transaction_count,
			   w.balance as balance,
			   w.tags as tags
		ORDER BY
			CASE
				WHEN w.address = $query THEN 1
				WHEN w.address STARTS WITH $query THEN 2
				WHEN w.label = $query THEN 3
				WHEN w.label STARTS WITH $query THEN 4
				ELSE 5
			END,
			w.confidence_score DESC
		LIMIT $limit
	`

	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, cypherQuery, map[string]interface{}{
			"query": query,
			"limit": limit,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}

		var wallets []map[string]interface{}
		for _, record := range records {
			wallets = append(wallets, record.AsMap())
		}

		return wallets, nil
	})

	if err != nil {
		c.logger.Error("Failed to search wallets",
			zap.String("query", query),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}

// Health checks the health of the Neo4j connection
func (c *Neo4jClient) Health(ctx context.Context) error {
	_, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, "RETURN 1 as health", nil)
		if err != nil {
			return nil, err
		}

		_, err = result.Single(ctx)
		return nil, err
	})

	return err
}
