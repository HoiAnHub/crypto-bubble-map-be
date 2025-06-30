#!/bin/bash

# Test script for Crypto Bubble Map Backend
set -e

echo "🚀 Testing Crypto Bubble Map Backend Server"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Server configuration
SERVER_URL="http://localhost:8080"
TIMEOUT=30

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

# Function to check if server is running
check_server() {
    print_status "Checking if server is running..."
    
    if curl -s -f "$SERVER_URL/health" > /dev/null; then
        print_success "Server is running at $SERVER_URL"
        return 0
    else
        print_error "Server is not running at $SERVER_URL"
        return 1
    fi
}

# Function to test health endpoint
test_health() {
    print_status "Testing health endpoint..."
    
    response=$(curl -s "$SERVER_URL/health")
    if echo "$response" | grep -q "healthy"; then
        print_success "Health endpoint is working"
        echo "Response: $response"
    else
        print_error "Health endpoint failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test readiness endpoint
test_readiness() {
    print_status "Testing readiness endpoint..."
    
    response=$(curl -s "$SERVER_URL/ready")
    if echo "$response" | grep -q "ready\|not ready"; then
        print_success "Readiness endpoint is working"
        echo "Response: $response"
    else
        print_error "Readiness endpoint failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test GraphQL endpoint
test_graphql() {
    print_status "Testing GraphQL endpoint..."
    
    response=$(curl -s -X POST "$SERVER_URL/graphql" \
        -H "Content-Type: application/json" \
        -d '{"query": "{ __typename }"}')
    
    if echo "$response" | grep -q "message\|data"; then
        print_success "GraphQL endpoint is working"
        echo "Response: $response"
    else
        print_error "GraphQL endpoint failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test CORS
test_cors() {
    print_status "Testing CORS headers..."
    
    response=$(curl -s -I -X OPTIONS "$SERVER_URL/graphql" \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: POST")
    
    if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
        print_success "CORS is configured"
    else
        print_warning "CORS headers not found"
    fi
}

# Function to test rate limiting
test_rate_limiting() {
    print_status "Testing rate limiting..."
    
    # Make multiple requests quickly
    for i in {1..5}; do
        curl -s "$SERVER_URL/health" > /dev/null
    done
    
    # Check if rate limiting is working (this is a basic test)
    response=$(curl -s -w "%{http_code}" "$SERVER_URL/health")
    if [[ "$response" == *"200"* ]]; then
        print_success "Rate limiting test passed (or limit not reached)"
    else
        print_warning "Rate limiting may be active or server error"
    fi
}

# Function to run all tests
run_all_tests() {
    echo ""
    print_status "Running all endpoint tests..."
    echo ""
    
    local failed=0
    
    # Test each endpoint
    test_health || ((failed++))
    echo ""
    
    test_readiness || ((failed++))
    echo ""
    
    test_graphql || ((failed++))
    echo ""
    
    test_cors || ((failed++))
    echo ""
    
    test_rate_limiting || ((failed++))
    echo ""
    
    # Summary
    echo "=========================================="
    if [ $failed -eq 0 ]; then
        print_success "All tests passed! ✅"
    else
        print_error "$failed test(s) failed ❌"
        return 1
    fi
}

# Function to show server info
show_server_info() {
    print_status "Server Information:"
    echo "  URL: $SERVER_URL"
    echo "  Health: $SERVER_URL/health"
    echo "  Readiness: $SERVER_URL/ready"
    echo "  GraphQL: $SERVER_URL/graphql"
    echo "  Playground: $SERVER_URL/playground"
    echo ""
}

# Function to wait for server to start
wait_for_server() {
    print_status "Waiting for server to start..."
    
    local count=0
    while [ $count -lt $TIMEOUT ]; do
        if check_server 2>/dev/null; then
            return 0
        fi
        
        echo -n "."
        sleep 1
        ((count++))
    done
    
    echo ""
    print_error "Server did not start within $TIMEOUT seconds"
    return 1
}

# Main execution
main() {
    show_server_info
    
    # Check if server is already running
    if ! check_server 2>/dev/null; then
        print_warning "Server is not running. Please start it first:"
        echo "  make run"
        echo "  # or"
        echo "  make up"
        echo ""
        print_status "Waiting for server to start..."
        wait_for_server || exit 1
    fi
    
    # Run all tests
    run_all_tests
    
    echo ""
    print_status "Test completed!"
}

# Handle command line arguments
case "${1:-}" in
    "health")
        test_health
        ;;
    "ready")
        test_readiness
        ;;
    "graphql")
        test_graphql
        ;;
    "cors")
        test_cors
        ;;
    "rate")
        test_rate_limiting
        ;;
    "wait")
        wait_for_server
        ;;
    "info")
        show_server_info
        ;;
    *)
        main
        ;;
esac
