package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationChannel represents the notification channel enum
type NotificationChannel string

const (
	ChannelDashboard NotificationChannel = "dashboard"
	ChannelEmail     NotificationChannel = "email"
)

// ValidChannels contains all valid notification channels
var ValidChannels = []NotificationChannel{ChannelDashboard, ChannelEmail}

// IsValid checks if the channel is a valid notification channel
func (c NotificationChannel) IsValid() bool {
	for _, valid := range ValidChannels {
		if c == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the channel
func (c NotificationChannel) String() string {
	return string(c)
}

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusEnviado  NotificationStatus = "enviado"
	NotificationStatusFalha    NotificationStatus = "falha"
	NotificationStatusPendente NotificationStatus = "pendente"
)

// Notification represents a sent notification
type Notification struct {
	ID            uuid.UUID           `json:"id" db:"id"`
	OccurrenceID  uuid.UUID           `json:"occurrence_id" db:"occurrence_id" validate:"required"`
	UserID        *uuid.UUID          `json:"user_id,omitempty" db:"user_id"`
	Canal         NotificationChannel `json:"canal" db:"canal" validate:"required,oneof=dashboard email"`
	EnviadoEm     time.Time           `json:"enviado_em" db:"enviado_em"`
	StatusEnvio   NotificationStatus  `json:"status_envio" db:"status_envio"`
	ErroMensagem  *string             `json:"erro_mensagem,omitempty" db:"erro_mensagem"`
	Metadata      json.RawMessage     `json:"metadata,omitempty" db:"metadata"`

	// Related data (populated by queries)
	Occurrence *Occurrence `json:"occurrence,omitempty" db:"-"`
	User       *User       `json:"user,omitempty" db:"-"`
}

// NotificationMetadata represents additional notification data
type NotificationMetadata struct {
	EmailTo       string `json:"email_to,omitempty"`
	EmailSubject  string `json:"email_subject,omitempty"`
	HospitalNome  string `json:"hospital_nome,omitempty"`
	Setor         string `json:"setor,omitempty"`
	TempoRestante string `json:"tempo_restante,omitempty"`
}

// CreateNotificationInput represents input for creating a notification
type CreateNotificationInput struct {
	OccurrenceID uuid.UUID           `json:"occurrence_id" validate:"required"`
	UserID       *uuid.UUID          `json:"user_id,omitempty"`
	Canal        NotificationChannel `json:"canal" validate:"required,oneof=dashboard email"`
	StatusEnvio  NotificationStatus  `json:"status_envio,omitempty"`
	ErroMensagem *string             `json:"erro_mensagem,omitempty"`
	Metadata     json.RawMessage     `json:"metadata,omitempty"`
}

// NotificationResponse represents the API response for a notification
type NotificationResponse struct {
	ID           uuid.UUID           `json:"id"`
	OccurrenceID uuid.UUID           `json:"occurrence_id"`
	UserID       *uuid.UUID          `json:"user_id,omitempty"`
	Canal        NotificationChannel `json:"canal"`
	EnviadoEm    time.Time           `json:"enviado_em"`
	StatusEnvio  NotificationStatus  `json:"status_envio"`
	ErroMensagem *string             `json:"erro_mensagem,omitempty"`
}

// ToResponse converts Notification to NotificationResponse
func (n *Notification) ToResponse() NotificationResponse {
	return NotificationResponse{
		ID:           n.ID,
		OccurrenceID: n.OccurrenceID,
		UserID:       n.UserID,
		Canal:        n.Canal,
		EnviadoEm:    n.EnviadoEm,
		StatusEnvio:  n.StatusEnvio,
		ErroMensagem: n.ErroMensagem,
	}
}

// SSEEvent represents a Server-Sent Event for dashboard notifications
type SSEEvent struct {
	Type         string    `json:"type"`
	OccurrenceID uuid.UUID `json:"occurrence_id"`
	HospitalNome string    `json:"hospital_nome"`
	Setor        string    `json:"setor"`
	DataObito    time.Time `json:"data_obito"`
	TempoRestante string   `json:"tempo_restante"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewOccurrenceSSEEvent creates a new SSE event for a new occurrence
func NewOccurrenceSSEEvent(occurrence *Occurrence, hospitalNome string) SSEEvent {
	setor := ""
	var data OccurrenceCompleteData
	if err := json.Unmarshal(occurrence.DadosCompletos, &data); err == nil {
		setor = data.Setor
	}

	return SSEEvent{
		Type:          "new_occurrence",
		OccurrenceID:  occurrence.ID,
		HospitalNome:  hospitalNome,
		Setor:         setor,
		DataObito:     occurrence.DataObito,
		TempoRestante: occurrence.FormatTimeRemaining(),
		CreatedAt:     time.Now(),
	}
}
