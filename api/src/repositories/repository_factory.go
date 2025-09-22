// Package repositories contains the repository layer implementations for the TT Stock Backend API.
// It provides data access interfaces and implementations using GORM for database operations.
package repositories

import (
	"tt-stock-api/src/database"
)

// RepositoryFactory creates and manages repository instances
type RepositoryFactory struct {
	connectionManager *database.ConnectionManager
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(connectionManager *database.ConnectionManager) *RepositoryFactory {
	return &RepositoryFactory{
		connectionManager: connectionManager,
	}
}

// GetUserRepository returns a user repository instance
func (f *RepositoryFactory) GetUserRepository() UserRepository {
	return &userRepository{
		db: f.connectionManager.GetDB(),
	}
}

// GetProductRepository returns a product repository instance
func (f *RepositoryFactory) GetProductRepository() ProductRepository {
	return &productRepository{
		db: f.connectionManager.GetDB(),
	}
}

// GetStockMovementRepository returns a stock movement repository instance
func (f *RepositoryFactory) GetStockMovementRepository() StockMovementRepository {
	return &stockMovementRepository{
		db: f.connectionManager.GetDB(),
	}
}

// GetSessionRepository returns a session repository instance
func (f *RepositoryFactory) GetSessionRepository() SessionRepository {
	return &sessionRepository{
		db: f.connectionManager.GetDB(),
	}
}

// GetAlertRepository returns an alert repository instance
func (f *RepositoryFactory) GetAlertRepository() AlertRepository {
	return &alertRepository{
		db: f.connectionManager.GetDB(),
	}
}

// GetConnectionManager returns the database connection manager
func (f *RepositoryFactory) GetConnectionManager() *database.ConnectionManager {
	return f.connectionManager
}

// Close closes the database connection
func (f *RepositoryFactory) Close() error {
	return f.connectionManager.Close()
}
