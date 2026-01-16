package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Severity represents the severity level of an audit log entry
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarn     Severity = "WARN"
	SeverityCritical Severity = "CRITICAL"
)

// ValidSeverities contains all valid severity levels
var ValidSeverities = []Severity{
	SeverityInfo,
	SeverityWarn,
	SeverityCritical,
}

// IsValid checks if the severity is a valid type
func (s Severity) IsValid() bool {
	for _, valid := range ValidSeverities {
		if s == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the severity
func (s Severity) String() string {
	return string(s)
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	Timestamp    time.Time       `json:"timestamp" db:"timestamp"`
	UsuarioID    *uuid.UUID      `json:"usuario_id,omitempty" db:"usuario_id"`
	ActorName    string          `json:"actor_name" db:"actor_name"`
	Acao         string          `json:"acao" db:"acao" validate:"required,min=2,max=100"`
	EntidadeTipo string          `json:"entidade_tipo" db:"entidade_tipo" validate:"required,min=2,max=100"`
	EntidadeID   string          `json:"entidade_id" db:"entidade_id" validate:"required"`
	HospitalID   *uuid.UUID      `json:"hospital_id,omitempty" db:"hospital_id"`
	Severity     Severity        `json:"severity" db:"severity" validate:"required,oneof=INFO WARN CRITICAL"`
	Detalhes     json.RawMessage `json:"detalhes,omitempty" db:"detalhes"`
	IPAddress    *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string         `json:"user_agent,omitempty" db:"user_agent"`
}

// CreateAuditLogInput represents input for creating an audit log entry
type CreateAuditLogInput struct {
	UsuarioID    *uuid.UUID      `json:"usuario_id,omitempty"`
	ActorName    string          `json:"actor_name" validate:"required"`
	Acao         string          `json:"acao" validate:"required,min=2,max=100"`
	EntidadeTipo string          `json:"entidade_tipo" validate:"required,min=2,max=100"`
	EntidadeID   string          `json:"entidade_id" validate:"required"`
	HospitalID   *uuid.UUID      `json:"hospital_id,omitempty"`
	Severity     Severity        `json:"severity" validate:"required"`
	Detalhes     json.RawMessage `json:"detalhes,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	DataInicio   *time.Time `json:"data_inicio,omitempty"`
	DataFim      *time.Time `json:"data_fim,omitempty"`
	UsuarioID    *uuid.UUID `json:"usuario_id,omitempty"`
	Acao         *string    `json:"acao,omitempty"`
	EntidadeTipo *string    `json:"entidade_tipo,omitempty"`
	EntidadeID   *string    `json:"entidade_id,omitempty"`
	Severity     *Severity  `json:"severity,omitempty"`
	HospitalID   *uuid.UUID `json:"hospital_id,omitempty"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
}

// DefaultAuditLogFilters returns default filters for audit logs
func DefaultAuditLogFilters() *AuditLogFilter {
	return &AuditLogFilter{
		Page:     1,
		PageSize: 20,
	}
}

// AuditLogResponse represents the API response for audit log entries
type AuditLogResponse struct {
	ID           uuid.UUID       `json:"id"`
	Timestamp    time.Time       `json:"timestamp"`
	UsuarioID    *uuid.UUID      `json:"usuario_id,omitempty"`
	ActorName    string          `json:"actor_name"`
	Acao         string          `json:"acao"`
	EntidadeTipo string          `json:"entidade_tipo"`
	EntidadeID   string          `json:"entidade_id"`
	HospitalID   *uuid.UUID      `json:"hospital_id,omitempty"`
	HospitalNome *string         `json:"hospital_nome,omitempty"`
	Severity     Severity        `json:"severity"`
	Detalhes     json.RawMessage `json:"detalhes,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
}

// ToResponse converts AuditLog to AuditLogResponse
func (a *AuditLog) ToResponse() AuditLogResponse {
	return AuditLogResponse{
		ID:           a.ID,
		Timestamp:    a.Timestamp,
		UsuarioID:    a.UsuarioID,
		ActorName:    a.ActorName,
		Acao:         a.Acao,
		EntidadeTipo: a.EntidadeTipo,
		EntidadeID:   a.EntidadeID,
		HospitalID:   a.HospitalID,
		Severity:     a.Severity,
		Detalhes:     a.Detalhes,
		IPAddress:    a.IPAddress,
		UserAgent:    a.UserAgent,
	}
}

// Common action strings for audit log entries
const (
	// Authentication actions
	ActionAuthLogin       = "auth.login"
	ActionAuthLogout      = "auth.logout"
	ActionAuthLoginFailed = "auth.login_failed"

	// Triagem rule actions
	ActionRegraCreate = "regra.create"
	ActionRegraUpdate = "regra.update"
	ActionRegraDelete = "regra.delete"

	// Occurrence actions
	ActionOcorrenciaVisualizar  = "ocorrencia.visualizar"
	ActionOcorrenciaAceitar     = "ocorrencia.aceitar"
	ActionOcorrenciaRecusar     = "ocorrencia.recusar"
	ActionOcorrenciaStatusChange = "ocorrencia.status_change"
	ActionTriagemRejeicao       = "triagem.rejeicao"

	// User actions
	ActionUsuarioCreate    = "usuario.create"
	ActionUsuarioUpdate    = "usuario.update"
	ActionUsuarioDesativar = "usuario.desativar"
)

// VitalConnectBotActor is the name used for system actions
const VitalConnectBotActor = "VitalConnect Bot"
