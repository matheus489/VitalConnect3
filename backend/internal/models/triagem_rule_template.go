package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrTriagemRuleTemplateNotFound is returned when a triagem rule template is not found
	ErrTriagemRuleTemplateNotFound = errors.New("triagem rule template not found")

	// ErrInvalidTriagemRuleTemplateType is returned when the template type is invalid
	ErrInvalidTriagemRuleTemplateType = errors.New("invalid triagem rule template type")
)

// TriagemRuleTemplateType represents the type of triagem rule template
type TriagemRuleTemplateType string

const (
	TemplateTypeIdadeMaxima              TriagemRuleTemplateType = "idade_maxima"
	TemplateTypeCausasExcludentes        TriagemRuleTemplateType = "causas_excludentes"
	TemplateTypeJanelaHoras              TriagemRuleTemplateType = "janela_horas"
	TemplateTypeIdentificacaoDesconhecida TriagemRuleTemplateType = "identificacao_desconhecida"
	TemplateTypeSetorPriorizacao         TriagemRuleTemplateType = "setor_priorizacao"
)

// ValidTriagemRuleTemplateTypes contains all valid template types
var ValidTriagemRuleTemplateTypes = []TriagemRuleTemplateType{
	TemplateTypeIdadeMaxima,
	TemplateTypeCausasExcludentes,
	TemplateTypeJanelaHoras,
	TemplateTypeIdentificacaoDesconhecida,
	TemplateTypeSetorPriorizacao,
}

// IsValid checks if the template type is valid
func (t TriagemRuleTemplateType) IsValid() bool {
	for _, valid := range ValidTriagemRuleTemplateTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the template type
func (t TriagemRuleTemplateType) String() string {
	return string(t)
}

// TriagemRuleTemplateCondition represents the condition configuration for a rule template
type TriagemRuleTemplateCondition struct {
	Valor interface{} `json:"valor"`
	Acao  string      `json:"acao"` // "rejeitar", "priorizar", "alertar"
}

// TriagemRuleTemplate represents a master triagem rule template (global, not tied to any tenant)
type TriagemRuleTemplate struct {
	ID         uuid.UUID               `json:"id" db:"id"`
	Nome       string                  `json:"nome" db:"nome" validate:"required,min=2,max=255"`
	Tipo       TriagemRuleTemplateType `json:"tipo" db:"tipo" validate:"required"`
	Condicao   json.RawMessage         `json:"condicao" db:"condicao" validate:"required"`
	Descricao  *string                 `json:"descricao,omitempty" db:"descricao"`
	Ativo      bool                    `json:"ativo" db:"ativo"`
	Prioridade int                     `json:"prioridade" db:"prioridade"`
	CreatedAt  time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at" db:"updated_at"`
}

// CreateTriagemRuleTemplateInput represents input for creating a triagem rule template
type CreateTriagemRuleTemplateInput struct {
	Nome       string          `json:"nome" validate:"required,min=2,max=255"`
	Tipo       string          `json:"tipo" validate:"required"`
	Condicao   json.RawMessage `json:"condicao" validate:"required"`
	Descricao  *string         `json:"descricao,omitempty" validate:"omitempty,max=1000"`
	Ativo      *bool           `json:"ativo,omitempty"`
	Prioridade *int            `json:"prioridade,omitempty" validate:"omitempty,min=0,max=1000"`
}

// UpdateTriagemRuleTemplateInput represents input for updating a triagem rule template
type UpdateTriagemRuleTemplateInput struct {
	Nome       *string         `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	Tipo       *string         `json:"tipo,omitempty"`
	Condicao   json.RawMessage `json:"condicao,omitempty"`
	Descricao  *string         `json:"descricao,omitempty" validate:"omitempty,max=1000"`
	Ativo      *bool           `json:"ativo,omitempty"`
	Prioridade *int            `json:"prioridade,omitempty" validate:"omitempty,min=0,max=1000"`
}

// CloneTriagemRuleTemplateInput represents input for cloning a template to tenants
type CloneTriagemRuleTemplateInput struct {
	TenantIDs []uuid.UUID `json:"tenant_ids" validate:"required,min=1"`
}

// TriagemRuleTemplateResponse represents the API response for a triagem rule template
type TriagemRuleTemplateResponse struct {
	ID         uuid.UUID       `json:"id"`
	Nome       string          `json:"nome"`
	Tipo       string          `json:"tipo"`
	Condicao   json.RawMessage `json:"condicao"`
	Descricao  *string         `json:"descricao,omitempty"`
	Ativo      bool            `json:"ativo"`
	Prioridade int             `json:"prioridade"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// TriagemRuleTemplateWithUsage represents a template with usage information
type TriagemRuleTemplateWithUsage struct {
	TriagemRuleTemplate
	TenantCount int `json:"tenant_count" db:"tenant_count"`
}

// TriagemRuleTemplateWithUsageResponse represents the API response for a template with usage
type TriagemRuleTemplateWithUsageResponse struct {
	TriagemRuleTemplateResponse
	TenantCount int `json:"tenant_count"`
}

// TriagemRuleTemplateFilter represents filters for querying templates
type TriagemRuleTemplateFilter struct {
	Tipo     *string `json:"tipo,omitempty"`
	Ativo    *bool   `json:"ativo,omitempty"`
	Search   *string `json:"search,omitempty"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

// DefaultTriagemRuleTemplateFilters returns default filters
func DefaultTriagemRuleTemplateFilters() *TriagemRuleTemplateFilter {
	return &TriagemRuleTemplateFilter{
		Page:     1,
		PageSize: 20,
	}
}

// Validate validates the template data
func (t *TriagemRuleTemplate) Validate() error {
	if t.Nome == "" || len(t.Nome) < 2 || len(t.Nome) > 255 {
		return errors.New("template nome must be between 2 and 255 characters")
	}

	if !t.Tipo.IsValid() {
		return ErrInvalidTriagemRuleTemplateType
	}

	if len(t.Condicao) == 0 {
		return errors.New("template condicao is required")
	}

	// Validate that condicao is valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(t.Condicao, &js); err != nil {
		return errors.New("template condicao must be valid JSON")
	}

	return nil
}

// ToResponse converts TriagemRuleTemplate to TriagemRuleTemplateResponse
func (t *TriagemRuleTemplate) ToResponse() TriagemRuleTemplateResponse {
	return TriagemRuleTemplateResponse{
		ID:         t.ID,
		Nome:       t.Nome,
		Tipo:       string(t.Tipo),
		Condicao:   t.Condicao,
		Descricao:  t.Descricao,
		Ativo:      t.Ativo,
		Prioridade: t.Prioridade,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
}

// ToWithUsageResponse converts TriagemRuleTemplateWithUsage to TriagemRuleTemplateWithUsageResponse
func (t *TriagemRuleTemplateWithUsage) ToWithUsageResponse() TriagemRuleTemplateWithUsageResponse {
	return TriagemRuleTemplateWithUsageResponse{
		TriagemRuleTemplateResponse: t.TriagemRuleTemplate.ToResponse(),
		TenantCount:                 t.TenantCount,
	}
}

// GetCondition parses and returns the condition
func (t *TriagemRuleTemplate) GetCondition() (*TriagemRuleTemplateCondition, error) {
	var condition TriagemRuleTemplateCondition
	if err := json.Unmarshal(t.Condicao, &condition); err != nil {
		return nil, err
	}
	return &condition, nil
}

// SetCondition sets the condition from a struct
func (t *TriagemRuleTemplate) SetCondition(condition TriagemRuleTemplateCondition) error {
	data, err := json.Marshal(condition)
	if err != nil {
		return err
	}
	t.Condicao = data
	return nil
}

// Validate validates the CreateTriagemRuleTemplateInput
func (i *CreateTriagemRuleTemplateInput) Validate() error {
	if i.Nome == "" || len(i.Nome) < 2 || len(i.Nome) > 255 {
		return errors.New("template nome must be between 2 and 255 characters")
	}

	tipo := TriagemRuleTemplateType(i.Tipo)
	if !tipo.IsValid() {
		return ErrInvalidTriagemRuleTemplateType
	}

	if len(i.Condicao) == 0 {
		return errors.New("template condicao is required")
	}

	// Validate that condicao is valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(i.Condicao, &js); err != nil {
		return errors.New("template condicao must be valid JSON")
	}

	return nil
}

// Validate validates the UpdateTriagemRuleTemplateInput
func (i *UpdateTriagemRuleTemplateInput) Validate() error {
	if i.Nome != nil && (len(*i.Nome) < 2 || len(*i.Nome) > 255) {
		return errors.New("template nome must be between 2 and 255 characters")
	}

	if i.Tipo != nil {
		tipo := TriagemRuleTemplateType(*i.Tipo)
		if !tipo.IsValid() {
			return ErrInvalidTriagemRuleTemplateType
		}
	}

	if len(i.Condicao) > 0 {
		// Validate that condicao is valid JSON
		var js json.RawMessage
		if err := json.Unmarshal(i.Condicao, &js); err != nil {
			return errors.New("template condicao must be valid JSON")
		}
	}

	return nil
}

// Validate validates the CloneTriagemRuleTemplateInput
func (i *CloneTriagemRuleTemplateInput) Validate() error {
	if len(i.TenantIDs) == 0 {
		return errors.New("at least one tenant_id is required")
	}

	// Check for duplicate tenant IDs
	seen := make(map[uuid.UUID]bool)
	for _, id := range i.TenantIDs {
		if seen[id] {
			return errors.New("duplicate tenant_id found")
		}
		seen[id] = true
	}

	return nil
}

// ToTriagemRule converts a template to a tenant-specific triagem rule
// This is used when cloning a template to a tenant
func (t *TriagemRuleTemplate) ToTriagemRule(tenantID uuid.UUID) *TriagemRule {
	return &TriagemRule{
		ID:         uuid.New(),
		Nome:       t.Nome,
		Descricao:  t.Descricao,
		Regras:     t.Condicao, // Map condicao to regras
		Ativo:      t.Ativo,
		Prioridade: t.Prioridade,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
