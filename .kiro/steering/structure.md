# Project Structure & Organization

## Directory Layout

Follow Go standard project layout with clean architecture principles:

```
tt-stock-api/
├── cmd/api/                    # Application entry points
│   └── main.go                # Main application entry
├── internal/                   # Private application code
│   ├── auth/                  # Authentication domain
│   │   ├── handler.go         # HTTP handlers
│   │   ├── service.go         # Business logic
│   │   └── middleware.go      # JWT middleware
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
│           └── routes.go
├── pkg/                       # Public/shared code
│   ├── utils/                 # Utility functions
│   │   └── password.go        # Password hashing
│   └── response/              # API responses
│       └── response.go        # Response helpers
├── migrations/                # Database migrations
├── docs/                      # Documentation
├── .env.example              # Environment template
├── Makefile                  # Build commands
└── go.mod                    # Go module definition
```

## Naming Conventions

### Files & Directories
- Use lowercase with underscores for file names: `user_service.go`
- Use singular nouns for package names: `user`, `auth`, `config`
- Group related functionality in domain packages

### Go Code
- Use PascalCase for exported functions/types: `AuthenticateUser`
- Use camelCase for private functions/variables: `validateCredentials`
- Use descriptive names: `GenerateAccessToken` not `GenToken`

## Architecture Patterns

### Layered Architecture
1. **Handler Layer**: HTTP request/response handling
2. **Service Layer**: Business logic and validation
3. **Repository Layer**: Data access and persistence
4. **Model Layer**: Data structures and entities

### Dependency Injection
- Pass dependencies through constructors
- Use interfaces for testability
- Avoid global variables except for configuration

## API Conventions

### Endpoints
- Use RESTful patterns: `/auth/login`, `/auth/refresh`, `/auth/logout`
- Version APIs when needed: `/v1/auth/login`
- Use HTTP methods appropriately: POST for authentication

### Response Format
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

### Error Responses
```json
{
  "success": false,
  "message": "Invalid credentials",
  "error_code": "AUTH_INVALID_CREDENTIALS"
}
```

## Database Conventions

### Table Names
- Use snake_case: `users`, `token_blacklist`
- Use plural nouns for table names
- Include timestamps: `created_at`, `updated_at`

### Model Fields
- Use consistent field naming across models
- Include proper JSON tags: `json:"phone_number"`
- Include database tags: `db:"phone_number"`

## Testing Structure

### Test Files
- Place tests alongside source files: `service_test.go`
- Use table-driven tests for multiple scenarios
- Mock external dependencies (database, HTTP clients)

### Test Naming
- Use descriptive test names: `TestAuthenticateUser_ValidCredentials_ReturnsTokens`
- Group related tests in subtests using `t.Run()`