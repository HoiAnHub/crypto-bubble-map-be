#!/bin/bash

# Test database connections for Crypto Bubble Map Backend
set -e

echo "🔍 Testing Database Connections"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    print_status "Loaded environment variables from .env"
else
    print_warning ".env file not found, using default values"
fi

# Test Neo4j connection
test_neo4j() {
    print_status "Testing Neo4j connection..."
    
    # Extract connection details
    NEO4J_HOST=$(echo $NEO4J_URI | sed 's/neo4j:\/\///' | cut -d':' -f1)
    NEO4J_PORT=$(echo $NEO4J_URI | sed 's/neo4j:\/\///' | cut -d':' -f2)
    
    if command -v nc >/dev/null 2>&1; then
        if nc -z $NEO4J_HOST $NEO4J_PORT 2>/dev/null; then
            print_success "Neo4j is reachable at $NEO4J_HOST:$NEO4J_PORT"
        else
            print_error "Cannot connect to Neo4j at $NEO4J_HOST:$NEO4J_PORT"
            print_warning "Make sure Neo4j is running: docker run -p 7687:7687 -p 7474:7474 neo4j"
            return 1
        fi
    else
        print_warning "netcat (nc) not available, skipping Neo4j port check"
    fi
    
    # Test with cypher-shell if available
    if command -v cypher-shell >/dev/null 2>&1; then
        if cypher-shell -a $NEO4J_URI -u $NEO4J_USERNAME -p $NEO4J_PASSWORD "RETURN 1" >/dev/null 2>&1; then
            print_success "Neo4j authentication successful"
        else
            print_error "Neo4j authentication failed"
            return 1
        fi
    else
        print_warning "cypher-shell not available, skipping Neo4j auth test"
    fi
}

# Test MongoDB connection
test_mongodb() {
    print_status "Testing MongoDB connection..."
    
    # Test with mongosh if available
    if command -v mongosh >/dev/null 2>&1; then
        if mongosh "$MONGO_URI" --eval "db.runCommand('ping')" >/dev/null 2>&1; then
            print_success "MongoDB connection successful"
        else
            print_error "Cannot connect to MongoDB"
            print_warning "Check your MongoDB Atlas connection string"
            return 1
        fi
    elif command -v mongo >/dev/null 2>&1; then
        if mongo "$MONGO_URI" --eval "db.runCommand('ping')" >/dev/null 2>&1; then
            print_success "MongoDB connection successful"
        else
            print_error "Cannot connect to MongoDB"
            return 1
        fi
    else
        print_warning "MongoDB client not available, skipping connection test"
        print_status "MongoDB URI: ${MONGO_URI:0:50}..."
    fi
}

# Test PostgreSQL connection
test_postgresql() {
    print_status "Testing PostgreSQL connection..."
    
    if command -v psql >/dev/null 2>&1; then
        export PGPASSWORD=$POSTGRES_PASSWORD
        if psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1;" >/dev/null 2>&1; then
            print_success "PostgreSQL connection successful"
        else
            print_error "Cannot connect to PostgreSQL"
            print_warning "Make sure PostgreSQL is running on port $POSTGRES_PORT"
            return 1
        fi
        unset PGPASSWORD
    else
        print_warning "psql not available, testing with nc..."
        if command -v nc >/dev/null 2>&1; then
            if nc -z $POSTGRES_HOST $POSTGRES_PORT 2>/dev/null; then
                print_success "PostgreSQL is reachable at $POSTGRES_HOST:$POSTGRES_PORT"
            else
                print_error "Cannot connect to PostgreSQL at $POSTGRES_HOST:$POSTGRES_PORT"
                return 1
            fi
        else
            print_warning "Cannot test PostgreSQL connection (no psql or nc)"
        fi
    fi
}

# Test Redis connection
test_redis() {
    print_status "Testing Redis connection..."
    
    if command -v redis-cli >/dev/null 2>&1; then
        if [ -n "$REDIS_PASSWORD" ]; then
            if redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping >/dev/null 2>&1; then
                print_success "Redis connection successful (with auth)"
            else
                print_error "Cannot connect to Redis with authentication"
                return 1
            fi
        else
            if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping >/dev/null 2>&1; then
                print_success "Redis connection successful"
            else
                print_error "Cannot connect to Redis"
                print_warning "Make sure Redis is running: docker run -p 6379:6379 redis"
                return 1
            fi
        fi
    else
        print_warning "redis-cli not available, testing with nc..."
        if command -v nc >/dev/null 2>&1; then
            if nc -z $REDIS_HOST $REDIS_PORT 2>/dev/null; then
                print_success "Redis is reachable at $REDIS_HOST:$REDIS_PORT"
            else
                print_error "Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
                return 1
            fi
        else
            print_warning "Cannot test Redis connection (no redis-cli or nc)"
        fi
    fi
}

# Show configuration summary
show_config() {
    print_status "Database Configuration Summary:"
    echo "  Neo4j:      $NEO4J_URI"
    echo "  MongoDB:    ${MONGO_URI:0:50}..."
    echo "  PostgreSQL: $POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB"
    echo "  Redis:      $REDIS_HOST:$REDIS_PORT"
    echo ""
}

# Test all databases
test_all() {
    local failed=0
    
    show_config
    
    test_neo4j || ((failed++))
    echo ""
    
    test_mongodb || ((failed++))
    echo ""
    
    test_postgresql || ((failed++))
    echo ""
    
    test_redis || ((failed++))
    echo ""
    
    echo "==============================="
    if [ $failed -eq 0 ]; then
        print_success "All database connections successful! ✅"
    else
        print_error "$failed database connection(s) failed ❌"
        echo ""
        print_status "Quick setup commands:"
        echo "  # Start Neo4j"
        echo "  docker run -d -p 7687:7687 -p 7474:7474 --env NEO4J_AUTH=neo4j/password neo4j"
        echo ""
        echo "  # Start Redis"
        echo "  docker run -d -p 6379:6379 redis"
        echo ""
        echo "  # PostgreSQL should be running on port 5433 with your credentials"
        echo "  # MongoDB Atlas should be accessible with your connection string"
        return 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    "neo4j")
        test_neo4j
        ;;
    "mongodb"|"mongo")
        test_mongodb
        ;;
    "postgresql"|"postgres")
        test_postgresql
        ;;
    "redis")
        test_redis
        ;;
    "config")
        show_config
        ;;
    *)
        test_all
        ;;
esac
