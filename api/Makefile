# TT Stock API Makefile
# Provides convenient commands for development, testing, and deployment

.PHONY: help build run test clean docker-build docker-run docker-dev docker-stop docker-logs

# Default target
help: ## Show this help message
	@echo "TT Stock API - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development Commands
build: ## Build the application
	@echo "Building TT Stock API..."
	go build -o bin/ttstock-api src/main.go

run: ## Run the application locally
	@echo "Running TT Stock API..."
	go run src/main.go

test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

test-performance: ## Run performance tests only
	@echo "Running performance tests..."
	go test -v ./tests/performance/...

test-load: ## Run load tests only
	@echo "Running load tests..."
	go test -v ./tests/load/...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Docker Commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t ttstock-api .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env ttstock-api

docker-dev: ## Start development environment with Docker Compose
	@echo "Starting development environment..."
	docker compose -f docker-compose.dev.yml up -d

docker-prod: ## Start production environment with Docker Compose
	@echo "Starting production environment..."
	docker compose up -d

docker-stop: ## Stop all Docker containers
	@echo "Stopping Docker containers..."
	docker compose -f docker-compose.dev.yml down
	docker compose down

docker-logs: ## Show Docker container logs
	@echo "Showing Docker logs..."
	docker compose -f docker-compose.dev.yml logs -f

docker-clean: ## Clean up Docker resources
	@echo "Cleaning up Docker resources..."
	docker compose -f docker-compose.dev.yml down -v
	docker compose down -v
	docker system prune -f

# Database Commands
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	go run src/main.go migrate

db-seed: ## Seed database with test data
	@echo "Seeding database..."
	go run src/main.go seed

db-reset: ## Reset database (drop, create, migrate, seed)
	@echo "Resetting database..."
	go run src/main.go db-reset

# Development Setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	cp .env.example .env
	@echo "Please edit .env file with your configuration"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Documentation
docs: ## Generate API documentation
	@echo "Generating API documentation..."
	swag init -g src/main.go -o docs/

docs-serve: ## Serve documentation locally
	@echo "Serving documentation at http://localhost:8080/swagger/index.html"
	@echo "Make sure the API is running first"

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Health Checks
health: ## Check API health
	@echo "Checking API health..."
	curl -f http://localhost:8080/health || echo "API is not running"

health-detailed: ## Check detailed API health
	@echo "Checking detailed API health..."
	curl -f http://localhost:8080/health/detailed || echo "API is not running"

# Development Workflow
dev: docker-dev ## Start full development environment
	@echo "Development environment started!"
	@echo "API: http://localhost:8080"
	@echo "pgAdmin: http://localhost:5050 (admin@ttstock.dev / admin123)"
	@echo "Redis Commander: http://localhost:8081 (admin / admin123)"
	@echo "Mailhog: http://localhost:8025"

dev-stop: docker-stop ## Stop development environment
	@echo "Development environment stopped!"

# Production Commands
prod-build: ## Build production Docker image
	@echo "Building production Docker image..."
	docker build -t ttstock-api:latest .

prod-build-linux: ## Build production Docker image for Linux AMD64
	@echo "Building production Docker image for Linux AMD64..."
	docker build --platform linux/amd64 -t ttstock-api:latest .

prod-build-multi: ## Build production Docker image for multiple platforms
	@echo "Building production Docker image for multiple platforms..."
	docker buildx build --platform linux/amd64,linux/arm64 -t ttstock-api:latest .

prod-push: ## Push production Docker image to Docker Hub
	@echo "Pushing production Docker image to Docker Hub..."
	docker tag ttstock-api:latest thisisteem/ttstock-api:latest
	docker push thisisteem/ttstock-api:latest

prod-build-push: prod-build-linux prod-push ## Build for Linux and push to Docker Hub
	@echo "Production image built and pushed successfully!"

prod-deploy: ## Deploy to production
	@echo "Deploying to production..."
	docker compose up -d

# Monitoring
logs: ## Show application logs
	@echo "Showing application logs..."
	docker compose -f docker-compose.dev.yml logs -f api-dev

monitor: ## Monitor all services
	@echo "Monitoring all services..."
	docker compose -f docker-compose.dev.yml ps

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	golangci-lint run --enable gosec

# Performance
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. ./tests/performance/...

load-test: ## Run load tests
	@echo "Running load tests..."
	go test -v ./tests/load/...

# Quick Commands
quick-start: setup docker-dev ## Quick start for new developers
	@echo "Quick start completed!"
	@echo "API: http://localhost:8080"
	@echo "pgAdmin: http://localhost:5050"

quick-test: test-unit test-integration ## Quick test suite
	@echo "Quick tests completed!"

quick-clean: clean docker-clean ## Quick cleanup
	@echo "Quick cleanup completed!"