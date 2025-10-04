#!/bin/sh
# wait-for-postgres.sh - Wait for PostgreSQL to be ready

set -e

host="$1"
port="$2"
user="$3"
database="$4"
shift 4
cmd="$@"

# Default values
host=${host:-postgres}
port=${port:-5432}
user=${user:-postgres}
database=${database:-postgres}

# Maximum wait time in seconds (5 minutes)
max_wait=300
wait_time=0
retry_interval=1

echo "Waiting for PostgreSQL at $host:$port..."

# Function to test PostgreSQL connection
test_postgres() {
    PGPASSWORD="$DB_PASSWORD" pg_isready -h "$host" -p "$port" -U "$user" -d "$database" >/dev/null 2>&1
}

# Wait for PostgreSQL with exponential backoff
while ! test_postgres; do
    if [ $wait_time -ge $max_wait ]; then
        echo "❌ Timeout: PostgreSQL at $host:$port did not become ready within $max_wait seconds"
        exit 1
    fi
    
    echo "PostgreSQL is unavailable - waiting $retry_interval seconds (waited ${wait_time}s total)..."
    sleep $retry_interval
    
    wait_time=$((wait_time + retry_interval))
    
    # Exponential backoff: double the retry interval, max 16 seconds
    if [ $retry_interval -lt 16 ]; then
        retry_interval=$((retry_interval * 2))
    fi
done

echo "✅ PostgreSQL at $host:$port is ready!"

# Execute the command if provided
if [ -n "$cmd" ]; then
    echo "Executing command: $cmd"
    exec $cmd
fi