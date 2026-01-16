package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/services/auth"
)

var (
	ErrUserExists = errors.New("user with this email already exists")
)

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail retrieves a user by email for authentication
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, password_hash, nome, role, ativo
		FROM users
		WHERE email = $1
	`

	var user auth.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nome,
		&user.Role,
		&user.Ativo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	// Get first hospital for auth claims (for backward compatibility)
	hospitals, err := r.GetUserHospitals(ctx, user.ID)
	if err == nil && len(hospitals) > 0 {
		user.HospitalID = &hospitals[0].ID
	}

	return &user, nil
}

// GetByID retrieves a user by ID for authentication
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `
		SELECT id, email, password_hash, nome, role, ativo
		FROM users
		WHERE id = $1
	`

	var user auth.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nome,
		&user.Role,
		&user.Ativo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	// Get first hospital for auth claims (for backward compatibility)
	hospitals, err := r.GetUserHospitals(ctx, user.ID)
	if err == nil && len(hospitals) > 0 {
		user.HospitalID = &hospitals[0].ID
	}

	return &user, nil
}

// Create creates a new user for auth service
func (r *UserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, nome, role, email_notifications, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, $6, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Nome,
		user.Role,
		user.Ativo,
	)

	if err != nil {
		return err
	}

	// Handle hospital association if provided
	if user.HospitalID != nil {
		return r.SetUserHospitals(ctx, user.ID, []uuid.UUID{*user.HospitalID})
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetUserHospitals retrieves all hospitals associated with a user
func (r *UserRepository) GetUserHospitals(ctx context.Context, userID uuid.UUID) ([]models.Hospital, error) {
	query := `
		SELECT h.id, h.nome, h.codigo, h.endereco, h.ativo, h.created_at, h.updated_at
		FROM hospitals h
		INNER JOIN user_hospitals uh ON h.id = uh.hospital_id
		WHERE uh.user_id = $1
		ORDER BY h.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []models.Hospital
	for rows.Next() {
		var h models.Hospital
		var endereco sql.NullString

		err := rows.Scan(
			&h.ID, &h.Nome, &h.Codigo, &endereco, &h.Ativo, &h.CreatedAt, &h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endereco.Valid {
			h.Endereco = &endereco.String
		}

		hospitals = append(hospitals, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hospitals, nil
}

// SetUserHospitals sets the hospitals for a user (replaces all existing associations)
func (r *UserRepository) SetUserHospitals(ctx context.Context, userID uuid.UUID, hospitalIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing associations
	_, err = tx.ExecContext(ctx, `DELETE FROM user_hospitals WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Insert new associations
	if len(hospitalIDs) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO user_hospitals (user_id, hospital_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, hospital_id) DO NOTHING
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, hospitalID := range hospitalIDs {
			_, err = stmt.ExecContext(ctx, userID, hospitalID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// ListWithPagination returns users with pagination, search, and filters
func (r *UserRepository) ListWithPagination(ctx context.Context, params *models.UserListParams) (*models.UserListResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Status == "" {
		params.Status = "all"
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(u.nome ILIKE $%d OR u.email ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if params.Status == "active" {
		conditions = append(conditions, fmt.Sprintf("u.ativo = $%d", argIndex))
		args = append(args, true)
		argIndex++
	} else if params.Status == "inactive" {
		conditions = append(conditions, fmt.Sprintf("u.ativo = $%d", argIndex))
		args = append(args, false)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM users u %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	offset := (params.Page - 1) * params.PerPage

	// Get users
	query := fmt.Sprintf(`
		SELECT u.id, u.email, u.nome, u.role, u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at
		FROM users u
		%s
		ORDER BY u.nome ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var mobilePhone sql.NullString

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &mobilePhone, &u.EmailNotifications, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if mobilePhone.Valid {
			u.MobilePhone = &mobilePhone.String
		}

		// Load hospitals for each user
		hospitals, err := r.GetUserHospitals(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		u.Hospitals = hospitals

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.UserListResult{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// List returns all users with hospital data (deprecated - use ListWithPagination)
func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	result, err := r.ListWithPagination(ctx, &models.UserListParams{
		Page:    1,
		PerPage: 1000, // High limit for backward compatibility
		Status:  "all",
	})
	if err != nil {
		return nil, err
	}
	return result.Users, nil
}

// ListByRole returns all active users with a specific role
func (r *UserRepository) ListByRole(ctx context.Context, role string) ([]models.User, error) {
	query := `
		SELECT u.id, u.email, u.nome, u.role, u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at
		FROM users u
		WHERE u.role = $1 AND u.ativo = true
		ORDER BY u.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var mobilePhone sql.NullString

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &mobilePhone, &u.EmailNotifications, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if mobilePhone.Valid {
			u.MobilePhone = &mobilePhone.String
		}

		// Load hospitals for each user
		hospitals, err := r.GetUserHospitals(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		u.Hospitals = hospitals

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ListByRoleAndHospital returns all active users with a specific role linked to a specific hospital
func (r *UserRepository) ListByRoleAndHospital(ctx context.Context, role string, hospitalID uuid.UUID) ([]models.User, error) {
	query := `
		SELECT u.id, u.email, u.nome, u.role, u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_hospitals uh ON u.id = uh.user_id
		WHERE u.role = $1 AND u.ativo = true AND uh.hospital_id = $2
		ORDER BY u.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, role, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var mobilePhone sql.NullString

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &mobilePhone, &u.EmailNotifications, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if mobilePhone.Valid {
			u.MobilePhone = &mobilePhone.String
		}

		// Load hospitals for each user
		hospitals, err := r.GetUserHospitals(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		u.Hospitals = hospitals

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ListByRoleWithEmailNotifications returns active users with a specific role that have email notifications enabled
func (r *UserRepository) ListByRoleWithEmailNotifications(ctx context.Context, role string) ([]models.User, error) {
	query := `
		SELECT u.id, u.email, u.nome, u.role, u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at
		FROM users u
		WHERE u.role = $1 AND u.ativo = true AND u.email_notifications = true
		ORDER BY u.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var mobilePhone sql.NullString

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &mobilePhone, &u.EmailNotifications, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if mobilePhone.Valid {
			u.MobilePhone = &mobilePhone.String
		}

		// Load hospitals for each user
		hospitals, err := r.GetUserHospitals(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		u.Hospitals = hospitals

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ListOperatorsByHospital returns all active operators linked to a specific hospital
func (r *UserRepository) ListOperatorsByHospital(ctx context.Context, hospitalID uuid.UUID) ([]models.User, error) {
	return r.ListByRoleAndHospital(ctx, string(models.RoleOperador), hospitalID)
}

// GetModelByID retrieves a user by ID with hospital data
func (r *UserRepository) GetModelByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.nome, u.role, u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at
		FROM users u
		WHERE u.id = $1
	`

	var u models.User
	var mobilePhone sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Nome, &u.Role, &mobilePhone, &u.EmailNotifications, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	if mobilePhone.Valid {
		u.MobilePhone = &mobilePhone.String
	}

	// Load hospitals
	hospitals, err := r.GetUserHospitals(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Hospitals = hospitals

	return &u, nil
}

// CreateUser creates a new user from input
func (r *UserRepository) CreateUser(ctx context.Context, input *models.CreateUserInput, passwordHash string) (*models.User, error) {
	// Check if email already exists
	exists, err := r.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Default email notifications to true
	emailNotifications := true
	if input.EmailNotifications != nil {
		emailNotifications = *input.EmailNotifications
	}

	user := &models.User{
		ID:                 uuid.New(),
		Email:              input.Email,
		PasswordHash:       passwordHash,
		Nome:               input.Nome,
		Role:               input.Role,
		MobilePhone:        input.MobilePhone,
		EmailNotifications: emailNotifications,
		Ativo:              true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users (id, email, password_hash, nome, role, mobile_phone, email_notifications, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = tx.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Nome,
		user.Role,
		user.MobilePhone,
		user.EmailNotifications,
		user.Ativo,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Set hospital associations
	if len(input.HospitalIDs) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO user_hospitals (user_id, hospital_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, hospital_id) DO NOTHING
		`)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		for _, hospitalID := range input.HospitalIDs {
			_, err = stmt.ExecContext(ctx, user.ID, hospitalID)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Load hospitals for response
	hospitals, err := r.GetUserHospitals(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Hospitals = hospitals

	return user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, id uuid.UUID, input *models.UpdateUserInput, passwordHash *string) (*models.User, error) {
	// Get existing user
	user, err := r.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Note: Email is not updatable per spec
	if input.Nome != nil {
		user.Nome = *input.Nome
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.MobilePhone != nil {
		user.MobilePhone = input.MobilePhone
	}
	if input.EmailNotifications != nil {
		user.EmailNotifications = *input.EmailNotifications
	}
	if input.Ativo != nil {
		user.Ativo = *input.Ativo
	}
	if passwordHash != nil {
		user.PasswordHash = *passwordHash
	}

	user.UpdatedAt = time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET nome = $1, role = $2, mobile_phone = $3, email_notifications = $4, ativo = $5, password_hash = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := tx.ExecContext(ctx, query,
		user.Nome,
		user.Role,
		user.MobilePhone,
		user.EmailNotifications,
		user.Ativo,
		user.PasswordHash,
		user.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, auth.ErrUserNotFound
	}

	// Update hospital associations if provided
	if input.HospitalIDs != nil {
		// Delete existing associations
		_, err = tx.ExecContext(ctx, `DELETE FROM user_hospitals WHERE user_id = $1`, id)
		if err != nil {
			return nil, err
		}

		// Insert new associations
		if len(input.HospitalIDs) > 0 {
			stmt, err := tx.PrepareContext(ctx, `
				INSERT INTO user_hospitals (user_id, hospital_id, created_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT (user_id, hospital_id) DO NOTHING
			`)
			if err != nil {
				return nil, err
			}
			defer stmt.Close()

			for _, hospitalID := range input.HospitalIDs {
				_, err = stmt.ExecContext(ctx, id, hospitalID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Load hospitals for response
	hospitals, err := r.GetUserHospitals(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Hospitals = hospitals

	return user, nil
}

// UpdateProfile updates a user's own profile (name and password only)
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, input *models.UpdateProfileInput, newPasswordHash *string) (*models.User, error) {
	// Get existing user
	user, err := r.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Nome != nil {
		user.Nome = *input.Nome
	}
	if newPasswordHash != nil {
		user.PasswordHash = *newPasswordHash
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET nome = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Nome,
		user.PasswordHash,
		user.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, auth.ErrUserNotFound
	}

	return user, nil
}

// DeactivateUser deactivates a user (soft delete)
func (r *UserRepository) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET ativo = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// GetUsersWithSMSEnabled returns all active users with SMS enabled and mobile phone set
func (r *UserRepository) GetUsersWithSMSEnabled(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT u.id, u.email, u.nome, u.role, u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN user_notification_preferences p ON u.id = p.user_id
		WHERE u.ativo = true
		AND u.mobile_phone IS NOT NULL
		AND u.mobile_phone != ''
		AND (p.sms_enabled IS NULL OR p.sms_enabled = true)
		ORDER BY u.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var mobilePhone sql.NullString

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &mobilePhone, &u.EmailNotifications, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if mobilePhone.Valid {
			u.MobilePhone = &mobilePhone.String
		}

		// Load hospitals for each user
		hospitals, err := r.GetUserHospitals(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		u.Hospitals = hospitals

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
