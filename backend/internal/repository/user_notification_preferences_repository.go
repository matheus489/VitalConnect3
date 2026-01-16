package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	ErrPreferencesNotFound = errors.New("notification preferences not found")
)

// UserNotificationPreferencesRepository handles database operations for user notification preferences
type UserNotificationPreferencesRepository struct {
	db *sql.DB
}

// NewUserNotificationPreferencesRepository creates a new UserNotificationPreferencesRepository
func NewUserNotificationPreferencesRepository(db *sql.DB) *UserNotificationPreferencesRepository {
	return &UserNotificationPreferencesRepository{db: db}
}

// Create creates a new notification preferences record
func (r *UserNotificationPreferencesRepository) Create(ctx context.Context, input *models.CreateNotificationPreferencesInput) (*models.UserNotificationPreferences, error) {
	prefs := &models.UserNotificationPreferences{
		ID:               uuid.New(),
		UserID:           input.UserID,
		SMSEnabled:       true,  // Default
		EmailEnabled:     true,  // Default
		DashboardEnabled: true,  // Always true
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Apply input overrides
	if input.SMSEnabled != nil {
		prefs.SMSEnabled = *input.SMSEnabled
	}
	if input.EmailEnabled != nil {
		prefs.EmailEnabled = *input.EmailEnabled
	}
	// DashboardEnabled is always true, ignore input

	query := `
		INSERT INTO user_notification_preferences (id, user_id, sms_enabled, email_enabled, dashboard_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, sms_enabled, email_enabled, dashboard_enabled, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.SMSEnabled,
		prefs.EmailEnabled,
		prefs.DashboardEnabled,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.SMSEnabled,
		&prefs.EmailEnabled,
		&prefs.DashboardEnabled,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return prefs, nil
}

// GetByUserID retrieves notification preferences by user ID
func (r *UserNotificationPreferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserNotificationPreferences, error) {
	query := `
		SELECT id, user_id, sms_enabled, email_enabled, dashboard_enabled, created_at, updated_at
		FROM user_notification_preferences
		WHERE user_id = $1
	`

	prefs := &models.UserNotificationPreferences{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.SMSEnabled,
		&prefs.EmailEnabled,
		&prefs.DashboardEnabled,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPreferencesNotFound
		}
		return nil, err
	}

	return prefs, nil
}

// Update updates notification preferences for a user
func (r *UserNotificationPreferencesRepository) Update(ctx context.Context, userID uuid.UUID, input *models.UpdateNotificationPreferencesInput) (*models.UserNotificationPreferences, error) {
	// Get existing preferences
	prefs, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.SMSEnabled != nil {
		prefs.SMSEnabled = *input.SMSEnabled
	}
	if input.EmailEnabled != nil {
		prefs.EmailEnabled = *input.EmailEnabled
	}
	// DashboardEnabled is always true, never update

	prefs.UpdatedAt = time.Now()

	query := `
		UPDATE user_notification_preferences
		SET sms_enabled = $2, email_enabled = $3, updated_at = $4
		WHERE user_id = $1
		RETURNING id, user_id, sms_enabled, email_enabled, dashboard_enabled, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		userID,
		prefs.SMSEnabled,
		prefs.EmailEnabled,
		prefs.UpdatedAt,
	).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.SMSEnabled,
		&prefs.EmailEnabled,
		&prefs.DashboardEnabled,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return prefs, nil
}

// EnsureExists creates preferences with defaults if they don't exist, or returns existing ones
func (r *UserNotificationPreferencesRepository) EnsureExists(ctx context.Context, userID uuid.UUID, hasMobilePhone bool) (*models.UserNotificationPreferences, error) {
	// Try to get existing preferences
	prefs, err := r.GetByUserID(ctx, userID)
	if err == nil {
		return prefs, nil
	}

	if !errors.Is(err, ErrPreferencesNotFound) {
		return nil, err
	}

	// Create with defaults
	input := &models.CreateNotificationPreferencesInput{
		UserID:       userID,
		SMSEnabled:   &hasMobilePhone,
		EmailEnabled: boolPtr(true),
	}

	return r.Create(ctx, input)
}

// Delete deletes notification preferences for a user (usually called when user is deleted)
func (r *UserNotificationPreferencesRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_notification_preferences WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPreferencesNotFound
	}

	return nil
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
