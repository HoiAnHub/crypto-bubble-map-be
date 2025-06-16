import dotenv from 'dotenv';

dotenv.config();

interface Config {
  nodeEnv: string;
  port: number;
  host: string;
  corsOrigin: string;
  
  database: {
    url: string;
    host: string;
    port: number;
    name: string;
    user: string;
    password: string;
  };
  
  neo4j: {
    uri: string;
    user: string;
    password: string;
  };
  
  redis: {
    url: string;
    host: string;
    port: number;
    password?: string;
  };
  
  ethereum: {
    rpcUrl: string;
    rpcUrlBackup: string;
    network: string;
  };
  
  apis: {
    etherscan: string;
    coingecko: string;
  };
  
  cache: {
    walletDetailsTtl: number;
    transactionsTtl: number;
    networkDataTtl: number;
  };
  
  rateLimiting: {
    windowMs: number;
    maxRequests: number;
  };
  
  performance: {
    maxNetworkDepth: number;
    maxTransactionsPerRequest: number;
    batchSize: number;
  };
  
  logging: {
    level: string;
    file: string;
  };
  
  security: {
    jwtSecret: string;
    apiKey: string;
  };
}

const config: Config = {
  nodeEnv: process.env.NODE_ENV || 'development',
  port: parseInt(process.env.PORT || '3001', 10),
  host: process.env.HOST || 'localhost',
  corsOrigin: process.env.CORS_ORIGIN || 'http://localhost:3000',
  
  database: {
    url: process.env.DATABASE_URL || 'postgresql://username:password@localhost:5432/crypto_bubble_map',
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT || '5432', 10),
    name: process.env.DB_NAME || 'crypto_bubble_map',
    user: process.env.DB_USER || 'username',
    password: process.env.DB_PASSWORD || 'password',
  },
  
  neo4j: {
    uri: process.env.NEO4J_URI || 'neo4j://localhost:7687',
    user: process.env.NEO4J_USER || 'neo4j',
    password: process.env.NEO4J_PASSWORD || 'password',
  },
  
  redis: {
    url: process.env.REDIS_URL || 'redis://localhost:6379',
    host: process.env.REDIS_HOST || 'localhost',
    port: parseInt(process.env.REDIS_PORT || '6379', 10),
    password: process.env.REDIS_PASSWORD || undefined,
  },
  
  ethereum: {
    rpcUrl: process.env.ETHEREUM_RPC_URL || '',
    rpcUrlBackup: process.env.ETHEREUM_RPC_URL_BACKUP || '',
    network: process.env.ETHEREUM_NETWORK || 'mainnet',
  },
  
  apis: {
    etherscan: process.env.ETHERSCAN_API_KEY || '',
    coingecko: process.env.COINGECKO_API_KEY || '',
  },
  
  cache: {
    walletDetailsTtl: parseInt(process.env.CACHE_TTL_WALLET_DETAILS || '300', 10),
    transactionsTtl: parseInt(process.env.CACHE_TTL_TRANSACTIONS || '600', 10),
    networkDataTtl: parseInt(process.env.CACHE_TTL_NETWORK_DATA || '900', 10),
  },
  
  rateLimiting: {
    windowMs: parseInt(process.env.API_RATE_LIMIT_WINDOW_MS || '900000', 10),
    maxRequests: parseInt(process.env.API_RATE_LIMIT_MAX_REQUESTS || '100', 10),
  },
  
  performance: {
    maxNetworkDepth: parseInt(process.env.MAX_NETWORK_DEPTH || '3', 10),
    maxTransactionsPerRequest: parseInt(process.env.MAX_TRANSACTIONS_PER_REQUEST || '100', 10),
    batchSize: parseInt(process.env.BATCH_SIZE || '50', 10),
  },
  
  logging: {
    level: process.env.LOG_LEVEL || 'info',
    file: process.env.LOG_FILE || 'logs/app.log',
  },
  
  security: {
    jwtSecret: process.env.JWT_SECRET || 'your-super-secret-jwt-key',
    apiKey: process.env.API_KEY || 'your-api-key-for-authentication',
  },
};

// Validation
const requiredEnvVars = [
  'ETHEREUM_RPC_URL',
];

const missingEnvVars = requiredEnvVars.filter(envVar => !process.env[envVar]);

if (missingEnvVars.length > 0) {
  console.error('Missing required environment variables:', missingEnvVars);
  if (process.env.NODE_ENV === 'production') {
    process.exit(1);
  }
}

export { config };
