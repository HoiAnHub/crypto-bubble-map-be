.PHONY: help setup build run test clean generate up down logs status

# Default target
help:
	@echo "Available commands:"
	@echo "  setup     - Initialize project dependencies"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application locally"
	@echo "  test      - Run tests"
	@echo "  generate  - Generate GraphQL code"
	@echo "  up        - Start with Docker Compose"
	@echo "  down      - Stop Docker Compose services"
	@echo "  logs      - View Docker logs"
	@echo "  clean     - Clean build artifacts"

# Setup project
setup:
	@echo "Setting up project..."
	go mod download
	go mod tidy
	cp .env.example .env
	@echo "Project setup complete!"

# Quick start with database setup
quick-start:
	@echo "Quick start with database setup..."
	@./scripts/quick-start.sh

# Build application
build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go
	@echo "Build complete!"

# Run application locally
run:
	@echo "Starting server..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Test database connections
test-db:
	@echo "Testing database connections..."
	@./scripts/test-db-connections.sh

# Test server endpoints
test-server:
	@echo "Testing server endpoints..."
	@./scripts/test-server.sh

# Generate GraphQL code
generate:
	@echo "Generating GraphQL code..."
	go run github.com/99designs/gqlgen generate

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Docker commands
up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

up-build:
	@echo "Building and starting services..."
	docker-compose up -d --build

down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

logs:
	@echo "Viewing Docker logs..."
	docker-compose logs -f

status:
	@echo "Checking service status..."
	docker-compose ps

# Database commands
db-migrate:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go

db-seed:
	@echo "Seeding database..."
	go run cmd/seed/main.go

# Development commands
dev:
	@echo "Starting development server with hot reload..."
	air

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run

# Production commands
deploy:
	@echo "Deploying to production..."
	docker-compose -f docker-compose.prod.yml up -d --build

# Utility commands
deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/99designs/gqlgen@latest
	go install github.com/cosmtrek/air@latest

check-env:
	@echo "Checking environment variables..."
	@if [ ! -f .env ]; then echo "Error: .env file not found. Run 'make setup' first."; exit 1; fi
	@echo "Environment file exists ✓"

# Health check
health:
	@echo "Checking service health..."
	curl -f http://localhost:8080/health || echo "Service is not running"

# GraphQL schema validation
validate-schema:
	@echo "Validating GraphQL schema..."
	go run github.com/99designs/gqlgen validate

# Security scan
security-scan:
	@echo "Running security scan..."
	gosec ./...

# Performance test
perf-test:
	@echo "Running performance tests..."
	go test -bench=. -benchmem ./...

# Docker cleanup
docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f
