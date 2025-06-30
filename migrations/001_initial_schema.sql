-- Initial database schema for Crypto Bubble Map Backend
-- PostgreSQL migration script

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    role VARCHAR(50) DEFAULT 'user',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Watched wallet tags table
CREATE TABLE IF NOT EXISTS watched_wallet_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    color VARCHAR(7) DEFAULT '#3B82F6',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Watched wallets table
CREATE TABLE IF NOT EXISTS watched_wallets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL,
    network_id VARCHAR(50) NOT NULL,
    label VARCHAR(255),
    notes TEXT,
    balance VARCHAR(255),
    risk_score DECIMAL(5,2) DEFAULT 0.0,
    quality_score DECIMAL(5,2) DEFAULT 0.0,
    wallet_type VARCHAR(50) DEFAULT 'regular',
    is_flagged BOOLEAN DEFAULT false,
    last_activity TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, address, network_id)
);

-- Junction table for watched wallets and tags (many-to-many)
CREATE TABLE IF NOT EXISTS watched_wallet_tag_associations (
    id SERIAL PRIMARY KEY,
    watched_wallet_id INTEGER NOT NULL REFERENCES watched_wallets(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES watched_wallet_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(watched_wallet_id, tag_id)
);

-- Wallet alerts table
CREATE TABLE IF NOT EXISTS wallet_alerts (
    id SERIAL PRIMARY KEY,
    wallet_id INTEGER NOT NULL REFERENCES watched_wallets(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_at TIMESTAMP,
    acknowledged_by INTEGER REFERENCES users(id),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_watched_wallets_user_id ON watched_wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_watched_wallets_address ON watched_wallets(address);
CREATE INDEX IF NOT EXISTS idx_watched_wallets_network_id ON watched_wallets(network_id);
CREATE INDEX IF NOT EXISTS idx_watched_wallets_risk_score ON watched_wallets(risk_score);
CREATE INDEX IF NOT EXISTS idx_watched_wallets_flagged ON watched_wallets(is_flagged);
CREATE INDEX IF NOT EXISTS idx_watched_wallets_last_activity ON watched_wallets(last_activity);

CREATE INDEX IF NOT EXISTS idx_wallet_alerts_wallet_id ON wallet_alerts(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_alerts_type ON wallet_alerts(type);
CREATE INDEX IF NOT EXISTS idx_wallet_alerts_severity ON wallet_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_wallet_alerts_acknowledged ON wallet_alerts(acknowledged);
CREATE INDEX IF NOT EXISTS idx_wallet_alerts_timestamp ON wallet_alerts(timestamp);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_sessions_updated_at BEFORE UPDATE ON user_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_watched_wallet_tags_updated_at BEFORE UPDATE ON watched_wallet_tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_watched_wallets_updated_at BEFORE UPDATE ON watched_wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_wallet_alerts_updated_at BEFORE UPDATE ON wallet_alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default tags
INSERT INTO watched_wallet_tags (name, color, description) VALUES
    ('High Risk', '#EF4444', 'Wallets with high risk scores'),
    ('Exchange', '#3B82F6', 'Exchange wallets'),
    ('DeFi', '#10B981', 'DeFi protocol wallets'),
    ('Whale', '#8B5CF6', 'High-value wallets'),
    ('Suspicious', '#F59E0B', 'Potentially suspicious activity'),
    ('Whitelist', '#06B6D4', 'Trusted wallets'),
    ('Bridge', '#EC4899', 'Cross-chain bridge wallets'),
    ('Contract', '#6B7280', 'Smart contract addresses'),
    ('Miner', '#F97316', 'Mining pool wallets'),
    ('Personal', '#84CC16', 'Personal wallet addresses')
ON CONFLICT (name) DO NOTHING;

-- Create admin user (password: admin123)
-- Note: In production, this should be done securely
INSERT INTO users (email, password_hash, first_name, last_name, role, is_verified) VALUES
    ('admin@cryptobubblemap.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', 'admin', true)
ON CONFLICT (email) DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE users IS 'User accounts and authentication';
COMMENT ON TABLE user_sessions IS 'User session management';
COMMENT ON TABLE watched_wallets IS 'User-watched wallet addresses';
COMMENT ON TABLE watched_wallet_tags IS 'Tags for categorizing watched wallets';
COMMENT ON TABLE watched_wallet_tag_associations IS 'Many-to-many relationship between wallets and tags';
COMMENT ON TABLE wallet_alerts IS 'Alerts generated for watched wallets';

COMMENT ON COLUMN users.role IS 'User role: user, admin, analyst';
COMMENT ON COLUMN watched_wallets.wallet_type IS 'Type: regular, exchange, contract, whale, defi, bridge, miner';
COMMENT ON COLUMN wallet_alerts.type IS 'Alert type: large_transaction, suspicious_activity, risk_change, etc.';
COMMENT ON COLUMN wallet_alerts.severity IS 'Alert severity: low, medium, high, critical';
