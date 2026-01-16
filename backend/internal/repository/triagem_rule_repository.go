package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	ErrTriagemRuleNotFound = errors.New("triagem rule not found")
)

const (
	triagemRulesCacheKey = "triagem_rules:all"
	triagemRulesCacheTTL = 5 * time.Minute
)

// TriagemRuleRepository handles triagem rule data access
type TriagemRuleRepository struct {
	db    *sql.DB
	redis *redis.Client
}

// NewTriagemRuleRepository creates a new triagem rule repository
func NewTriagemRuleRepository(db *sql.DB, redisClient *redis.Client) *TriagemRuleRepository {
	return &TriagemRuleRepository{
		db:    db,
		redis: redisClient,
	}
}

// List returns all triagem rules
func (r *TriagemRuleRepository) List(ctx context.Context) ([]models.TriagemRule, error) {
	query := `
		SELECT id, nome, descricao, regras, ativo, prioridade, created_at, updated_at
		FROM triagem_rules
		ORDER BY prioridade DESC, nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.TriagemRule
	for rows.Next() {
		var rule models.TriagemRule
		var descricao sql.NullString
		var regras string

		err := rows.Scan(
			&rule.ID,
			&rule.Nome,
			&descricao,
			&regras,
			&rule.Ativo,
			&rule.Prioridade,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if descricao.Valid {
			rule.Descricao = &descricao.String
		}
		rule.Regras = json.RawMessage(regras)

		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// ListActive returns only active triagem rules (with caching)
func (r *TriagemRuleRepository) ListActive(ctx context.Context) ([]models.TriagemRule, error) {
	// Try to get from cache first
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, triagemRulesCacheKey).Result()
		if err == nil && cached != "" {
			var rules []models.TriagemRule
			if err := json.Unmarshal([]byte(cached), &rules); err == nil {
				return rules, nil
			}
		}
	}

	// Query from database
	query := `
		SELECT id, nome, descricao, regras, ativo, prioridade, created_at, updated_at
		FROM triagem_rules
		WHERE ativo = true
		ORDER BY prioridade DESC, nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.TriagemRule
	for rows.Next() {
		var rule models.TriagemRule
		var descricao sql.NullString
		var regras string

		err := rows.Scan(
			&rule.ID,
			&rule.Nome,
			&descricao,
			&regras,
			&rule.Ativo,
			&rule.Prioridade,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if descricao.Valid {
			rule.Descricao = &descricao.String
		}
		rule.Regras = json.RawMessage(regras)

		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Cache the results
	if r.redis != nil && len(rules) > 0 {
		data, err := json.Marshal(rules)
		if err == nil {
			r.redis.Set(ctx, triagemRulesCacheKey, string(data), triagemRulesCacheTTL)
		}
	}

	return rules, nil
}

// GetByID retrieves a triagem rule by ID
func (r *TriagemRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TriagemRule, error) {
	query := `
		SELECT id, nome, descricao, regras, ativo, prioridade, created_at, updated_at
		FROM triagem_rules
		WHERE id = $1
	`

	var rule models.TriagemRule
	var descricao sql.NullString
	var regras string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.Nome,
		&descricao,
		&regras,
		&rule.Ativo,
		&rule.Prioridade,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTriagemRuleNotFound
		}
		return nil, err
	}

	if descricao.Valid {
		rule.Descricao = &descricao.String
	}
	rule.Regras = json.RawMessage(regras)

	return &rule, nil
}

// Create creates a new triagem rule
func (r *TriagemRuleRepository) Create(ctx context.Context, input *models.CreateTriagemRuleInput) (*models.TriagemRule, error) {
	rule := &models.TriagemRule{
		ID:         uuid.New(),
		Nome:       input.Nome,
		Descricao:  input.Descricao,
		Regras:     input.Regras,
		Ativo:      true,
		Prioridade: 50,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if input.Ativo != nil {
		rule.Ativo = *input.Ativo
	}
	if input.Prioridade != nil {
		rule.Prioridade = *input.Prioridade
	}

	query := `
		INSERT INTO triagem_rules (id, nome, descricao, regras, ativo, prioridade, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Nome,
		rule.Descricao,
		string(rule.Regras),
		rule.Ativo,
		rule.Prioridade,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Invalidate cache
	r.InvalidateCache(ctx)

	return rule, nil
}

// Update updates a triagem rule
func (r *TriagemRuleRepository) Update(ctx context.Context, id uuid.UUID, input *models.UpdateTriagemRuleInput) (*models.TriagemRule, error) {
	// Get existing rule
	rule, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Nome != nil {
		rule.Nome = *input.Nome
	}
	if input.Descricao != nil {
		rule.Descricao = input.Descricao
	}
	if input.Regras != nil {
		rule.Regras = input.Regras
	}
	if input.Ativo != nil {
		rule.Ativo = *input.Ativo
	}
	if input.Prioridade != nil {
		rule.Prioridade = *input.Prioridade
	}

	rule.UpdatedAt = time.Now()

	query := `
		UPDATE triagem_rules
		SET nome = $1, descricao = $2, regras = $3, ativo = $4, prioridade = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.Nome,
		rule.Descricao,
		string(rule.Regras),
		rule.Ativo,
		rule.Prioridade,
		rule.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrTriagemRuleNotFound
	}

	// Invalidate cache
	r.InvalidateCache(ctx)

	return rule, nil
}

// SoftDelete performs a soft delete by setting ativo to false
func (r *TriagemRuleRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE triagem_rules
		SET ativo = false, updated_at = $1
		WHERE id = $2 AND ativo = true
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTriagemRuleNotFound
	}

	// Invalidate cache
	r.InvalidateCache(ctx)

	return nil
}

// InvalidateCache invalidates the triagem rules cache (exported for use by handlers)
func (r *TriagemRuleRepository) InvalidateCache(ctx context.Context) {
	if r.redis != nil {
		r.redis.Del(ctx, triagemRulesCacheKey)
	}
}
