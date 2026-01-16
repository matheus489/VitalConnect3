package models

import (
	"time"

	"github.com/google/uuid"
)

// UserNotificationPreferences represents user preferences for notification channels
type UserNotificationPreferences struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id" validate:"required"`
	SMSEnabled       bool      `json:"sms_enabled" db:"sms_enabled"`
	EmailEnabled     bool      `json:"email_enabled" db:"email_enabled"`
	DashboardEnabled bool      `json:"dashboard_enabled" db:"dashboard_enabled"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// CreateNotificationPreferencesInput represents input for creating notification preferences
type CreateNotificationPreferencesInput struct {
	UserID           uuid.UUID `json:"user_id" validate:"required"`
	SMSEnabled       *bool     `json:"sms_enabled,omitempty"`
	EmailEnabled     *bool     `json:"email_enabled,omitempty"`
	DashboardEnabled *bool     `json:"dashboard_enabled,omitempty"`
}

// UpdateNotificationPreferencesInput represents input for updating notification preferences
// Note: DashboardEnabled is not included because it cannot be changed
type UpdateNotificationPreferencesInput struct {
	SMSEnabled   *bool `json:"sms_enabled,omitempty"`
	EmailEnabled *bool `json:"email_enabled,omitempty"`
}

// NotificationPreferencesResponse represents the API response for notification preferences
type NotificationPreferencesResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	SMSEnabled       bool      `json:"sms_enabled"`
	EmailEnabled     bool      `json:"email_enabled"`
	DashboardEnabled bool      `json:"dashboard_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToResponse converts UserNotificationPreferences to NotificationPreferencesResponse
func (p *UserNotificationPreferences) ToResponse() NotificationPreferencesResponse {
	return NotificationPreferencesResponse{
		ID:               p.ID,
		UserID:           p.UserID,
		SMSEnabled:       p.SMSEnabled,
		EmailEnabled:     p.EmailEnabled,
		DashboardEnabled: p.DashboardEnabled,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

// DefaultPreferences creates default notification preferences for a user
func DefaultPreferences(userID uuid.UUID, hasMobilePhone bool) *UserNotificationPreferences {
	return &UserNotificationPreferences{
		ID:               uuid.New(),
		UserID:           userID,
		SMSEnabled:       hasMobilePhone, // Default to true only if user has mobile phone
		EmailEnabled:     true,           // Default to true
		DashboardEnabled: true,           // Always true, not editable
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
