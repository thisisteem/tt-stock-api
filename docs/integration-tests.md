# Integration Tests

This document describes how to run and understand the integration tests for the authentication handlers.

## Overview

The integration tests (`internal/auth/handler_integration_test.go`) test the complete HTTP request/response cycle for authentication endpoints with real database connections. Unlike unit tests that use mocks, integration tests verify the entire system works together correctly.

## Test Coverage

The integration tests cover:

### Login Endpoint (`POST /auth/login`)
- ✅ Successful login with valid credentials
- ✅ Login with invalid phone number format
- ✅ Login with invalid PIN format  
- ✅ Login with non-existent phone number
- ✅ Login with wrong PIN
- ✅ Database updates (last_login_at timestamp)
- ✅ JWT token generation and validation

### Refresh Token Endpoint (`POST /auth/refresh`)
- ✅ Successful token refresh with valid refresh token
- ✅ Refresh with invalid token
- ✅ Refresh with access token instead of refresh token
- ✅ Refresh with already used (blacklisted) refresh token
- ✅ Token blacklisting verification
- ✅ New token generation and validation

### Logout Endpoint (`POST /auth/logout`)
- ✅ Successful logout with access token only
- ✅ Successful logout with both access and refresh tokens
- ✅ Logout with invalid access token
- ✅ Logout with refresh token in authorization header
- ✅ Logout with already blacklisted token
- ✅ Token blacklisting verification

### Complete Authentication Flow
- ✅ End-to-end flow: login → refresh → logout
- ✅ Token lifecycle management
- ✅ Database state verification

### Additional Scenarios
- ✅ Concurrent authentication requests
- ✅ Token expiration behavior
- ✅ Database cleanup and isolation

## Prerequisites

### Database Setup

Integration tests require a PostgreSQL test database. You can set this up in several ways:

#### Option 1: Local PostgreSQL
```bash
# Create test database
createdb tt_stock_test_db

# Set environment variable
export TEST_DB_URL="postgres://localhost:5432/tt_stock_test_db?sslmode=disable"
```

#### Option 2: Docker PostgreSQL
```bash
# Run PostgreSQL in Docker
docker run --name postgres-test -e POSTGRES_DB=tt_stock_test_db -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:15

# Set environment variable
export TEST_DB_URL="postgres://postgres:password@localhost:5432/tt_stock_test_db?sslmode=disable"
```

#### Option 3: In-Memory Database (SQLite)
For faster tests, you could modify the tests to use SQLite in-memory database, though this requires code changes to support multiple database drivers.

## Running Integration Tests

### Method 1: Using the Script
```bash
# Run all integration tests
./scripts/run-integration-tests.sh

# Or with custom database URL
TEST_DB_URL="your-test-db-url" ./scripts/run-integration-tests.sh
```

### Method 2: Direct Go Test
```bash
# Set test database URL
export TEST_DB_URL="postgres://localhost:5432/tt_stock_test_db?sslmode=disable"

# Run all integration tests
go test -v ./internal/auth -run Integration

# Run specific integration test
go test -v ./internal/auth -run TestLoginEndpoint_Integration

# Run with coverage
go test -v ./internal/auth -run Integration -cover
```

### Method 3: Skip Integration Tests
If no `TEST_DB_URL` is set, integration tests will be automatically skipped:

```bash
# This will skip integration tests
go test -v ./internal/auth
```

## Test Structure

Each integration test follows this pattern:

1. **Setup**: Create test database connection and tables
2. **Seed**: Insert test user data
3. **Execute**: Make HTTP requests to endpoints
4. **Verify**: Check responses and database state
5. **Cleanup**: Remove test data and close connections

## Key Features

### Real Database Operations
- Tests use actual PostgreSQL database
- Verifies database schema and constraints
- Tests transaction handling and data persistence

### Complete HTTP Cycle
- Tests full Fiber application setup
- Verifies request parsing and response formatting
- Tests middleware integration

### Token Lifecycle
- Verifies JWT token generation and validation
- Tests token blacklisting and expiration
- Validates token type enforcement

### Concurrent Testing
- Tests system behavior under concurrent requests
- Verifies database connection handling
- Tests for race conditions

## Troubleshooting

### Database Connection Issues
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Test connection manually
psql "postgres://localhost:5432/tt_stock_test_db?sslmode=disable"
```

### Permission Issues
```bash
# Make script executable
chmod +x scripts/run-integration-tests.sh

# Check database permissions
psql -c "SELECT current_user, current_database();" "postgres://localhost:5432/tt_stock_test_db?sslmode=disable"
```

### Test Failures
- Ensure test database is empty before running tests
- Check that all required environment variables are set
- Verify database schema matches application expectations

## Best Practices

### Test Isolation
- Each test cleans up its data
- Tests don't depend on each other
- Database state is reset between tests

### Performance
- Use transactions for faster cleanup
- Consider using test database templates
- Run integration tests separately from unit tests

### Maintenance
- Keep integration tests focused on critical paths
- Update tests when API contracts change
- Monitor test execution time and optimize as needed

## CI/CD Integration

For continuous integration, you can:

1. **Use Docker Compose** for test database
2. **Set up test database** in CI pipeline
3. **Run integration tests** as separate job
4. **Generate coverage reports** for integration tests

Example GitHub Actions workflow:
```yaml
name: Integration Tests
on: [push, pull_request]
jobs:
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: tt_stock_test_db
          POSTGRES_PASSWORD: password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Run Integration Tests
        env:
          TEST_DB_URL: postgres://postgres:password@localhost:5432/tt_stock_test_db?sslmode=disable
        run: go test -v ./internal/auth -run Integration
```