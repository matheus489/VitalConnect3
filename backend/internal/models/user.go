package models

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// MobilePhoneRegex validates phone numbers in E.164 format
var MobilePhoneRegex = regexp.MustCompile(`^\+[1-9]\d{10,14}$`)

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
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	Email              string     `json:"email" db:"email" validate:"required,email,max=255"`
	PasswordHash       string     `json:"-" db:"password_hash"`
	Nome               string     `json:"nome" db:"nome" validate:"required,min=2,max=255"`
	Role               UserRole   `json:"role" db:"role" validate:"required,oneof=operador gestor admin"`
	IsSuperAdmin       bool       `json:"is_super_admin" db:"is_super_admin"`
	MobilePhone        *string    `json:"mobile_phone,omitempty" db:"mobile_phone"`
	EmailNotifications bool       `json:"email_notifications" db:"email_notifications"`
	Ativo              bool       `json:"ativo" db:"ativo"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

	// Related data (populated by queries) - N:N relationship with hospitals
	Hospitals []Hospital `json:"hospitals,omitempty" db:"-"`
}

// UserWithTenant extends User with tenant information for admin views
type UserWithTenant struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	Email              string     `json:"email" db:"email"`
	Nome               string     `json:"nome" db:"nome"`
	Role               UserRole   `json:"role" db:"role"`
	IsSuperAdmin       bool       `json:"is_super_admin" db:"is_super_admin"`
	MobilePhone        *string    `json:"mobile_phone,omitempty" db:"mobile_phone"`
	EmailNotifications bool       `json:"email_notifications" db:"email_notifications"`
	Ativo              bool       `json:"ativo" db:"ativo"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

	// Tenant info (populated by admin queries)
	TenantName *string `json:"tenant_name,omitempty" db:"tenant_name"`
	TenantSlug *string `json:"tenant_slug,omitempty" db:"tenant_slug"`

	// Related data (populated by queries) - N:N relationship with hospitals
	Hospitals []Hospital `json:"hospitals,omitempty" db:"-"`
}

// ToResponse converts UserWithTenant to UserWithTenantResponse
func (u *UserWithTenant) ToResponse() UserWithTenantResponse {
	resp := UserWithTenantResponse{
		ID:                 u.ID,
		TenantID:           u.TenantID,
		Email:              u.Email,
		Nome:               u.Nome,
		Role:               u.Role,
		IsSuperAdmin:       u.IsSuperAdmin,
		Hospitals:          make([]HospitalResponse, 0, len(u.Hospitals)),
		MobilePhone:        u.MobilePhone,
		EmailNotifications: u.EmailNotifications,
		Ativo:              u.Ativo,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
		TenantName:         u.TenantName,
		TenantSlug:         u.TenantSlug,
	}

	for _, h := range u.Hospitals {
		resp.Hospitals = append(resp.Hospitals, h.ToResponse())
	}

	return resp
}

// UserWithTenantResponse represents the API response for a user with tenant info
type UserWithTenantResponse struct {
	ID                 uuid.UUID          `json:"id"`
	TenantID           *uuid.UUID         `json:"tenant_id,omitempty"`
	Email              string             `json:"email"`
	Nome               string             `json:"nome"`
	Role               UserRole           `json:"role"`
	IsSuperAdmin       bool               `json:"is_super_admin,omitempty"`
	Hospitals          []HospitalResponse `json:"hospitals"`
	MobilePhone        *string            `json:"mobile_phone,omitempty"`
	EmailNotifications bool               `json:"email_notifications"`
	Ativo              bool               `json:"ativo"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	TenantName         *string            `json:"tenant_name,omitempty"`
	TenantSlug         *string            `json:"tenant_slug,omitempty"`
}

// AdminUpdateUserRoleInput represents input for updating a user's role (super admin)
type AdminUpdateUserRoleInput struct {
	Role         *UserRole `json:"role,omitempty" validate:"omitempty,oneof=operador gestor admin"`
	IsSuperAdmin *bool     `json:"is_super_admin,omitempty"`
}

// AdminBanUserInput represents input for banning/unbanning a user
type AdminBanUserInput struct {
	Banned    bool    `json:"banned"`
	BanReason *string `json:"ban_reason,omitempty"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email              string      `json:"email" validate:"required,email,max=255"`
	Password           string      `json:"password" validate:"required,min=8,max=72"`
	Nome               string      `json:"nome" validate:"required,min=2,max=255"`
	Role               UserRole    `json:"role" validate:"required,oneof=operador gestor admin"`
	HospitalIDs        []uuid.UUID `json:"hospital_ids,omitempty"`
	MobilePhone        *string     `json:"mobile_phone,omitempty"`
	EmailNotifications *bool       `json:"email_notifications,omitempty"`
}

// UpdateUserInput represents input for updating a user (admin only)
type UpdateUserInput struct {
	Password           *string     `json:"password,omitempty" validate:"omitempty,min=8,max=72"`
	Nome               *string     `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	Role               *UserRole   `json:"role,omitempty" validate:"omitempty,oneof=operador gestor admin"`
	HospitalIDs        []uuid.UUID `json:"hospital_ids,omitempty"`
	MobilePhone        *string     `json:"mobile_phone,omitempty"`
	EmailNotifications *bool       `json:"email_notifications,omitempty"`
	Ativo              *bool       `json:"ativo,omitempty"`
}

// UpdateProfileInput represents input for updating own profile
type UpdateProfileInput struct {
	Nome            *string `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	CurrentPassword *string `json:"current_password,omitempty" validate:"omitempty,min=8,max=72"`
	NewPassword     *string `json:"new_password,omitempty" validate:"omitempty,min=8,max=72"`
}

// UserResponse represents the API response for a user
type UserResponse struct {
	ID                 uuid.UUID          `json:"id"`
	TenantID           *uuid.UUID         `json:"tenant_id,omitempty"`
	Email              string             `json:"email"`
	Nome               string             `json:"nome"`
	Role               UserRole           `json:"role"`
	IsSuperAdmin       bool               `json:"is_super_admin,omitempty"`
	Hospitals          []HospitalResponse `json:"hospitals"`
	MobilePhone        *string            `json:"mobile_phone,omitempty"`
	EmailNotifications bool               `json:"email_notifications"`
	Ativo              bool               `json:"ativo"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	resp := UserResponse{
		ID:                 u.ID,
		TenantID:           u.TenantID,
		Email:              u.Email,
		Nome:               u.Nome,
		Role:               u.Role,
		IsSuperAdmin:       u.IsSuperAdmin,
		Hospitals:          make([]HospitalResponse, 0, len(u.Hospitals)),
		MobilePhone:        u.MobilePhone,
		EmailNotifications: u.EmailNotifications,
		Ativo:              u.Ativo,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}

	for _, h := range u.Hospitals {
		resp.Hospitals = append(resp.Hospitals, h.ToResponse())
	}

	return resp
}

// GetHospitalIDs returns a slice of hospital IDs
func (u *User) GetHospitalIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(u.Hospitals))
	for _, h := range u.Hospitals {
		ids = append(ids, h.ID)
	}
	return ids
}

// HasHospital checks if user is linked to a specific hospital
func (u *User) HasHospital(hospitalID uuid.UUID) bool {
	for _, h := range u.Hospitals {
		if h.ID == hospitalID {
			return true
		}
	}
	return false
}

// ValidateMobilePhone validates a mobile phone number in E.164 format
func ValidateMobilePhone(phone string) bool {
	if phone == "" {
		return true // Empty is valid (optional field)
	}
	return MobilePhoneRegex.MatchString(phone)
}

// MaskMobilePhone masks a phone number for logging (+55119****9999)
func MaskMobilePhone(phone string) string {
	if phone == "" || len(phone) < 8 {
		return phone
	}
	// Keep first 5 chars and last 4 chars, mask the middle
	prefix := phone[:5]
	suffix := phone[len(phone)-4:]
	middle := ""
	for i := 0; i < len(phone)-9; i++ {
		middle += "*"
	}
	return prefix + middle + suffix
}

// CanManageUsers returns true if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin || u.IsSuperAdmin
}

// CanManageHospitals returns true if the user can manage hospitals
func (u *User) CanManageHospitals() bool {
	return u.Role == RoleAdmin || u.IsSuperAdmin
}

// CanManageTriagemRules returns true if the user can manage triagem rules
func (u *User) CanManageTriagemRules() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor || u.IsSuperAdmin
}

// CanViewMetrics returns true if the user can view dashboard metrics
func (u *User) CanViewMetrics() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor || u.IsSuperAdmin
}

// CanOperateOccurrences returns true if the user can operate occurrences
func (u *User) CanOperateOccurrences() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor || u.Role == RoleOperador || u.IsSuperAdmin
}

// CanReceiveSMSNotifications returns true if the user can receive SMS notifications
func (u *User) CanReceiveSMSNotifications() bool {
	return u.Ativo && u.MobilePhone != nil && *u.MobilePhone != ""
}

// CanReceiveEmailNotifications returns true if the user can receive email notifications
func (u *User) CanReceiveEmailNotifications() bool {
	return u.Ativo && u.EmailNotifications
}

// CanManageShifts returns true if the user can create, update, or delete shifts
// Admin can manage all shifts, Gestor can manage shifts for their hospital
func (u *User) CanManageShifts() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor || u.IsSuperAdmin
}

// CanViewShifts returns true if the user can view shifts
// All authenticated users can view shifts
func (u *User) CanViewShifts() bool {
	return u.Role == RoleAdmin || u.Role == RoleGestor || u.Role == RoleOperador || u.IsSuperAdmin
}

// CanManageShiftsForHospital returns true if the user can manage shifts for a specific hospital
func (u *User) CanManageShiftsForHospital(hospitalID uuid.UUID) bool {
	if u.Role == RoleAdmin || u.IsSuperAdmin {
		return true
	}
	if u.Role == RoleGestor && u.HasHospital(hospitalID) {
		return true
	}
	return false
}

// CanSwitchTenantContext returns true if user can switch tenant context (super-admin only)
func (u *User) CanSwitchTenantContext() bool {
	return u.IsSuperAdmin
}

// UserListParams represents parameters for listing users with pagination and filtering
type UserListParams struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty,max=255"`
	Status  string `form:"status" binding:"omitempty,oneof=all active inactive"`
}

// UserListResult represents the result of a paginated user list
type UserListResult struct {
	Users      []User `json:"users"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}
