// Package repositories contains the repository layer implementations for the TT Stock Backend API.
// It provides data access interfaces and implementations using GORM for database operations.
package repositories

import (
	"context"
	"errors"

	"tt-stock-api/src/models"

	"gorm.io/gorm"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *models.Product) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id uint) (*models.Product, error)

	// GetBySKU retrieves a product by SKU
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *models.Product) error

	// Delete soft deletes a product
	Delete(ctx context.Context, id uint) error

	// List retrieves products with pagination and filtering
	List(ctx context.Context, req *models.ProductSearchRequest) (*models.ProductSearchResponse, error)

	// GetActiveProducts retrieves all active products
	GetActiveProducts(ctx context.Context) ([]models.Product, error)

	// GetProductsByType retrieves products by type
	GetProductsByType(ctx context.Context, productType models.ProductType) ([]models.Product, error)

	// GetProductsByBrand retrieves products by brand
	GetProductsByBrand(ctx context.Context, brand string) ([]models.Product, error)

	// GetLowStockProducts retrieves products with low stock
	GetLowStockProducts(ctx context.Context) ([]models.Product, error)

	// GetOutOfStockProducts retrieves products that are out of stock
	GetOutOfStockProducts(ctx context.Context) ([]models.Product, error)

	// UpdateQuantity updates the quantity of a product
	UpdateQuantity(ctx context.Context, productID uint, quantity int) error

	// AdjustQuantity adjusts the quantity of a product
	AdjustQuantity(ctx context.Context, productID uint, adjustment int) error

	// Count returns the total number of products
	Count(ctx context.Context) (int64, error)

	// Exists checks if a product exists by SKU
	Exists(ctx context.Context, sku string) (bool, error)

	// GetProductsWithSpecifications retrieves products with their specifications
	GetProductsWithSpecifications(ctx context.Context, productIDs []uint) ([]models.Product, error)
}

// productRepository implements the ProductRepository interface
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository instance
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	// Check if product with SKU already exists
	exists, err := r.Exists(ctx, product.SKU)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("product with this SKU already exists")
	}

	// Create product with specifications
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	if id == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	var product models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return &product, nil
}

// GetBySKU retrieves a product by SKU
func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	if sku == "" {
		return nil, errors.New("SKU cannot be empty")
	}

	var product models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("sku = ?", sku).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return &product, nil
}

// Update updates an existing product
func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	if product.ID == 0 {
		return errors.New("product ID cannot be zero")
	}

	// Check if product exists
	_, err := r.GetByID(ctx, product.ID)
	if err != nil {
		return err
	}

	// Update product
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return err
	}

	return nil
}

// Delete soft deletes a product
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("product ID cannot be zero")
	}

	// Check if product exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete by setting IsActive to false
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves products with pagination and filtering
func (r *productRepository) List(ctx context.Context, req *models.ProductSearchRequest) (*models.ProductSearchResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.Product{}).Preload("Specifications")

	// Apply filters
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Brand != nil && *req.Brand != "" {
		query = query.Where("brand = ?", *req.Brand)
	}
	if req.StockStatus != nil {
		switch *req.StockStatus {
		case models.StockStatusAvailable:
			query = query.Where("quantity_on_hand > low_stock_threshold")
		case models.StockStatusLowStock:
			query = query.Where("quantity_on_hand <= low_stock_threshold AND quantity_on_hand > 0")
		case models.StockStatusOutOfStock:
			query = query.Where("quantity_on_hand = 0")
		}
	}
	if req.MinPrice != nil {
		query = query.Where("selling_price >= ?", *req.MinPrice)
	}
	if req.MaxPrice != nil {
		query = query.Where("selling_price <= ?", *req.MaxPrice)
	}
	if req.Query != nil && *req.Query != "" {
		searchTerm := "%" + *req.Query + "%"
		query = query.Where("brand ILIKE ? OR model ILIKE ? OR sku ILIKE ? OR description ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Only show active products
	query = query.Where("is_active = ?", true)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (req.Page - 1) * req.Limit
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	// Apply pagination and ordering
	var products []models.Product
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.Limit).Find(&products).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
	}

	return &models.ProductSearchResponse{
		Products:       productResponses,
		TotalCount:     total,
		SearchCriteria: *req,
		Pagination: models.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

// GetActiveProducts retrieves all active products
func (r *productRepository) GetActiveProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("is_active = ?", true).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// GetProductsByType retrieves products by type
func (r *productRepository) GetProductsByType(ctx context.Context, productType models.ProductType) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("type = ? AND is_active = ?", productType, true).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// GetProductsByBrand retrieves products by brand
func (r *productRepository) GetProductsByBrand(ctx context.Context, brand string) ([]models.Product, error) {
	if brand == "" {
		return nil, errors.New("brand cannot be empty")
	}

	var products []models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("brand = ? AND is_active = ?", brand, true).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// GetLowStockProducts retrieves products with low stock
func (r *productRepository) GetLowStockProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("quantity_on_hand <= low_stock_threshold AND quantity_on_hand > 0 AND is_active = ?", true).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// GetOutOfStockProducts retrieves products that are out of stock
func (r *productRepository) GetOutOfStockProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("quantity_on_hand = 0 AND is_active = ?", true).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// UpdateQuantity updates the quantity of a product
func (r *productRepository) UpdateQuantity(ctx context.Context, productID uint, quantity int) error {
	if productID == 0 {
		return errors.New("product ID cannot be zero")
	}

	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	// Check if product exists
	_, err := r.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// Update quantity
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("quantity_on_hand", quantity).Error; err != nil {
		return err
	}

	return nil
}

// AdjustQuantity adjusts the quantity of a product
func (r *productRepository) AdjustQuantity(ctx context.Context, productID uint, adjustment int) error {
	if productID == 0 {
		return errors.New("product ID cannot be zero")
	}

	// Get current product
	product, err := r.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// Calculate new quantity
	newQuantity := product.QuantityOnHand + adjustment
	if newQuantity < 0 {
		return errors.New("insufficient stock for this operation")
	}

	// Update quantity
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("quantity_on_hand", newQuantity).Error; err != nil {
		return err
	}

	return nil
}

// Count returns the total number of products
func (r *productRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// Exists checks if a product exists by SKU
func (r *productRepository) Exists(ctx context.Context, sku string) (bool, error) {
	if sku == "" {
		return false, errors.New("SKU cannot be empty")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetProductsWithSpecifications retrieves products with their specifications
func (r *productRepository) GetProductsWithSpecifications(ctx context.Context, productIDs []uint) ([]models.Product, error) {
	if len(productIDs) == 0 {
		return []models.Product{}, nil
	}

	var products []models.Product
	if err := r.db.WithContext(ctx).Preload("Specifications").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}
