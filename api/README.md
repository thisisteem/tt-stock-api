# TT Stock Backend API

A comprehensive tire and wheel inventory management system API built with Go and Gin framework, following Clean Architecture principles and SOLID design patterns.

## 🚀 Features

- **User Authentication & Authorization**: JWT-based authentication with role-based access control (Admin, Owner, Staff)
- **Product Management**: Complete CRUD operations for tires and wheels with specifications
- **Stock Movement Tracking**: Real-time inventory tracking with movement history
- **Low Stock Alerts**: Automated notifications for inventory management
- **Business Intelligence**: Comprehensive reporting and analytics
- **High Performance**: Sub-200ms response times with support for 1000+ concurrent users
- **Comprehensive Testing**: Unit tests, integration tests, performance tests, and load tests

## 🏗️ Architecture

This API follows Clean Architecture principles with clear separation of concerns:

```
├── src/
│   ├── models/          # Domain models and entities
│   ├── repositories/    # Data access layer
│   ├── services/        # Business logic layer
│   ├── handlers/        # HTTP delivery layer
│   ├── middleware/      # Cross-cutting concerns
│   ├── validators/      # Input validation
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   └── router/          # HTTP routing
├── tests/
│   ├── unit/           # Unit tests
│   ├── integration/    # Integration tests
│   ├── performance/    # Performance tests
│   └── load/           # Load tests
├── docs/               # API documentation
└── migrations/         # Database migrations
```

## 📋 Prerequisites

- **Go 1.21+**: [Download and install Go](https://golang.org/dl/)
- **PostgreSQL 13+**: [Download and install PostgreSQL](https://www.postgresql.org/download/)
- **Docker & Docker Compose**: [Download and install Docker](https://www.docker.com/get-started) (optional, for containerized development)

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd tt-stock-api/api
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Environment Configuration

Create a `.env` file in the `api` directory:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=60s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=ttstock_user
DB_PASSWORD=ttstock_password
DB_NAME=ttstock_db
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_IDLE_TIME=30m

# JWT Configuration
JWT_SECRET_KEY=your-super-secret-jwt-key-change-this-in-production
JWT_TOKEN_LIFETIME=24h

# Application Configuration
APP_NAME=TT Stock API
APP_VERSION=1.0.0
APP_ENVIRONMENT=development
APP_LOG_LEVEL=info

# Security Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST=10

# Middleware Configuration
REQUEST_SIZE_LIMIT=10MB
REQUEST_TIMEOUT=30s
```

### 4. Database Setup

#### Option A: Using Docker Compose (Recommended)

```bash
# Start PostgreSQL database
docker-compose up -d postgres

# Wait for database to be ready, then run migrations
go run src/main.go migrate
```

#### Option B: Manual PostgreSQL Setup

```bash
# Create database and user
sudo -u postgres psql
CREATE DATABASE ttstock_db;
CREATE USER ttstock_user WITH PASSWORD 'ttstock_password';
GRANT ALL PRIVILEGES ON DATABASE ttstock_db TO ttstock_user;
\q

# Run migrations
go run src/main.go migrate
```

### 5. Run the Application

```bash
# Development mode
go run src/main.go

# Or build and run
go build -o ttstock-api src/main.go
./ttstock-api
```

The API will be available at `http://localhost:8080`

## 🧪 Testing

### Run All Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Run Specific Test Suites

```bash
# Unit tests
go test ./tests/unit/...

# Integration tests
go test ./tests/integration/...

# Performance tests
go test ./tests/performance/...

# Load tests
go test ./tests/load/...
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📚 API Documentation

### Interactive Documentation

Once the server is running, you can access:

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI Spec**: `http://localhost:8080/swagger/doc.json`

### API Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh token

#### Products
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/{id}` - Get product
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product
- `GET /api/v1/products/search` - Search products
- `GET /api/v1/products/low-stock` - Get low stock products

#### Stock Management
- `GET /api/v1/stock/movements` - List stock movements
- `POST /api/v1/stock/movements` - Create stock movement
- `GET /api/v1/stock/movements/{id}` - Get stock movement
- `POST /api/v1/stock/sale` - Process sale

#### Alerts
- `GET /api/v1/alerts` - List alerts
- `PUT /api/v1/alerts/{id}/read` - Mark alert as read

#### Health Checks
- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed health check

## 🔐 Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/api/v1/products
```

### User Roles

- **Admin**: Full system access, can manage users and all data
- **Owner**: Business management access, can manage products and view reports
- **Staff**: Limited access, can manage inventory and process sales

## 🚀 Performance

### Response Time Requirements
- All API endpoints respond within **200ms** under normal load
- Health checks respond within **50ms**
- Search operations respond within **150ms**

### Load Testing
The API is tested to handle:
- **1000+ concurrent users**
- **100+ requests per second**
- **95%+ success rate** under load

### Performance Testing

```bash
# Run performance tests
go test -v ./tests/performance/...

# Run load tests
go test -v ./tests/load/...

# Benchmark specific endpoints
go test -bench=. ./tests/performance/...
```

## 🐳 Docker Support

### Development with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

### Build Docker Image

```bash
# Build image
docker build -t ttstock-api .

# Run container
docker run -p 8080:8080 --env-file .env ttstock-api
```

## 🔧 Development

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Run go vet
go vet ./...
```

### Database Migrations

```bash
# Create new migration
go run src/main.go migrate create <migration_name>

# Run migrations
go run src/main.go migrate up

# Rollback migration
go run src/main.go migrate down
```

### Adding New Features

1. **Models**: Add domain models in `src/models/`
2. **Repositories**: Implement data access in `src/repositories/`
3. **Services**: Add business logic in `src/services/`
4. **Handlers**: Create HTTP endpoints in `src/handlers/`
5. **Tests**: Write comprehensive tests in `tests/`

## 📊 Monitoring & Logging

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed
```

### Logging

The API provides structured logging with different levels:
- **DEBUG**: Detailed information for debugging
- **INFO**: General information about application flow
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages for failures

## 🚨 Error Handling

The API returns consistent error responses:

```json
{
  "error": "Validation Error",
  "message": "Invalid request data",
  "details": {
    "field": "phone_number",
    "reason": "Phone number is required"
  }
}
```

## 🔒 Security

- **JWT Authentication**: Secure token-based authentication
- **Input Validation**: Comprehensive input validation and sanitization
- **CORS Protection**: Configurable CORS policies
- **Rate Limiting**: Protection against abuse
- **SQL Injection Prevention**: Parameterized queries with GORM
- **XSS Protection**: Input sanitization and output encoding

## 📈 Business Intelligence

The API provides comprehensive reporting capabilities:

- **Inventory Reports**: Stock levels, movements, and trends
- **Sales Analytics**: Revenue, profit margins, and top products
- **User Activity**: Login statistics and user behavior
- **Financial Reports**: Profitability and revenue breakdowns

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting PR
- Follow Clean Architecture principles

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:

- **Documentation**: Check the [API documentation](docs/openapi.yaml)
- **Issues**: Open an issue on GitHub
- **Email**: support@ttstock.com

## 🗺️ Roadmap

- [ ] Real-time notifications with WebSockets
- [ ] Mobile app API endpoints
- [ ] Advanced analytics dashboard
- [ ] Multi-tenant support
- [ ] API versioning
- [ ] GraphQL endpoint
- [ ] Event sourcing for audit trails

---

**Built with ❤️ using Go, Gin, PostgreSQL, and Clean Architecture principles.**
