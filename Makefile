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
	@if [ -n "$$DB_URL" ]; then \
		psql "$$DB_URL" -c "DROP TABLE IF EXISTS token_blacklist CASCADE; DROP TABLE IF EXISTS users CASCADE;"; \
		echo "Tables dropped successfully"; \
	else \
		echo "DB_URL environment variable not set. Please set it and try again."; \
		exit 1; \
	fi

# Database reset (drop and recreate tables)
migrate-reset: migrate-down
	@echo "Recreating tables..."
	$(MAKE) run &
	@sleep 3
	@pkill -f "$(BINARY_NAME)" || true
	@echo "Database reset completed"

# Create a new user (requires psql and DB_URL environment variable)
create-user:
	@echo "Creating a new user..."
	@if [ -z "$$PHONE" ] || [ -z "$$PIN" ]; then \
		echo "Usage: make create-user PHONE=0123456789 PIN=123456"; \
		exit 1; \
	fi
	@if [ -z "$$DB_URL" ]; then \
		echo "DB_URL environment variable not set. Please set it and try again."; \
		exit 1; \
	fi
	@echo "Creating user with phone: $$PHONE"
	@PIN_HASH=$$(echo -n "$$PIN" | $(GOCMD) run -c 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { hash, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 12); fmt.Print(string(hash)) }' "$$PIN" 2>/dev/null || echo "Error generating hash"); \
	if [ "$$PIN_HASH" = "Error generating hash" ]; then \
		echo "Failed to generate PIN hash. Using psql to create user..."; \
		psql "$$DB_URL" -c "INSERT INTO users (phone_number, pin_hash) VALUES ('$$PHONE', crypt('$$PIN', gen_salt('bf', 12)));"; \
	else \
		psql "$$DB_URL" -c "INSERT INTO users (phone_number, pin_hash) VALUES ('$$PHONE', '$$PIN_HASH');"; \
	fi
	@echo "User created successfully"

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
	@echo "  DB_URL         PostgreSQL connection string"
	@echo "  JWT_SECRET     Secret key for JWT tokens"
	@echo "  PORT           Server port (default: 8080)"
	@echo "  ENV            Environment (development/production)"

.PHONY: build build-prod run dev clean test test-coverage test-coverage-html test-watch deps deps-update fmt vet lint security migrate-up migrate-down migrate-reset create-user check install-tools help