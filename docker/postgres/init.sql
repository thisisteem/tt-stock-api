-- PostgreSQL initialization script for tt-stock-api
-- This script runs automatically when the PostgreSQL container starts for the first time

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the application database if it doesn't exist
-- Note: The database is already created by POSTGRES_DB environment variable
-- This is just for documentation and potential future use

-- Create the application user if it doesn't exist
-- Note: The user is already created by POSTGRES_USER environment variable
-- This is just for documentation and potential future use

-- Set proper permissions
-- Grant all privileges on the database to the application user
GRANT ALL PRIVILEGES ON DATABASE tt_stock_db TO tt_stock_user;

-- Grant usage on schema
GRANT USAGE ON SCHEMA public TO tt_stock_user;
GRANT CREATE ON SCHEMA public TO tt_stock_user;

-- Grant privileges on all tables in public schema (for future tables)
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO tt_stock_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO tt_stock_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO tt_stock_user;

-- Set up proper authentication method
-- This ensures secure password authentication
-- Note: This is handled by POSTGRES_INITDB_ARGS in docker-compose.yml

-- Log initialization completion
DO $$
BEGIN
    RAISE NOTICE 'TT Stock API database initialization completed successfully';
    RAISE NOTICE 'Database: tt_stock_db';
    RAISE NOTICE 'User: tt_stock_user';
    RAISE NOTICE 'Extensions enabled: uuid-ossp, pgcrypto';
END $$;