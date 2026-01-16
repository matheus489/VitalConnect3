package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Hospital represents a hospital integrated with VitalConnect
type Hospital struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Nome          string          `json:"nome" db:"nome" validate:"required,min=2,max=255"`
	Codigo        string          `json:"codigo" db:"codigo" validate:"required,min=2,max=50"`
	Endereco      *string         `json:"endereco,omitempty" db:"endereco"`
	Latitude      *float64        `json:"latitude,omitempty" db:"latitude"`
	Longitude     *float64        `json:"longitude,omitempty" db:"longitude"`
	ConfigConexao json.RawMessage `json:"config_conexao,omitempty" db:"config_conexao"`
	Ativo         bool            `json:"ativo" db:"ativo"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
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
	Endereco      *string         `json:"endereco,omitempty" validate:"omitempty,max=500"`
	Latitude      *float64        `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude     *float64        `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	ConfigConexao json.RawMessage `json:"config_conexao,omitempty"`
}

// UpdateHospitalInput represents input for updating a hospital
type UpdateHospitalInput struct {
	Nome          *string         `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	Codigo        *string         `json:"codigo,omitempty" validate:"omitempty,min=2,max=50,alphanum"`
	Endereco      *string         `json:"endereco,omitempty" validate:"omitempty,max=500"`
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
