package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Hospital represents a hospital integrated with SIDOT
type Hospital struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Nome          string          `json:"nome" db:"nome" validate:"required,min=2,max=255"`
	Codigo        string          `json:"codigo" db:"codigo" validate:"required,min=2,max=50"`
	Endereco      *string         `json:"endereco,omitempty" db:"endereco"`
	Telefone      *string         `json:"telefone,omitempty" db:"telefone" validate:"omitempty,max=20"`
	Latitude      *float64        `json:"latitude,omitempty" db:"latitude"`
	Longitude     *float64        `json:"longitude,omitempty" db:"longitude"`
	ConfigConexao json.RawMessage `json:"config_conexao,omitempty" db:"config_conexao"`
	Ativo         bool            `json:"ativo" db:"ativo"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

// HospitalWithTenant extends Hospital with tenant information for admin views
type HospitalWithTenant struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Nome          string          `json:"nome" db:"nome"`
	Codigo        string          `json:"codigo" db:"codigo"`
	Endereco      *string         `json:"endereco,omitempty" db:"endereco"`
	Telefone      *string         `json:"telefone,omitempty" db:"telefone"`
	Latitude      *float64        `json:"latitude,omitempty" db:"latitude"`
	Longitude     *float64        `json:"longitude,omitempty" db:"longitude"`
	ConfigConexao json.RawMessage `json:"config_conexao,omitempty" db:"config_conexao"`
	Ativo         bool            `json:"ativo" db:"ativo"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`

	// Tenant info (populated by admin queries)
	TenantName *string `json:"tenant_name,omitempty" db:"tenant_name"`
	TenantSlug *string `json:"tenant_slug,omitempty" db:"tenant_slug"`
}

// ToResponse converts HospitalWithTenant to HospitalWithTenantResponse
func (h *HospitalWithTenant) ToResponse() HospitalWithTenantResponse {
	return HospitalWithTenantResponse{
		ID:         h.ID,
		TenantID:   h.TenantID,
		Nome:       h.Nome,
		Codigo:     h.Codigo,
		Endereco:   h.Endereco,
		Telefone:   h.Telefone,
		Latitude:   h.Latitude,
		Longitude:  h.Longitude,
		Ativo:      h.Ativo,
		CreatedAt:  h.CreatedAt,
		UpdatedAt:  h.UpdatedAt,
		TenantName: h.TenantName,
		TenantSlug: h.TenantSlug,
	}
}

// HospitalWithTenantResponse represents the API response for a hospital with tenant info
type HospitalWithTenantResponse struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Nome       string    `json:"nome"`
	Codigo     string    `json:"codigo"`
	Endereco   *string   `json:"endereco,omitempty"`
	Telefone   *string   `json:"telefone,omitempty"`
	Latitude   *float64  `json:"latitude,omitempty"`
	Longitude  *float64  `json:"longitude,omitempty"`
	Ativo      bool      `json:"ativo"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	TenantName *string   `json:"tenant_name,omitempty"`
	TenantSlug *string   `json:"tenant_slug,omitempty"`
}

// AdminReassignHospitalInput represents input for reassigning a hospital to a different tenant
type AdminReassignHospitalInput struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

// HospitalConfig represents the connection configuration for a hospital
type HospitalConfig struct {
	Tipo         string `json:"tipo,omitempty"`          // "simulado", "hl7", "fhir"
	Host         string `json:"host,omitempty"`          // Hostname for integration
	Port         int    `json:"port,omitempty"`          // Port number
	Database     string `json:"database,omitempty"`      // Database name
	PollInterval int    `json:"poll_interval,omitempty"` // Polling interval in seconds
}

// CreateHospitalInput represents input for creating a hospital
type CreateHospitalInput struct {
	Nome          string          `json:"nome" validate:"required,min=2,max=255"`
	Codigo        string          `json:"codigo" validate:"required,min=2,max=50,alphanum"`
	Endereco      string          `json:"endereco" validate:"required,max=500"`
	Telefone      *string         `json:"telefone,omitempty" validate:"omitempty,max=20"`
	Latitude      float64         `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude     float64         `json:"longitude" validate:"required,min=-180,max=180"`
	ConfigConexao json.RawMessage `json:"config_conexao,omitempty"`
	Ativo         *bool           `json:"ativo,omitempty"`
}

// UpdateHospitalInput represents input for updating a hospital
type UpdateHospitalInput struct {
	Nome          *string         `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	Codigo        *string         `json:"codigo,omitempty" validate:"omitempty,min=2,max=50,alphanum"`
	Endereco      *string         `json:"endereco,omitempty" validate:"omitempty,max=500"`
	Telefone      *string         `json:"telefone,omitempty" validate:"omitempty,max=20"`
	Latitude      *float64        `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude     *float64        `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	ConfigConexao json.RawMessage `json:"config_conexao,omitempty"`
	Ativo         *bool           `json:"ativo,omitempty"`
}

// HospitalResponse represents the API response for a hospital
type HospitalResponse struct {
	ID        uuid.UUID `json:"id"`
	Nome      string    `json:"nome"`
	Codigo    string    `json:"codigo"`
	Endereco  *string   `json:"endereco,omitempty"`
	Telefone  *string   `json:"telefone,omitempty"`
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	Ativo     bool      `json:"ativo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Hospital to HospitalResponse
func (h *Hospital) ToResponse() HospitalResponse {
	return HospitalResponse{
		ID:        h.ID,
		Nome:      h.Nome,
		Codigo:    h.Codigo,
		Endereco:  h.Endereco,
		Telefone:  h.Telefone,
		Latitude:  h.Latitude,
		Longitude: h.Longitude,
		Ativo:     h.Ativo,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

// IsActive returns true if the hospital is active and not soft deleted
func (h *Hospital) IsActive() bool {
	return h.Ativo && h.DeletedAt == nil
}

// HasCoordinates returns true if the hospital has valid coordinates
func (h *Hospital) HasCoordinates() bool {
	return h.Latitude != nil && h.Longitude != nil
}
