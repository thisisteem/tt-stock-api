# Implementation Plan

- [ ] 1. Set up project structure and core configuration
  - Create Go module and directory structure following the specified layout
  - Set up go.mod with required dependencies (Fiber, JWT, bcrypt, PostgreSQL driver)
  - Create .env file template with JWT_SECRET, DB_URL configuration
  - Implement config loader in `internal/config/config.go` to read environment variables
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 2. Implement database connection and user model
  - [ ] 2.1 Create database connection utilities in `internal/db/db.go`
    - Implement PostgreSQL connection with proper error handling
    - Create database migration for users and token_blacklist tables
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [ ] 2.2 Implement User model in `internal/user/model.go`
    - Define User struct with proper JSON and database tags
    - Include ID, PhoneNumber, PinHash, CreatedAt, UpdatedAt, LastLoginAt fields
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [ ] 2.3 Create TokenBlacklist model for logout functionality
    - Define TokenBlacklist struct for managing invalidated tokens
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 3. Implement PIN hashing utilities
  - [ ] 3.1 Create password utilities in `pkg/utils/password.go`
    - Implement HashPin function using bcrypt with proper work factor
    - Implement CheckPin function to verify hashed PINs
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [ ]* 3.2 Write unit tests for PIN hashing functions
    - Test PIN hashing with different inputs
    - Test PIN verification with correct and incorrect PINs
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 4. Create user repository layer
  - [ ] 4.1 Implement user repository in `internal/user/repository.go`
    - Create FindByPhoneNumber method to retrieve users from database
    - Implement proper error handling for database operations
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [ ]* 4.2 Write unit tests for user repository
    - Test user retrieval with valid and invalid phone numbers
    - Use database mocks for testing
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

- [ ] 5. Implement authentication service layer
  - [ ] 5.1 Create authentication service in `internal/auth/service.go`
    - Implement ValidatePhoneNumber function for Thai phone number format (^0[0-9]{9}$)
    - Implement ValidatePin function for 6-digit PIN format (^[0-9]{6}$)
    - Implement AuthenticateUser function to validate credentials
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [ ] 5.2 Implement JWT token generation and validation
    - Create GenerateAccessToken function (15-minute expiration)
    - Create GenerateRefreshToken function (7-day expiration)
    - Implement token validation and parsing functions
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [ ] 5.3 Implement token blacklist functionality
    - Create BlacklistToken function to invalidate tokens on logout
    - Implement IsTokenBlacklisted function to check token validity
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ]* 5.4 Write unit tests for authentication service
    - Test phone number and PIN validation functions
    - Test token generation and validation
    - Test authentication logic with mocked dependencies
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 6. Create API response utilities
  - [ ] 6.1 Implement response helpers in `pkg/response/response.go`
    - Create SuccessResponse, ErrorResponse, and LoginResponse structs
    - Implement helper functions for consistent JSON responses
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 7. Implement authentication handlers
  - [ ] 7.1 Create authentication handlers in `internal/auth/handler.go`
    - Implement Login handler for POST /auth/login endpoint
    - Implement Refresh handler for POST /auth/refresh endpoint
    - Implement Logout handler for POST /auth/logout endpoint
    - Include proper input validation and error handling
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ]* 7.2 Write integration tests for authentication handlers
    - Test login endpoint with valid and invalid credentials
    - Test refresh token functionality
    - Test logout endpoint
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 8. Create JWT middleware for protected routes
  - [ ] 8.1 Implement JWT middleware in `internal/auth/middleware.go`
    - Create JWTProtected middleware function
    - Implement token extraction and validation from Authorization header
    - Add user context extraction for protected routes
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [ ]* 8.2 Write tests for JWT middleware
    - Test middleware with valid and invalid tokens
    - Test token expiration handling
    - Test blacklisted token rejection
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 9. Set up Fiber application and routing
  - [ ] 9.1 Create Fiber server configuration in `internal/app/server.go`
    - Initialize Fiber app with proper middleware (CORS, logging, recovery)
    - Configure JSON parsing and response settings
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ] 9.2 Implement route registration in `internal/app/routes/routes.go`
    - Register authentication routes (/auth/login, /auth/refresh, /auth/logout)
    - Set up dependency injection for handlers and services
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 10. Create application entry point and wire everything together
  - [ ] 10.1 Implement main.go in `cmd/api/main.go`
    - Load configuration from environment variables
    - Initialize database connection
    - Set up dependency injection for all layers
    - Start Fiber server with graceful shutdown
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ] 10.2 Create Makefile for build and run commands
    - Add commands for build, run, test, and lint
    - Include database migration commands
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 11. Create documentation and setup instructions
  - [ ] 11.1 Write comprehensive README.md
    - Document API endpoints with request/response examples
    - Include setup and run instructions
    - Add environment variable configuration guide
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_