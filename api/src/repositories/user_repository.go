// Package repositories contains the repository layer implementations for the TT Stock Backend API.
// It provides data access interfaces and implementations using GORM for database operations.
package repositories

import (
	"context"
	"errors"
	"time"

	"tt-stock-api/src/models"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*models.User, error)

	// GetByPhoneNumber retrieves a user by phone number
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *models.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uint) error

	// List retrieves users with pagination and filtering
	List(ctx context.Context, req *models.UserListRequest) (*models.UserListResponse, error)

	// GetActiveUsers retrieves all active users
	GetActiveUsers(ctx context.Context) ([]models.User, error)

	// GetUsersByRole retrieves users by role
	GetUsersByRole(ctx context.Context, role models.UserRole) ([]models.User, error)

	// UpdateLastLogin updates the last login timestamp for a user
	UpdateLastLogin(ctx context.Context, userID uint) error

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// Exists checks if a user exists by phone number
	Exists(ctx context.Context, phoneNumber string) (bool, error)
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Check if user with phone number already exists
	exists, err := r.Exists(ctx, user.PhoneNumber)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user with this phone number already exists")
	}

	// Create user
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if id == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetByPhoneNumber retrieves a user by phone number
func (r *userRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	if phoneNumber == "" {
		return nil, errors.New("phone number cannot be empty")
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if user.ID == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Check if user exists
	_, err := r.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Update user
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return err
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Check if user exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete by setting IsActive to false
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves users with pagination and filtering
func (r *userRepository) List(ctx context.Context, req *models.UserListRequest) (*models.UserListResponse, error) {
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

	query := r.db.WithContext(ctx).Model(&models.User{})

	// Apply filters
	if req.Role != nil {
		query = query.Where("role = ?", *req.Role)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	if req.Search != nil && *req.Search != "" {
		searchTerm := "%" + *req.Search + "%"
		query = query.Where("name ILIKE ? OR phone_number ILIKE ?", searchTerm, searchTerm)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (req.Page - 1) * req.Limit
	totalPages := (total + int64(req.Limit) - 1) / int64(req.Limit)

	// Apply pagination and ordering
	var users []models.User
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return &models.UserListResponse{
		Users: userResponses,
		Pagination: models.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: int(totalPages),
			HasNext:    req.Page < int(totalPages),
			HasPrev:    req.Page > 1,
		},
	}, nil
}

// GetActiveUsers retrieves all active users
func (r *userRepository) GetActiveUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetUsersByRole retrieves users by role
func (r *userRepository) GetUsersByRole(ctx context.Context, role models.UserRole) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("role = ? AND is_active = ?", role, true).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}

	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", now).Error; err != nil {
		return err
	}

	return nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// Exists checks if a user exists by phone number
func (r *userRepository) Exists(ctx context.Context, phoneNumber string) (bool, error) {
	if phoneNumber == "" {
		return false, errors.New("phone number cannot be empty")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("phone_number = ?", phoneNumber).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
