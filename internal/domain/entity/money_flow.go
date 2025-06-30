package entity

import (
	"time"
)

// MoneyFlowType represents the direction of money flow
type MoneyFlowType string

const (
	MoneyFlowTypeInbound  MoneyFlowType = "INBOUND"
	MoneyFlowTypeOutbound MoneyFlowType = "OUTBOUND"
	MoneyFlowTypeBoth     MoneyFlowType = "BOTH"
)

// TransferType represents the type of transfer
type TransferType string

const (
	TransferTypeETH   TransferType = "ETH"
	TransferTypeToken TransferType = "TOKEN"
	TransferTypeBoth  TransferType = "BOTH"
)

// MoneyFlowAccount represents an account in money flow analysis
type MoneyFlowAccount struct {
	Address         string    `json:"address"`
	Label           *string   `json:"label,omitempty"`
	TotalValue      string    `json:"total_value"`
	TotalUsdValue   string    `json:"total_usd_value"`
	TransactionCount int64    `json:"transaction_count"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	RiskScore       int       `json:"risk_score"`
	Tags            []string  `json:"tags"`
	IsExchange      bool      `json:"is_exchange"`
	IsContract      bool      `json:"is_contract"`
}

// MoneyFlowTransaction represents a transaction in money flow analysis
type MoneyFlowTransaction struct {
	ID              string            `json:"id"`
	Hash            string            `json:"hash"`
	From            string            `json:"from"`
	To              string            `json:"to"`
	Value           string            `json:"value"`
	Token           *string           `json:"token,omitempty"`
	TokenSymbol     *string           `json:"token_symbol,omitempty"`
	TokenDecimals   *int              `json:"token_decimals,omitempty"`
	UsdValue        *string           `json:"usd_value,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
	BlockNumber     int64             `json:"block_number"`
	GasUsed         string            `json:"gas_used"`
	GasPrice        string            `json:"gas_price"`
	GasFee          string            `json:"gas_fee"`
	Method          *string           `json:"method,omitempty"`
	TransactionType TransactionType   `json:"transaction_type"`
	RiskLevel       RiskLevel         `json:"risk_level"`
	Status          TransactionStatus `json:"status"`
	Direction       string            `json:"direction"`
	IsInternal      bool              `json:"is_internal"`
	ContractAddress *string           `json:"contract_address,omitempty"`
}

// TopToken represents a top token in money flow analysis
type TopToken struct {
	Symbol           string `json:"symbol"`
	Volume           string `json:"volume"`
	UsdVolume        string `json:"usd_volume"`
	TransactionCount int64  `json:"transaction_count"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// MoneyFlowSummary represents summary statistics for money flow
type MoneyFlowSummary struct {
	TotalInbound         string     `json:"total_inbound"`
	TotalOutbound        string     `json:"total_outbound"`
	TotalInboundUsd      string     `json:"total_inbound_usd"`
	TotalOutboundUsd     string     `json:"total_outbound_usd"`
	UniqueCounterparties int        `json:"unique_counterparties"`
	TimeRange            TimeRange  `json:"time_range"`
	TopTokens            []TopToken `json:"top_tokens"`
}

// SankeyNode represents a node in Sankey diagram
type SankeyNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Value    string `json:"value"`
	Color    string `json:"color"`
}

// SankeyLink represents a link in Sankey diagram
type SankeyLink struct {
	Source       string                 `json:"source"`
	Target       string                 `json:"target"`
	Value        string                 `json:"value"`
	Color        string                 `json:"color"`
	Transactions []MoneyFlowTransaction `json:"transactions"`
}

// SankeyData represents data for Sankey diagram
type SankeyData struct {
	Nodes []SankeyNode `json:"nodes"`
	Links []SankeyLink `json:"links"`
}

// MoneyFlowData represents complete money flow analysis data
type MoneyFlowData struct {
	CenterAccount    MoneyFlowAccount       `json:"center_account"`
	InboundAccounts  []MoneyFlowAccount     `json:"inbound_accounts"`
	OutboundAccounts []MoneyFlowAccount     `json:"outbound_accounts"`
	Transactions     []MoneyFlowTransaction `json:"transactions"`
	Summary          MoneyFlowSummary       `json:"summary"`
	SankeyData       SankeyData             `json:"sankey_data"`
}

// MoneyFlowFilters represents filters for money flow analysis
type MoneyFlowFilters struct {
	FlowType      MoneyFlowType  `json:"flow_type"`
	TransferType  TransferType   `json:"transfer_type"`
	TopN          int            `json:"top_n"`
	TimeRange     *TimeRange     `json:"time_range,omitempty"`
	BlockRange    *BlockRange    `json:"block_range,omitempty"`
	TokenFilter   *string        `json:"token_filter,omitempty"`
	SearchQuery   *string        `json:"search_query,omitempty"`
	MinValue      *string        `json:"min_value,omitempty"`
	MaxValue      *string        `json:"max_value,omitempty"`
	RiskLevel     *RiskLevel     `json:"risk_level,omitempty"`
}

// BlockRange represents a block range
type BlockRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// PairwiseTransactionSummary represents summary of transactions between two wallets
type PairwiseTransactionSummary struct {
	WalletA            string            `json:"wallet_a"`
	WalletB            string            `json:"wallet_b"`
	TotalTransactions  int64             `json:"total_transactions"`
	TotalVolume        string            `json:"total_volume"`
	TotalVolumeUSD     string            `json:"total_volume_usd"`
	FirstTransaction   time.Time         `json:"first_transaction"`
	LastTransaction    time.Time         `json:"last_transaction"`
	TopTokens          []TopToken        `json:"top_tokens"`
	RiskDistribution   RiskDistribution  `json:"risk_distribution"`
	TransactionTypes   map[string]int64  `json:"transaction_types"`
}

// PairwiseTransactionResult represents result of pairwise transaction query
type PairwiseTransactionResult struct {
	Summary      PairwiseTransactionSummary `json:"summary"`
	Transactions []Transaction              `json:"transactions"`
	TotalCount   int64                      `json:"total_count"`
}

// RiskDistribution represents distribution of risk levels
type RiskDistribution struct {
	Low      int64 `json:"low"`
	Medium   int64 `json:"medium"`
	High     int64 `json:"high"`
	Critical int64 `json:"critical"`
}

// TransactionFilters represents filters for transaction queries
type TransactionFilters struct {
	TimeRange       *TimeRange       `json:"time_range,omitempty"`
	BlockRange      *BlockRange      `json:"block_range,omitempty"`
	TokenFilter     *string          `json:"token_filter,omitempty"`
	MinValue        *string          `json:"min_value,omitempty"`
	MaxValue        *string          `json:"max_value,omitempty"`
	TransactionType *TransactionType `json:"transaction_type,omitempty"`
	RiskLevel       *RiskLevel       `json:"risk_level,omitempty"`
}
