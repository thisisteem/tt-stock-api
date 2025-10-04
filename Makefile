# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=tt-stock-api
BINARY_PATH=./bin/$(BINARY_NAME)

# Default target
.DEFAULT_GOAL := help

# Build the application
build: $(BINARY_PATH)
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -ldflags="-s -w" -o $(BINARY_PATH) ./cmd/api
	@echo "Build completed: $(BINARY_PATH)"

# Build for production with optimizations
build-prod: $(BINARY_PATH)
	@echo "Building $(BINARY_NAME) for production..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w -X main.version=$(shell git describe --tags --always)" -o $(BINARY_PATH) ./cmd/api
	@echo "Production build completed: $(BINARY_PATH)"

# Run the application
run:
	@echo "Starting $(BINARY_NAME)..."
	$(GOCMD) run ./cmd/api/main.go

# Run the application with hot reload (requires air to be installed)
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf ./bin
	@echo "Clean completed"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

# Run tests with detailed coverage report
test-coverage-html:
	@echo "Generating HTML coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests in watch mode (requires gotestsum to be installed)
test-watch:
	@echo "Running tests in watch mode..."
	@if command -v gotestsum > /dev/null; then \
		gotestsum --watch ./...; \
	else \
		echo "gotestsum not installed. Install with: go install gotest.tools/gotestsum@latest"; \
		$(MAKE) test; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Code formatted"

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...
	@echo "Code vetted"

# Lint code (requires golangci-lint to be installed)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from: https://golangci-lint.run/usage/install/"; \
		$(MAKE) vet; \
	fi

# Security audit (requires gosec to be installed)
security:
	@echo "Running security audit..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Database migration up (creates tables)
migrate-up:
	@echo "Running database migrations (up)..."
	@echo "Note: This application uses embedded migrations in the code."
	@echo "Tables will be created automatically when the application starts."
	@echo "To manually create tables, run: make run (tables are created on startup)"

# Database migration down (drops tables) - WARNING: This will delete all data
migrate-down:
	@echo "WARNING: This will drop all tables and delete all data!"
	@echo "This is a destructive operation. Make sure you have backups."
	@read -p "Are you sure you want to continue? (y/N): " confirm && [ "$$confirm" = "y" ]
	@echo "Dropping tables..."
	@if ! docker ps | grep -q tt-stock-postgres; then \
		echo "❌ PostgreSQL container not running. Start it with: make docker-dev"; \
		exit 1; \
	fi
	@docker-compose exec -T postgres psql -U tt_stock_user -d tt_stock_db -c \
		"DROP TABLE IF EXISTS token_blacklist CASCADE; DROP TABLE IF EXISTS users CASCADE;" \
		&& echo "✅ Tables dropped successfully" \
		|| echo "❌ Failed to drop tables"

# Database reset (drop and recreate tables)
migrate-reset: migrate-down
	@echo "Recreating tables..."
	$(MAKE) run &
	@sleep 3
	@pkill -f "$(BINARY_NAME)" || true
	@echo "Database reset completed"

# Create a new user (uses Docker containers)
create-user:
	@echo "Creating a new user..."
	@if [ -z "$(PHONE)" ] || [ -z "$(PIN)" ]; then \
		echo "Usage: make create-user PHONE=0123456789 PIN=123456"; \
		exit 1; \
	fi
	@if ! docker ps | grep -q tt-stock-postgres; then \
		echo "❌ PostgreSQL container not running. Start it with: make docker-dev"; \
		exit 1; \
	fi
	@echo "Creating user with phone: $(PHONE)"
	@docker-compose exec -T postgres psql -U tt_stock_user -d tt_stock_db -c \
		"INSERT INTO users (phone_number, pin_hash, created_at, updated_at) VALUES ('$(PHONE)', crypt('$(PIN)', gen_salt('bf', 12)), NOW(), NOW());" \
		&& echo "✅ User created successfully" \
		|| echo "❌ Failed to create user (user may already exist)"

# Check code quality (runs multiple checks)
check: fmt vet lint test
	@echo "All quality checks completed"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) gotest.tools/gotestsum@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Development tools installed"
	@echo "Note: You may also want to install golangci-lint from: https://golangci-lint.run/usage/install/"

# Create binary directory
$(BINARY_PATH):
	@mkdir -p ./bin

# Docker Commands
docker-build:
	@echo "Building Docker images..."
	docker build --target production -t tt-stock-api:latest .
	docker build --target development -t tt-stock-api:dev .
	@echo "Docker images built successfully"

docker-build-prod:
	@echo "Building production Docker image..."
	docker build --target production -t tt-stock-api:latest .
	@echo "Production Docker image built successfully"

docker-build-dev:
	@echo "Building development Docker image..."
	docker build --target development -t tt-stock-api:dev .
	@echo "Development Docker image built successfully"

docker-up:
	@echo "Starting Docker services..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Please create one based on .env.example"; \
		echo "Required variables: JWT_SECRET, DB_PASSWORD"; \
		exit 1; \
	fi
	docker-compose up -d
	@echo "Docker services started. Use 'make docker-logs' to view logs."

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Docker services stopped"

docker-dev:
	@echo "Starting development environment..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Please create one based on .env.example"; \
		echo "Required variables: JWT_SECRET, DB_PASSWORD"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
	@echo "Development environment started"

docker-dev-build:
	@echo "Building and starting development environment..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Please create one based on .env.example"; \
		echo "Required variables: JWT_SECRET, DB_PASSWORD"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

docker-logs:
	@echo "Showing Docker container logs..."
	docker-compose logs -f

docker-logs-api:
	@echo "Showing API container logs..."
	docker-compose logs -f api

docker-logs-db:
	@echo "Showing database container logs..."
	docker-compose logs -f postgres

docker-exec-api:
	@echo "Accessing API container shell..."
	docker-compose exec api sh

docker-exec-db:
	@echo "Accessing database container..."
	docker-compose exec postgres psql -U $(shell grep DB_USER .env | cut -d '=' -f2) -d $(shell grep DB_NAME .env | cut -d '=' -f2)

docker-test:
	@echo "Running tests in Docker container..."
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml exec api go test ./...

docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f
	@echo "Docker cleanup completed"

docker-clean-all:
	@echo "WARNING: This will remove all Docker containers, images, and volumes!"
	@read -p "Are you sure you want to continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker-compose down -v --remove-orphans
	docker system prune -a -f --volumes
	@echo "Complete Docker cleanup completed"

docker-reset:
	@echo "Resetting Docker environment..."
	$(MAKE) docker-down
	docker-compose down -v
	$(MAKE) docker-up
	@echo "Docker environment reset completed"

# Show help
help:
	@echo "TT Stock API - Available Make Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the application"
	@echo "  build-prod     Build for production with optimizations"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Development Commands:"
	@echo "  run            Run the application"
	@echo "  dev            Run with hot reload (requires air)"
	@echo "  deps           Download and tidy dependencies"
	@echo "  deps-update    Update all dependencies"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build       Build both production and development Docker images"
	@echo "  docker-build-prod  Build production Docker image only"
	@echo "  docker-build-dev   Build development Docker image only"
	@echo "  docker-up          Start Docker services (production mode)"
	@echo "  docker-down        Stop Docker services"
	@echo "  docker-dev         Start development environment with hot reload"
	@echo "  docker-dev-build   Build and start development environment"
	@echo "  docker-logs        Show logs from all containers"
	@echo "  docker-logs-api    Show API container logs only"
	@echo "  docker-logs-db     Show database container logs only"
	@echo "  docker-exec-api    Access API container shell"
	@echo "  docker-exec-db     Access database container"
	@echo "  docker-test        Run tests in Docker container"
	@echo "  docker-clean       Clean up Docker resources"
	@echo "  docker-clean-all   Remove all Docker containers, images, and volumes"
	@echo "  docker-reset       Reset Docker environment"
	@echo ""
	@echo "Testing Commands:"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo "  test-coverage-html  Generate HTML coverage report"
	@echo "  test-watch     Run tests in watch mode (requires gotestsum)"
	@echo ""
	@echo "Code Quality Commands:"
	@echo "  fmt            Format code"
	@echo "  vet            Vet code"
	@echo "  lint           Lint code (requires golangci-lint)"
	@echo "  security       Run security audit (requires gosec)"
	@echo "  check          Run all quality checks"
	@echo ""
	@echo "Database Commands:"
	@echo "  migrate-up     Create database tables"
	@echo "  migrate-down   Drop database tables (WARNING: destructive)"
	@echo "  migrate-reset  Reset database (drop and recreate)"
	@echo "  create-user    Create a new user (Usage: make create-user PHONE=0123456789 PIN=123456)"
	@echo ""
	@echo "Setup Commands:"
	@echo "  install-tools  Install development tools"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Environment Variables:"
	@echo "  JWT_SECRET     Secret key for JWT tokens (required, min 32 chars)"
	@echo "  DB_PASSWORD    Database password (required)"
	@echo "  DB_NAME        Database name (default: tt_stock_db)"
	@echo "  DB_USER        Database user (default: tt_stock_user)"
	@echo "  DB_HOST        Database host (default: postgres for Docker)"
	@echo "  DB_PORT        Database port (default: 5432)"
	@echo "  PORT           Server port (default: 8080)"
	@echo "  ENV            Environment (development/production)"

.PHONY: build build-prod run dev clean test test-coverage test-coverage-html test-watch deps deps-update fmt vet lint security migrate-up migrate-down migrate-reset create-user check install-tools docker-build docker-build-prod docker-build-dev docker-up docker-down docker-dev docker-dev-build docker-logs docker-logs-api docker-logs-db docker-exec-api docker-exec-db docker-test docker-clean docker-clean-all docker-reset help