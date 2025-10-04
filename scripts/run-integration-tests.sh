#!/bin/bash

# Script to run integration tests with a test database
# This script sets up a test database URL and runs the integration tests

set -e

# Default test database URL (modify as needed)
TEST_DB_URL=${TEST_DB_URL:-"postgres://localhost:5432/tt_stock_test_db?sslmode=disable"}

echo "Running integration tests with database: $TEST_DB_URL"

# Export the test database URL
export TEST_DB_URL="$TEST_DB_URL"

# Run integration tests
echo "Running authentication handler integration tests..."
go test -v ./internal/auth -run Integration

echo "Integration tests completed!"