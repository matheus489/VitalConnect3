package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/services"
)

var (
	// ErrAdminSettingNotFound is returned when a system setting is not found
	ErrAdminSettingNotFound = errors.New("system setting not found")

	// ErrAdminSettingKeyExists is returned when a system setting key already exists
	ErrAdminSettingKeyExists = errors.New("system setting with this key already exists")
)

// AdminSettingsRepository handles admin-level system settings data access
type AdminSettingsRepository struct {
	db                *sql.DB
	encryptionService *services.EncryptionService
}

// NewAdminSettingsRepository creates a new admin settings repository
func NewAdminSettingsRepository(db *sql.DB, encSvc *services.EncryptionService) *AdminSettingsRepository {
	return &AdminSettingsRepository{
		db:                db,
		encryptionService: encSvc,
	}
}

// GetAllSettings retrieves all system settings
// For encrypted settings, the value is returned as encrypted (masked in handler)
func (r *AdminSettingsRepository) GetAllSettings(ctx context.Context) ([]models.SystemSetting, error) {
	query := `
		SELECT id, key, value, description, is_encrypted, created_at, updated_at
		FROM system_settings
		ORDER BY key ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	var settings []models.SystemSetting
	for rows.Next() {
		var s models.SystemSetting
		var description sql.NullString
		var value string

		err := rows.Scan(
			&s.ID,
			&s.Key,
			&value,
			&description,
			&s.IsEncrypted,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		s.Value = json.RawMessage(value)
		if description.Valid {
			s.Description = &description.String
		}

		settings = append(settings, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}

// GetSettingByKey retrieves a single setting by its key
func (r *AdminSettingsRepository) GetSettingByKey(ctx context.Context, key string) (*models.SystemSetting, error) {
	query := `
		SELECT id, key, value, description, is_encrypted, created_at, updated_at
		FROM system_settings
		WHERE key = $1
	`

	var s models.SystemSetting
	var description sql.NullString
	var value string

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&s.ID,
		&s.Key,
		&value,
		&description,
		&s.IsEncrypted,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminSettingNotFound
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	s.Value = json.RawMessage(value)
	if description.Valid {
		s.Description = &description.String
	}

	return &s, nil
}

// UpsertSetting creates or updates a system setting
// If is_encrypted is true and encryption service is available, the value is encrypted before storage
func (r *AdminSettingsRepository) UpsertSetting(ctx context.Context, input *models.CreateSystemSettingInput) (*models.SystemSetting, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Prepare the value for storage
	valueToStore := string(input.Value)

	// Encrypt if needed and encryption service is available
	if input.IsEncrypted && r.encryptionService != nil {
		encrypted, err := r.encryptionService.EncryptValue(valueToStore)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt value: %w", err)
		}
		// Store encrypted value as JSON string
		encryptedJSON, _ := json.Marshal(encrypted)
		valueToStore = string(encryptedJSON)
	}

	// Check if setting exists
	existing, err := r.GetSettingByKey(ctx, input.Key)
	if err != nil && !errors.Is(err, ErrAdminSettingNotFound) {
		return nil, err
	}

	now := time.Now()

	if existing != nil {
		// Update existing setting
		query := `
			UPDATE system_settings
			SET value = $1, description = $2, is_encrypted = $3, updated_at = $4
			WHERE key = $5
			RETURNING id, key, value, description, is_encrypted, created_at, updated_at
		`

		var s models.SystemSetting
		var desc sql.NullString
		var value string

		err := r.db.QueryRowContext(ctx, query, valueToStore, input.Description, input.IsEncrypted, now, input.Key).Scan(
			&s.ID,
			&s.Key,
			&value,
			&desc,
			&s.IsEncrypted,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update setting: %w", err)
		}

		s.Value = json.RawMessage(value)
		if desc.Valid {
			s.Description = &desc.String
		}

		return &s, nil
	}

	// Create new setting
	id := uuid.New()
	query := `
		INSERT INTO system_settings (id, key, value, description, is_encrypted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, key, value, description, is_encrypted, created_at, updated_at
	`

	var s models.SystemSetting
	var desc sql.NullString
	var value string

	err = r.db.QueryRowContext(ctx, query, id, input.Key, valueToStore, input.Description, input.IsEncrypted, now, now).Scan(
		&s.ID,
		&s.Key,
		&value,
		&desc,
		&s.IsEncrypted,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create setting: %w", err)
	}

	s.Value = json.RawMessage(value)
	if desc.Valid {
		s.Description = &desc.String
	}

	return &s, nil
}

// UpdateSetting updates an existing system setting
func (r *AdminSettingsRepository) UpdateSetting(ctx context.Context, key string, input *models.UpdateSystemSettingInput) (*models.SystemSetting, error) {
	// Get existing setting
	existing, err := r.GetSettingByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// Prepare value if provided
	var valueToStore *string
	if input.Value != nil {
		v := string(input.Value)
		isEncrypted := existing.IsEncrypted
		if input.IsEncrypted != nil {
			isEncrypted = *input.IsEncrypted
		}

		// Encrypt if needed
		if isEncrypted && r.encryptionService != nil {
			encrypted, err := r.encryptionService.EncryptValue(v)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt value: %w", err)
			}
			encryptedJSON, _ := json.Marshal(encrypted)
			encV := string(encryptedJSON)
			valueToStore = &encV
		} else {
			valueToStore = &v
		}
	}

	// Build dynamic update
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if valueToStore != nil {
		setClauses = append(setClauses, fmt.Sprintf("value = $%d", argIndex))
		args = append(args, *valueToStore)
		argIndex++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *input.Description)
		argIndex++
	}
	if input.IsEncrypted != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_encrypted = $%d", argIndex))
		args = append(args, *input.IsEncrypted)
		argIndex++
	}

	if len(setClauses) == 0 {
		return existing, nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, key)

	query := fmt.Sprintf(`
		UPDATE system_settings
		SET %s
		WHERE key = $%d
		RETURNING id, key, value, description, is_encrypted, created_at, updated_at
	`, fmt.Sprintf("%s", implode(setClauses, ", ")), argIndex)

	var s models.SystemSetting
	var desc sql.NullString
	var value string

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&s.ID,
		&s.Key,
		&value,
		&desc,
		&s.IsEncrypted,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	s.Value = json.RawMessage(value)
	if desc.Valid {
		s.Description = &desc.String
	}

	return &s, nil
}

// DeleteSetting deletes a system setting by key
func (r *AdminSettingsRepository) DeleteSetting(ctx context.Context, key string) error {
	query := `DELETE FROM system_settings WHERE key = $1`

	result, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAdminSettingNotFound
	}

	return nil
}

// GetDecryptedSetting retrieves a setting and decrypts it if encrypted
// This is used internally when the actual value is needed
func (r *AdminSettingsRepository) GetDecryptedSetting(ctx context.Context, key string) (*models.SystemSetting, error) {
	setting, err := r.GetSettingByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	if !setting.IsEncrypted || r.encryptionService == nil {
		return setting, nil
	}

	// Decrypt the value
	var encryptedStr string
	if err := json.Unmarshal(setting.Value, &encryptedStr); err != nil {
		return setting, nil // Return as-is if can't parse as string
	}

	decrypted, err := r.encryptionService.DecryptValue(encryptedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt setting: %w", err)
	}

	setting.Value = json.RawMessage(decrypted)
	return setting, nil
}

// implode joins strings with a separator
func implode(arr []string, sep string) string {
	result := ""
	for i, s := range arr {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
