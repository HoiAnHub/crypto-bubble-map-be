// Wallet and Transaction Types
export interface WalletNode {
  id: string;
  address: string;
  label?: string;
  balance?: string;
  balanceUSD?: number;
  transactionCount?: number;
  tags?: string[];
  color?: string;
  size?: number;
  isContract?: boolean;
  contractType?: string;
  firstSeen?: Date;
  lastActivity?: Date;
  riskScore?: number;
}

export interface TransactionLink {
  source: string;
  target: string;
  value: number;
  valueUSD?: number;
  type?: string;
  timestamp?: number;
  transactionHash?: string;
  gasUsed?: number;
  gasPrice?: string;
  blockNumber?: number;
}

export interface GraphData {
  nodes: WalletNode[];
  links: TransactionLink[];
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  value: string;
  valueUSD?: number;
  gasUsed: number;
  gasPrice: string;
  timestamp: number;
  blockNumber: number;
  status: number;
  methodId?: string;
  functionName?: string;
  tokenTransfers?: TokenTransfer[];
}

export interface TokenTransfer {
  from: string;
  to: string;
  value: string;
  tokenAddress: string;
  tokenSymbol?: string;
  tokenName?: string;
  tokenDecimals?: number;
}

export interface WalletDetails {
  address: string;
  balance: string;
  balanceUSD: number;
  transactionCount: number;
  firstTransaction?: Date;
  lastTransaction?: Date;
  isContract: boolean;
  contractType?: string;
  label?: string;
  tags: string[];
  riskScore: number;
  tokenBalances?: TokenBalance[];
}

export interface TokenBalance {
  tokenAddress: string;
  symbol: string;
  name: string;
  decimals: number;
  balance: string;
  balanceFormatted: string;
  valueUSD?: number;
}

// API Response Types
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
  timestamp: string;
}

export interface PaginatedResponse<T = any> extends ApiResponse<T> {
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// Search Types
export interface SearchResult {
  address: string;
  label?: string;
  type: 'wallet' | 'contract';
  transactionCount: number;
  balance?: string;
  relevanceScore: number;
}

// Network Analysis Types
export interface NetworkAnalysis {
  centralityScores: Record<string, number>;
  clusters: string[][];
  riskAssessment: Record<string, RiskAssessment>;
}

export interface RiskAssessment {
  score: number;
  factors: string[];
  level: 'low' | 'medium' | 'high' | 'critical';
}

// Database Types
export interface DatabaseWallet {
  id: number;
  address: string;
  balance: string;
  transaction_count: number;
  is_contract: boolean;
  contract_type?: string;
  label?: string;
  tags: string[];
  risk_score: number;
  first_seen: Date;
  last_activity: Date;
  created_at: Date;
  updated_at: Date;
}

export interface DatabaseTransaction {
  id: number;
  hash: string;
  from_address: string;
  to_address: string;
  value: string;
  gas_used: number;
  gas_price: string;
  timestamp: Date;
  block_number: number;
  status: number;
  method_id?: string;
  function_name?: string;
  created_at: Date;
}

// Cache Types
export interface CacheEntry<T = any> {
  data: T;
  timestamp: number;
  ttl: number;
}

// Error Types
export interface ApiError extends Error {
  statusCode: number;
  code?: string;
  details?: any;
}

// Request Types
export interface WalletNetworkRequest {
  address: string;
  depth?: number;
}

export interface WalletSearchRequest {
  query: string;
  limit?: number;
  offset?: number;
}

export interface TransactionHistoryRequest {
  address: string;
  limit?: number;
  offset?: number;
  startBlock?: number;
  endBlock?: number;
}
