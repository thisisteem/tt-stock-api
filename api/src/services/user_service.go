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

// UserService defines the interface for user management operations
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, req *models.UserCreateRequest, createdBy *models.User) (*models.UserResponse, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id uint, requester *models.User) (*models.UserResponse, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id uint, req *models.UserUpdateRequest, requester *models.User) (*models.UserResponse, error)

	// DeleteUser soft deletes a user
	DeleteUser(ctx context.Context, id uint, requester *models.User) error

	// ListUsers retrieves users with pagination and filtering
	ListUsers(ctx context.Context, req *models.UserListRequest, requester *models.User) (*models.UserListResponse, error)

	// GetUserProfile retrieves the current user's profile
	GetUserProfile(ctx context.Context, user *models.User) (*models.UserResponse, error)

	// UpdateUserProfile updates the current user's profile
	UpdateUserProfile(ctx context.Context, user *models.User, req *models.UserUpdateRequest) (*models.UserResponse, error)

	// ChangeUserPIN changes a user's PIN
	ChangeUserPIN(ctx context.Context, user *models.User, currentPIN, newPIN string) error

	// GetUsersByRole retrieves users by role
	GetUsersByRole(ctx context.Context, role models.UserRole, requester *models.User) ([]models.UserResponse, error)

	// ActivateUser activates a user account
	ActivateUser(ctx context.Context, id uint, requester *models.User) error

	// DeactivateUser deactivates a user account
	DeactivateUser(ctx context.Context, id uint, requester *models.User) error
}

// userService implements the UserService interface
type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *models.UserCreateRequest, createdBy *models.User) (*models.UserResponse, error) {
	if req == nil {
		return nil, errors.New("create request cannot be nil")
	}

	// Check permissions
	if !s.canManageUsers(createdBy) {
		return nil, errors.New("insufficient permissions to create users")
	}

	// Validate role assignment
	if err := s.validateRoleAssignment(req.Role, createdBy); err != nil {
		return nil, err
	}

	// Create user model
	user := &models.User{
		PhoneNumber: req.PhoneNumber,
		PIN:         req.PIN, // Will be hashed in BeforeCreate hook
		Role:        req.Role,
		Name:        req.Name,
		IsActive:    true,
	}

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uint, requester *models.User) (*models.UserResponse, error) {
	if id == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	// Check permissions
	if !s.canViewUser(requester, id) {
		return nil, errors.New("insufficient permissions to view this user")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id uint, req *models.UserUpdateRequest, requester *models.User) (*models.UserResponse, error) {
	if id == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// Check permissions
	if !s.canManageUser(requester, id) {
		return nil, errors.New("insufficient permissions to update this user")
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate role changes
	if req.Role != nil {
		if err := s.validateRoleAssignment(*req.Role, requester); err != nil {
			return nil, err
		}
		user.Role = *req.Role
	}

	// Update fields
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.IsActive != nil {
		// Only admins and owners can activate/deactivate users
		if !s.canManageUsers(requester) {
			return nil, errors.New("insufficient permissions to change user status")
		}
		user.IsActive = *req.IsActive
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// DeleteUser soft deletes a user
func (s *userService) DeleteUser(ctx context.Context, id uint, requester *models.User) error {
	if id == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Check permissions
	if !s.canManageUsers(requester) {
		return errors.New("insufficient permissions to delete users")
	}

	// Prevent self-deletion
	if requester.ID == id {
		return errors.New("cannot delete your own account")
	}

	// Get user to check if they exist
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

// ListUsers retrieves users with pagination and filtering
func (s *userService) ListUsers(ctx context.Context, req *models.UserListRequest, requester *models.User) (*models.UserListResponse, error) {
	if req == nil {
		return nil, errors.New("list request cannot be nil")
	}

	// Check permissions
	if !s.canViewUsers(requester) {
		return nil, errors.New("insufficient permissions to list users")
	}

	// Apply role-based filtering
	if err := s.applyRoleFiltering(req, requester); err != nil {
		return nil, err
	}

	// List users
	return s.userRepo.List(ctx, req)
}

// GetUserProfile retrieves the current user's profile
func (s *userService) GetUserProfile(ctx context.Context, user *models.User) (*models.UserResponse, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUserProfile updates the current user's profile
func (s *userService) UpdateUserProfile(ctx context.Context, user *models.User, req *models.UserUpdateRequest) (*models.UserResponse, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}

	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// Users can only update their own profile
	// Role and IsActive cannot be changed through profile update
	if req.Role != nil || req.IsActive != nil {
		return nil, errors.New("role and status cannot be changed through profile update")
	}

	// Update allowed fields
	if req.Name != nil {
		user.Name = *req.Name
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// ChangeUserPIN changes a user's PIN
func (s *userService) ChangeUserPIN(ctx context.Context, user *models.User, currentPIN, newPIN string) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if currentPIN == "" {
		return errors.New("current PIN is required")
	}

	if newPIN == "" {
		return errors.New("new PIN is required")
	}

	// Verify current PIN
	if !user.VerifyPIN(currentPIN) {
		return errors.New("current PIN is incorrect")
	}

	// Validate new PIN
	if len(newPIN) < 4 || len(newPIN) > 6 {
		return errors.New("new PIN must be 4-6 digits")
	}

	// Update PIN (will be hashed in BeforeUpdate hook)
	user.PIN = newPIN

	// Update user
	return s.userRepo.Update(ctx, user)
}

// GetUsersByRole retrieves users by role
func (s *userService) GetUsersByRole(ctx context.Context, role models.UserRole, requester *models.User) ([]models.UserResponse, error) {
	// Check permissions
	if !s.canViewUsers(requester) {
		return nil, errors.New("insufficient permissions to view users by role")
	}

	// Get users by role
	users, err := s.userRepo.GetUsersByRole(ctx, role)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

// ActivateUser activates a user account
func (s *userService) ActivateUser(ctx context.Context, id uint, requester *models.User) error {
	if id == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Check permissions
	if !s.canManageUsers(requester) {
		return errors.New("insufficient permissions to activate users")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Activate user
	user.IsActive = true
	return s.userRepo.Update(ctx, user)
}

// DeactivateUser deactivates a user account
func (s *userService) DeactivateUser(ctx context.Context, id uint, requester *models.User) error {
	if id == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Check permissions
	if !s.canManageUsers(requester) {
		return errors.New("insufficient permissions to deactivate users")
	}

	// Prevent self-deactivation
	if requester.ID == id {
		return errors.New("cannot deactivate your own account")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Deactivate user
	user.IsActive = false
	return s.userRepo.Update(ctx, user)
}

// Permission checking methods

// canManageUsers checks if user can manage other users
func (s *userService) canManageUsers(user *models.User) bool {
	return user != nil && (user.IsAdmin() || user.IsOwner())
}

// canViewUsers checks if user can view user list
func (s *userService) canViewUsers(user *models.User) bool {
	return user != nil && (user.IsAdmin() || user.IsOwner())
}

// canViewUser checks if user can view a specific user
func (s *userService) canViewUser(user *models.User, targetUserID uint) bool {
	if user == nil {
		return false
	}
	// Users can view their own profile, admins and owners can view all
	return user.ID == targetUserID || user.IsAdmin() || user.IsOwner()
}

// canManageUser checks if user can manage a specific user
func (s *userService) canManageUser(user *models.User, targetUserID uint) bool {
	if user == nil {
		return false
	}
	// Users can manage their own profile (limited), admins and owners can manage all
	return user.ID == targetUserID || user.IsAdmin() || user.IsOwner()
}

// validateRoleAssignment validates if a user can assign a specific role
func (s *userService) validateRoleAssignment(role models.UserRole, assigner *models.User) error {
	if assigner == nil {
		return errors.New("assigner cannot be nil")
	}

	// Only admins can assign admin roles
	if role == models.UserRoleAdmin && !assigner.IsAdmin() {
		return errors.New("only admins can assign admin roles")
	}

	// Only admins and owners can assign owner roles
	if role == models.UserRoleOwner && !assigner.IsAdmin() && !assigner.IsOwner() {
		return errors.New("only admins and owners can assign owner roles")
	}

	return nil
}

// applyRoleFiltering applies role-based filtering to user list requests
func (s *userService) applyRoleFiltering(req *models.UserListRequest, requester *models.User) error {
	if requester == nil {
		return errors.New("requester cannot be nil")
	}

	// Staff users can only see other staff users
	if requester.IsStaff() {
		staffRole := models.UserRoleStaff
		req.Role = &staffRole
	}

	return nil
}
