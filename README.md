# TT Stock API

A secure authentication API for tire & wheel stock management system designed for tire shops in Thailand. This API provides JWT-based authentication with Thai phone number validation and 6-digit PIN security.

## 🚀 Features

- **Secure Authentication**: Thai phone number + 6-digit PIN authentication
- **JWT Token Management**: Access tokens (15 min) and refresh tokens (1 day)
- **Token Blacklisting**: Secure logout with token invalidation
- **Input Validation**: Thai phone number format and PIN validation
- **Security**: bcrypt PIN hashing with salt
- **Mobile-Optimized**: Lightweight responses for mobile applications

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Environment Configuration](#environment-configuration)
- [Development](#development)
- [Testing](#testing)
- [Database Setup](#database-setup)
- [Security](#security)
- [Contributing](#contributing)

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Make (optional, for using Makefile commands)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd tt-stock-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   # or using make
   make deps
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Set up PostgreSQL database**
   ```bash
   # Create database
   createdb tt_stock_db
   
   # Update DB_URL in .env file
   DB_URL=postgres://username:password@localhost:5432/tt_stock_db?sslmode=disable
   ```

5. **Run the application**
   ```bash
   make run
   # or
   go run cmd/api/main.go
   ```

The API will be available at `http://localhost:8080`

## 📚 API Documentation

### Base URL
```
http://localhost:8080
```

### Authentication Endpoints

#### 1. Login
Authenticate user with phone number and PIN.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "phone_number": "0123456789",
  "pin": "123456"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "phone_number": "0123456789"
    }
  }
}
```

**Error Response (401):**
```json
{
  "success": false,
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid phone number or PIN"
  }
}
```

#### 2. Refresh Token
Get new access and refresh tokens using a valid refresh token.

**Endpoint:** `POST /auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "phone_number": "0123456789"
    }
  }
}
```

#### 3. Logout
Invalidate access and refresh tokens.

**Endpoint:** `POST /auth/logout`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body (Optional):**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Logout successful"
}
```

### Protected Routes

For accessing protected endpoints, include the access token in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Invalid request data or format |
| `AUTHENTICATION_ERROR` | Invalid credentials or token |
| `TOKEN_EXPIRED` | Access token has expired |
| `NOT_FOUND` | Resource not found |
| `INTERNAL_SERVER_ERROR` | Server error |

### Validation Rules

- **Phone Number**: Must be Thai format `^0[0-9]{9}$` (10 digits starting with 0)
- **PIN**: Must be exactly 6 digits `^[0-9]{6}$`

## ⚙️ Environment Configuration

Create a `.env` file in the root directory with the following variables:

```bash
# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Database Configuration
DB_URL=postgres://username:password@localhost:5432/tt_stock_db?sslmode=disable

# Server Configuration
PORT=8080

# Environment
ENV=development
```

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET` | Secret key for JWT token signing | - | ✅ |
| `DB_URL` | PostgreSQL connection string | - | ✅ |
| `PORT` | Server port | 8080 | ❌ |
| `ENV` | Environment (development/production) | development | ❌ |

### Security Notes

- **JWT_SECRET**: Use a strong, random secret key (minimum 32 characters)
- **DB_URL**: Use SSL in production (`sslmode=require`)
- Never commit `.env` file to version control

## 🛠️ Development

### Available Make Commands

```bash
# Development
make run              # Run the application
make dev              # Run with hot reload (requires air)
make build            # Build the application
make build-prod       # Build for production

# Testing
make test             # Run tests
make test-coverage    # Run tests with coverage
make test-watch       # Run tests in watch mode

# Code Quality
make fmt              # Format code
make vet              # Vet code
make lint             # Lint code (requires golangci-lint)
make check            # Run all quality checks

# Database
make migrate-up       # Create database tables
make migrate-down     # Drop database tables (WARNING: destructive)
make create-user      # Create a new user

# Dependencies
make deps             # Download dependencies
make deps-update      # Update dependencies

# Tools
make install-tools    # Install development tools
make help             # Show all available commands
```

### Hot Reload Development

Install Air for hot reload during development:

```bash
go install github.com/cosmtrek/air@latest
make dev
```

### Code Quality Tools

Install recommended development tools:

```bash
make install-tools
```

This installs:
- `air` - Hot reload
- `gotestsum` - Enhanced test runner
- `gosec` - Security analyzer
- `golangci-lint` - Comprehensive linter (manual install required)

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make test-coverage-html

# Run tests in watch mode
make test-watch
```

### Test Structure

- Unit tests are located alongside source files (`*_test.go`)
- Integration tests use test database
- Mocked dependencies for isolated testing

### Example Test Commands

```bash
# Test specific package
go test ./internal/auth/...

# Test with verbose output
go test -v ./...

# Test with race detection
go test -race ./...
```

## 🗄️ Database Setup

### Automatic Setup

The application automatically creates required tables on startup. Simply ensure your PostgreSQL database exists and is accessible.

### Manual Setup

If you need to manually manage the database:

```bash
# Create tables
make migrate-up

# Drop tables (WARNING: This deletes all data)
make migrate-down

# Reset database (drop and recreate)
make migrate-reset
```

### Database Schema

The application creates the following tables:

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(10) UNIQUE NOT NULL,
    pin_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);
```

#### Token Blacklist Table
```sql
CREATE TABLE token_blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token TEXT NOT NULL,
    user_id UUID REFERENCES users(id),
    token_type VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    blacklisted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Creating Users

Users must be created manually by administrators:

```bash
# Using make command
make create-user PHONE=0123456789 PIN=123456

# Using psql directly
psql $DB_URL -c "INSERT INTO users (phone_number, pin_hash) VALUES ('0123456789', crypt('123456', gen_salt('bf', 12)));"
```

## 🔒 Security

### Authentication Security

- **PIN Hashing**: bcrypt with work factor 12
- **Token Security**: JWT with configurable expiration
- **Token Blacklisting**: Secure logout implementation
- **Input Validation**: Strict format validation for phone numbers and PINs

### Token Expiration

- **Access Token**: 15 minutes (for API requests)
- **Refresh Token**: 1 day (for token renewal)

### Security Best Practices

1. **Use HTTPS in production**
2. **Set strong JWT_SECRET** (minimum 32 characters)
3. **Enable SSL for database connections**
4. **Regularly rotate JWT secrets**
5. **Monitor failed authentication attempts**
6. **Use rate limiting in production**

### Production Security Checklist

- [ ] HTTPS enabled
- [ ] Strong JWT_SECRET configured
- [ ] Database SSL enabled
- [ ] Rate limiting implemented
- [ ] Logging configured
- [ ] Security headers configured
- [ ] CORS properly configured

## 🏗️ Project Structure

```
tt-stock-api/
├── cmd/api/                    # Application entry point
│   └── main.go
├── internal/                   # Private application code
│   ├── auth/                  # Authentication domain
│   │   ├── handler.go         # HTTP handlers
│   │   ├── service.go         # Business logic
│   │   ├── middleware.go      # JWT middleware
│   │   └── model.go           # Auth models
│   ├── user/                  # User domain
│   │   ├── model.go           # User data structures
│   │   └── repository.go      # Data access layer
│   ├── db/                    # Database utilities
│   │   └── db.go              # Connection management
│   ├── config/                # Configuration
│   │   └── config.go          # Environment config
│   └── app/                   # Application setup
│       ├── server.go          # Server configuration
│       └── routes/            # Route definitions
├── pkg/                       # Public/shared code
│   ├── utils/                 # Utility functions
│   │   └── password.go        # Password hashing
│   └── response/              # API responses
│       └── response.go        # Response helpers
├── .env.example              # Environment template
├── Makefile                  # Build commands
└── go.mod                    # Go module definition
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run quality checks (`make check`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Write tests for new functionality
- Update documentation for API changes
- Run `make check` before committing
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

1. Check the [API Documentation](#api-documentation)
2. Review [Common Issues](#common-issues)
3. Create an issue in the repository

### Common Issues

**Q: "Failed to connect to database"**
A: Check your `DB_URL` in `.env` and ensure PostgreSQL is running.

**Q: "Invalid JWT secret"**
A: Ensure `JWT_SECRET` is set in `.env` and is at least 32 characters long.

**Q: "Token expired" errors****
A: Use the refresh token endpoint to get new tokens, or re-authenticate.

**Q: "Invalid phone number format"**
A: Phone numbers must be Thai format: 10 digits starting with 0 (e.g., 0123456789).

---

Built with ❤️ for Thai tire & wheel shops