import { Pool, PoolClient } from 'pg';
import { config } from '@/config/config';
import { logger } from '@/utils/logger';

export class DatabaseService {
  private pool: Pool | null = null;
  private isConnected = false;

  constructor() {
    // Initialize pool but don't connect yet
  }

  public async connect(): Promise<void> {
    if (this.isConnected && this.pool) {
      return;
    }

    try {
      this.pool = new Pool({
        connectionString: config.database.url,
        host: config.database.host,
        port: config.database.port,
        database: config.database.name,
        user: config.database.user,
        password: config.database.password,
        max: 20, // Maximum number of clients in the pool
        idleTimeoutMillis: 30000, // Close idle clients after 30 seconds
        connectionTimeoutMillis: 2000, // Return an error after 2 seconds if connection could not be established
        ssl: config.nodeEnv === 'production' ? { rejectUnauthorized: false } : false,
      });

      // Test the connection
      const client = await this.pool.connect();
      await client.query('SELECT NOW()');
      client.release();

      this.isConnected = true;
      logger.info('✅ PostgreSQL connected successfully');

      // Set up event handlers
      this.pool.on('error', (err) => {
        logger.error('Unexpected error on idle client:', err);
        this.isConnected = false;
      });

      this.pool.on('connect', () => {
        logger.debug('New client connected to PostgreSQL');
      });

      this.pool.on('remove', () => {
        logger.debug('Client removed from PostgreSQL pool');
      });

    } catch (error) {
      logger.error('Failed to connect to PostgreSQL:', error);
      this.isConnected = false;
      throw error;
    }
  }

  public async disconnect(): Promise<void> {
    if (this.pool) {
      try {
        await this.pool.end();
        this.isConnected = false;
        logger.info('PostgreSQL disconnected');
      } catch (error) {
        logger.error('Error disconnecting from PostgreSQL:', error);
      }
    }
  }

  public async getClient(): Promise<PoolClient> {
    if (!this.pool || !this.isConnected) {
      throw new Error('Database not connected');
    }

    return await this.pool.connect();
  }

  public async query(text: string, params?: any[]): Promise<any> {
    if (!this.pool || !this.isConnected) {
      throw new Error('Database not connected');
    }

    const start = Date.now();
    try {
      const result = await this.pool.query(text, params);
      const duration = Date.now() - start;
      
      logger.debug('Executed query', {
        text: text.substring(0, 100) + (text.length > 100 ? '...' : ''),
        duration: `${duration}ms`,
        rows: result.rowCount,
      });

      return result;
    } catch (error) {
      logger.error('Database query error:', {
        text: text.substring(0, 100) + (text.length > 100 ? '...' : ''),
        error: error instanceof Error ? error.message : error,
        params,
      });
      throw error;
    }
  }

  public async transaction<T>(callback: (client: PoolClient) => Promise<T>): Promise<T> {
    const client = await this.getClient();
    
    try {
      await client.query('BEGIN');
      const result = await callback(client);
      await client.query('COMMIT');
      return result;
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  public isReady(): boolean {
    return this.isConnected && this.pool !== null;
  }

  // Health check
  public async healthCheck(): Promise<boolean> {
    try {
      if (!this.pool || !this.isConnected) {
        return false;
      }

      const result = await this.pool.query('SELECT 1 as health');
      return result.rows[0]?.health === 1;
    } catch (error) {
      logger.error('Database health check failed:', error);
      return false;
    }
  }

  // Database initialization
  public async initializeTables(): Promise<void> {
    if (!this.isReady()) {
      throw new Error('Database not connected');
    }

    try {
      // Create wallets table
      await this.query(`
        CREATE TABLE IF NOT EXISTS wallets (
          id SERIAL PRIMARY KEY,
          address VARCHAR(42) UNIQUE NOT NULL,
          balance DECIMAL(36, 18) DEFAULT 0,
          balance_usd DECIMAL(15, 2) DEFAULT 0,
          transaction_count INTEGER DEFAULT 0,
          is_contract BOOLEAN DEFAULT FALSE,
          contract_type VARCHAR(50),
          label VARCHAR(255),
          tags TEXT[],
          risk_score INTEGER DEFAULT 0,
          first_seen TIMESTAMP WITH TIME ZONE,
          last_activity TIMESTAMP WITH TIME ZONE,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
          updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
      `);

      // Create transactions table
      await this.query(`
        CREATE TABLE IF NOT EXISTS transactions (
          id SERIAL PRIMARY KEY,
          hash VARCHAR(66) UNIQUE NOT NULL,
          from_address VARCHAR(42) NOT NULL,
          to_address VARCHAR(42),
          value DECIMAL(36, 18) NOT NULL,
          value_usd DECIMAL(15, 2),
          gas_used INTEGER,
          gas_price DECIMAL(36, 18),
          timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
          block_number INTEGER NOT NULL,
          status INTEGER NOT NULL,
          method_id VARCHAR(10),
          function_name VARCHAR(255),
          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
      `);

      // Create token_transfers table
      await this.query(`
        CREATE TABLE IF NOT EXISTS token_transfers (
          id SERIAL PRIMARY KEY,
          transaction_hash VARCHAR(66) NOT NULL,
          from_address VARCHAR(42) NOT NULL,
          to_address VARCHAR(42) NOT NULL,
          token_address VARCHAR(42) NOT NULL,
          value DECIMAL(36, 18) NOT NULL,
          token_symbol VARCHAR(20),
          token_name VARCHAR(255),
          token_decimals INTEGER,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
      `);

      // Create indexes
      await this.query('CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_wallets_balance ON wallets(balance DESC);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_wallets_transaction_count ON wallets(transaction_count DESC);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_wallets_last_activity ON wallets(last_activity DESC);');
      
      await this.query('CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_transactions_from ON transactions(from_address);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_transactions_to ON transactions(to_address);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_transactions_block ON transactions(block_number DESC);');
      
      await this.query('CREATE INDEX IF NOT EXISTS idx_token_transfers_hash ON token_transfers(transaction_hash);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_token_transfers_from ON token_transfers(from_address);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_token_transfers_to ON token_transfers(to_address);');
      await this.query('CREATE INDEX IF NOT EXISTS idx_token_transfers_token ON token_transfers(token_address);');

      logger.info('Database tables initialized successfully');
    } catch (error) {
      logger.error('Failed to initialize database tables:', error);
      throw error;
    }
  }
}
