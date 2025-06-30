#!/bin/bash

# Quick start script for Crypto Bubble Map Backend
set -e

echo "🚀 Crypto Bubble Map Backend - Quick Start"
echo "=========================================="

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

# Check if Docker is available
check_docker() {
    if command -v docker >/dev/null 2>&1; then
        print_success "Docker is available"
        return 0
    else
        print_error "Docker is not available"
        return 1
    fi
}

# Start Neo4j if not running
start_neo4j() {
    print_status "Checking Neo4j..."
    
    if nc -z localhost 7687 2>/dev/null; then
        print_success "Neo4j is already running"
    else
        print_status "Starting Neo4j..."
        if check_docker; then
            docker run -d --name crypto-neo4j \
                -p 7687:7687 -p 7474:7474 \
                --env NEO4J_AUTH=neo4j/password \
                neo4j:latest
            
            print_status "Waiting for Neo4j to start..."
            sleep 10
            
            if nc -z localhost 7687 2>/dev/null; then
                print_success "Neo4j started successfully"
                print_status "Neo4j Browser: http://localhost:7474"
            else
                print_error "Failed to start Neo4j"
                return 1
            fi
        else
            print_error "Cannot start Neo4j without Docker"
            return 1
        fi
    fi
}

# Start Redis if not running
start_redis() {
    print_status "Checking Redis..."
    
    if nc -z localhost 6379 2>/dev/null; then
        print_success "Redis is already running"
    else
        print_status "Starting Redis..."
        if check_docker; then
            docker run -d --name crypto-redis \
                -p 6379:6379 \
                redis:latest
            
            print_status "Waiting for Redis to start..."
            sleep 5
            
            if nc -z localhost 6379 2>/dev/null; then
                print_success "Redis started successfully"
            else
                print_error "Failed to start Redis"
                return 1
            fi
        else
            print_error "Cannot start Redis without Docker"
            return 1
        fi
    fi
}

# Check database connections
check_databases() {
    print_status "Testing database connections..."
    
    if [ -f scripts/test-db-connections.sh ]; then
        ./scripts/test-db-connections.sh
    else
        print_warning "Database test script not found"
    fi
}

# Build the application
build_app() {
    print_status "Building application..."
    
    if make build; then
        print_success "Application built successfully"
    else
        print_error "Failed to build application"
        return 1
    fi
}

# Start the server
start_server() {
    print_status "Starting server..."
    
    print_status "Server will start on http://localhost:8080"
    print_status "GraphQL endpoint: http://localhost:8080/graphql"
    print_status "Health check: http://localhost:8080/health"
    print_status ""
    print_status "Press Ctrl+C to stop the server"
    print_status ""
    
    make run
}

# Show helpful information
show_info() {
    echo ""
    print_success "🎉 Setup completed successfully!"
    echo ""
    print_status "Available endpoints:"
    echo "  • Health check:    http://localhost:8080/health"
    echo "  • Readiness check: http://localhost:8080/ready"
    echo "  • GraphQL API:     http://localhost:8080/graphql"
    echo "  • GraphQL Playground: http://localhost:8080/playground"
    echo ""
    print_status "Database interfaces:"
    echo "  • Neo4j Browser:   http://localhost:7474 (neo4j/password)"
    echo "  • MongoDB Atlas:   Already configured in cloud"
    echo "  • PostgreSQL:      localhost:5433 (hoianhub_user/hoianhub_password)"
    echo "  • Redis:           localhost:6379"
    echo ""
    print_status "Useful commands:"
    echo "  • Test databases:  make test-db"
    echo "  • Test server:     make test-server"
    echo "  • View logs:       docker logs crypto-neo4j"
    echo "  • Stop services:   docker stop crypto-neo4j crypto-redis"
    echo ""
}

# Cleanup function
cleanup() {
    print_status "Cleaning up..."
    docker stop crypto-neo4j crypto-redis 2>/dev/null || true
    docker rm crypto-neo4j crypto-redis 2>/dev/null || true
}

# Main execution
main() {
    # Check prerequisites
    print_status "Checking prerequisites..."
    
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v make >/dev/null 2>&1; then
        print_error "Make is not installed"
        exit 1
    fi
    
    if ! command -v nc >/dev/null 2>&1; then
        print_warning "netcat (nc) not available, some checks may be skipped"
    fi
    
    print_success "Prerequisites check passed"
    echo ""
    
    # Start databases
    start_neo4j || exit 1
    echo ""
    
    start_redis || exit 1
    echo ""
    
    # Test connections
    check_databases
    echo ""
    
    # Build application
    build_app || exit 1
    echo ""
    
    # Show info
    show_info
    
    # Start server
    start_server
}

# Handle command line arguments
case "${1:-}" in
    "start")
        main
        ;;
    "stop")
        cleanup
        ;;
    "info")
        show_info
        ;;
    "databases"|"db")
        start_neo4j
        start_redis
        check_databases
        ;;
    *)
        echo "Usage: $0 [start|stop|info|db]"
        echo ""
        echo "Commands:"
        echo "  start    - Full setup and start server"
        echo "  stop     - Stop and cleanup Docker containers"
        echo "  info     - Show helpful information"
        echo "  db       - Start databases only"
        echo ""
        echo "Default: start"
        echo ""
        main
        ;;
esac
