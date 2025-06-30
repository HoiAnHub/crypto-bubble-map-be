#!/bin/bash

# Migration runner script for Crypto Bubble Map Backend
# This script runs all database migrations and seeds data

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="$PROJECT_ROOT/migrations"

# Default database connection parameters
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-crypto_bubble_map}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"

MONGODB_HOST="${MONGODB_HOST:-localhost}"
MONGODB_PORT="${MONGODB_PORT:-27017}"
MONGODB_DB="${MONGODB_DB:-crypto_bubble_map}"
MONGODB_USER="${MONGODB_USER:-}"
MONGODB_PASSWORD="${MONGODB_PASSWORD:-}"

NEO4J_HOST="${NEO4J_HOST:-localhost}"
NEO4J_PORT="${NEO4J_PORT:-7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-password}"

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Crypto Bubble Map Migrations  ${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

print_step() {
    echo -e "${YELLOW}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

check_dependencies() {
    print_step "Checking dependencies..."
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        print_error "psql is not installed. Please install PostgreSQL client."
        exit 1
    fi
    
    # Check if mongosh is available
    if ! command -v mongosh &> /dev/null; then
        print_error "mongosh is not installed. Please install MongoDB Shell."
        exit 1
    fi
    
    # Check if cypher-shell is available
    if ! command -v cypher-shell &> /dev/null; then
        print_error "cypher-shell is not installed. Please install Neo4j."
        exit 1
    fi
    
    print_success "All dependencies are available"
}

test_connections() {
    print_step "Testing database connections..."
    
    # Test PostgreSQL connection
    print_info "Testing PostgreSQL connection..."
    if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "SELECT 1;" &> /dev/null; then
        print_success "PostgreSQL connection successful"
    else
        print_error "Failed to connect to PostgreSQL"
        exit 1
    fi
    
    # Test MongoDB connection
    print_info "Testing MongoDB connection..."
    if [ -n "$MONGODB_USER" ] && [ -n "$MONGODB_PASSWORD" ]; then
        MONGO_URI="mongodb://$MONGODB_USER:$MONGODB_PASSWORD@$MONGODB_HOST:$MONGODB_PORT/$MONGODB_DB"
    else
        MONGO_URI="mongodb://$MONGODB_HOST:$MONGODB_PORT/$MONGODB_DB"
    fi
    
    if mongosh "$MONGO_URI" --eval "db.runCommand('ping')" &> /dev/null; then
        print_success "MongoDB connection successful"
    else
        print_error "Failed to connect to MongoDB"
        exit 1
    fi
    
    # Test Neo4j connection
    print_info "Testing Neo4j connection..."
    if cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" "RETURN 1;" &> /dev/null; then
        print_success "Neo4j connection successful"
    else
        print_error "Failed to connect to Neo4j"
        exit 1
    fi
}

create_databases() {
    print_step "Creating databases if they don't exist..."
    
    # Create PostgreSQL database
    print_info "Creating PostgreSQL database..."
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE $POSTGRES_DB;" 2>/dev/null || true
    print_success "PostgreSQL database ready"
    
    # MongoDB database is created automatically when first document is inserted
    print_success "MongoDB database will be created automatically"
    
    # Neo4j database is created automatically
    print_success "Neo4j database ready"
}

run_postgresql_migrations() {
    print_step "Running PostgreSQL migrations..."
    
    if [ -f "$MIGRATIONS_DIR/001_initial_schema.sql" ]; then
        print_info "Running initial schema migration..."
        PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$MIGRATIONS_DIR/001_initial_schema.sql"
        print_success "PostgreSQL migrations completed"
    else
        print_error "PostgreSQL migration file not found"
        exit 1
    fi
}

run_mongodb_migrations() {
    print_step "Running MongoDB migrations..."
    
    if [ -f "$MIGRATIONS_DIR/mongodb_seed.js" ]; then
        print_info "Running MongoDB seed script..."
        mongosh "$MONGO_URI" < "$MIGRATIONS_DIR/mongodb_seed.js"
        print_success "MongoDB migrations completed"
    else
        print_error "MongoDB migration file not found"
        exit 1
    fi
}

run_neo4j_migrations() {
    print_step "Running Neo4j migrations..."
    
    if [ -f "$MIGRATIONS_DIR/neo4j_seed.cypher" ]; then
        print_info "Running Neo4j seed script..."
        cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" -f "$MIGRATIONS_DIR/neo4j_seed.cypher"
        print_success "Neo4j migrations completed"
    else
        print_error "Neo4j migration file not found"
        exit 1
    fi
}

verify_migrations() {
    print_step "Verifying migrations..."
    
    # Verify PostgreSQL tables
    print_info "Verifying PostgreSQL tables..."
    table_count=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
    if [ "$table_count" -gt 0 ]; then
        print_success "PostgreSQL tables created successfully ($table_count tables)"
    else
        print_error "No PostgreSQL tables found"
        exit 1
    fi
    
    # Verify MongoDB collections
    print_info "Verifying MongoDB collections..."
    collection_count=$(mongosh "$MONGO_URI" --quiet --eval "db.runCommand('listCollections').cursor.firstBatch.length")
    if [ "$collection_count" -gt 0 ]; then
        print_success "MongoDB collections created successfully ($collection_count collections)"
    else
        print_error "No MongoDB collections found"
        exit 1
    fi
    
    # Verify Neo4j nodes
    print_info "Verifying Neo4j nodes..."
    node_count=$(cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" --format plain "MATCH (n) RETURN count(n);" | tail -n 1)
    if [ "$node_count" -gt 0 ]; then
        print_success "Neo4j nodes created successfully ($node_count nodes)"
    else
        print_error "No Neo4j nodes found"
        exit 1
    fi
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  --skip-deps             Skip dependency checks"
    echo "  --skip-test             Skip connection tests"
    echo "  --postgresql-only       Run only PostgreSQL migrations"
    echo "  --mongodb-only          Run only MongoDB migrations"
    echo "  --neo4j-only            Run only Neo4j migrations"
    echo "  --skip-verify           Skip migration verification"
    echo ""
    echo "Environment Variables:"
    echo "  POSTGRES_HOST           PostgreSQL host (default: localhost)"
    echo "  POSTGRES_PORT           PostgreSQL port (default: 5432)"
    echo "  POSTGRES_DB             PostgreSQL database (default: crypto_bubble_map)"
    echo "  POSTGRES_USER           PostgreSQL user (default: postgres)"
    echo "  POSTGRES_PASSWORD       PostgreSQL password (default: password)"
    echo "  MONGODB_HOST            MongoDB host (default: localhost)"
    echo "  MONGODB_PORT            MongoDB port (default: 27017)"
    echo "  MONGODB_DB              MongoDB database (default: crypto_bubble_map)"
    echo "  NEO4J_HOST              Neo4j host (default: localhost)"
    echo "  NEO4J_PORT              Neo4j port (default: 7687)"
    echo "  NEO4J_USER              Neo4j user (default: neo4j)"
    echo "  NEO4J_PASSWORD          Neo4j password (default: password)"
}

# Main execution
main() {
    local skip_deps=false
    local skip_test=false
    local postgresql_only=false
    local mongodb_only=false
    local neo4j_only=false
    local skip_verify=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            --skip-deps)
                skip_deps=true
                shift
                ;;
            --skip-test)
                skip_test=true
                shift
                ;;
            --postgresql-only)
                postgresql_only=true
                shift
                ;;
            --mongodb-only)
                mongodb_only=true
                shift
                ;;
            --neo4j-only)
                neo4j_only=true
                shift
                ;;
            --skip-verify)
                skip_verify=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_header
    
    if [ "$skip_deps" = false ]; then
        check_dependencies
    fi
    
    if [ "$skip_test" = false ]; then
        test_connections
    fi
    
    create_databases
    
    # Run migrations based on options
    if [ "$postgresql_only" = true ]; then
        run_postgresql_migrations
    elif [ "$mongodb_only" = true ]; then
        run_mongodb_migrations
    elif [ "$neo4j_only" = true ]; then
        run_neo4j_migrations
    else
        run_postgresql_migrations
        run_mongodb_migrations
        run_neo4j_migrations
    fi
    
    if [ "$skip_verify" = false ]; then
        verify_migrations
    fi
    
    echo ""
    print_success "All migrations completed successfully!"
    echo ""
    print_info "Your databases are now ready for the Crypto Bubble Map application."
}

# Run main function with all arguments
main "$@"
