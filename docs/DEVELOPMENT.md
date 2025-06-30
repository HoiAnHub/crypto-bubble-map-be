# Development Guide

This guide covers the development setup and workflow for the Crypto Bubble Map Backend.

## 🏗️ Architecture Overview

The backend follows a **Domain-Driven Design (DDD)** architecture with clean separation of concerns:

```
crypto-bubble-map-be/
├── cmd/                    # Application entry points
│   └── server/            # Main server application
├── internal/              # Private application code
│   ├── domain/           # Domain layer (business logic)
│   │   ├── entity/       # Domain entities
│   │   └── repository/   # Repository interfaces
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── database/     # Database clients
│   │   ├── cache/        # Cache implementations
│   │   ├── config/       # Configuration
│   │   ├── logger/       # Logging
│   │   └── repository/   # Repository implementations
│   └── interfaces/       # Interface layer
│       └── graphql/      # GraphQL handlers
├── graph/                # GraphQL schema and resolvers
├── scripts/              # Utility scripts
└── docs/                 # Documentation
```

## 🛠️ Development Setup

### 1. Prerequisites

- **Go 1.23+**: [Install Go](https://golang.org/doc/install)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **Make**: Usually pre-installed on Unix systems

### 2. Environment Setup

```bash
# Clone the repository
git clone <repository-url>
cd crypto-bubble-map-be

# Copy environment configuration
cp .env.example .env

# Install dependencies
go mod download

# Setup development environment
make setup
```

### 3. Database Setup

The application uses multiple databases:

- **Neo4j**: Graph database for wallet relationships
- **MongoDB**: Document store for transactions
- **PostgreSQL**: Relational database for user data
- **Redis**: Cache and session store

```bash
# Start all databases with Docker
make up-db

# Or start individual services
docker-compose up neo4j mongodb postgresql redis
```

### 4. Configuration

Edit `.env` file with your database credentials:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_MODE=debug

# Database Configuration
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=password

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=crypto_bubble_map

POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=5432
POSTGRESQL_DATABASE=crypto_bubble_map
POSTGRESQL_USERNAME=postgres
POSTGRESQL_PASSWORD=password

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

## 🚀 Running the Application

### Development Mode

```bash
# Run with hot reload (if air is installed)
make dev

# Or run directly
make run

# Build and run
make build && ./bin/server
```

### Production Mode

```bash
# Build optimized binary
make build-prod

# Run with production settings
ENV=production ./bin/server
```

### Docker

```bash
# Build and run with Docker Compose
make up

# View logs
make logs

# Stop services
make down
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Test server endpoints
./scripts/test-server.sh
```

### Test Structure

```
tests/
├── unit/           # Unit tests
├── integration/    # Integration tests
└── fixtures/       # Test data
```

## 📊 Database Operations

### Migrations

```bash
# Run PostgreSQL migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create name=add_user_table
```

### Seeding Data

```bash
# Seed development data
make seed

# Seed test data
make seed-test
```

## 🔧 Development Workflow

### 1. Adding New Features

1. **Create Domain Entity** (if needed):
   ```go
   // internal/domain/entity/new_entity.go
   type NewEntity struct {
       ID   string `json:"id"`
       Name string `json:"name"`
   }
   ```

2. **Define Repository Interface**:
   ```go
   // internal/domain/repository/new_repository.go
   type NewRepository interface {
       Create(ctx context.Context, entity *NewEntity) error
       GetByID(ctx context.Context, id string) (*NewEntity, error)
   }
   ```

3. **Implement Repository**:
   ```go
   // internal/infrastructure/repository/new_repository_impl.go
   type newRepositoryImpl struct {
       db *database.Client
   }
   ```

4. **Add GraphQL Schema**:
   ```graphql
   # graph/schema.graphqls
   type NewEntity {
       id: ID!
       name: String!
   }
   ```

5. **Implement Resolver**:
   ```go
   // graph/resolvers.go
   func (r *queryResolver) NewEntity(ctx context.Context, id string) (*entity.NewEntity, error) {
       return r.newRepo.GetByID(ctx, id)
   }
   ```

### 2. Code Style

- Follow Go conventions and use `gofmt`
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused

### 3. Git Workflow

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes and commit
git add .
git commit -m "feat: add new feature"

# Push and create PR
git push origin feature/new-feature
```

## 🔍 Debugging

### Logging

The application uses structured logging with Zap:

```go
logger.Info("Processing request",
    zap.String("user_id", userID),
    zap.String("action", "create_wallet"),
)
```

### Debug Mode

Set `SERVER_MODE=debug` in `.env` for detailed logging.

### Database Debugging

```bash
# Neo4j Browser
open http://localhost:7474

# MongoDB Compass
open mongodb://localhost:27017

# PostgreSQL
psql -h localhost -U postgres -d crypto_bubble_map
```

## 📈 Performance

### Profiling

```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Monitoring

- Health check: `GET /health`
- Readiness check: `GET /ready`
- Metrics: `GET /metrics` (if enabled)

## 🚀 Deployment

### Building

```bash
# Build for current platform
make build

# Build for Linux (for Docker)
make build-linux

# Build with version info
make build VERSION=v1.0.0
```

### Docker

```bash
# Build Docker image
make docker-build

# Push to registry
make docker-push
```

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [Neo4j Documentation](https://neo4j.com/docs/)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📞 Support

For questions or issues:
- Create an issue in the repository
- Contact the development team
- Check the documentation
