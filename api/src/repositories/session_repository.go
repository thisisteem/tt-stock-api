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

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *models.Session) error

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id uint) (*models.Session, error)

	// GetByToken retrieves a session by JWT token
	GetByToken(ctx context.Context, token string) (*models.Session, error)

	// GetByRefreshToken retrieves a session by refresh token
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *models.Session) error

	// Delete deletes a session
	Delete(ctx context.Context, id uint) error

	// List retrieves sessions with pagination and filtering
	List(ctx context.Context, req *models.SessionListRequest) (*models.SessionListResponse, error)

	// GetSessionsByUser retrieves sessions for a specific user
	GetSessionsByUser(ctx context.Context, userID uint) ([]models.Session, error)

	// GetActiveSessions retrieves all active sessions
	GetActiveSessions(ctx context.Context) ([]models.Session, error)

	// GetExpiredSessions retrieves all expired sessions
	GetExpiredSessions(ctx context.Context) ([]models.Session, error)

	// RevokeSession revokes a session
	RevokeSession(ctx context.Context, id uint) error

	// RevokeAllUserSessions revokes all sessions for a user
	RevokeAllUserSessions(ctx context.Context, userID uint) error

	// CleanupExpiredSessions removes expired sessions
	CleanupExpiredSessions(ctx context.Context) error

	// Count returns the total number of sessions
	Count(ctx context.Context) (int64, error)

	// CountActiveSessions returns the number of active sessions
	CountActiveSessions(ctx context.Context) (int64, error)
}

// sessionRepository implements the SessionRepository interface
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new SessionRepository instance
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	// Create session
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id uint) (*models.Session, error) {
	if id == 0 {
		return nil, errors.New("session ID cannot be zero")
	}

	var session models.Session
	if err := r.db.WithContext(ctx).Preload("User").First(&session, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &session, nil
}

// GetByToken retrieves a session by JWT token
func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	var session models.Session
	if err := r.db.WithContext(ctx).Preload("User").Where("token = ?", token).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &session, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token cannot be empty")
	}

	var session models.Session
	if err := r.db.WithContext(ctx).Preload("User").Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &session, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	if session.ID == 0 {
		return errors.New("session ID cannot be zero")
	}

	// Check if session exists
	_, err := r.GetByID(ctx, session.ID)
	if err != nil {
		return err
	}

	// Update session
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		return err
	}

	return nil
}

// Delete deletes a session
func (r *sessionRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("session ID cannot be zero")
	}

	// Check if session exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete session
	if err := r.db.WithContext(ctx).Delete(&models.Session{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves sessions with pagination and filtering
func (r *sessionRepository) List(ctx context.Context, req *models.SessionListRequest) (*models.SessionListResponse, error) {
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

	query := r.db.WithContext(ctx).Model(&models.Session{}).Preload("User")

	// Apply filters
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Expired != nil {
		if *req.Expired {
			query = query.Where("expires_at < ?", time.Now())
		} else {
			query = query.Where("expires_at >= ?", time.Now())
		}
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
	var sessions []models.Session
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.Limit).Find(&sessions).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	sessionResponses := make([]models.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = session.ToResponse()
	}

	return &models.SessionListResponse{
		Sessions: sessionResponses,
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

// GetSessionsByUser retrieves sessions for a specific user
func (r *sessionRepository) GetSessionsByUser(ctx context.Context, userID uint) ([]models.Session, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var sessions []models.Session
	if err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetActiveSessions retrieves all active sessions
func (r *sessionRepository) GetActiveSessions(ctx context.Context) ([]models.Session, error) {
	var sessions []models.Session
	if err := r.db.WithContext(ctx).Preload("User").Where("status = ? AND expires_at > ?", models.SessionStatusActive, time.Now()).Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetExpiredSessions retrieves all expired sessions
func (r *sessionRepository) GetExpiredSessions(ctx context.Context) ([]models.Session, error) {
	var sessions []models.Session
	if err := r.db.WithContext(ctx).Preload("User").Where("expires_at < ?", time.Now()).Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

// RevokeSession revokes a session
func (r *sessionRepository) RevokeSession(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("session ID cannot be zero")
	}

	// Check if session exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update session status to revoked
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Where("id = ?", id).Update("status", models.SessionStatusRevoked).Error; err != nil {
		return err
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *sessionRepository) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Update all user sessions to revoked
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Where("user_id = ?", userID).Update("status", models.SessionStatusRevoked).Error; err != nil {
		return err
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	// Delete expired sessions
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error; err != nil {
		return err
	}

	return nil
}

// Count returns the total number of sessions
func (r *sessionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountActiveSessions returns the number of active sessions
func (r *sessionRepository) CountActiveSessions(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Where("status = ? AND expires_at > ?", models.SessionStatusActive, time.Now()).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
