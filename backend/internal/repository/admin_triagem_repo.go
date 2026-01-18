package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	// ErrAdminTriagemTemplateNotFound is returned when a triagem template is not found
	ErrAdminTriagemTemplateNotFound = errors.New("triagem template not found")
)

// AdminTriagemTemplateRepository handles admin-level triagem template data access
type AdminTriagemTemplateRepository struct {
	db *sql.DB
}

// NewAdminTriagemTemplateRepository creates a new admin triagem template repository
func NewAdminTriagemTemplateRepository(db *sql.DB) *AdminTriagemTemplateRepository {
	return &AdminTriagemTemplateRepository{db: db}
}

// AdminTriagemTemplateListParams contains parameters for listing templates
type AdminTriagemTemplateListParams struct {
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
	Search   string `form:"search"`
	Tipo     string `form:"tipo"`
	Ativo    *bool  `form:"ativo"`
}

// AdminTriagemTemplateListResult contains paginated template results
type AdminTriagemTemplateListResult struct {
	Templates  []models.TriagemRuleTemplateWithUsage `json:"templates"`
	Total      int                                   `json:"total"`
	Page       int                                   `json:"page"`
	PerPage    int                                   `json:"per_page"`
	TotalPages int                                   `json:"total_pages"`
}

// TemplateUsage represents which tenants use similar rules to a template
type TemplateUsage struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	TenantSlug string    `json:"tenant_slug"`
	RuleCount  int       `json:"rule_count"`
}

// TemplateUsageResult contains the usage information for a template
type TemplateUsageResult struct {
	TemplateID  uuid.UUID       `json:"template_id"`
	TotalTenants int            `json:"total_tenants"`
	Tenants     []TemplateUsage `json:"tenants"`
}

// CloneResult contains the result of cloning a template to tenants
type CloneResult struct {
	TemplateID    uuid.UUID   `json:"template_id"`
	ClonedToTenants []uuid.UUID `json:"cloned_to_tenants"`
	SuccessCount  int         `json:"success_count"`
	FailedTenants []uuid.UUID `json:"failed_tenants,omitempty"`
}

// ListTemplates returns all triagem rule templates with optional filters and pagination
func (r *AdminTriagemTemplateRepository) ListTemplates(ctx context.Context, params *AdminTriagemTemplateListParams) (*AdminTriagemTemplateListResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(t.nome ILIKE $%d OR t.descricao ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if params.Tipo != "" {
		conditions = append(conditions, fmt.Sprintf("t.tipo = $%d", argIndex))
		args = append(args, params.Tipo)
		argIndex++
	}

	if params.Ativo != nil {
		conditions = append(conditions, fmt.Sprintf("t.ativo = $%d", argIndex))
		args = append(args, *params.Ativo)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM triagem_rule_templates t %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count templates: %w", err)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	offset := (params.Page - 1) * params.PerPage

	// Get templates with tenant usage count
	query := fmt.Sprintf(`
		SELECT
			t.id, t.nome, t.tipo, t.condicao, t.descricao, t.ativo, t.prioridade, t.created_at, t.updated_at,
			COALESCE((
				SELECT COUNT(DISTINCT tr.tenant_id)
				FROM triagem_rules tr
				WHERE tr.nome = t.nome
			), 0) AS tenant_count
		FROM triagem_rule_templates t
		%s
		ORDER BY t.prioridade DESC, t.nome ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []models.TriagemRuleTemplateWithUsage
	for rows.Next() {
		var t models.TriagemRuleTemplateWithUsage
		var descricao sql.NullString
		var condicao string

		err := rows.Scan(
			&t.ID,
			&t.Nome,
			&t.Tipo,
			&condicao,
			&descricao,
			&t.Ativo,
			&t.Prioridade,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.TenantCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		t.Condicao = json.RawMessage(condicao)
		if descricao.Valid {
			t.Descricao = &descricao.String
		}

		templates = append(templates, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return &AdminTriagemTemplateListResult{
		Templates:  templates,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetTemplateByID retrieves a single template by ID
func (r *AdminTriagemTemplateRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.TriagemRuleTemplateWithUsage, error) {
	query := `
		SELECT
			t.id, t.nome, t.tipo, t.condicao, t.descricao, t.ativo, t.prioridade, t.created_at, t.updated_at,
			COALESCE((
				SELECT COUNT(DISTINCT tr.tenant_id)
				FROM triagem_rules tr
				WHERE tr.nome = t.nome
			), 0) AS tenant_count
		FROM triagem_rule_templates t
		WHERE t.id = $1
	`

	var t models.TriagemRuleTemplateWithUsage
	var descricao sql.NullString
	var condicao string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Nome,
		&t.Tipo,
		&condicao,
		&descricao,
		&t.Ativo,
		&t.Prioridade,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.TenantCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminTriagemTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	t.Condicao = json.RawMessage(condicao)
	if descricao.Valid {
		t.Descricao = &descricao.String
	}

	return &t, nil
}

// CreateTemplate creates a new triagem rule template
func (r *AdminTriagemTemplateRepository) CreateTemplate(ctx context.Context, input *models.CreateTriagemRuleTemplateInput) (*models.TriagemRuleTemplate, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	template := &models.TriagemRuleTemplate{
		ID:         uuid.New(),
		Nome:       input.Nome,
		Tipo:       models.TriagemRuleTemplateType(input.Tipo),
		Condicao:   input.Condicao,
		Descricao:  input.Descricao,
		Ativo:      true,
		Prioridade: 50,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if input.Ativo != nil {
		template.Ativo = *input.Ativo
	}
	if input.Prioridade != nil {
		template.Prioridade = *input.Prioridade
	}

	query := `
		INSERT INTO triagem_rule_templates (id, nome, tipo, condicao, descricao, ativo, prioridade, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Nome,
		string(template.Tipo),
		string(template.Condicao),
		template.Descricao,
		template.Ativo,
		template.Prioridade,
		template.CreatedAt,
		template.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// UpdateTemplate updates an existing triagem rule template
func (r *AdminTriagemTemplateRepository) UpdateTemplate(ctx context.Context, id uuid.UUID, input *models.UpdateTriagemRuleTemplateInput) (*models.TriagemRuleTemplate, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Get existing template
	existing, err := r.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build dynamic update
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if input.Nome != nil {
		setClauses = append(setClauses, fmt.Sprintf("nome = $%d", argIndex))
		args = append(args, *input.Nome)
		existing.Nome = *input.Nome
		argIndex++
	}
	if input.Tipo != nil {
		setClauses = append(setClauses, fmt.Sprintf("tipo = $%d", argIndex))
		args = append(args, *input.Tipo)
		existing.Tipo = models.TriagemRuleTemplateType(*input.Tipo)
		argIndex++
	}
	if len(input.Condicao) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("condicao = $%d", argIndex))
		args = append(args, string(input.Condicao))
		existing.Condicao = input.Condicao
		argIndex++
	}
	if input.Descricao != nil {
		setClauses = append(setClauses, fmt.Sprintf("descricao = $%d", argIndex))
		args = append(args, *input.Descricao)
		existing.Descricao = input.Descricao
		argIndex++
	}
	if input.Ativo != nil {
		setClauses = append(setClauses, fmt.Sprintf("ativo = $%d", argIndex))
		args = append(args, *input.Ativo)
		existing.Ativo = *input.Ativo
		argIndex++
	}
	if input.Prioridade != nil {
		setClauses = append(setClauses, fmt.Sprintf("prioridade = $%d", argIndex))
		args = append(args, *input.Prioridade)
		existing.Prioridade = *input.Prioridade
		argIndex++
	}

	if len(setClauses) == 0 {
		// No updates, return existing
		return &existing.TriagemRuleTemplate, nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE triagem_rule_templates
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrAdminTriagemTemplateNotFound
	}

	existing.UpdatedAt = time.Now()
	return &existing.TriagemRuleTemplate, nil
}

// CloneToTenant copies a template to one or more tenant's triagem_rules table
func (r *AdminTriagemTemplateRepository) CloneToTenant(ctx context.Context, templateID uuid.UUID, tenantIDs []uuid.UUID) (*CloneResult, error) {
	// Get the template
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	result := &CloneResult{
		TemplateID:      templateID,
		ClonedToTenants: make([]uuid.UUID, 0),
		FailedTenants:   make([]uuid.UUID, 0),
	}

	// Clone to each tenant
	for _, tenantID := range tenantIDs {
		query := `
			INSERT INTO triagem_rules (id, tenant_id, nome, descricao, regras, ativo, prioridade, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		_, err := r.db.ExecContext(ctx, query,
			uuid.New(),
			tenantID,
			template.Nome,
			template.Descricao,
			string(template.Condicao),
			template.Ativo,
			template.Prioridade,
			time.Now(),
			time.Now(),
		)

		if err != nil {
			result.FailedTenants = append(result.FailedTenants, tenantID)
		} else {
			result.ClonedToTenants = append(result.ClonedToTenants, tenantID)
			result.SuccessCount++
		}
	}

	return result, nil
}

// GetTemplateUsage returns which tenants have rules similar to a template
func (r *AdminTriagemTemplateRepository) GetTemplateUsage(ctx context.Context, templateID uuid.UUID) (*TemplateUsageResult, error) {
	// Get template first
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Find tenants with rules matching this template's nome
	query := `
		SELECT
			t.id AS tenant_id,
			t.name AS tenant_name,
			t.slug AS tenant_slug,
			COUNT(tr.id) AS rule_count
		FROM tenants t
		INNER JOIN triagem_rules tr ON tr.tenant_id = t.id
		WHERE tr.nome = $1
		GROUP BY t.id, t.name, t.slug
		ORDER BY t.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, template.Nome)
	if err != nil {
		return nil, fmt.Errorf("failed to query template usage: %w", err)
	}
	defer rows.Close()

	result := &TemplateUsageResult{
		TemplateID: templateID,
		Tenants:    make([]TemplateUsage, 0),
	}

	for rows.Next() {
		var usage TemplateUsage
		err := rows.Scan(
			&usage.TenantID,
			&usage.TenantName,
			&usage.TenantSlug,
			&usage.RuleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}
		result.Tenants = append(result.Tenants, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage: %w", err)
	}

	result.TotalTenants = len(result.Tenants)
	return result, nil
}

// DeleteTemplate soft-deletes a template by setting ativo to false
func (r *AdminTriagemTemplateRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE triagem_rule_templates
		SET ativo = false, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAdminTriagemTemplateNotFound
	}

	return nil
}
