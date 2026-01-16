package models

import (
	"time"

	"github.com/google/uuid"
)

// OutcomeType represents the outcome type enum
type OutcomeType string

const (
	OutcomeSucessoCaptacao       OutcomeType = "sucesso_captacao"
	OutcomeFamiliaRecusou        OutcomeType = "familia_recusou"
	OutcomeContraindicacaoMedica OutcomeType = "contraindicacao_medica"
	OutcomeTempoExcedido         OutcomeType = "tempo_excedido"
	OutcomeOutro                 OutcomeType = "outro"
)

// ValidOutcomes contains all valid outcome types
var ValidOutcomes = []OutcomeType{
	OutcomeSucessoCaptacao,
	OutcomeFamiliaRecusou,
	OutcomeContraindicacaoMedica,
	OutcomeTempoExcedido,
	OutcomeOutro,
}

// IsValid checks if the outcome is a valid type
func (o OutcomeType) IsValid() bool {
	for _, valid := range ValidOutcomes {
		if o == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the outcome
func (o OutcomeType) String() string {
	return string(o)
}

// DisplayName returns a human-readable name for the outcome
func (o OutcomeType) DisplayName() string {
	switch o {
	case OutcomeSucessoCaptacao:
		return "Sucesso na captacao"
	case OutcomeFamiliaRecusou:
		return "Familia recusou"
	case OutcomeContraindicacaoMedica:
		return "Contraindicacao medica"
	case OutcomeTempoExcedido:
		return "Tempo excedido"
	case OutcomeOutro:
		return "Outro"
	default:
		return string(o)
	}
}

// OccurrenceHistory represents a history entry for an occurrence
type OccurrenceHistory struct {
	ID             uuid.UUID         `json:"id" db:"id"`
	OccurrenceID   uuid.UUID         `json:"occurrence_id" db:"occurrence_id" validate:"required"`
	UserID         *uuid.UUID        `json:"user_id,omitempty" db:"user_id"`
	Acao           string            `json:"acao" db:"acao" validate:"required,min=2,max=100"`
	StatusAnterior *OccurrenceStatus `json:"status_anterior,omitempty" db:"status_anterior"`
	StatusNovo     *OccurrenceStatus `json:"status_novo,omitempty" db:"status_novo"`
	Observacoes    *string           `json:"observacoes,omitempty" db:"observacoes"`
	Desfecho       *OutcomeType      `json:"desfecho,omitempty" db:"desfecho"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`

	// Related data (populated by queries)
	User *User `json:"user,omitempty" db:"-"`
}

// CreateHistoryInput represents input for creating a history entry
type CreateHistoryInput struct {
	OccurrenceID   uuid.UUID         `json:"occurrence_id" validate:"required"`
	UserID         *uuid.UUID        `json:"user_id,omitempty"`
	Acao           string            `json:"acao" validate:"required,min=2,max=100"`
	StatusAnterior *OccurrenceStatus `json:"status_anterior,omitempty"`
	StatusNovo     *OccurrenceStatus `json:"status_novo,omitempty"`
	Observacoes    *string           `json:"observacoes,omitempty" validate:"omitempty,max=1000"`
	Desfecho       *OutcomeType      `json:"desfecho,omitempty"`
}

// OccurrenceHistoryResponse represents the API response for history entries
type OccurrenceHistoryResponse struct {
	ID             uuid.UUID         `json:"id"`
	OccurrenceID   uuid.UUID         `json:"occurrence_id"`
	UserID         *uuid.UUID        `json:"user_id,omitempty"`
	UserNome       *string           `json:"user_nome,omitempty"`
	Acao           string            `json:"acao"`
	StatusAnterior *OccurrenceStatus `json:"status_anterior,omitempty"`
	StatusNovo     *OccurrenceStatus `json:"status_novo,omitempty"`
	Observacoes    *string           `json:"observacoes,omitempty"`
	Desfecho       *OutcomeType      `json:"desfecho,omitempty"`
	DesfechoNome   *string           `json:"desfecho_nome,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// ToResponse converts OccurrenceHistory to OccurrenceHistoryResponse
func (h *OccurrenceHistory) ToResponse() OccurrenceHistoryResponse {
	resp := OccurrenceHistoryResponse{
		ID:             h.ID,
		OccurrenceID:   h.OccurrenceID,
		UserID:         h.UserID,
		Acao:           h.Acao,
		StatusAnterior: h.StatusAnterior,
		StatusNovo:     h.StatusNovo,
		Observacoes:    h.Observacoes,
		Desfecho:       h.Desfecho,
		CreatedAt:      h.CreatedAt,
	}

	if h.User != nil {
		resp.UserNome = &h.User.Nome
	}

	if h.Desfecho != nil {
		displayName := h.Desfecho.DisplayName()
		resp.DesfechoNome = &displayName
	}

	return resp
}

// Common action strings for history entries
const (
	ActionOccurrenceCreated     = "Ocorrencia criada automaticamente"
	ActionStatusChanged         = "Status alterado"
	ActionOccurrenceAssigned    = "Ocorrencia assumida"
	ActionOccurrenceAccepted    = "Ocorrencia aceita"
	ActionOccurrenceRefused     = "Ocorrencia recusada"
	ActionOccurrenceCanceled    = "Ocorrencia cancelada"
	ActionOccurrenceConcluded   = "Ocorrencia concluida"
	ActionOutcomeRegistered     = "Desfecho registrado"
	ActionNotificationSent      = "Notificacao enviada"
)
