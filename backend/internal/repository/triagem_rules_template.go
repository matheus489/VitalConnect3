package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// TriagemRuleTemplate represents a template triagem rule for new tenants
type TriagemRuleTemplate struct {
	Nome       string
	Descricao  string
	Regras     map[string]interface{}
	Ativo      bool
	Prioridade int
}

// DefaultTriagemRulesTemplates returns the default set of triagem rules for new tenants
// These are copied from the SES-GO configuration as the canonical template
func DefaultTriagemRulesTemplates() []TriagemRuleTemplate {
	return []TriagemRuleTemplate{
		{
			Nome:      "Idade Maxima",
			Descricao: "Rejeita potenciais doadores acima de 80 anos",
			Regras: map[string]interface{}{
				"tipo":  "idade_maxima",
				"valor": 80,
				"acao":  "rejeitar",
			},
			Ativo:      true,
			Prioridade: 100,
		},
		{
			Nome:      "Janela de Tempo",
			Descricao: "Rejeita obitos com mais de 6 horas desde a ocorrencia",
			Regras: map[string]interface{}{
				"tipo":  "janela_horas",
				"valor": 6,
				"acao":  "rejeitar",
			},
			Ativo:      true,
			Prioridade: 90,
		},
		{
			Nome:      "Identificacao Desconhecida",
			Descricao: "Rejeita potenciais doadores sem identificacao",
			Regras: map[string]interface{}{
				"tipo":  "identificacao_desconhecida",
				"valor": true,
				"acao":  "rejeitar",
			},
			Ativo:      true,
			Prioridade: 95,
		},
		{
			Nome:      "Causas Excludentes",
			Descricao: "Rejeita obitos com causas que impossibilitam doacao",
			Regras: map[string]interface{}{
				"tipo": "causas_excludentes",
				"valor": []string{
					"Neoplasia maligna disseminada",
					"Sepse grave",
					"HIV/AIDS",
					"Hepatite B ou C ativa",
					"Tuberculose ativa",
					"Raiva",
					"Doenca de Creutzfeldt-Jakob",
					"Uso de drogas intravenosas",
				},
				"acao": "rejeitar",
			},
			Ativo:      true,
			Prioridade: 85,
		},
		{
			Nome:      "Priorizacao por Setor",
			Descricao: "Define prioridade de atendimento baseado no setor hospitalar",
			Regras: map[string]interface{}{
				"tipo": "setor_priorizacao",
				"valor": map[string]int{
					"UTI":         100,
					"Emergencia":  80,
					"Centro Cirurgico": 70,
					"Enfermaria":  50,
					"Outros":      30,
				},
				"acao": "priorizar",
			},
			Ativo:      true,
			Prioridade: 50,
		},
	}
}

// CopyTriagemRulesToTenant copies the default triagem rules template to a new tenant
func (r *TriagemRuleRepository) CopyTriagemRulesToTenant(ctx context.Context, tenantID uuid.UUID) error {
	templates := DefaultTriagemRulesTemplates()

	for _, template := range templates {
		regrasJSON, err := json.Marshal(template.Regras)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO triagem_rules (id, tenant_id, nome, descricao, regras, ativo, prioridade, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`

		_, err = r.db.ExecContext(ctx, query,
			uuid.New(),
			tenantID,
			template.Nome,
			template.Descricao,
			string(regrasJSON),
			template.Ativo,
			template.Prioridade,
		)

		if err != nil {
			return err
		}
	}

	// Invalidate cache for this tenant's rules
	r.InvalidateCache(ctx)

	return nil
}

// CopyTriagemRulesFromTenant copies triagem rules from one tenant to another
// This is useful when onboarding a new tenant that wants to start with another tenant's configuration
func (r *TriagemRuleRepository) CopyTriagemRulesFromTenant(ctx context.Context, sourceTenantID, targetTenantID uuid.UUID) error {
	query := `
		INSERT INTO triagem_rules (id, tenant_id, nome, descricao, regras, ativo, prioridade, created_at, updated_at)
		SELECT
			gen_random_uuid(),
			$2,
			nome,
			descricao,
			regras,
			ativo,
			prioridade,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		FROM triagem_rules
		WHERE tenant_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, sourceTenantID, targetTenantID)
	if err != nil {
		return err
	}

	// Invalidate cache
	r.InvalidateCache(ctx)

	return nil
}
