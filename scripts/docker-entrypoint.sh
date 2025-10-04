#!/bin/sh
# Docker entrypoint script for tt-stock-api

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to log with timestamp
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ENTRYPOINT] $1"
}

log_info() {
    echo "${BLUE}$(date '+%Y-%m-%d %H:%M:%S') [INFO]${NC} $1"
}

log_warn() {
    echo "${YELLOW}$(date '+%Y-%m-%d %H:%M:%S') [WARN]${NC} $1"
}

log_error() {
    echo "${RED}$(date '+%Y-%m-%d %H:%M:%S') [ERROR]${NC} $1"
}

log_success() {
    echo "${GREEN}$(date '+%Y-%m-%d %H:%M:%S') [SUCCESS]${NC} $1"
}

# Function to wait for PostgreSQL
wait_for_postgres() {
    log_info "Waiting for PostgreSQL to be ready..."
    
    # Use the wait-for-postgres script if available, otherwise use a simple approach
    if [ -f "/usr/local/bin/wait-for-postgres.sh" ]; then
        /usr/local/bin/wait-for-postgres.sh "$DB_HOST" "$DB_PORT" "$DB_USER" "$DB_NAME"
    else
        # Simple wait approach using nc or timeout
        max_attempts=30
        attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
                log_success "PostgreSQL is ready!"
                return 0
            fi
            
            log_info "PostgreSQL not ready yet (attempt $attempt/$max_attempts), waiting 2 seconds..."
            sleep 2
            attempt=$((attempt + 1))
        done
        
        log_error "PostgreSQL did not become ready within the timeout period"
        exit 1
    fi
}

# Function to validate environment variables
validate_environment() {
    log_info "Validating environment configuration..."
    
    # Check if we're in a Go environment (development) or binary environment (production)
    if command -v go >/dev/null 2>&1; then
        # Development environment - run Go validation
        log_info "Running in development mode - using Go validation"
        if ! go run -C /app ./cmd/api -validate-env-only 2>/dev/null; then
            # If the app doesn't support -validate-env-only flag, we'll let it validate during startup
            log_warn "Environment validation will be performed during application startup"
        fi
    else
        # Production environment - basic validation
        log_info "Running in production mode - performing basic validation"
        
        # Check critical environment variables
        if [ -z "$JWT_SECRET" ]; then
            log_error "JWT_SECRET environment variable is required"
            exit 1
        fi
        
        if [ -z "$DB_PASSWORD" ]; then
            log_error "DB_PASSWORD environment variable is required"
            exit 1
        fi
        
        if [ ${#JWT_SECRET} -lt 32 ]; then
            log_error "JWT_SECRET must be at least 32 characters long"
            exit 1
        fi
        
        log_success "Basic environment validation passed"
    fi
}

# Function to setup development environment
setup_development() {
    log_info "Setting up development environment..."
    
    # Create tmp directory for Air
    mkdir -p /app/tmp
    
    # Ensure proper permissions
    if [ "$(id -u)" = "0" ]; then
        # Running as root, fix permissions
        chown -R 1000:1000 /app/tmp 2>/dev/null || true
    fi
    
    log_success "Development environment setup complete"
}

# Main entrypoint logic
main() {
    log_info "Starting TT Stock API container..."
    log_info "Environment: ${ENV:-production}"
    
    # Validate environment variables first
    validate_environment
    
    # Wait for PostgreSQL if DB_HOST is set
    if [ -n "$DB_HOST" ] && [ "$DB_HOST" != "localhost" ] && [ "$DB_HOST" != "127.0.0.1" ]; then
        wait_for_postgres
    else
        log_info "Skipping PostgreSQL wait (DB_HOST not set or is localhost)"
    fi
    
    # Setup development environment if needed
    if [ "$ENV" = "development" ] || [ "$HOT_RELOAD" = "true" ]; then
        setup_development
    fi
    
    log_success "Container initialization complete"
    log_info "Executing command: $*"
    
    # Execute the main command
    exec "$@"
}

# Handle special cases
case "$1" in
    --help|-h)
        echo "TT Stock API Docker Entrypoint"
        echo ""
        echo "This script initializes the container environment and starts the application."
        echo ""
        echo "Environment variables:"
        echo "  JWT_SECRET     - JWT signing secret (required, min 32 chars)"
        echo "  DB_PASSWORD    - Database password (required)"
        echo "  DB_HOST        - Database host (default: postgres)"
        echo "  DB_PORT        - Database port (default: 5432)"
        echo "  DB_NAME        - Database name (default: tt_stock_db)"
        echo "  DB_USER        - Database user (default: tt_stock_user)"
        echo "  ENV            - Environment (development/production)"
        echo "  HOT_RELOAD     - Enable hot reload (true/false)"
        echo ""
        exit 0
        ;;
    --validate-only)
        validate_environment
        log_success "Environment validation completed successfully"
        exit 0
        ;;
    *)
        # Normal startup
        main "$@"
        ;;
esac