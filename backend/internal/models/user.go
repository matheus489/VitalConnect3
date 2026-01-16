package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the user role enum
type UserRole string

const (
	RoleOperador UserRole = "operador"
	RoleGestor   UserRole = "gestor"
	RoleAdmin    UserRole = "admin"
)

// ValidRoles contains all valid user roles
var ValidRoles = []UserRole{RoleOperador, RoleGestor, RoleAdmin}

// IsValid checks if the role is a valid user role
func (r UserRole) IsValid() bool {
	for _, valid := range ValidRoles {
		if r == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the role
func (r UserRole) String() string {
	return string(r)
}

// User represents a system user
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email" validate:"required,email,max=255"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Nome         string     `json:"nome" db:"nome" validate:"required,min=2,max=255"`
	Role         UserRole   `json:"role" db:"role" validate:"required,oneof=operador gestor admin"`
	HospitalID   *uuid.UUID `json:"hospital_id,omitempty" db:"hospital_id"`
	Ativo        bool       `json:"ativo" db:"ativo"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// Related data (populated by queries)
	Hospital *Hospital `json:"hospital,omitempty" db:"-"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email      string     `json:"email" validate:"required,email,max=255"`
	Password   string     `json:"password" validate:"required,min=8,max=72"`
	Nome       string     `json:"nome" validate:"required,min=2,max=255"`
	Role       UserRole   `json:"role" validate:"required,oneof=operador gestor admin"`
	HospitalID *uuid.UUID `json:"hospital_id,omitempty"`
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Email      *string    `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Password   *string    `json:"password,omitempty" validate:"omitempty,min=8,max=72"`
	Nome       *string    `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	Role       *UserRole  `json:"role,omitempty" validate:"omitempty,oneof=operador gestor admin"`
	HospitalID *uuid.UUID `json:"hospital_id,omitempty"`
	Ativo      *bool      `json:"ativo,omitempty"`
}

// UserResponse represents the API response for a user
type UserResponse struct {
	ID         uuid.UUID         `json:"id"`
	Email      string            `json:"email"`
	Nome       string            `json:"nome"`
	Role       UserRole          `json:"role"`
	HospitalID *uuid.UUID        `json:"hospital_id,omitempty"`
	Hospital   *HospitalResponse `json:"hospital,omitempty"`
	Ativo      bool              `json:"ativo"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	resp := UserResponse{
		ID:         u.ID,
		Email:      u.Email,
		Nome:       u.Nome,
		Role:       u.Role,
		HospitalID: u.HospitalID,
		Ativo:      u.Ativo,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}

	if u.Hospital != nil {
		hospitalResp := u.Hospital.ToResponse()
		resp.Hospital = &hospitalResp
	}

	return resp
}

// CanManageUsers returns true if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}

// CanManageHospitals returns true if the user can manage hospitals
func (u *User) CanManageHospitals() bool {
	return u.Role == RoleAdmin
}

// CanManageTriagemRules returns true if the user can manage triagem rules
func (u *User) CanManageTriagemRules() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor
}

// CanViewMetrics returns true if the user can view dashboard metrics
func (u *User) CanViewMetrics() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor
}

// CanOperateOccurrences returns true if the user can operate occurrences
func (u *User) CanOperateOccurrences() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor || u.Role == RoleOperador
}
