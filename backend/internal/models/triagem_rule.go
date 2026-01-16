package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TriagemRule represents a triagem rule configuration
type TriagemRule struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	Nome       string          `json:"nome" db:"nome" validate:"required,min=2,max=255"`
	Descricao  *string         `json:"descricao,omitempty" db:"descricao"`
	Regras     json.RawMessage `json:"regras" db:"regras" validate:"required"`
	Ativo      bool            `json:"ativo" db:"ativo"`
	Prioridade int             `json:"prioridade" db:"prioridade"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// RuleType represents the type of triagem rule
type RuleType string

const (
	RuleTypeIdadeMaxima              RuleType = "idade_maxima"
	RuleTypeCausasExcludentes        RuleType = "causas_excludentes"
	RuleTypeJanelaHoras              RuleType = "janela_horas"
	RuleTypeIdentificacaoDesconhecida RuleType = "identificacao_desconhecida"
	RuleTypeSetorPriorizacao         RuleType = "setor_priorizacao"
)

// RuleAction represents the action to take when a rule matches
type RuleAction string

const (
	RuleActionRejeitar   RuleAction = "rejeitar"
	RuleActionPriorizar  RuleAction = "priorizar"
	RuleActionAlertar    RuleAction = "alertar"
)

// RuleConfig represents the configuration of a single rule
type RuleConfig struct {
	Tipo  RuleType    `json:"tipo"`
	Valor interface{} `json:"valor"`
	Acao  RuleAction  `json:"acao"`
}

// IdadeMaximaRule represents an age-based rule
type IdadeMaximaRule struct {
	Tipo  RuleType   `json:"tipo"`
	Valor int        `json:"valor"` // Maximum age in years
	Acao  RuleAction `json:"acao"`
}

// CausasExcludentesRule represents a rule for excluding certain causes of death
type CausasExcludentesRule struct {
	Tipo  RuleType   `json:"tipo"`
	Valor []string   `json:"valor"` // List of excluded causes
	Acao  RuleAction `json:"acao"`
}

// JanelaHorasRule represents a time window rule
type JanelaHorasRule struct {
	Tipo  RuleType   `json:"tipo"`
	Valor int        `json:"valor"` // Window in hours
	Acao  RuleAction `json:"acao"`
}

// IdentificacaoDesconhecidaRule represents an unknown identification rule
type IdentificacaoDesconhecidaRule struct {
	Tipo  RuleType   `json:"tipo"`
	Valor bool       `json:"valor"` // true = reject unknown identifications
	Acao  RuleAction `json:"acao"`
}

// SetorPriorizacaoRule represents a sector prioritization rule
type SetorPriorizacaoRule struct {
	Tipo  RuleType          `json:"tipo"`
	Valor map[string]int    `json:"valor"` // Sector -> score mapping
	Acao  RuleAction        `json:"acao"`
}

// CreateTriagemRuleInput represents input for creating a triagem rule
type CreateTriagemRuleInput struct {
	Nome       string          `json:"nome" validate:"required,min=2,max=255"`
	Descricao  *string         `json:"descricao,omitempty" validate:"omitempty,max=1000"`
	Regras     json.RawMessage `json:"regras" validate:"required"`
	Ativo      *bool           `json:"ativo,omitempty"`
	Prioridade *int            `json:"prioridade,omitempty" validate:"omitempty,min=0,max=1000"`
}

// UpdateTriagemRuleInput represents input for updating a triagem rule
type UpdateTriagemRuleInput struct {
	Nome       *string         `json:"nome,omitempty" validate:"omitempty,min=2,max=255"`
	Descricao  *string         `json:"descricao,omitempty" validate:"omitempty,max=1000"`
	Regras     json.RawMessage `json:"regras,omitempty"`
	Ativo      *bool           `json:"ativo,omitempty"`
	Prioridade *int            `json:"prioridade,omitempty" validate:"omitempty,min=0,max=1000"`
}

// TriagemRuleResponse represents the API response for a triagem rule
type TriagemRuleResponse struct {
	ID         uuid.UUID       `json:"id"`
	Nome       string          `json:"nome"`
	Descricao  *string         `json:"descricao,omitempty"`
	Regras     json.RawMessage `json:"regras"`
	Ativo      bool            `json:"ativo"`
	Prioridade int             `json:"prioridade"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ToResponse converts TriagemRule to TriagemRuleResponse
func (r *TriagemRule) ToResponse() TriagemRuleResponse {
	return TriagemRuleResponse{
		ID:         r.ID,
		Nome:       r.Nome,
		Descricao:  r.Descricao,
		Regras:     r.Regras,
		Ativo:      r.Ativo,
		Prioridade: r.Prioridade,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// ParseRuleConfig parses the rule configuration from JSON
func (r *TriagemRule) ParseRuleConfig() (*RuleConfig, error) {
	var config RuleConfig
	if err := json.Unmarshal(r.Regras, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Default sector scores for prioritization
var DefaultSectorScores = map[string]int{
	"UTI":         100,
	"Emergencia":  80,
	"Enfermaria":  50,
	"Centro Cirurgico": 70,
	"Outros":      40,
}

// GetSectorScore returns the score for a given sector
func GetSectorScore(setor string) int {
	if score, exists := DefaultSectorScores[setor]; exists {
		return score
	}
	return DefaultSectorScores["Outros"]
}
