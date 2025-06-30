# Database Migrations and Seed Data

This directory contains database migrations and seed data for the Crypto Bubble Map Backend application.

## Overview

The application uses three databases:
- **PostgreSQL**: User management, watched wallets, alerts
- **MongoDB**: Transactions, security alerts, compliance reports
- **Neo4j**: Wallet networks, relationships, graph analysis

## Files

### Migration Scripts

- `001_initial_schema.sql` - PostgreSQL initial schema and seed data
- `mongodb_seed.js` - MongoDB collections, indexes, and sample data
- `neo4j_seed.cypher` - Neo4j nodes, relationships, and sample data

### Utility Scripts

- `../scripts/run_migrations.sh` - Automated migration runner script

## Quick Start

### Prerequisites

Make sure you have the following installed:
- PostgreSQL client (`psql`)
- MongoDB Shell (`mongosh`)
- Neo4j Shell (`cypher-shell`)

### Running All Migrations

```bash
# Make the script executable
chmod +x scripts/run_migrations.sh

# Run all migrations with default settings
./scripts/run_migrations.sh

# Run with custom database settings
POSTGRES_PASSWORD=mypassword MONGODB_HOST=remote-mongo ./scripts/run_migrations.sh
```

### Running Individual Migrations

#### PostgreSQL

```bash
# Set environment variables
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=crypto_bubble_map
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password

# Run migration
psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f migrations/001_initial_schema.sql
```

#### MongoDB

```bash
# Set environment variables
export MONGODB_HOST=localhost
export MONGODB_PORT=27017
export MONGODB_DB=crypto_bubble_map

# Run migration
mongosh mongodb://$MONGODB_HOST:$MONGODB_PORT/$MONGODB_DB < migrations/mongodb_seed.js
```

#### Neo4j

```bash
# Set environment variables
export NEO4J_HOST=localhost
export NEO4J_PORT=7687
export NEO4J_USER=neo4j
export NEO4J_PASSWORD=password

# Run migration
cypher-shell -a bolt://$NEO4J_HOST:$NEO4J_PORT -u $NEO4J_USER -p $NEO4J_PASSWORD -f migrations/neo4j_seed.cypher
```

## Database Schema

### PostgreSQL Tables

- `users` - User accounts and authentication
- `user_sessions` - Session management
- `watched_wallets` - User-watched wallet addresses
- `watched_wallet_tags` - Tags for categorizing wallets
- `watched_wallet_tag_associations` - Many-to-many wallet-tag relationships
- `wallet_alerts` - Alerts for watched wallets

### MongoDB Collections

- `security_alerts` - Security alerts and compliance violations
- `compliance_reports` - AML/KYC compliance reports
- `transactions` - Transaction data and analysis

### Neo4j Node Types

- `Wallet` - Blockchain wallet addresses
- `Network` - Blockchain networks (Ethereum, Polygon, etc.)
- `Transaction` - Transaction relationships
- `Cluster` - Wallet clusters and groups
- `NetworkStats` - Network statistics

## Sample Data

The migrations include sample data for development and testing:

### PostgreSQL Sample Data

- Admin user account (email: `admin@cryptobubblemap.com`, password: `admin123`)
- Default wallet tags (High Risk, Exchange, DeFi, etc.)

### MongoDB Sample Data

- 3 security alerts with different severity levels
- 2 compliance reports (AML and Risk Assessment)
- 2 sample transactions

### Neo4j Sample Data

- 4 blockchain networks (Ethereum, Polygon, BSC, Arbitrum)
- 5 sample wallets with different types and risk profiles
- 3 transactions showing wallet relationships
- 2 wallet clusters demonstrating network analysis

## Environment Variables

### PostgreSQL
- `POSTGRES_HOST` - Database host (default: localhost)
- `POSTGRES_PORT` - Database port (default: 5432)
- `POSTGRES_DB` - Database name (default: crypto_bubble_map)
- `POSTGRES_USER` - Database user (default: postgres)
- `POSTGRES_PASSWORD` - Database password (default: password)

### MongoDB
- `MONGODB_HOST` - Database host (default: localhost)
- `MONGODB_PORT` - Database port (default: 27017)
- `MONGODB_DB` - Database name (default: crypto_bubble_map)
- `MONGODB_USER` - Database user (optional)
- `MONGODB_PASSWORD` - Database password (optional)

### Neo4j
- `NEO4J_HOST` - Database host (default: localhost)
- `NEO4J_PORT` - Database port (default: 7687)
- `NEO4J_USER` - Database user (default: neo4j)
- `NEO4J_PASSWORD` - Database password (default: password)

## Migration Script Options

The `run_migrations.sh` script supports several options:

```bash
# Show help
./scripts/run_migrations.sh --help

# Skip dependency checks
./scripts/run_migrations.sh --skip-deps

# Skip connection tests
./scripts/run_migrations.sh --skip-test

# Run only PostgreSQL migrations
./scripts/run_migrations.sh --postgresql-only

# Run only MongoDB migrations
./scripts/run_migrations.sh --mongodb-only

# Run only Neo4j migrations
./scripts/run_migrations.sh --neo4j-only

# Skip verification step
./scripts/run_migrations.sh --skip-verify
```

## Troubleshooting

### Common Issues

1. **Connection refused errors**
   - Ensure all database services are running
   - Check firewall settings
   - Verify connection parameters

2. **Authentication failures**
   - Check username/password combinations
   - Ensure users have necessary permissions
   - For PostgreSQL, check `pg_hba.conf` settings

3. **Permission denied errors**
   - Ensure database users have CREATE/INSERT permissions
   - For PostgreSQL, the user needs CREATEDB permission
   - For MongoDB, ensure user has readWrite role

4. **Command not found errors**
   - Install missing database clients
   - Ensure clients are in your PATH
   - Use package managers: `brew install postgresql mongodb/brew/mongodb-community neo4j`

### Verification

After running migrations, you can verify the setup:

```bash
# Check PostgreSQL tables
psql -h localhost -U postgres -d crypto_bubble_map -c "\dt"

# Check MongoDB collections
mongosh crypto_bubble_map --eval "show collections"

# Check Neo4j nodes
cypher-shell -u neo4j -p password "MATCH (n) RETURN labels(n), count(n)"
```

## Development

### Adding New Migrations

1. Create new migration files with incremental numbers
2. Update the migration runner script if needed
3. Test migrations on a clean database
4. Document any new environment variables or dependencies

### Best Practices

- Always backup databases before running migrations in production
- Test migrations on development/staging environments first
- Use transactions where possible to ensure atomicity
- Include rollback scripts for complex migrations
- Document any manual steps required

## Production Deployment

For production deployments:

1. Review and customize all default passwords
2. Use environment variables for sensitive configuration
3. Run migrations during maintenance windows
4. Monitor database performance after migrations
5. Have rollback procedures ready

## Support

For issues with migrations:
1. Check the troubleshooting section above
2. Review database logs for specific error messages
3. Ensure all prerequisites are installed and configured
4. Test individual migration scripts to isolate issues
