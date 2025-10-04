# Technology Stack

## Backend Framework
- **Go**: Primary programming language
- **Fiber**: Web framework for HTTP API
- **PostgreSQL**: Primary database for user data and inventory

## Authentication & Security
- **JWT**: Token-based authentication (access + refresh tokens)
- **bcrypt**: PIN hashing with salt
- **Token blacklisting**: Secure logout implementation

## Key Dependencies
- Fiber web framework
- JWT library for Go
- bcrypt for password hashing
- PostgreSQL driver (pq or pgx)

## Project Structure
```
cmd/api/           # Application entry point
internal/
  auth/           # Authentication logic
  user/           # User models and repository
  db/             # Database connection utilities
  config/         # Configuration management
  app/            # Server setup and routing
pkg/
  utils/          # Shared utilities (password, response)
  response/       # API response helpers
```

## Common Commands

### Development
```bash
# Run the application
go run cmd/api/main.go

# Build the application
go build -o bin/api cmd/api/main.go

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Database
```bash
# Run migrations (when implemented)
make migrate-up

# Rollback migrations
make migrate-down
```

## Environment Configuration
- `JWT_SECRET`: Secret key for JWT token signing
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`: PostgreSQL connection details
- `PORT`: Server port (default: 8080)

## Validation Rules
- **Thai Phone Numbers**: `^0[0-9]{9}$` (10 digits starting with 0)
- **PIN Format**: `^[0-9]{6}$` (exactly 6 digits)

## Token Expiration
- **Access Token**: 15 minutes
- **Refresh Token**: 1 day