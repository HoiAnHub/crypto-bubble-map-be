# Crypto Bubble Map Backend

A GraphQL API backend for the Crypto Bubble Map application, providing comprehensive blockchain analytics and wallet relationship data.

## ✨ Features

- **GraphQL API**: Comprehensive GraphQL schema for all frontend needs
- **Multi-Database Integration**: Neo4j (graph), MongoDB (raw data), PostgreSQL (user data), Redis (cache)
- **Real-time Subscriptions**: WebSocket-based real-time updates
- **Advanced Analytics**: Risk scoring, wallet classification, transaction analysis
- **Authentication & Authorization**: JWT-based auth with role-based access
- **Performance Optimized**: DataLoader, caching, and query optimization
- **Microservice Architecture**: Clean, scalable, and maintainable design

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Frontend      │───▶│   GraphQL    │───▶│    Neo4j        │
│   React App     │    │   Gateway    │    │   Graph DB      │
└─────────────────┘    └──────┬───────┘    └─────────────────┘
                              │
                              ├─────────────▶┌─────────────────┐
                              │              │    MongoDB      │
                              │              │   Raw Data      │
                              │              └─────────────────┘
                              │
                              ├─────────────▶┌─────────────────┐
                              │              │  PostgreSQL     │
                              │              │   User Data     │
                              │              └─────────────────┘
                              │
                              └─────────────▶┌─────────────────┐
                                             │     Redis       │
                                             │     Cache       │
                                             └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- Neo4j Database (local or Docker)
- MongoDB Atlas (cloud) - already configured
- PostgreSQL (local on port 5433) - already configured
- Redis (local or Docker)
- Docker and Docker Compose (optional)

### Setup

1. **Clone and setup environment:**
   ```bash
   cd crypto-bubble-map-be
   make setup
   ```

2. **Database configuration is already set up:**
   - ✅ **MongoDB Atlas**: Cloud database already configured
   - ✅ **PostgreSQL**: Local database on port 5433 already configured
   - ⚠️ **Neo4j**: Need to start local instance
   - ⚠️ **Redis**: Need to start local instance

3. **Start required databases:**
   ```bash
   # Start Neo4j (if not running)
   docker run -d -p 7687:7687 -p 7474:7474 --env NEO4J_AUTH=neo4j/password neo4j

   # Start Redis (if not running)
   docker run -d -p 6379:6379 redis

   # Test all database connections
   make test-db
   ```

4. **Start the server:**
   ```bash
   # Build and run locally
   make build && make run

   # Test server endpoints
   make test-server
   ```

### 🎯 Current Implementation Status

✅ **Completed Features:**
- ✅ Complete project structure with Go modules
- ✅ Multi-database integration (Neo4j, MongoDB, PostgreSQL, Redis)
- ✅ Configuration management with environment variables
- ✅ Structured logging with Zap
- ✅ Docker containerization with docker-compose
- ✅ Repository pattern implementation
- ✅ Domain-driven design architecture
- ✅ Cache layer with Redis
- ✅ Health check endpoints
- ✅ Graceful shutdown handling
- ✅ Rate limiting middleware
- ✅ CORS support
- ✅ Basic GraphQL endpoint structure

🚧 **In Progress:**
- 🚧 GraphQL schema generation (gqlgen setup)
- 🚧 Complete resolver implementations
- 🚧 Authentication & JWT middleware
- 🚧 Watch list functionality
- 🚧 Security alerts system

📋 **Next Steps:**
- [ ] Complete GraphQL resolvers
- [ ] Add authentication system
- [ ] Implement real-time subscriptions
- [ ] Add comprehensive testing
- [ ] Performance optimization
- [ ] API documentation
- [ ] Monitoring & metrics

## 📋 Usage

### Using Makefile

```bash
# Build server
make build

# Run locally
make run

# Start with Docker Compose
make up

# Generate GraphQL code
make generate

# Run tests
make test

# View logs
make logs

# Stop services
make down
```

## ⚙️ Configuration

Key environment variables in `.env`:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=release

# Neo4j Configuration
NEO4J_URI=neo4j://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=password
NEO4J_DATABASE=neo4j

# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017/ethereum_raw_data

# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=crypto_bubble_map

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Application Configuration
APP_ENV=production
LOG_LEVEL=info
```

## 📊 GraphQL Schema

### Core Queries

```graphql
# Get wallet network for bubble map
query GetWalletNetwork($address: String!, $depth: Int!) {
  walletNetwork(address: $address, depth: $depth) {
    nodes {
      id
      address
      label
      balance
      riskScore {
        totalScore
        riskLevel
      }
    }
    links {
      source
      target
      value
      transactionCount
    }
  }
}

# Get wallet rankings
query GetWalletRankings($category: RankingCategory!, $limit: Int!) {
  walletRankings(category: $category, limit: $limit) {
    rank
    wallet {
      address
      qualityScore
      riskScore
    }
    score
  }
}
```

### Real-time Subscriptions

```graphql
# Subscribe to wallet updates
subscription WalletUpdates($addresses: [String!]!) {
  walletUpdates(addresses: $addresses) {
    address
    balance
    transactionCount
    riskScore
  }
}
```

## 🔧 Development

### Project Structure

```
crypto-bubble-map-be/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   ├── application/
│   │   ├── service/
│   │   └── usecase/
│   ├── infrastructure/
│   │   ├── config/
│   │   ├── database/
│   │   ├── cache/
│   │   └── logger/
│   ├── interfaces/
│   │   ├── graphql/
│   │   ├── rest/
│   │   └── middleware/
│   └── pkg/
├── graph/
│   ├── schema.graphqls
│   ├── resolver.go
│   └── generated/
├── scripts/
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## 🐳 Docker

### Development
```bash
docker-compose up -d
```

### Production
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
