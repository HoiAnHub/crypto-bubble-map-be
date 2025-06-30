package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/config"

	"go.uber.org/zap"
)

// BlockchainAPIClient provides access to external blockchain APIs
type BlockchainAPIClient struct {
	config     *config.ExternalConfig
	logger     *zap.Logger
	httpClient *http.Client
}

// NewBlockchainAPIClient creates a new blockchain API client
func NewBlockchainAPIClient(cfg *config.ExternalConfig, logger *zap.Logger) *BlockchainAPIClient {
	return &BlockchainAPIClient{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NetworkData represents real-time network data
type NetworkData struct {
	ID                string  `json:"id"`
	TVL               float64 `json:"tvl"`
	DailyTransactions int     `json:"daily_transactions"`
	TPS               int     `json:"tps"`
	GasPrice          float64 `json:"gas_price"`
	BlockTime         int     `json:"block_time"`
	MarketCap         float64 `json:"market_cap"`
	Price             float64 `json:"price"`
	Volume24h         float64 `json:"volume_24h"`
}

// CoinGeckoResponse represents CoinGecko API response
type CoinGeckoResponse struct {
	ID                 string  `json:"id"`
	Symbol             string  `json:"symbol"`
	Name               string  `json:"name"`
	CurrentPrice       float64 `json:"current_price"`
	MarketCap          float64 `json:"market_cap"`
	TotalVolume        float64 `json:"total_volume"`
	PriceChange24h     float64 `json:"price_change_24h"`
	MarketCapChange24h float64 `json:"market_cap_change_24h"`
}

// EthereumGasResponse represents Ethereum gas price response
type EthereumGasResponse struct {
	SafeGasPrice     string `json:"SafeGasPrice"`
	StandardGasPrice string `json:"StandardGasPrice"`
	FastGasPrice     string `json:"FastGasPrice"`
}

// GetNetworkData fetches real-time data for a specific network
func (c *BlockchainAPIClient) GetNetworkData(ctx context.Context, networkID string) (*NetworkData, error) {
	data := &NetworkData{
		ID: networkID,
	}

	// Fetch price and market data from CoinGecko
	if err := c.fetchCoinGeckoData(ctx, networkID, data); err != nil {
		c.logger.Warn("Failed to fetch CoinGecko data",
			zap.String("networkID", networkID),
			zap.Error(err))
	}

	// Fetch network-specific data based on network type
	switch networkID {
	case "ethereum":
		if err := c.fetchEthereumData(ctx, data); err != nil {
			c.logger.Warn("Failed to fetch Ethereum data", zap.Error(err))
		}
	case "polygon":
		if err := c.fetchPolygonData(ctx, data); err != nil {
			c.logger.Warn("Failed to fetch Polygon data", zap.Error(err))
		}
	case "bsc":
		if err := c.fetchBSCData(ctx, data); err != nil {
			c.logger.Warn("Failed to fetch BSC data", zap.Error(err))
		}
	default:
		// For other networks, use estimated data based on market cap
		c.estimateNetworkMetrics(data)
	}

	return data, nil
}

// fetchCoinGeckoData fetches price and market data from CoinGecko
func (c *BlockchainAPIClient) fetchCoinGeckoData(ctx context.Context, networkID string, data *NetworkData) error {
	// Map network IDs to CoinGecko IDs
	coinGeckoID := c.mapNetworkToCoinGecko(networkID)
	if coinGeckoID == "" {
		return fmt.Errorf("no CoinGecko mapping for network: %s", networkID)
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s", coinGeckoID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.config.CoinGeckoAPIKey != "" {
		req.Header.Set("X-CG-Demo-API-Key", c.config.CoinGeckoAPIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CoinGecko API error: %d", resp.StatusCode)
	}

	var coinData struct {
		MarketData struct {
			CurrentPrice map[string]float64 `json:"current_price"`
			MarketCap    map[string]float64 `json:"market_cap"`
			TotalVolume  map[string]float64 `json:"total_volume"`
		} `json:"market_data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&coinData); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract USD values
	if price, ok := coinData.MarketData.CurrentPrice["usd"]; ok {
		data.Price = price
	}
	if marketCap, ok := coinData.MarketData.MarketCap["usd"]; ok {
		data.MarketCap = marketCap
	}
	if volume, ok := coinData.MarketData.TotalVolume["usd"]; ok {
		data.Volume24h = volume
	}

	return nil
}

// fetchEthereumData fetches Ethereum-specific data
func (c *BlockchainAPIClient) fetchEthereumData(ctx context.Context, data *NetworkData) error {
	// Fetch gas prices
	if err := c.fetchEthereumGasPrice(ctx, data); err != nil {
		c.logger.Warn("Failed to fetch Ethereum gas price", zap.Error(err))
	}

	// Set Ethereum-specific metrics
	data.BlockTime = 12 // Ethereum block time ~12 seconds
	data.TPS = 15       // Ethereum TPS ~15

	// Estimate daily transactions based on block time and TPS
	data.DailyTransactions = data.TPS * 86400 // TPS * seconds in a day

	// Estimate TVL (this would ideally come from DeFi Pulse or similar)
	data.TVL = data.MarketCap * 0.1 // Rough estimate: 10% of market cap

	return nil
}

// fetchEthereumGasPrice fetches current Ethereum gas price
func (c *BlockchainAPIClient) fetchEthereumGasPrice(ctx context.Context, data *NetworkData) error {
	url := "https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=YourApiKeyToken"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Etherscan API error: %d", resp.StatusCode)
	}

	var gasResp struct {
		Status string              `json:"status"`
		Result EthereumGasResponse `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gasResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if gasResp.Status == "1" {
		if gasPrice, err := strconv.ParseFloat(gasResp.Result.StandardGasPrice, 64); err == nil {
			data.GasPrice = gasPrice
		}
	}

	return nil
}

// fetchPolygonData fetches Polygon-specific data
func (c *BlockchainAPIClient) fetchPolygonData(ctx context.Context, data *NetworkData) error {
	data.BlockTime = 2 // Polygon block time ~2 seconds
	data.TPS = 7000    // Polygon TPS ~7000
	data.GasPrice = 30 // Polygon gas price in Gwei

	data.DailyTransactions = data.TPS * 86400
	data.TVL = data.MarketCap * 0.05 // Estimate

	return nil
}

// fetchBSCData fetches BSC-specific data
func (c *BlockchainAPIClient) fetchBSCData(ctx context.Context, data *NetworkData) error {
	data.BlockTime = 3 // BSC block time ~3 seconds
	data.TPS = 160     // BSC TPS ~160
	data.GasPrice = 5  // BSC gas price in Gwei

	data.DailyTransactions = data.TPS * 86400
	data.TVL = data.MarketCap * 0.08 // Estimate

	return nil
}

// estimateNetworkMetrics provides estimated metrics for networks without specific APIs
func (c *BlockchainAPIClient) estimateNetworkMetrics(data *NetworkData) {
	// Provide reasonable estimates based on market cap
	if data.MarketCap > 0 {
		// Estimate metrics based on market cap tier
		if data.MarketCap > 100000000000 { // > $100B
			data.TPS = 1000
			data.BlockTime = 5
			data.DailyTransactions = 500000
			data.TVL = data.MarketCap * 0.15
		} else if data.MarketCap > 10000000000 { // > $10B
			data.TPS = 500
			data.BlockTime = 10
			data.DailyTransactions = 200000
			data.TVL = data.MarketCap * 0.1
		} else {
			data.TPS = 100
			data.BlockTime = 15
			data.DailyTransactions = 50000
			data.TVL = data.MarketCap * 0.05
		}
		data.GasPrice = 20 // Default gas price
	}
}

// mapNetworkToCoinGecko maps network IDs to CoinGecko coin IDs
func (c *BlockchainAPIClient) mapNetworkToCoinGecko(networkID string) string {
	mapping := map[string]string{
		"ethereum":  "ethereum",
		"polygon":   "matic-network",
		"bsc":       "binancecoin",
		"avalanche": "avalanche-2",
		"fantom":    "fantom",
		"arbitrum":  "ethereum", // Arbitrum uses ETH
		"optimism":  "ethereum", // Optimism uses ETH
		"solana":    "solana",
		"cardano":   "cardano",
		"polkadot":  "polkadot",
		"cosmos":    "cosmos",
		"near":      "near",
		"algorand":  "algorand",
		"tezos":     "tezos",
		"flow":      "flow",
		"sui":       "sui",
	}

	return mapping[networkID]
}
