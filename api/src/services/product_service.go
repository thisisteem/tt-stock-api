// Package services contains the business logic layer implementations for the TT Stock Backend API.
// It provides service interfaces and implementations that orchestrate repository operations
// and implement business rules and validation logic.
package services

import (
	"context"
	"errors"
	"fmt"

	"tt-stock-api/src/models"
	"tt-stock-api/src/repositories"
)

// ProductService defines the interface for product management operations
type ProductService interface {
	// CreateProduct creates a new product
	CreateProduct(ctx context.Context, req *models.ProductCreateRequest, user *models.User) (*models.ProductResponse, error)

	// GetProduct retrieves a product by ID
	GetProduct(ctx context.Context, id uint, user *models.User) (*models.ProductResponse, error)

	// GetProductBySKU retrieves a product by SKU
	GetProductBySKU(ctx context.Context, sku string, user *models.User) (*models.ProductResponse, error)

	// UpdateProduct updates an existing product
	UpdateProduct(ctx context.Context, id uint, req *models.ProductUpdateRequest, user *models.User) (*models.ProductResponse, error)

	// DeleteProduct soft deletes a product
	DeleteProduct(ctx context.Context, id uint, user *models.User) error

	// ListProducts retrieves products with pagination and filtering
	ListProducts(ctx context.Context, req *models.ProductSearchRequest, user *models.User) (*models.ProductSearchResponse, error)

	// SearchProducts searches products with advanced filters
	SearchProducts(ctx context.Context, req *models.ProductSearchRequest, user *models.User) (*models.ProductSearchResponse, error)

	// GetLowStockProducts retrieves products with low stock
	GetLowStockProducts(ctx context.Context, user *models.User) ([]models.ProductResponse, error)

	// GetOutOfStockProducts retrieves products that are out of stock
	GetOutOfStockProducts(ctx context.Context, user *models.User) ([]models.ProductResponse, error)

	// GetProductsByType retrieves products by type
	GetProductsByType(ctx context.Context, productType models.ProductType, user *models.User) ([]models.ProductResponse, error)

	// GetProductsByBrand retrieves products by brand
	GetProductsByBrand(ctx context.Context, brand string, user *models.User) ([]models.ProductResponse, error)

	// UpdateProductQuantity updates the quantity of a product
	UpdateProductQuantity(ctx context.Context, productID uint, quantity int, user *models.User) (*models.ProductResponse, error)

	// AdjustProductQuantity adjusts the quantity of a product
	AdjustProductQuantity(ctx context.Context, productID uint, adjustment int, user *models.User) (*models.ProductResponse, error)

	// GetProductStatistics retrieves product statistics
	GetProductStatistics(ctx context.Context, user *models.User) (*models.ProductStatistics, error)
}

// productService implements the ProductService interface
type productService struct {
	productRepo repositories.ProductRepository
}

// NewProductService creates a new ProductService instance
func NewProductService(productRepo repositories.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

// CreateProduct creates a new product
func (s *productService) CreateProduct(ctx context.Context, req *models.ProductCreateRequest, user *models.User) (*models.ProductResponse, error) {
	if req == nil {
		return nil, errors.New("create request cannot be nil")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to create products")
	}

	// Create product model
	product := &models.Product{
		Type:              req.Type,
		Brand:             req.Brand,
		Model:             req.Model,
		SKU:               req.SKU,
		Description:       req.Description,
		ImageBase64:       req.ImageBase64,
		CostPrice:         req.CostPrice,
		SellingPrice:      req.SellingPrice,
		QuantityOnHand:    req.QuantityOnHand,
		LowStockThreshold: req.LowStockThreshold,
		IsActive:          true,
	}

	// Create product
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Create specifications if provided
	if len(req.Specifications) > 0 {
		if err := s.createProductSpecifications(ctx, product.ID, req.Specifications); err != nil {
			// Log error but don't fail product creation
			// TODO: Add proper logging
		}
	}

	// Get the created product with specifications
	createdProduct, err := s.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created product: %w", err)
	}

	response := createdProduct.ToResponse()
	return &response, nil
}

// GetProduct retrieves a product by ID
func (s *productService) GetProduct(ctx context.Context, id uint, user *models.User) (*models.ProductResponse, error) {
	if id == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view products")
	}

	// Get product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

// GetProductBySKU retrieves a product by SKU
func (s *productService) GetProductBySKU(ctx context.Context, sku string, user *models.User) (*models.ProductResponse, error) {
	if sku == "" {
		return nil, errors.New("SKU cannot be empty")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view products")
	}

	// Get product
	product, err := s.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

// UpdateProduct updates an existing product
func (s *productService) UpdateProduct(ctx context.Context, id uint, req *models.ProductUpdateRequest, user *models.User) (*models.ProductResponse, error) {
	if id == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to update products")
	}

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Type != nil {
		product.Type = *req.Type
	}
	if req.Brand != nil {
		product.Brand = *req.Brand
	}
	if req.Model != nil {
		product.Model = *req.Model
	}
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.ImageBase64 != nil {
		product.ImageBase64 = req.ImageBase64
	}
	if req.CostPrice != nil {
		product.CostPrice = *req.CostPrice
	}
	if req.SellingPrice != nil {
		product.SellingPrice = *req.SellingPrice
	}
	if req.QuantityOnHand != nil {
		product.QuantityOnHand = *req.QuantityOnHand
	}
	if req.LowStockThreshold != nil {
		product.LowStockThreshold = *req.LowStockThreshold
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	// Update product
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	response := product.ToResponse()
	return &response, nil
}

// DeleteProduct soft deletes a product
func (s *productService) DeleteProduct(ctx context.Context, id uint, user *models.User) error {
	if id == 0 {
		return errors.New("product ID cannot be zero")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return errors.New("insufficient permissions to delete products")
	}

	// Get product to check if it exists and has stock
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Prevent deletion of products with stock
	if product.QuantityOnHand > 0 {
		return errors.New("cannot delete product with remaining stock")
	}

	// Delete product
	return s.productRepo.Delete(ctx, id)
}

// ListProducts retrieves products with pagination and filtering
func (s *productService) ListProducts(ctx context.Context, req *models.ProductSearchRequest, user *models.User) (*models.ProductSearchResponse, error) {
	if req == nil {
		return nil, errors.New("list request cannot be nil")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to list products")
	}

	// List products
	return s.productRepo.List(ctx, req)
}

// SearchProducts searches products with advanced filters
func (s *productService) SearchProducts(ctx context.Context, req *models.ProductSearchRequest, user *models.User) (*models.ProductSearchResponse, error) {
	if req == nil {
		return nil, errors.New("search request cannot be nil")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to search products")
	}

	// Search products
	return s.productRepo.List(ctx, req)
}

// GetLowStockProducts retrieves products with low stock
func (s *productService) GetLowStockProducts(ctx context.Context, user *models.User) ([]models.ProductResponse, error) {
	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view low stock products")
	}

	// Get low stock products
	products, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

// GetOutOfStockProducts retrieves products that are out of stock
func (s *productService) GetOutOfStockProducts(ctx context.Context, user *models.User) ([]models.ProductResponse, error) {
	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view out of stock products")
	}

	// Get out of stock products
	products, err := s.productRepo.GetOutOfStockProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

// GetProductsByType retrieves products by type
func (s *productService) GetProductsByType(ctx context.Context, productType models.ProductType, user *models.User) ([]models.ProductResponse, error) {
	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view products by type")
	}

	// Get products by type
	products, err := s.productRepo.GetProductsByType(ctx, productType)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

// GetProductsByBrand retrieves products by brand
func (s *productService) GetProductsByBrand(ctx context.Context, brand string, user *models.User) ([]models.ProductResponse, error) {
	if brand == "" {
		return nil, errors.New("brand cannot be empty")
	}

	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view products by brand")
	}

	// Get products by brand
	products, err := s.productRepo.GetProductsByBrand(ctx, brand)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

// UpdateProductQuantity updates the quantity of a product
func (s *productService) UpdateProductQuantity(ctx context.Context, productID uint, quantity int, user *models.User) (*models.ProductResponse, error) {
	if productID == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to update product quantity")
	}

	// Update quantity
	if err := s.productRepo.UpdateQuantity(ctx, productID, quantity); err != nil {
		return nil, err
	}

	// Get updated product
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

// AdjustProductQuantity adjusts the quantity of a product
func (s *productService) AdjustProductQuantity(ctx context.Context, productID uint, adjustment int, user *models.User) (*models.ProductResponse, error) {
	if productID == 0 {
		return nil, errors.New("product ID cannot be zero")
	}

	// Check permissions
	if !s.canManageInventory(user) {
		return nil, errors.New("insufficient permissions to adjust product quantity")
	}

	// Adjust quantity
	if err := s.productRepo.AdjustQuantity(ctx, productID, adjustment); err != nil {
		return nil, err
	}

	// Get updated product
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

// GetProductStatistics retrieves product statistics
func (s *productService) GetProductStatistics(ctx context.Context, user *models.User) (*models.ProductStatistics, error) {
	// Check permissions
	if !s.canViewInventory(user) {
		return nil, errors.New("insufficient permissions to view product statistics")
	}

	// Get total count
	_, err := s.productRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get low stock products
	lowStockProducts, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Get out of stock products
	outOfStockProducts, err := s.productRepo.GetOutOfStockProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Get active products
	activeProducts, err := s.productRepo.GetActiveProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate total value
	var totalValue float64
	for _, product := range activeProducts {
		totalValue += product.GetTotalValue()
	}

	// Calculate average stock value
	var averageStockValue float64
	if len(activeProducts) > 0 {
		averageStockValue = totalValue / float64(len(activeProducts))
	}

	return &models.ProductStatistics{
		TotalProducts:     int64(len(activeProducts)),
		TotalValue:        totalValue,
		LowStockCount:     int64(len(lowStockProducts)),
		OutOfStockCount:   int64(len(outOfStockProducts)),
		AverageStockValue: averageStockValue,
	}, nil
}

// Helper methods

// canViewInventory checks if user can view inventory
func (s *productService) canViewInventory(user *models.User) bool {
	return user != nil && user.CanManageInventory()
}

// canManageInventory checks if user can manage inventory
func (s *productService) canManageInventory(user *models.User) bool {
	return user != nil && user.CanManageInventory()
}

// createProductSpecifications creates product specifications
func (s *productService) createProductSpecifications(ctx context.Context, productID uint, specs []models.ProductSpecificationCreateRequest) error {
	// This would typically involve creating specifications in a separate repository
	// For now, we'll return nil as the specifications are handled in the product model
	// TODO: Implement specification creation logic
	return nil
}
