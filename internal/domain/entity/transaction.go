package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionType represents different types of transactions
type TransactionType string

const (
	TransactionTypeTransfer     TransactionType = "TRANSFER"
	TransactionTypeSwap         TransactionType = "SWAP"
	TransactionTypeDeposit      TransactionType = "DEPOSIT"
	TransactionTypeWithdraw     TransactionType = "WITHDRAW"
	TransactionTypeContractCall TransactionType = "CONTRACT_CALL"
	TransactionTypeNFTTransfer  TransactionType = "NFT_TRANSFER"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
	TransactionStatusPending TransactionStatus = "PENDING"
)

// TransactionDirection represents the direction of a transaction relative to a wallet
type TransactionDirection string

const (
	TransactionDirectionIncoming TransactionDirection = "INCOMING"
	TransactionDirectionOutgoing TransactionDirection = "OUTGOING"
)

// Transaction represents an Ethereum transaction from MongoDB
type Transaction struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Hash                 string             `bson:"hash" json:"hash"`
	BlockHash            string             `bson:"block_hash" json:"block_hash"`
	BlockNumber          string             `bson:"block_number" json:"block_number"`
	TransactionIndex     uint               `bson:"transaction_index" json:"transaction_index"`
	From                 string             `bson:"from" json:"from"`
	To                   *string            `bson:"to" json:"to"`
	Value                string             `bson:"value" json:"value"`
	Gas                  uint64             `bson:"gas" json:"gas"`
	GasPrice             string             `bson:"gas_price" json:"gas_price"`
	GasUsed              uint64             `bson:"gas_used" json:"gas_used"`
	CumulativeGasUsed    uint64             `bson:"cumulative_gas_used" json:"cumulative_gas_used"`
	Data                 string             `bson:"data" json:"data"`
	Nonce                uint64             `bson:"nonce" json:"nonce"`
	Status               uint64             `bson:"status" json:"status"`
	MaxFeePerGas         *string            `bson:"max_fee_per_gas,omitempty" json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas *string            `bson:"max_priority_fee_per_gas,omitempty" json:"max_priority_fee_per_gas,omitempty"`
	ContractAddress      *string            `bson:"contract_address,omitempty" json:"contract_address,omitempty"`
	CrawledAt            time.Time          `bson:"crawled_at" json:"crawled_at"`
	Network              string             `bson:"network" json:"network"`
	ProcessedAt          *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"`

	// Enhanced fields for GraphQL
	Timestamp       time.Time         `json:"timestamp"`
	GasFee          string            `json:"gas_fee"`
	TransactionType TransactionType   `json:"transaction_type"`
	Method          *string           `json:"method,omitempty"`
	RiskLevel       RiskLevel         `json:"risk_level"`
	TxStatus        TransactionStatus `json:"tx_status"`
}

// PairwiseTransaction represents a transaction between two specific wallets
type PairwiseTransaction struct {
	ID              string               `json:"id"`
	Hash            string               `json:"hash"`
	From            string               `json:"from"`
	To              string               `json:"to"`
	Value           string               `json:"value"`
	Token           string               `json:"token"`
	TokenSymbol     string               `json:"token_symbol"`
	TokenDecimals   *int                 `json:"token_decimals,omitempty"`
	USDValue        *float64             `json:"usd_value,omitempty"`
	Timestamp       time.Time            `json:"timestamp"`
	BlockNumber     string               `json:"block_number"`
	GasUsed         string               `json:"gas_used"`
	GasPrice        string               `json:"gas_price"`
	GasFee          string               `json:"gas_fee"`
	TransactionType TransactionType      `json:"transaction_type"`
	Method          *string              `json:"method,omitempty"`
	RiskLevel       RiskLevel            `json:"risk_level"`
	RiskFactors     []string             `json:"risk_factors"`
	Status          TransactionStatus    `json:"status"`
	Direction       TransactionDirection `json:"direction"`
	IsInternal      bool                 `json:"is_internal"`
	ContractAddress *string              `json:"contract_address,omitempty"`
	Logs            []TransactionLog     `json:"logs,omitempty"`
}

// TransactionLog represents a transaction log entry
type TransactionLog struct {
	Address string      `json:"address"`
	Topics  []string    `json:"topics"`
	Data    string      `json:"data"`
	Decoded *DecodedLog `json:"decoded,omitempty"`
}

// DecodedLog represents a decoded transaction log
type DecodedLog struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// PairwiseTransactionSummary represents a summary of transactions between two wallets
type PairwiseTransactionSummary struct {
	WalletA           string                      `json:"wallet_a"`
	WalletB           string                      `json:"wallet_b"`
	TotalTransactions int64                       `json:"total_transactions"`
	TotalVolume       string                      `json:"total_volume"`
	TotalVolumeUSD    float64                     `json:"total_volume_usd"`
	FirstTransaction  time.Time                   `json:"first_transaction"`
	LastTransaction   time.Time                   `json:"last_transaction"`
	TopTokens         []TokenSummary              `json:"top_tokens"`
	RiskDistribution  RiskDistribution            `json:"risk_distribution"`
	TransactionTypes  TransactionTypeDistribution `json:"transaction_types"`
}

// TokenSummary represents a summary of token transactions
type TokenSummary struct {
	Symbol           string  `json:"symbol"`
	Volume           string  `json:"volume"`
	VolumeUSD        float64 `json:"volume_usd"`
	TransactionCount int64   `json:"transaction_count"`
}

// RiskDistribution represents the distribution of risk levels
type RiskDistribution struct {
	Low      int64 `json:"low"`
	Medium   int64 `json:"medium"`
	High     int64 `json:"high"`
	Critical int64 `json:"critical"`
}

// TransactionTypeDistribution represents the distribution of transaction types
type TransactionTypeDistribution struct {
	Transfer     int64 `json:"transfer"`
	Swap         int64 `json:"swap"`
	Deposit      int64 `json:"deposit"`
	Withdraw     int64 `json:"withdraw"`
	ContractCall int64 `json:"contract_call"`
}

// MoneyFlowTransaction represents a transaction in money flow analysis
type MoneyFlowTransaction struct {
	ID              string          `json:"id"`
	Hash            string          `json:"hash"`
	From            string          `json:"from"`
	To              string          `json:"to"`
	Value           string          `json:"value"`
	Token           *string         `json:"token,omitempty"`
	TokenSymbol     *string         `json:"token_symbol,omitempty"`
	USDValue        *float64        `json:"usd_value,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
	BlockNumber     string          `json:"block_number"`
	GasUsed         string          `json:"gas_used"`
	GasPrice        string          `json:"gas_price"`
	GasFee          string          `json:"gas_fee"`
	Method          *string         `json:"method,omitempty"`
	TransactionType TransactionType `json:"transaction_type"`
	RiskLevel       RiskLevel       `json:"risk_level"`
}

// MoneyFlowAccount represents an account in money flow analysis
type MoneyFlowAccount struct {
	Address          string    `json:"address"`
	Label            *string   `json:"label,omitempty"`
	TotalValue       string    `json:"total_value"`
	TotalUSDValue    float64   `json:"total_usd_value"`
	TransactionCount int64     `json:"transaction_count"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
	RiskScore        float64   `json:"risk_score"`
	Tags             []string  `json:"tags"`
	IsExchange       bool      `json:"is_exchange"`
	IsContract       bool      `json:"is_contract"`
}

// MoneyFlowData represents comprehensive money flow analysis data
type MoneyFlowData struct {
	CenterAccount    MoneyFlowAccount       `json:"center_account"`
	InboundAccounts  []MoneyFlowAccount     `json:"inbound_accounts"`
	OutboundAccounts []MoneyFlowAccount     `json:"outbound_accounts"`
	Transactions     []MoneyFlowTransaction `json:"transactions"`
	Summary          MoneyFlowSummary       `json:"summary"`
	SankeyData       SankeyData             `json:"sankey_data"`
}

// MoneyFlowSummary represents a summary of money flow analysis
type MoneyFlowSummary struct {
	TotalInbound         string         `json:"total_inbound"`
	TotalOutbound        string         `json:"total_outbound"`
	TotalInboundUSD      float64        `json:"total_inbound_usd"`
	TotalOutboundUSD     float64        `json:"total_outbound_usd"`
	UniqueCounterparties int64          `json:"unique_counterparties"`
	TimeRange            TimeRange      `json:"time_range"`
	TopTokens            []TokenSummary `json:"top_tokens"`
}

// SankeyData represents data for Sankey diagram visualization
type SankeyData struct {
	Nodes []SankeyNode `json:"nodes"`
	Links []SankeyLink `json:"links"`
}

// SankeyNode represents a node in Sankey diagram
type SankeyNode struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Category SankeyNodeCategory `json:"category"`
	Value    string             `json:"value"`
	Color    string             `json:"color"`
}

// SankeyNodeCategory represents the category of a Sankey node
type SankeyNodeCategory string

const (
	SankeyNodeCategorySource SankeyNodeCategory = "SOURCE"
	SankeyNodeCategoryCenter SankeyNodeCategory = "CENTER"
	SankeyNodeCategoryTarget SankeyNodeCategory = "TARGET"
)

// SankeyLink represents a link in Sankey diagram
type SankeyLink struct {
	Source       string                 `json:"source"`
	Target       string                 `json:"target"`
	Value        string                 `json:"value"`
	Color        string                 `json:"color"`
	Transactions []MoneyFlowTransaction `json:"transactions"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TransactionFilters represents filters for transaction queries
type TransactionFilters struct {
	Direction       *TransactionDirection `json:"direction,omitempty"`
	TokenFilter     *string               `json:"token_filter,omitempty"`
	RiskLevel       *RiskLevel            `json:"risk_level,omitempty"`
	TransactionType *TransactionType      `json:"transaction_type,omitempty"`
	TimeRange       *TimeRange            `json:"time_range,omitempty"`
	MinValue        *string               `json:"min_value,omitempty"`
	MaxValue        *string               `json:"max_value,omitempty"`
}

// MoneyFlowFilters represents filters for money flow analysis
type MoneyFlowFilters struct {
	FlowType     FlowType     `json:"flow_type"`
	TransferType TransferType `json:"transfer_type"`
	TopN         int          `json:"top_n"`
	TimeRange    *TimeRange   `json:"time_range,omitempty"`
	BlockRange   *BlockRange  `json:"block_range,omitempty"`
	TokenFilter  *string      `json:"token_filter,omitempty"`
	SearchQuery  *string      `json:"search_query,omitempty"`
	MinValue     *string      `json:"min_value,omitempty"`
	MaxValue     *string      `json:"max_value,omitempty"`
	RiskLevel    *RiskLevel   `json:"risk_level,omitempty"`
}

// FlowType represents the type of money flow
type FlowType string

const (
	FlowTypeInbound  FlowType = "INBOUND"
	FlowTypeOutbound FlowType = "OUTBOUND"
	FlowTypeBoth     FlowType = "BOTH"
)

// TransferType represents the type of transfer
type TransferType string

const (
	TransferTypeETH   TransferType = "ETH"
	TransferTypeToken TransferType = "TOKEN"
	TransferTypeBoth  TransferType = "BOTH"
)

// BlockRange represents a range of blocks
type BlockRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TransactionUpdate represents real-time transaction updates
type TransactionUpdate struct {
	Transaction   Transaction `json:"transaction"`
	WalletAddress string      `json:"wallet_address"`
}

// PairwiseTransactionResult represents paginated pairwise transaction results
type PairwiseTransactionResult struct {
	Transactions []PairwiseTransaction      `json:"transactions"`
	Summary      PairwiseTransactionSummary `json:"summary"`
	HasMore      bool                       `json:"has_more"`
	Total        int64                      `json:"total"`
}

// Helper methods
func (t *Transaction) GetGasFee() string {
	// Calculate gas fee from gas used and gas price
	// This would be implemented based on the specific calculation logic
	return t.GasFee
}

func (t *Transaction) GetTransactionStatus() TransactionStatus {
	if t.Status == 1 {
		return TransactionStatusSuccess
	}
	return TransactionStatusFailed
}

func (t *Transaction) GetTimestamp() time.Time {
	if !t.Timestamp.IsZero() {
		return t.Timestamp
	}
	return t.CrawledAt
}
