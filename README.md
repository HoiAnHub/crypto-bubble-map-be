# Crypto Bubble Map Backend

A Node.js backend service that provides Ethereum blockchain data for the crypto-bubble-map frontend visualization tool.

## Features

- **Proactive Data Crawling**: Background job system that pre-crawls cryptocurrency data
- **Ethereum Blockchain Integration**: Real-time data fetching from Ethereum mainnet
- **Wallet Network Analysis**: Discover relationships between wallet addresses
- **Transaction History**: Detailed transaction data and analysis
- **Graph Database**: Neo4j integration for complex relationship queries
- **Intelligent Caching**: Multi-layer Redis caching with pre-populated data
- **Background Job Scheduler**: Automated data collection and refresh system
- **Market Data Integration**: Real-time price feeds and gas tracker
- **Popular Wallet Discovery**: Automated identification of trending addresses
- **RESTful API**: Clean API endpoints serving pre-crawled data for optimal performance

## Technology Stack

- **Runtime**: Node.js with TypeScript
- **Framework**: Express.js
- **Blockchain**: Ethers.js for Ethereum integration
- **Databases**: PostgreSQL (primary data) + Neo4j (graph relationships)
- **Caching**: Redis
- **API Documentation**: Swagger/OpenAPI
- **Job Scheduling**: Node-cron for automated background tasks

## API Endpoints

### Wallet Data (Served from Pre-crawled Cache)
- `GET /api/wallets/network?address={address}&depth={depth}` - Get wallet network relationships
- `GET /api/wallets/{address}` - Get detailed wallet information
- `GET /api/wallets/search?q={query}` - Search wallets by address/label
- `GET /api/wallets/{address}/transactions?limit={limit}` - Get transaction history
- `GET /api/wallets/popular` - Get popular/trending wallets (pre-crawled)
- `GET /api/wallets/market-data` - Get current market data (pre-crawled)
- `GET /api/wallets/stats` - Get system statistics
- `POST /api/wallets/batch` - Get multiple wallet details

### Background Job Management
- `GET /api/jobs/stats` - Get job queue statistics and status
- `GET /api/jobs/health` - Get job system health status
- `POST /api/jobs/trigger/{jobType}` - Manually trigger specific job types

### Health Checks
- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed service health including job system

## Proactive Data Crawling System

The backend now features a comprehensive proactive data crawling system that pre-fetches and caches cryptocurrency data to improve response times and user experience.

### Background Job Types

1. **Market Data Crawling** (Every 5 minutes)
   - ETH price and market cap from CoinGecko
   - Gas price tracking from Etherscan
   - Network statistics and block information

2. **Popular Wallet Discovery** (Every 6 hours)
   - Identifies trending and high-activity wallet addresses
   - Analyzes transaction patterns and volumes
   - Maintains a curated list of interesting wallets

3. **Network Statistics** (Every 10 minutes)
   - Current block number and network health
   - Transaction pool status
   - Network difficulty and hash rate

4. **Wallet Data Refresh** (Every 2 hours)
   - Updates existing wallet information
   - Refreshes transaction counts and balances
   - Prioritizes high-activity and popular wallets

5. **Data Cleanup** (Daily at 2 AM)
   - Removes old cached entries
   - Cleans up completed job records
   - Maintains database performance

### Job Queue System

- **Redis-based Queue**: Reliable job storage and processing
- **Priority Levels**: Critical, High, Medium, Low priority jobs
- **Retry Logic**: Automatic retry with exponential backoff
- **Job Monitoring**: Real-time statistics and health monitoring
- **Graceful Shutdown**: Proper cleanup on application restart

### Configuration

Background jobs can be configured via environment variables:

```bash
# Enable/disable background jobs
JOBS_ENABLED=true

# Job intervals (cron format)
JOB_INTERVAL_MARKET_DATA=*/5 * * * *      # Every 5 minutes
JOB_INTERVAL_POPULAR_WALLETS=0 */6 * * *  # Every 6 hours
JOB_INTERVAL_NETWORK_STATS=*/10 * * * *   # Every 10 minutes
JOB_INTERVAL_WALLET_REFRESH=0 */2 * * *   # Every 2 hours
JOB_INTERVAL_CLEANUP=0 2 * * *            # Daily at 2 AM

# Batch sizes for processing
JOB_BATCH_SIZE_WALLETS=20
JOB_BATCH_SIZE_TRANSACTIONS=100

# Retry configuration
JOB_MAX_RETRIES=3
JOB_RETRY_DELAY=5000

# Priority wallet lists
JOB_HIGH_PRIORITY_WALLETS=0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6,0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

### Benefits

- **Improved Response Times**: API endpoints serve pre-crawled data from cache
- **Reduced External API Calls**: Batch processing minimizes rate limiting
- **Better User Experience**: Instant data availability for popular requests
- **Scalable Architecture**: Job queue handles high-volume data processing
- **Fault Tolerance**: Retry logic and error handling ensure data consistency

## Getting Started

### Prerequisites

- Node.js (v18 or later)
- PostgreSQL database
- Neo4j database
- Redis server
- Ethereum RPC endpoint (Infura, Alchemy, etc.)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/HoiAnHub/crypto-bubble-map-be.git
   cd crypto-bubble-map-be
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run database migrations:
   ```bash
   npm run migrate
   ```

5. Start the development server:
   ```bash
   npm run dev
   ```

The server will start on `http://localhost:3001`

## Environment Variables

See `.env.example` for required environment variables including:
- Database connection strings
- Ethereum RPC endpoints
- Redis configuration
- API keys and secrets

### Redis Configuration

The service supports both URL-based and individual parameter Redis configuration:

**Individual Parameters (Recommended):**
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=dev
REDIS_DB=0
```

**URL-based (Backward Compatibility):**
```bash
REDIS_URL=redis://localhost:6379
```

The service will automatically use individual parameters if `REDIS_HOST` and `REDIS_PORT` are provided, otherwise it falls back to `REDIS_URL` for backward compatibility.

## Development

- `npm run dev` - Start development server with hot reload
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run test` - Run tests
- `npm run lint` - Run ESLint

## License

ISC