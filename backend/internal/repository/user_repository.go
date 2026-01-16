package repository

import (
	"context"
	"database/sql"
	"errors"
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

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, password_hash, nome, role, hospital_id, ativo
		FROM users
		WHERE email = $1
	`

	var user auth.User
	var hospitalID sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nome,
		&user.Role,
		&hospitalID,
		&user.Ativo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	if hospitalID.Valid {
		hID, err := uuid.Parse(hospitalID.String)
		if err == nil {
			user.HospitalID = &hID
		}
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `
		SELECT id, email, password_hash, nome, role, hospital_id, ativo
		FROM users
		WHERE id = $1
	`

	var user auth.User
	var hospitalID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nome,
		&user.Role,
		&hospitalID,
		&user.Ativo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	if hospitalID.Valid {
		hID, err := uuid.Parse(hospitalID.String)
		if err == nil {
			user.HospitalID = &hID
		}
	}

	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, nome, role, hospital_id, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	var hospitalID interface{}
	if user.HospitalID != nil {
		hospitalID = user.HospitalID
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Nome,
		user.Role,
		hospitalID,
		user.Ativo,
	)

	return err
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

// List returns all users with optional hospital data
func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT
			u.id, u.email, u.nome, u.role, u.hospital_id, u.ativo, u.created_at, u.updated_at,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM users u
		LEFT JOIN hospitals h ON u.hospital_id = h.id
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
		var hospitalID sql.NullString
		var hID sql.NullString
		var hNome, hCodigo sql.NullString
		var hEndereco sql.NullString
		var hAtivo sql.NullBool

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &hospitalID, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
			&hID, &hNome, &hCodigo, &hEndereco, &hAtivo,
		)
		if err != nil {
			return nil, err
		}

		if hospitalID.Valid {
			hIDParsed, err := uuid.Parse(hospitalID.String)
			if err == nil {
				u.HospitalID = &hIDParsed
			}
		}

		if hID.Valid && hNome.Valid {
			hospital := &models.Hospital{
				Nome:   hNome.String,
				Codigo: hCodigo.String,
				Ativo:  hAtivo.Bool,
			}
			if hIDParsed, err := uuid.Parse(hID.String); err == nil {
				hospital.ID = hIDParsed
			}
			if hEndereco.Valid {
				hospital.Endereco = &hEndereco.String
			}
			u.Hospital = hospital
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ListByRole returns all active users with a specific role
func (r *UserRepository) ListByRole(ctx context.Context, role string) ([]models.User, error) {
	query := `
		SELECT
			u.id, u.email, u.nome, u.role, u.hospital_id, u.ativo, u.created_at, u.updated_at,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM users u
		LEFT JOIN hospitals h ON u.hospital_id = h.id
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
		var hospitalID sql.NullString
		var hID sql.NullString
		var hNome, hCodigo sql.NullString
		var hEndereco sql.NullString
		var hAtivo sql.NullBool

		err := rows.Scan(
			&u.ID, &u.Email, &u.Nome, &u.Role, &hospitalID, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
			&hID, &hNome, &hCodigo, &hEndereco, &hAtivo,
		)
		if err != nil {
			return nil, err
		}

		if hospitalID.Valid {
			hIDParsed, err := uuid.Parse(hospitalID.String)
			if err == nil {
				u.HospitalID = &hIDParsed
			}
		}

		if hID.Valid && hNome.Valid {
			hospital := &models.Hospital{
				Nome:   hNome.String,
				Codigo: hCodigo.String,
				Ativo:  hAtivo.Bool,
			}
			if hIDParsed, err := uuid.Parse(hID.String); err == nil {
				hospital.ID = hIDParsed
			}
			if hEndereco.Valid {
				hospital.Endereco = &hEndereco.String
			}
			u.Hospital = hospital
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetModelByID retrieves a user by ID with hospital data
func (r *UserRepository) GetModelByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT
			u.id, u.email, u.password_hash, u.nome, u.role, u.hospital_id, u.ativo, u.created_at, u.updated_at,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM users u
		LEFT JOIN hospitals h ON u.hospital_id = h.id
		WHERE u.id = $1
	`

	var u models.User
	var hospitalID sql.NullString
	var hID sql.NullString
	var hNome, hCodigo sql.NullString
	var hEndereco sql.NullString
	var hAtivo sql.NullBool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Nome, &u.Role, &hospitalID, &u.Ativo, &u.CreatedAt, &u.UpdatedAt,
		&hID, &hNome, &hCodigo, &hEndereco, &hAtivo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	if hospitalID.Valid {
		hIDParsed, err := uuid.Parse(hospitalID.String)
		if err == nil {
			u.HospitalID = &hIDParsed
		}
	}

	if hID.Valid && hNome.Valid {
		hospital := &models.Hospital{
			Nome:   hNome.String,
			Codigo: hCodigo.String,
			Ativo:  hAtivo.Bool,
		}
		if hIDParsed, err := uuid.Parse(hID.String); err == nil {
			hospital.ID = hIDParsed
		}
		if hEndereco.Valid {
			hospital.Endereco = &hEndereco.String
		}
		u.Hospital = hospital
	}

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

	user := &models.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: passwordHash,
		Nome:         input.Nome,
		Role:         input.Role,
		HospitalID:   input.HospitalID,
		Ativo:        true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, password_hash, nome, role, hospital_id, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Nome,
		user.Role,
		user.HospitalID,
		user.Ativo,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, id uuid.UUID, input *models.UpdateUserInput, passwordHash *string) (*models.User, error) {
	// Get existing user
	user, err := r.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check email uniqueness if changing
	if input.Email != nil && *input.Email != user.Email {
		exists, err := r.ExistsByEmail(ctx, *input.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUserExists
		}
		user.Email = *input.Email
	}

	if input.Nome != nil {
		user.Nome = *input.Nome
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.HospitalID != nil {
		user.HospitalID = input.HospitalID
	}
	if input.Ativo != nil {
		user.Ativo = *input.Ativo
	}
	if passwordHash != nil {
		user.PasswordHash = *passwordHash
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = $1, nome = $2, role = $3, hospital_id = $4, ativo = $5, password_hash = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Nome,
		user.Role,
		user.HospitalID,
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
