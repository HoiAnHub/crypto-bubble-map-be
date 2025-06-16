# Crypto Bubble Map Backend Setup Guide

This guide will help you set up the complete backend infrastructure for the crypto bubble map application.

## Prerequisites

Before starting, ensure you have the following installed:

- **Node.js** (v18 or later)
- **PostgreSQL** (v12 or later)
- **Neo4j** (v4.4 or later)
- **Redis** (v6 or later)

## Quick Start (Development)

### 1. Install Dependencies

```bash
cd crypto-bubble-map-be
npm install
```

### 2. Environment Configuration

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Ethereum Configuration - Get free RPC endpoints from:
# - Infura: https://infura.io/
# - Alchemy: https://www.alchemy.com/
# - Or use public endpoints (limited rate)
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROJECT_ID
ETHEREUM_RPC_URL_BACKUP=https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY

# External APIs (optional but recommended)
ETHERSCAN_API_KEY=YOUR_ETHERSCAN_API_KEY
COINGECKO_API_KEY=YOUR_COINGECKO_API_KEY
```

### 3. Database Setup

#### PostgreSQL Setup

1. Install PostgreSQL and create a database:
```sql
CREATE DATABASE crypto_bubble_map;
CREATE USER crypto_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE crypto_bubble_map TO crypto_user;
```

2. Update your `.env` file:
```env
DATABASE_URL=postgresql://crypto_user:your_password@localhost:5432/crypto_bubble_map
```

#### Neo4j Setup

1. Install Neo4j Desktop or Community Edition
2. Create a new database with credentials:
   - Username: `neo4j`
   - Password: `your_neo4j_password`

3. Update your `.env` file:
```env
NEO4J_URI=neo4j://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=your_neo4j_password
```

#### Redis Setup

1. Install and start Redis:
```bash
# macOS
brew install redis
brew services start redis

# Ubuntu/Debian
sudo apt install redis-server
sudo systemctl start redis-server

# Windows
# Download from https://redis.io/download
```

2. Configure Redis connection in `.env`:
```bash
# Individual parameters (recommended)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=dev
REDIS_DB=0

# Or use URL format (backward compatibility)
REDIS_URL=redis://localhost:6379
```

3. For development with password protection:
```bash
# Edit redis.conf
requirepass dev

# Restart Redis
brew services restart redis  # macOS
sudo systemctl restart redis-server  # Linux
```

### 4. Start the Backend

```bash
# Development mode with hot reload
npm run dev

# Or build and run production
npm run build
npm start
```

The server will start on `http://localhost:3001`

### 5. Test the API

Run the test script to verify everything is working:

```bash
node test-api.js
```

## Docker Setup (Alternative)

If you prefer using Docker, create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: crypto_bubble_map
      POSTGRES_USER: crypto_user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  neo4j:
    image: neo4j:5.0
    environment:
      NEO4J_AUTH: neo4j/password
      NEO4J_PLUGINS: '["apoc"]'
    ports:
      - "7474:7474"
      - "7687:7687"
    volumes:
      - neo4j_data:/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --requirepass dev

volumes:
  postgres_data:
  neo4j_data:
  redis_data:
```

Then run:
```bash
docker-compose up -d
```

## API Endpoints

Once running, the following endpoints will be available:

### Health Checks
- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed service health
- `GET /health/ready` - Kubernetes readiness probe
- `GET /health/live` - Kubernetes liveness probe

### Wallet APIs
- `GET /api/wallets/network?address={address}&depth={depth}` - Get wallet network
- `GET /api/wallets/{address}` - Get wallet details
- `GET /api/wallets/search?q={query}` - Search wallets
- `GET /api/wallets/{address}/transactions?limit={limit}` - Get transactions
- `POST /api/wallets/batch` - Get multiple wallet details
- `GET /api/wallets/stats` - Get system statistics

## Frontend Integration

Update your frontend's API configuration to point to the backend:

```typescript
// In your frontend .env file
NEXT_PUBLIC_API_URL=http://localhost:3001/api
```

## Production Deployment

### Environment Variables

For production, ensure you set:

```env
NODE_ENV=production
PORT=3001

# Use secure database connections
DATABASE_URL=postgresql://user:pass@prod-db:5432/crypto_bubble_map
NEO4J_URI=neo4j+s://prod-neo4j:7687

# Redis configuration (use individual parameters for better control)
REDIS_HOST=prod-redis
REDIS_PORT=6379
REDIS_PASS=your_secure_redis_password
REDIS_DB=0

# Or use URL format (backward compatibility)
# REDIS_URL=redis://prod-redis:6379

# Use production-grade RPC endpoints
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROD_PROJECT_ID
ETHERSCAN_API_KEY=YOUR_PROD_ETHERSCAN_KEY

# Security
JWT_SECRET=your-super-secure-jwt-secret
API_KEY=your-production-api-key

# CORS
CORS_ORIGIN=https://your-frontend-domain.com
```

### Performance Tuning

1. **Database Connections**: Adjust pool sizes based on your load
2. **Caching**: Configure Redis with appropriate TTL values
3. **Rate Limiting**: Adjust rate limits based on your usage patterns
4. **Logging**: Set appropriate log levels for production

### Monitoring

The backend includes health check endpoints for monitoring:

- Use `/health/ready` for Kubernetes readiness probes
- Use `/health/live` for liveness probes
- Monitor `/health/detailed` for service status

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Verify database credentials and connectivity
   - Check if databases are running
   - Ensure firewall allows connections

2. **Ethereum RPC Errors**
   - Verify RPC URLs are correct
   - Check API key limits
   - Try backup RPC endpoint

3. **Memory Issues**
   - Increase Node.js memory limit: `node --max-old-space-size=4096`
   - Optimize database queries
   - Implement proper caching

4. **Rate Limiting**
   - Adjust rate limits in configuration
   - Implement API key authentication
   - Use multiple RPC endpoints

### Logs

Check application logs for detailed error information:

```bash
# Development
tail -f logs/app.log

# Production
pm2 logs crypto-bubble-map-be
```

## Next Steps

1. **Configure External APIs**: Set up Etherscan and CoinGecko API keys for enhanced functionality
2. **Data Population**: Run initial data sync to populate the databases
3. **Monitoring**: Set up application monitoring and alerting
4. **Scaling**: Configure load balancing and horizontal scaling as needed

For more detailed information, refer to the API documentation and individual service configurations.
