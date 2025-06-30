// Neo4j seed data for Crypto Bubble Map Backend
// Run with: cypher-shell -u neo4j -p password < neo4j_seed.cypher

// Create constraints and indexes
CREATE CONSTRAINT wallet_address_unique IF NOT EXISTS FOR (w:Wallet) REQUIRE w.address IS UNIQUE;
CREATE CONSTRAINT network_id_unique IF NOT EXISTS FOR (n:Network) REQUIRE n.id IS UNIQUE;

CREATE INDEX wallet_risk_score IF NOT EXISTS FOR (w:Wallet) ON (w.risk_score);
CREATE INDEX wallet_balance IF NOT EXISTS FOR (w:Wallet) ON (w.balance);
CREATE INDEX wallet_type IF NOT EXISTS FOR (w:Wallet) ON (w.type);
CREATE INDEX wallet_network IF NOT EXISTS FOR (w:Wallet) ON (w.network_id);
CREATE INDEX wallet_last_activity IF NOT EXISTS FOR (w:Wallet) ON (w.last_activity);

CREATE INDEX transaction_timestamp IF NOT EXISTS FOR (t:Transaction) ON (t.timestamp);
CREATE INDEX transaction_value IF NOT EXISTS FOR (t:Transaction) ON (t.value_usd);
CREATE INDEX transaction_risk IF NOT EXISTS FOR (t:Transaction) ON (t.risk_score);

// Create sample networks
CREATE (eth:Network {
  id: 'ethereum',
  name: 'Ethereum',
  symbol: 'ETH',
  category: 'layer1',
  is_active: true,
  created_at: datetime()
});

CREATE (poly:Network {
  id: 'polygon',
  name: 'Polygon',
  symbol: 'MATIC',
  category: 'layer2',
  is_active: true,
  created_at: datetime()
});

CREATE (bsc:Network {
  id: 'bsc',
  name: 'BNB Smart Chain',
  symbol: 'BNB',
  category: 'layer1',
  is_active: true,
  created_at: datetime()
});

CREATE (arb:Network {
  id: 'arbitrum',
  name: 'Arbitrum',
  symbol: 'ETH',
  category: 'layer2',
  is_active: true,
  created_at: datetime()
});

// Create sample wallets with different types and risk profiles
CREATE (w1:Wallet {
  address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1',
  network_id: 'ethereum',
  type: 'whale',
  balance: '15000.5',
  balance_usd: 45000000,
  risk_score: 75.5,
  quality_score: 85.2,
  transaction_count: 1250,
  first_seen: datetime('2020-01-15T10:30:00Z'),
  last_activity: datetime('2024-01-15T10:30:00Z'),
  is_flagged: true,
  tags: ['high-value', 'suspicious-activity'],
  metadata: {
    ens_name: 'whale.eth',
    labels: ['DeFi Trader', 'High Volume'],
    social_profiles: {
      twitter: '@cryptowhale'
    }
  },
  created_at: datetime()
});

CREATE (w2:Wallet {
  address: '0x8ba1f109551bD432803012645Hac136c22C501e',
  network_id: 'ethereum',
  type: 'exchange',
  balance: '250000.0',
  balance_usd: 750000000,
  risk_score: 25.0,
  quality_score: 95.8,
  transaction_count: 50000,
  first_seen: datetime('2019-05-20T08:00:00Z'),
  last_activity: datetime('2024-01-15T12:45:00Z'),
  is_flagged: false,
  tags: ['exchange', 'high-volume', 'verified'],
  metadata: {
    exchange_name: 'Binance',
    labels: ['Hot Wallet', 'Exchange'],
    verification_status: 'verified'
  },
  created_at: datetime()
});

CREATE (w3:Wallet {
  address: '0x9876543210fedcba9876543210fedcba98765432',
  network_id: 'polygon',
  type: 'defi',
  balance: '5000.25',
  balance_usd: 12500000,
  risk_score: 35.2,
  quality_score: 78.5,
  transaction_count: 850,
  first_seen: datetime('2021-03-10T14:20:00Z'),
  last_activity: datetime('2024-01-14T16:20:00Z'),
  is_flagged: false,
  tags: ['defi', 'liquidity-provider'],
  metadata: {
    protocols: ['Uniswap', 'Aave', 'Compound'],
    labels: ['DeFi Power User']
  },
  created_at: datetime()
});

CREATE (w4:Wallet {
  address: '0x1234567890abcdef1234567890abcdef12345678',
  network_id: 'ethereum',
  type: 'contract',
  balance: '0.0',
  balance_usd: 0,
  risk_score: 90.0,
  quality_score: 15.0,
  transaction_count: 10000,
  first_seen: datetime('2022-01-01T00:00:00Z'),
  last_activity: datetime('2024-01-10T09:15:00Z'),
  is_flagged: true,
  tags: ['blacklisted', 'malicious', 'ransomware'],
  metadata: {
    contract_type: 'malicious',
    labels: ['Ransomware', 'Blacklisted'],
    blacklist_reason: 'Ransomware payments'
  },
  created_at: datetime()
});

CREATE (w5:Wallet {
  address: '0xabcdef1234567890abcdef1234567890abcdef12',
  network_id: 'bsc',
  type: 'regular',
  balance: '100.75',
  balance_usd: 25000,
  risk_score: 15.5,
  quality_score: 92.0,
  transaction_count: 125,
  first_seen: datetime('2023-06-15T11:30:00Z'),
  last_activity: datetime('2024-01-12T14:10:00Z'),
  is_flagged: false,
  tags: ['retail', 'low-risk'],
  metadata: {
    labels: ['Retail Investor'],
    activity_pattern: 'regular'
  },
  created_at: datetime()
});

// Create relationships between wallets and networks
MATCH (w:Wallet), (n:Network)
WHERE w.network_id = n.id
CREATE (w)-[:ON_NETWORK]->(n);

// Create sample transactions between wallets
CREATE (t1:Transaction {
  hash: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
  from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1',
  to_address: '0x8ba1f109551bD432803012645Hac136c22C501e',
  network_id: 'ethereum',
  value: '1500.0',
  value_usd: 4500000,
  gas_used: 21000,
  gas_price: '20',
  block_number: 18500000,
  timestamp: datetime('2024-01-15T10:30:00Z'),
  risk_score: 75.0,
  is_flagged: true,
  transaction_type: 'transfer',
  created_at: datetime()
});

CREATE (t2:Transaction {
  hash: '0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890',
  from_address: '0x9876543210fedcba9876543210fedcba98765432',
  to_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1',
  network_id: 'polygon',
  value: '50.0',
  value_usd: 50000,
  gas_used: 21000,
  gas_price: '30',
  block_number: 52000000,
  timestamp: datetime('2024-01-14T16:20:00Z'),
  risk_score: 25.0,
  is_flagged: false,
  transaction_type: 'transfer',
  created_at: datetime()
});

CREATE (t3:Transaction {
  hash: '0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321',
  from_address: '0x8ba1f109551bD432803012645Hac136c22C501e',
  to_address: '0x1234567890abcdef1234567890abcdef12345678',
  network_id: 'ethereum',
  value: '0.1',
  value_usd: 300,
  gas_used: 21000,
  gas_price: '25',
  block_number: 18499500,
  timestamp: datetime('2024-01-14T15:45:00Z'),
  risk_score: 95.0,
  is_flagged: true,
  transaction_type: 'transfer',
  created_at: datetime()
});

// Create transaction relationships
MATCH (w1:Wallet {address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1'}),
      (w2:Wallet {address: '0x8ba1f109551bD432803012645Hac136c22C501e'}),
      (t1:Transaction {hash: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef'})
CREATE (w1)-[:SENT {amount: '1500.0', timestamp: datetime('2024-01-15T10:30:00Z')}]->(t1)
CREATE (t1)-[:RECEIVED {amount: '1500.0', timestamp: datetime('2024-01-15T10:30:00Z')}]->(w2);

MATCH (w3:Wallet {address: '0x9876543210fedcba9876543210fedcba98765432'}),
      (w1:Wallet {address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1'}),
      (t2:Transaction {hash: '0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890'})
CREATE (w3)-[:SENT {amount: '50.0', timestamp: datetime('2024-01-14T16:20:00Z')}]->(t2)
CREATE (t2)-[:RECEIVED {amount: '50.0', timestamp: datetime('2024-01-14T16:20:00Z')}]->(w1);

MATCH (w2:Wallet {address: '0x8ba1f109551bD432803012645Hac136c22C501e'}),
      (w4:Wallet {address: '0x1234567890abcdef1234567890abcdef12345678'}),
      (t3:Transaction {hash: '0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321'})
CREATE (w2)-[:SENT {amount: '0.1', timestamp: datetime('2024-01-14T15:45:00Z')}]->(t3)
CREATE (t3)-[:RECEIVED {amount: '0.1', timestamp: datetime('2024-01-14T15:45:00Z')}]->(w4);

// Create wallet clustering relationships
MATCH (w1:Wallet {address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1'}),
      (w3:Wallet {address: '0x9876543210fedcba9876543210fedcba98765432'})
CREATE (w1)-[:CONNECTED_TO {
  strength: 0.75,
  transaction_count: 5,
  total_value: '250.0',
  first_interaction: datetime('2023-12-01T10:00:00Z'),
  last_interaction: datetime('2024-01-14T16:20:00Z'),
  relationship_type: 'frequent_counterparty'
}]->(w3);

MATCH (w2:Wallet {address: '0x8ba1f109551bD432803012645Hac136c22C501e'}),
      (w4:Wallet {address: '0x1234567890abcdef1234567890abcdef12345678'})
CREATE (w2)-[:CONNECTED_TO {
  strength: 0.95,
  transaction_count: 3,
  total_value: '0.3',
  first_interaction: datetime('2024-01-10T09:00:00Z'),
  last_interaction: datetime('2024-01-14T15:45:00Z'),
  relationship_type: 'suspicious_activity'
}]->(w4);

// Create some sample wallet clusters
CREATE (c1:Cluster {
  id: 'cluster_001',
  name: 'High-Value Trading Group',
  description: 'Cluster of high-value wallets with frequent interactions',
  risk_score: 65.0,
  wallet_count: 3,
  total_value_usd: 807500000,
  created_at: datetime(),
  last_updated: datetime()
});

CREATE (c2:Cluster {
  id: 'cluster_002',
  name: 'Suspicious Activity Network',
  description: 'Cluster showing potential money laundering patterns',
  risk_score: 90.0,
  wallet_count: 2,
  total_value_usd: 750000300,
  created_at: datetime(),
  last_updated: datetime()
});

// Connect wallets to clusters
MATCH (w1:Wallet {address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1'}),
      (w2:Wallet {address: '0x8ba1f109551bD432803012645Hac136c22C501e'}),
      (w3:Wallet {address: '0x9876543210fedcba9876543210fedcba98765432'}),
      (c1:Cluster {id: 'cluster_001'})
CREATE (w1)-[:BELONGS_TO {confidence: 0.85}]->(c1)
CREATE (w2)-[:BELONGS_TO {confidence: 0.90}]->(c1)
CREATE (w3)-[:BELONGS_TO {confidence: 0.75}]->(c1);

MATCH (w2:Wallet {address: '0x8ba1f109551bD432803012645Hac136c22C501e'}),
      (w4:Wallet {address: '0x1234567890abcdef1234567890abcdef12345678'}),
      (c2:Cluster {id: 'cluster_002'})
CREATE (w2)-[:BELONGS_TO {confidence: 0.95}]->(c2)
CREATE (w4)-[:BELONGS_TO {confidence: 0.98}]->(c2);

// Create some sample statistics nodes for quick queries
CREATE (stats:NetworkStats {
  network_id: 'ethereum',
  total_wallets: 200000000,
  total_transactions: 2000000000,
  total_volume_usd: 5000000000000,
  flagged_wallets: 2000000,
  last_updated: datetime()
});

CREATE (stats2:NetworkStats {
  network_id: 'polygon',
  total_wallets: 50000000,
  total_transactions: 800000000,
  total_volume_usd: 500000000000,
  flagged_wallets: 500000,
  last_updated: datetime()
});

// Create indexes for better performance
CREATE INDEX wallet_cluster_id IF NOT EXISTS FOR (c:Cluster) ON (c.id);
CREATE INDEX network_stats_id IF NOT EXISTS FOR (s:NetworkStats) ON (s.network_id);
