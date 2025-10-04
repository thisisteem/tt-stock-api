# Docker Setup Guide

This guide explains how to run the TT Stock API using Docker and Docker Compose.

## Prerequisites

- Docker (version 20.0 or higher)
- Docker Compose (version 2.0 or higher)
- Git (to clone the repository)

## Quick Start

1. **Clone the repository** (if not already done):
   ```bash
   git clone <repository-url>
   cd tt-stock-api
   ```

2. **Create environment file**:
   ```bash
   cp .env.example .env
   ```

3. **Configure required environment variables**:
   Edit the `.env` file and set the required security variables:
   ```bash
   # Generate a secure JWT secret (32+ characters)
   JWT_SECRET=your-secure-32-character-minimum-jwt-secret-key-here
   
   # Set a strong database password
   DB_PASSWORD=your-secure-database-password-here
   ```

4. **Start the development environment**:
   ```bash
   make docker-dev
   ```

5. **Access the API**:
   - API: http://localhost:8080
   - Health check: http://localhost:8080/health
   - Database: localhost:5432 (in development mode)

## Environment Configuration

### Required Variables (No Defaults)

The application will **refuse to start** if these variables are not set:

- `JWT_SECRET`: Must be at least 32 characters long
- `DB_PASSWORD`: Must be at least 8 characters long

### Optional Variables

- `DB_HOST`: Database host (default: `postgres`)
- `DB_PORT`: Database port (default: `5432`)
- `DB_NAME`: Database name (default: `tt_stock_db`)
- `DB_USER`: Database user (default: `tt_stock_user`)
- `PORT`: API server port (default: `8080`)
- `ENV`: Environment mode (`development` or `production`)

### Generating Secure Values

```bash
# Generate a secure JWT secret
openssl rand -base64 32

# Generate a secure database password
openssl rand -base64 24
```

## Docker Commands

### Development Commands

```bash
# Start development environment with hot reload
make docker-dev

# Build and start development environment
make docker-dev-build

# View logs from all containers
make docker-logs

# View API logs only
make docker-logs-api

# View database logs only
make docker-logs-db
```

### Production Commands

```bash
# Build Docker images
make docker-build

# Start production environment
make docker-up

# Stop all services
make docker-down
```

### Utility Commands

```bash
# Access API container shell
make docker-exec-api

# Access database container
make docker-exec-db

# Run tests in container
make docker-test

# Clean up Docker resources
make docker-clean

# Complete cleanup (removes all Docker data)
make docker-clean-all

# Reset environment (stop, clean volumes, restart)
make docker-reset
```

## Development Workflow

### Hot Reload Development

The development environment includes hot reload using Air:

1. Start development environment:
   ```bash
   make docker-dev
   ```

2. Make changes to your Go code

3. The application will automatically rebuild and restart

4. View logs to see the restart:
   ```bash
   make docker-logs-api
   ```

### Database Access

In development mode, PostgreSQL is exposed on port 5432:

```bash
# Connect using psql
psql -h localhost -p 5432 -U tt_stock_user -d tt_stock_db

# Or use the make command
make docker-exec-db
```

### Running Tests

```bash
# Run tests in the container
make docker-test

# Run tests with coverage
docker-compose -f docker-compose.yml -f docker-compose.dev.yml exec api go test -cover ./...
```

## Production Deployment

### Environment Setup

1. Create production environment file:
   ```bash
   cp .env.example .env.prod
   ```

2. Configure production values:
   ```bash
   ENV=production
   LOG_LEVEL=info
   HOT_RELOAD=false
   JWT_SECRET=<secure-production-secret>
   DB_PASSWORD=<secure-production-password>
   ```

3. Start production environment:
   ```bash
   docker-compose --env-file .env.prod up -d
   ```

### Security Considerations

- **Never use default passwords** in production
- **Use strong, unique secrets** for JWT_SECRET and DB_PASSWORD
- **Limit network access** to the database container
- **Use HTTPS** in production (configure reverse proxy)
- **Regular security updates** for base images

## Health Monitoring

The application provides multiple health check endpoints:

- `/health`: Comprehensive health check with database connectivity
- `/ready`: Kubernetes readiness probe endpoint
- `/live`: Kubernetes liveness probe endpoint

### Health Check Response

```json
{
  "success": true,
  "message": "Health check completed",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.0.0",
    "uptime": "1h30m45s",
    "database": {
      "status": "healthy",
      "connected": true,
      "response_time": "2.5ms"
    },
    "system": {
      "environment": "development",
      "port": "8080"
    }
  }
}
```

## Troubleshooting

### Common Issues

#### 1. Environment Variable Errors

**Error**: `Required environment variable JWT_SECRET is not set`

**Solution**: Ensure your `.env` file contains all required variables:
```bash
JWT_SECRET=your-32-character-minimum-secret-here
DB_PASSWORD=your-secure-password-here
```

#### 2. Database Connection Issues

**Error**: `Failed to connect to database`

**Solution**: 
- Check if PostgreSQL container is running: `docker-compose ps`
- Verify database credentials in `.env` file
- Check container logs: `make docker-logs-db`

#### 3. Port Already in Use

**Error**: `Port 8080 is already in use`

**Solution**: 
- Change the port in `.env` file: `PORT=8081`
- Or stop the conflicting service

#### 4. Permission Issues

**Error**: Permission denied errors in containers

**Solution**:
- Ensure Docker has proper permissions
- On Linux, add user to docker group: `sudo usermod -aG docker $USER`

### Viewing Logs

```bash
# All container logs
make docker-logs

# API container only
make docker-logs-api

# Database container only
make docker-logs-db

# Follow logs in real-time
docker-compose logs -f
```

### Container Status

```bash
# Check container status
docker-compose ps

# Check container health
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
```

### Database Debugging

```bash
# Connect to database
make docker-exec-db

# Check database tables
docker-compose exec postgres psql -U tt_stock_user -d tt_stock_db -c "\dt"

# Check database logs
make docker-logs-db
```

## File Structure

```
├── Dockerfile                 # Multi-stage Docker build
├── docker-compose.yml         # Production Docker Compose
├── docker-compose.dev.yml     # Development overrides
├── .env.example              # Environment template
├── .env.example              # Environment template (works for all setups)
├── .dockerignore             # Docker build exclusions
├── scripts/
│   ├── docker-entrypoint.sh  # Container startup script
│   └── wait-for-postgres.sh  # Database readiness script
└── docker/
    ├── postgres/
    │   └── init.sql          # Database initialization
    └── api/
        └── .air.toml         # Hot reload configuration
```

## Support

If you encounter issues:

1. Check this troubleshooting guide
2. Review container logs: `make docker-logs`
3. Verify environment configuration
4. Check Docker and Docker Compose versions
5. Ensure all required ports are available

For additional help, refer to the main project README or create an issue in the repository.