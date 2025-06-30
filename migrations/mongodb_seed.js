// MongoDB seed data for Crypto Bubble Map Backend
// Run with: mongosh crypto_bubble_map < mongodb_seed.js

// Switch to the database
use crypto_bubble_map;

// Create collections with validation schemas

// Security alerts collection
db.createCollection("security_alerts", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["id", "type", "severity", "wallet_address", "timestamp"],
      properties: {
        id: { bsonType: "string" },
        type: { 
          bsonType: "string",
          enum: ["suspicious_transaction", "high_risk_counterparty", "unusual_pattern", "compliance_violation", "blacklist_match"]
        },
        severity: {
          bsonType: "string", 
          enum: ["low", "medium", "high", "critical"]
        },
        status: {
          bsonType: "string",
          enum: ["active", "investigating", "resolved", "false_positive"]
        },
        wallet_address: { bsonType: "string" },
        network_id: { bsonType: "string" },
        title: { bsonType: "string" },
        description: { bsonType: "string" },
        confidence: { bsonType: "number", minimum: 0, maximum: 1 },
        action_required: { bsonType: "bool" },
        metadata: { bsonType: "object" },
        timestamp: { bsonType: "date" },
        acknowledged_at: { bsonType: "date" },
        resolved_at: { bsonType: "date" }
      }
    }
  }
});

// Compliance reports collection
db.createCollection("compliance_reports", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["id", "wallet_address", "report_type", "generated_at"],
      properties: {
        id: { bsonType: "string" },
        wallet_address: { bsonType: "string" },
        network_id: { bsonType: "string" },
        report_type: {
          bsonType: "string",
          enum: ["aml", "kyc", "sar", "risk_assessment"]
        },
        status: {
          bsonType: "string",
          enum: ["draft", "pending", "completed", "approved", "rejected"]
        },
        generated_at: { bsonType: "date" },
        generated_by: { bsonType: "string" },
        time_range: {
          bsonType: "object",
          properties: {
            start: { bsonType: "date" },
            end: { bsonType: "date" }
          }
        }
      }
    }
  }
});

// Transactions collection
db.createCollection("transactions", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["hash", "from_address", "to_address", "network_id", "timestamp"],
      properties: {
        hash: { bsonType: "string" },
        from_address: { bsonType: "string" },
        to_address: { bsonType: "string" },
        network_id: { bsonType: "string" },
        value: { bsonType: "string" },
        value_usd: { bsonType: "number" },
        gas_used: { bsonType: "number" },
        gas_price: { bsonType: "string" },
        status: { bsonType: "string" },
        block_number: { bsonType: "number" },
        timestamp: { bsonType: "date" },
        risk_score: { bsonType: "number", minimum: 0, maximum: 100 },
        is_flagged: { bsonType: "bool" }
      }
    }
  }
});

// Create indexes for performance
db.security_alerts.createIndex({ "wallet_address": 1 });
db.security_alerts.createIndex({ "type": 1 });
db.security_alerts.createIndex({ "severity": 1 });
db.security_alerts.createIndex({ "status": 1 });
db.security_alerts.createIndex({ "timestamp": -1 });
db.security_alerts.createIndex({ "network_id": 1 });

db.compliance_reports.createIndex({ "wallet_address": 1 });
db.compliance_reports.createIndex({ "report_type": 1 });
db.compliance_reports.createIndex({ "status": 1 });
db.compliance_reports.createIndex({ "generated_at": -1 });
db.compliance_reports.createIndex({ "network_id": 1 });

db.transactions.createIndex({ "hash": 1 }, { unique: true });
db.transactions.createIndex({ "from_address": 1 });
db.transactions.createIndex({ "to_address": 1 });
db.transactions.createIndex({ "network_id": 1 });
db.transactions.createIndex({ "timestamp": -1 });
db.transactions.createIndex({ "block_number": -1 });
db.transactions.createIndex({ "is_flagged": 1 });
db.transactions.createIndex({ "risk_score": -1 });

// Compound indexes
db.transactions.createIndex({ "from_address": 1, "timestamp": -1 });
db.transactions.createIndex({ "to_address": 1, "timestamp": -1 });
db.transactions.createIndex({ "network_id": 1, "timestamp": -1 });

// Insert sample security alerts
db.security_alerts.insertMany([
  {
    id: "alert_001",
    type: "suspicious_transaction",
    severity: "high",
    status: "active",
    wallet_address: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
    network_id: "ethereum",
    title: "Large Unusual Transaction",
    description: "Transaction amount significantly higher than historical average",
    confidence: 0.85,
    action_required: true,
    metadata: {
      transaction_hash: "0x1234567890abcdef",
      amount_usd: 1500000,
      historical_avg: 50000,
      deviation_factor: 30
    },
    timestamp: new Date("2024-01-15T10:30:00Z")
  },
  {
    id: "alert_002",
    type: "high_risk_counterparty",
    severity: "critical",
    status: "investigating",
    wallet_address: "0x8ba1f109551bD432803012645Hac136c22C501e",
    network_id: "ethereum",
    title: "Transaction with Blacklisted Address",
    description: "Interaction detected with known malicious address",
    confidence: 0.95,
    action_required: true,
    metadata: {
      blacklisted_address: "0x1234567890abcdef1234567890abcdef12345678",
      blacklist_reason: "Ransomware payments",
      transaction_count: 3
    },
    timestamp: new Date("2024-01-14T15:45:00Z")
  },
  {
    id: "alert_003",
    type: "unusual_pattern",
    severity: "medium",
    status: "active",
    wallet_address: "0x9876543210fedcba9876543210fedcba98765432",
    network_id: "polygon",
    title: "Rapid Sequential Transactions",
    description: "Unusually high frequency of transactions in short time period",
    confidence: 0.72,
    action_required: false,
    metadata: {
      transaction_count: 45,
      time_window_minutes: 10,
      normal_frequency: 2
    },
    timestamp: new Date("2024-01-13T08:20:00Z")
  }
]);

// Insert sample compliance reports
db.compliance_reports.insertMany([
  {
    id: "report_001",
    wallet_address: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
    network_id: "ethereum",
    report_type: "aml",
    status: "completed",
    generated_at: new Date("2024-01-10T12:00:00Z"),
    generated_by: "system",
    time_range: {
      start: new Date("2023-12-01T00:00:00Z"),
      end: new Date("2024-01-01T00:00:00Z")
    },
    summary: {
      total_transactions: 156,
      total_volume: "2500000.50",
      total_volume_usd: 2500000.50,
      high_risk_transactions: 3,
      suspicious_patterns: 1,
      regulatory_violations: 0,
      overall_risk_score: 65,
      compliance_score: 85
    },
    findings: [
      {
        type: "high_volume_transaction",
        severity: "medium",
        description: "Single transaction exceeding $1M threshold",
        recommendation: "Enhanced due diligence recommended"
      }
    ]
  },
  {
    id: "report_002",
    wallet_address: "0x8ba1f109551bD432803012645Hac136c22C501e",
    network_id: "ethereum",
    report_type: "risk_assessment",
    status: "completed",
    generated_at: new Date("2024-01-12T14:30:00Z"),
    generated_by: "analyst_001",
    time_range: {
      start: new Date("2023-11-01T00:00:00Z"),
      end: new Date("2024-01-01T00:00:00Z")
    },
    summary: {
      total_transactions: 89,
      total_volume: "750000.25",
      total_volume_usd: 750000.25,
      high_risk_transactions: 8,
      suspicious_patterns: 3,
      regulatory_violations: 1,
      overall_risk_score: 85,
      compliance_score: 45
    },
    findings: [
      {
        type: "blacklist_interaction",
        severity: "critical",
        description: "Multiple interactions with sanctioned addresses",
        recommendation: "Immediate investigation and potential account freeze"
      },
      {
        type: "mixing_service",
        severity: "high",
        description: "Use of cryptocurrency mixing services detected",
        recommendation: "Enhanced monitoring and source of funds verification"
      }
    ]
  }
]);

// Insert sample transactions
db.transactions.insertMany([
  {
    hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    from_address: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
    to_address: "0x8ba1f109551bD432803012645Hac136c22C501e",
    network_id: "ethereum",
    value: "1500000000000000000000000",
    value_usd: 1500000,
    gas_used: 21000,
    gas_price: "20000000000",
    status: "success",
    block_number: 18500000,
    timestamp: new Date("2024-01-15T10:30:00Z"),
    risk_score: 75,
    is_flagged: true,
    metadata: {
      method: "transfer",
      internal_transactions: 0,
      token_transfers: []
    }
  },
  {
    hash: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
    from_address: "0x9876543210fedcba9876543210fedcba98765432",
    to_address: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
    network_id: "polygon",
    value: "50000000000000000000",
    value_usd: 50000,
    gas_used: 21000,
    gas_price: "30000000000",
    status: "success",
    block_number: 52000000,
    timestamp: new Date("2024-01-14T16:20:00Z"),
    risk_score: 25,
    is_flagged: false,
    metadata: {
      method: "transfer",
      internal_transactions: 0,
      token_transfers: []
    }
  }
]);

print("MongoDB seed data inserted successfully!");
print("Collections created: security_alerts, compliance_reports, transactions");
print("Indexes created for optimal query performance");
print("Sample data inserted for testing and development");
