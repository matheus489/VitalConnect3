package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OccurrenceStatus represents the status enum for occurrences
type OccurrenceStatus string

const (
	StatusPendente    OccurrenceStatus = "PENDENTE"
	StatusEmAndamento OccurrenceStatus = "EM_ANDAMENTO"
	StatusAceita      OccurrenceStatus = "ACEITA"
	StatusRecusada    OccurrenceStatus = "RECUSADA"
	StatusCancelada   OccurrenceStatus = "CANCELADA"
	StatusConcluida   OccurrenceStatus = "CONCLUIDA"
)

// ValidStatuses contains all valid occurrence statuses
var ValidStatuses = []OccurrenceStatus{
	StatusPendente,
	StatusEmAndamento,
	StatusAceita,
	StatusRecusada,
	StatusCancelada,
	StatusConcluida,
}

// StatusTransitions defines valid status transitions
var StatusTransitions = map[OccurrenceStatus][]OccurrenceStatus{
	StatusPendente:    {StatusEmAndamento, StatusCancelada},
	StatusEmAndamento: {StatusAceita, StatusRecusada, StatusCancelada},
	StatusAceita:      {StatusConcluida, StatusCancelada},
	StatusRecusada:    {StatusConcluida, StatusCancelada},
	StatusCancelada:   {}, // Terminal state
	StatusConcluida:   {}, // Terminal state
}

// IsValid checks if the status is a valid occurrence status
func (s OccurrenceStatus) IsValid() bool {
	for _, valid := range ValidStatuses {
		if s == valid {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if transition to target status is valid
func (s OccurrenceStatus) CanTransitionTo(target OccurrenceStatus) bool {
	validTargets, exists := StatusTransitions[s]
	if !exists {
		return false
	}

	for _, valid := range validTargets {
		if target == valid {
			return true
		}
	}

	return false
}

// String returns the string representation of the status
func (s OccurrenceStatus) String() string {
	return string(s)
}

// IsTerminal returns true if this is a terminal status
func (s OccurrenceStatus) IsTerminal() bool {
	return s == StatusConcluida || s == StatusCancelada
}

// Occurrence represents an eligible death occurrence
type Occurrence struct {
	ID                    uuid.UUID        `json:"id" db:"id"`
	TenantID              uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	ObitoID               uuid.UUID        `json:"obito_id" db:"obito_id" validate:"required"`
	HospitalID            uuid.UUID        `json:"hospital_id" db:"hospital_id" validate:"required"`
	Status                OccurrenceStatus `json:"status" db:"status" validate:"required,oneof=PENDENTE EM_ANDAMENTO ACEITA RECUSADA CANCELADA CONCLUIDA"`
	ScorePriorizacao      int              `json:"score_priorizacao" db:"score_priorizacao"`
	NomePacienteMascarado string           `json:"nome_paciente_mascarado" db:"nome_paciente_mascarado" validate:"required"`
	DadosCompletos        json.RawMessage  `json:"-" db:"dados_completos"` // Hidden by default (LGPD)
	CreatedAt             time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at" db:"updated_at"`
	NotificadoEm          *time.Time       `json:"notificado_em,omitempty" db:"notificado_em"`
	DataObito             time.Time        `json:"data_obito" db:"data_obito"`
	JanelaExpiraEm        time.Time        `json:"janela_expira_em" db:"janela_expira_em"`

	// Related data (populated by queries)
	Hospital *Hospital      `json:"hospital,omitempty" db:"-"`
	Obito    *ObitoSimulado `json:"-" db:"-"` // Hidden by default (LGPD)
}

// OccurrenceCompleteData represents the complete data stored in dados_completos
type OccurrenceCompleteData struct {
	ObitoID                   uuid.UUID `json:"obito_id"`
	HospitalID                uuid.UUID `json:"hospital_id"`
	NomePaciente              string    `json:"nome_paciente"`
	DataNascimento            time.Time `json:"data_nascimento"`
	DataObito                 time.Time `json:"data_obito"`
	CausaMortis               string    `json:"causa_mortis"`
	Idade                     int       `json:"idade"`
	Prontuario                string    `json:"prontuario,omitempty"`
	Setor                     string    `json:"setor,omitempty"`
	Leito                     string    `json:"leito,omitempty"`
	IdentificacaoDesconhecida bool      `json:"identificacao_desconhecida"`
}

// CreateOccurrenceInput represents input for creating an occurrence
type CreateOccurrenceInput struct {
	ObitoID               uuid.UUID       `json:"obito_id" validate:"required"`
	HospitalID            uuid.UUID       `json:"hospital_id" validate:"required"`
	ScorePriorizacao      int             `json:"score_priorizacao" validate:"min=0,max=100"`
	NomePacienteMascarado string          `json:"nome_paciente_mascarado" validate:"required"`
	DadosCompletos        json.RawMessage `json:"dados_completos" validate:"required"`
	DataObito             time.Time       `json:"data_obito" validate:"required"`
}

// UpdateStatusInput represents input for updating occurrence status
type UpdateStatusInput struct {
	Status      OccurrenceStatus `json:"status" validate:"required,oneof=PENDENTE EM_ANDAMENTO ACEITA RECUSADA CANCELADA CONCLUIDA"`
	Observacoes *string          `json:"observacoes,omitempty" validate:"omitempty,max=1000"`
}

// RegisterOutcomeInput represents input for registering an occurrence outcome
type RegisterOutcomeInput struct {
	Desfecho    OutcomeType `json:"desfecho" validate:"required,oneof=sucesso_captacao familia_recusou contraindicacao_medica tempo_excedido outro"`
	Observacoes *string     `json:"observacoes,omitempty" validate:"omitempty,max=1000"`
}

// OccurrenceListResponse represents the API response for listing occurrences
type OccurrenceListResponse struct {
	ID                    uuid.UUID         `json:"id"`
	HospitalID            uuid.UUID         `json:"hospital_id"`
	Hospital              *HospitalResponse `json:"hospital,omitempty"`
	Status                OccurrenceStatus  `json:"status"`
	ScorePriorizacao      int               `json:"score_priorizacao"`
	NomePacienteMascarado string            `json:"nome_paciente_mascarado"`
	CreatedAt             time.Time         `json:"created_at"`
	NotificadoEm          *time.Time        `json:"notificado_em,omitempty"`
	DataObito             time.Time         `json:"data_obito"`
	JanelaExpiraEm        time.Time         `json:"janela_expira_em"`
	TempoRestante         string            `json:"tempo_restante"`
	Setor                 string            `json:"setor,omitempty"`
}

// OccurrenceDetailResponse represents the API response for occurrence details (includes unmasked data)
type OccurrenceDetailResponse struct {
	ID                    uuid.UUID               `json:"id"`
	HospitalID            uuid.UUID               `json:"hospital_id"`
	Hospital              *HospitalResponse       `json:"hospital,omitempty"`
	Status                OccurrenceStatus        `json:"status"`
	ScorePriorizacao      int                     `json:"score_priorizacao"`
	NomePacienteMascarado string                  `json:"nome_paciente_mascarado"`
	DadosCompletos        *OccurrenceCompleteData `json:"dados_completos"` // Visible in detail view (LGPD)
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	NotificadoEm          *time.Time              `json:"notificado_em,omitempty"`
	DataObito             time.Time               `json:"data_obito"`
	JanelaExpiraEm        time.Time               `json:"janela_expira_em"`
	TempoRestante         string                  `json:"tempo_restante"`
}

// ToListResponse converts Occurrence to OccurrenceListResponse
func (o *Occurrence) ToListResponse() OccurrenceListResponse {
	resp := OccurrenceListResponse{
		ID:                    o.ID,
		HospitalID:            o.HospitalID,
		Status:                o.Status,
		ScorePriorizacao:      o.ScorePriorizacao,
		NomePacienteMascarado: o.NomePacienteMascarado,
		CreatedAt:             o.CreatedAt,
		NotificadoEm:          o.NotificadoEm,
		DataObito:             o.DataObito,
		JanelaExpiraEm:        o.JanelaExpiraEm,
		TempoRestante:         o.FormatTimeRemaining(),
	}

	if o.Hospital != nil {
		hospitalResp := o.Hospital.ToResponse()
		resp.Hospital = &hospitalResp
	}

	// Extract setor from dados_completos
	var data OccurrenceCompleteData
	if err := json.Unmarshal(o.DadosCompletos, &data); err == nil {
		resp.Setor = data.Setor
	}

	return resp
}

// ToDetailResponse converts Occurrence to OccurrenceDetailResponse (includes unmasked data)
func (o *Occurrence) ToDetailResponse() OccurrenceDetailResponse {
	resp := OccurrenceDetailResponse{
		ID:                    o.ID,
		HospitalID:            o.HospitalID,
		Status:                o.Status,
		ScorePriorizacao:      o.ScorePriorizacao,
		NomePacienteMascarado: o.NomePacienteMascarado,
		CreatedAt:             o.CreatedAt,
		UpdatedAt:             o.UpdatedAt,
		NotificadoEm:          o.NotificadoEm,
		DataObito:             o.DataObito,
		JanelaExpiraEm:        o.JanelaExpiraEm,
		TempoRestante:         o.FormatTimeRemaining(),
	}

	if o.Hospital != nil {
		hospitalResp := o.Hospital.ToResponse()
		resp.Hospital = &hospitalResp
	}

	// Unmarshal complete data for detail view
	var data OccurrenceCompleteData
	if err := json.Unmarshal(o.DadosCompletos, &data); err == nil {
		resp.DadosCompletos = &data
	}

	return resp
}

// TimeRemaining returns the time remaining in the capture window
func (o *Occurrence) TimeRemaining() time.Duration {
	remaining := o.JanelaExpiraEm.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// FormatTimeRemaining returns a human-readable string for the time remaining
func (o *Occurrence) FormatTimeRemaining() string {
	remaining := o.TimeRemaining()

	if remaining <= 0 {
		return "Expirado"
	}

	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dmin", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dmin", minutes)
}

// IsExpired returns true if the capture window has expired
func (o *Occurrence) IsExpired() bool {
	return time.Now().After(o.JanelaExpiraEm)
}
