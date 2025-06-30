-- Initialize PostgreSQL database for Crypto Bubble Map Backend

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create indexes for better performance
-- These will be created automatically by GORM, but we can add custom ones here

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- This trigger will be applied to tables after they are created by GORM
-- For now, we'll just prepare the database

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE crypto_bubble_map TO postgres;

-- Create schema for future use
CREATE SCHEMA IF NOT EXISTS crypto_bubble_map;

-- Set default search path
ALTER DATABASE crypto_bubble_map SET search_path TO crypto_bubble_map, public;

-- Log initialization
INSERT INTO pg_stat_statements_info (dealloc) VALUES (0) ON CONFLICT DO NOTHING;
