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
	TransactionTypeMint         TransactionType = "MINT"
	TransactionTypeBurn         TransactionType = "BURN"
	TransactionTypeApprove      TransactionType = "APPROVE"
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

// TokenSummary represents a summary of token transactions
type TokenSummary struct {
	Symbol           string  `json:"symbol"`
	Volume           string  `json:"volume"`
	VolumeUSD        float64 `json:"volume_usd"`
	TransactionCount int64   `json:"transaction_count"`
}

// TransactionTypeDistribution represents the distribution of transaction types
type TransactionTypeDistribution struct {
	Transfer     int64 `json:"transfer"`
	Swap         int64 `json:"swap"`
	Mint         int64 `json:"mint"`
	Burn         int64 `json:"burn"`
	Approve      int64 `json:"approve"`
	ContractCall int64 `json:"contract_call"`
}

// TransactionUpdate represents real-time transaction updates
type TransactionUpdate struct {
	Hash            string          `json:"hash"`
	From            string          `json:"from"`
	To              string          `json:"to"`
	Value           string          `json:"value"`
	Timestamp       time.Time       `json:"timestamp"`
	TransactionType TransactionType `json:"transaction_type"`
	Detected        time.Time       `json:"detected"`
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
