package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrSystemSettingNotFound is returned when a system setting is not found
	ErrSystemSettingNotFound = errors.New("system setting not found")

	// ErrSystemSettingKeyExists is returned when a system setting key already exists
	ErrSystemSettingKeyExists = errors.New("system setting with this key already exists")

	// ErrInvalidSettingKey is returned when a setting key is invalid
	ErrInvalidSettingKey = errors.New("invalid setting key: must be alphanumeric with underscores, 2-100 characters")
)

// Common system setting keys
const (
	SettingKeySMTPConfig   = "smtp_config"
	SettingKeyTwilioConfig = "twilio_config"
	SettingKeyFCMConfig    = "fcm_config"
)

// SMTPConfig represents the SMTP configuration for email sending
type SMTPConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
}

// TwilioConfig represents the Twilio configuration for SMS sending
type TwilioConfig struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	FromNumber string `json:"from_number"`
}

// FCMConfig represents the Firebase Cloud Messaging configuration
type FCMConfig struct {
	ServerKey string `json:"server_key"`
}

// SystemSetting represents a global system configuration
type SystemSetting struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Key         string          `json:"key" db:"key" validate:"required,min=2,max=100"`
	Value       json.RawMessage `json:"value" db:"value" validate:"required"`
	Description *string         `json:"description,omitempty" db:"description"`
	IsEncrypted bool            `json:"is_encrypted" db:"is_encrypted"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateSystemSettingInput represents input for creating a system setting
type CreateSystemSettingInput struct {
	Key         string          `json:"key" validate:"required,min=2,max=100"`
	Value       json.RawMessage `json:"value" validate:"required"`
	Description *string         `json:"description,omitempty"`
	IsEncrypted bool            `json:"is_encrypted,omitempty"`
}

// UpdateSystemSettingInput represents input for updating a system setting
type UpdateSystemSettingInput struct {
	Value       json.RawMessage `json:"value,omitempty"`
	Description *string         `json:"description,omitempty"`
	IsEncrypted *bool           `json:"is_encrypted,omitempty"`
}

// SystemSettingResponse represents the API response for a system setting
type SystemSettingResponse struct {
	ID          uuid.UUID       `json:"id"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description *string         `json:"description,omitempty"`
	IsEncrypted bool            `json:"is_encrypted"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// SystemSettingMaskedResponse represents the API response with masked sensitive values
type SystemSettingMaskedResponse struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"` // Shows masked value for encrypted settings
	Description *string   `json:"description,omitempty"`
	IsEncrypted bool      `json:"is_encrypted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate validates the system setting data
func (s *SystemSetting) Validate() error {
	if s.Key == "" || len(s.Key) < 2 || len(s.Key) > 100 {
		return ErrInvalidSettingKey
	}

	if len(s.Value) == 0 {
		return errors.New("system setting value is required")
	}

	// Validate that value is valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(s.Value, &js); err != nil {
		return errors.New("system setting value must be valid JSON")
	}

	return nil
}

// ToResponse converts SystemSetting to SystemSettingResponse
func (s *SystemSetting) ToResponse() SystemSettingResponse {
	return SystemSettingResponse{
		ID:          s.ID,
		Key:         s.Key,
		Value:       s.Value,
		Description: s.Description,
		IsEncrypted: s.IsEncrypted,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// ToMaskedResponse converts SystemSetting to SystemSettingMaskedResponse
// For encrypted settings, the value is masked as "********"
func (s *SystemSetting) ToMaskedResponse() SystemSettingMaskedResponse {
	valueStr := string(s.Value)
	if s.IsEncrypted {
		valueStr = "********"
	}

	return SystemSettingMaskedResponse{
		ID:          s.ID,
		Key:         s.Key,
		Value:       valueStr,
		Description: s.Description,
		IsEncrypted: s.IsEncrypted,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// Validate validates the CreateSystemSettingInput
func (i *CreateSystemSettingInput) Validate() error {
	if i.Key == "" || len(i.Key) < 2 || len(i.Key) > 100 {
		return ErrInvalidSettingKey
	}

	if len(i.Value) == 0 {
		return errors.New("system setting value is required")
	}

	// Validate that value is valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(i.Value, &js); err != nil {
		return errors.New("system setting value must be valid JSON")
	}

	return nil
}

// Validate validates the UpdateSystemSettingInput
func (i *UpdateSystemSettingInput) Validate() error {
	if i.Value != nil {
		// Validate that value is valid JSON if provided
		var js json.RawMessage
		if err := json.Unmarshal(i.Value, &js); err != nil {
			return errors.New("system setting value must be valid JSON")
		}
	}

	return nil
}

// GetSMTPConfig parses the value as SMTPConfig
func (s *SystemSetting) GetSMTPConfig() (*SMTPConfig, error) {
	var config SMTPConfig
	if err := json.Unmarshal(s.Value, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetTwilioConfig parses the value as TwilioConfig
func (s *SystemSetting) GetTwilioConfig() (*TwilioConfig, error) {
	var config TwilioConfig
	if err := json.Unmarshal(s.Value, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetFCMConfig parses the value as FCMConfig
func (s *SystemSetting) GetFCMConfig() (*FCMConfig, error) {
	var config FCMConfig
	if err := json.Unmarshal(s.Value, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetValue sets the value from a struct
func (s *SystemSetting) SetValue(value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.Value = data
	return nil
}
