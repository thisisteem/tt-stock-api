#!/bin/bash

# Script to run integration tests with a test database
# This script sets up a test database URL and runs the integration tests

set -e

# Default test database configuration (modify as needed)
TEST_DB_HOST=${TEST_DB_HOST:-"localhost"}
TEST_DB_PORT=${TEST_DB_PORT:-"5432"}
TEST_DB_NAME=${TEST_DB_NAME:-"tt_stock_test_db"}
TEST_DB_USER=${TEST_DB_USER:-"postgres"}
TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-""}

echo "Running integration tests with database: $TEST_DB_HOST:$TEST_DB_PORT/$TEST_DB_NAME"

# Export the test database configuration
export TEST_DB_HOST="$TEST_DB_HOST"
export TEST_DB_PORT="$TEST_DB_PORT"
export TEST_DB_NAME="$TEST_DB_NAME"
export TEST_DB_USER="$TEST_DB_USER"
export TEST_DB_PASSWORD="$TEST_DB_PASSWORD"

# Run integration tests
echo "Running authentication handler integration tests..."
go test -v ./internal/auth -run Integration

echo "Integration tests completed!"