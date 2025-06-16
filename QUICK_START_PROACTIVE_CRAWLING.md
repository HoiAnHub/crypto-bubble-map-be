# Quick Start Guide: Proactive Data Crawling

## Prerequisites

Ensure you have the following services running:
- **PostgreSQL** (for primary data storage)
- **Neo4j** (for graph relationships)
- **Redis** (for caching and job queue)

## Setup Steps

### 1. Install Dependencies
```bash
cd crypto-bubble-map-be
npm install
```

### 2. Configure Environment Variables
Copy and configure the environment file:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```bash
# Required: External API keys
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
ETHERSCAN_API_KEY=YOUR_ETHERSCAN_API_KEY
COINGECKO_API_KEY=YOUR_COINGECKO_API_KEY

# Database connections
DATABASE_URL=postgresql://username:password@localhost:5432/crypto_bubble_map
NEO4J_URI=neo4j://localhost:7687

# Redis configuration (individual parameters recommended)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=dev
REDIS_DB=0

# Or use URL format (backward compatibility)
# REDIS_URL=redis://localhost:6379

# Background jobs (enabled by default)
JOBS_ENABLED=true
JOB_INTERVAL_MARKET_DATA=*/5 * * * *
JOB_INTERVAL_POPULAR_WALLETS=0 */6 * * *
```

### 3. Start the Server
```bash
# Development mode
npm run dev

# Production mode
npm run build
npm start
```

## Testing the Implementation

### 1. Run the Test Script
```bash
node test-proactive-crawling.js
```

This will test:
- ✅ Basic health checks
- ✅ Job system status
- ✅ Pre-crawled data endpoints
- ✅ Performance comparisons

### 2. Manual API Testing

#### Check Job System Health
```bash
curl http://localhost:3001/api/jobs/health
```

#### Get Job Statistics
```bash
curl http://localhost:3001/api/jobs/stats
```

#### Test Pre-crawled Market Data
```bash
curl http://localhost:3001/api/wallets/market-data
```

#### Test Popular Wallets
```bash
curl http://localhost:3001/api/wallets/popular
```

### 3. Monitor Job Execution

#### Check Server Logs
The server logs will show background job activity:
```
🔄 Background job scheduler started
📊 Scheduling market data crawl job
✅ Job 1234567890-abc123 completed successfully
```

#### Monitor Job Queue
```bash
# Get real-time job statistics
curl http://localhost:3001/api/jobs/stats | jq
```

## Expected Behavior

### Initial Startup (First 10 minutes)
1. **Server starts** with job scheduler enabled
2. **Database tables** are created automatically
3. **First jobs are scheduled** according to cron intervals
4. **Market data job** runs within 5 minutes
5. **Cache begins populating** with fresh data

### After 1 Hour
1. **Market data** is refreshed multiple times
2. **Popular wallets** discovery may have run (if 6-hour interval passed)
3. **API endpoints** serve cached data with fast response times
4. **Job statistics** show processing history

### Steady State
1. **Sub-100ms response times** for cached endpoints
2. **Regular job execution** according to configured intervals
3. **Automatic data refresh** without manual intervention
4. **Error handling** and retry logic in action

## Troubleshooting

### Common Issues

#### Jobs Not Running
- Check `JOBS_ENABLED=true` in `.env`
- Verify Redis connection
- Check server logs for error messages

#### No Cached Data
- Wait for initial job execution (up to 5 minutes for market data)
- Check external API keys are configured
- Verify network connectivity to external APIs

#### Performance Issues
- Monitor Redis memory usage
- Check database connection pool settings
- Review job batch sizes in configuration

### Debug Commands

#### Check Redis Job Queue
```bash
# Connect to Redis CLI (with password if configured)
redis-cli
AUTH dev  # if password is set

# Or connect directly with password
redis-cli -a dev

# Check job queues
LLEN jobs:queue:high
LLEN jobs:queue:medium
LLEN jobs:queue:low

# Check scheduled jobs
ZRANGE jobs:scheduled 0 -1 WITHSCORES

# Check current database
SELECT 0  # switch to database 0 if needed
```

#### Check Database Tables
```sql
-- Check wallet data
SELECT COUNT(*) FROM wallets;

-- Check market data history
SELECT * FROM market_data_history ORDER BY created_at DESC LIMIT 5;

-- Check job execution history
SELECT * FROM job_execution_history ORDER BY created_at DESC LIMIT 10;
```

## Performance Verification

### Response Time Testing
```bash
# Test cached endpoint (should be <100ms)
time curl http://localhost:3001/api/wallets/market-data

# Test traditional endpoint (may be slower on first call)
time curl http://localhost:3001/api/wallets/0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
```

### Cache Hit Rate
Monitor the logs for cache hit/miss messages:
```
Cache hit for market data
Cache hit for popular wallets
Cache miss for wallet details: 0x...
```

## Configuration Tuning

### Job Intervals
Adjust based on your needs:
```bash
# More frequent market data updates
JOB_INTERVAL_MARKET_DATA=*/2 * * * *

# Less frequent popular wallet discovery
JOB_INTERVAL_POPULAR_WALLETS=0 */12 * * *
```

### Batch Sizes
Optimize for your infrastructure:
```bash
# Larger batches for better throughput
JOB_BATCH_SIZE_WALLETS=50

# Smaller batches to reduce memory usage
JOB_BATCH_SIZE_WALLETS=10
```

### Priority Wallets
Add important addresses for priority processing:
```bash
JOB_HIGH_PRIORITY_WALLETS=0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6,0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

## Next Steps

1. **Monitor Performance**: Use the job statistics endpoints to track system health
2. **Tune Configuration**: Adjust intervals and batch sizes based on usage patterns
3. **Scale Infrastructure**: Add more Redis memory or database resources as needed
4. **Implement Monitoring**: Set up Grafana dashboards for visual monitoring
5. **Add Custom Jobs**: Extend the system with application-specific crawling jobs

## Support

If you encounter issues:
1. Check the server logs for detailed error messages
2. Verify all external services (PostgreSQL, Neo4j, Redis) are running
3. Test external API connectivity (Etherscan, CoinGecko)
4. Review the configuration in `.env` file
5. Run the test script to identify specific problems

The proactive crawling system is designed to be robust and self-healing, but proper configuration and monitoring are essential for optimal performance.
