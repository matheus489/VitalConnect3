package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	ErrObitoNotFound = errors.New("obito not found")
)

// ObitoRepository handles obito data access
type ObitoRepository struct {
	db *sql.DB
}

// NewObitoRepository creates a new obito repository
func NewObitoRepository(db *sql.DB) *ObitoRepository {
	return &ObitoRepository{db: db}
}

// GetUnprocessed returns unprocessed obitos since a given timestamp
func (r *ObitoRepository) GetUnprocessed(ctx context.Context, since time.Time) ([]models.ObitoSimulado, error) {
	query := `
		SELECT
			o.id, o.hospital_id, o.nome_paciente, o.data_nascimento, o.data_obito,
			o.causa_mortis, o.prontuario, o.setor, o.leito, o.identificacao_desconhecida,
			o.processado, o.processado_em, o.created_at,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM obitos_simulados o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		WHERE o.processado = false
		AND o.created_at >= $1
		ORDER BY o.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obitos []models.ObitoSimulado
	for rows.Next() {
		var o models.ObitoSimulado
		var h models.Hospital
		var prontuario, setor, leito, hEndereco sql.NullString
		var processadoEm sql.NullTime

		err := rows.Scan(
			&o.ID, &o.HospitalID, &o.NomePaciente, &o.DataNascimento, &o.DataObito,
			&o.CausaMortis, &prontuario, &setor, &leito, &o.IdentificacaoDesconhecida,
			&o.Processado, &processadoEm, &o.CreatedAt,
			&h.ID, &h.Nome, &h.Codigo, &hEndereco, &h.Ativo,
		)
		if err != nil {
			return nil, err
		}

		if prontuario.Valid {
			o.Prontuario = &prontuario.String
		}
		if setor.Valid {
			o.Setor = &setor.String
		}
		if leito.Valid {
			o.Leito = &leito.String
		}
		if processadoEm.Valid {
			o.ProcessadoEm = &processadoEm.Time
		}
		if hEndereco.Valid {
			h.Endereco = &hEndereco.String
		}
		o.Hospital = &h

		obitos = append(obitos, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return obitos, nil
}

// GetByID retrieves an obito by ID
func (r *ObitoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ObitoSimulado, error) {
	query := `
		SELECT
			o.id, o.hospital_id, o.nome_paciente, o.data_nascimento, o.data_obito,
			o.causa_mortis, o.prontuario, o.setor, o.leito, o.identificacao_desconhecida,
			o.processado, o.processado_em, o.created_at,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM obitos_simulados o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		WHERE o.id = $1
	`

	var o models.ObitoSimulado
	var h models.Hospital
	var prontuario, setor, leito, hEndereco sql.NullString
	var processadoEm sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.HospitalID, &o.NomePaciente, &o.DataNascimento, &o.DataObito,
		&o.CausaMortis, &prontuario, &setor, &leito, &o.IdentificacaoDesconhecida,
		&o.Processado, &processadoEm, &o.CreatedAt,
		&h.ID, &h.Nome, &h.Codigo, &hEndereco, &h.Ativo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrObitoNotFound
		}
		return nil, err
	}

	if prontuario.Valid {
		o.Prontuario = &prontuario.String
	}
	if setor.Valid {
		o.Setor = &setor.String
	}
	if leito.Valid {
		o.Leito = &leito.String
	}
	if processadoEm.Valid {
		o.ProcessadoEm = &processadoEm.Time
	}
	if hEndereco.Valid {
		h.Endereco = &hEndereco.String
	}
	o.Hospital = &h

	return &o, nil
}

// MarkAsProcessed marks an obito as processed
func (r *ObitoRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE obitos_simulados
		SET processado = true, processado_em = $1
		WHERE id = $2
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
		return ErrObitoNotFound
	}

	return nil
}

// IsProcessed checks if an obito has already been processed
func (r *ObitoRepository) IsProcessed(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT processado FROM obitos_simulados WHERE id = $1`

	var processado bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&processado)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrObitoNotFound
		}
		return false, err
	}

	return processado, nil
}

// GetByHospitalID returns unprocessed obitos for a specific hospital
func (r *ObitoRepository) GetByHospitalID(ctx context.Context, hospitalID uuid.UUID, since time.Time) ([]models.ObitoSimulado, error) {
	query := `
		SELECT
			o.id, o.hospital_id, o.nome_paciente, o.data_nascimento, o.data_obito,
			o.causa_mortis, o.prontuario, o.setor, o.leito, o.identificacao_desconhecida,
			o.processado, o.processado_em, o.created_at,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM obitos_simulados o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		WHERE o.hospital_id = $1
		AND o.processado = false
		AND o.created_at >= $2
		ORDER BY o.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obitos []models.ObitoSimulado
	for rows.Next() {
		var o models.ObitoSimulado
		var h models.Hospital
		var prontuario, setor, leito, hEndereco sql.NullString
		var processadoEm sql.NullTime

		err := rows.Scan(
			&o.ID, &o.HospitalID, &o.NomePaciente, &o.DataNascimento, &o.DataObito,
			&o.CausaMortis, &prontuario, &setor, &leito, &o.IdentificacaoDesconhecida,
			&o.Processado, &processadoEm, &o.CreatedAt,
			&h.ID, &h.Nome, &h.Codigo, &hEndereco, &h.Ativo,
		)
		if err != nil {
			return nil, err
		}

		if prontuario.Valid {
			o.Prontuario = &prontuario.String
		}
		if setor.Valid {
			o.Setor = &setor.String
		}
		if leito.Valid {
			o.Leito = &leito.String
		}
		if processadoEm.Valid {
			o.ProcessadoEm = &processadoEm.Time
		}
		if hEndereco.Valid {
			h.Endereco = &hEndereco.String
		}
		o.Hospital = &h

		obitos = append(obitos, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return obitos, nil
}

// CountTodayDetected returns the count of obitos detected today
func (r *ObitoRepository) CountTodayDetected(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) FROM obitos_simulados
		WHERE processado = true
		AND DATE(processado_em) = CURRENT_DATE
	`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// GetLastProcessedTime returns the last time an obito was processed
func (r *ObitoRepository) GetLastProcessedTime(ctx context.Context) (*time.Time, error) {
	query := `
		SELECT MAX(processado_em) FROM obitos_simulados
		WHERE processado = true
	`

	var lastProcessed sql.NullTime
	err := r.db.QueryRowContext(ctx, query).Scan(&lastProcessed)
	if err != nil {
		return nil, err
	}

	if lastProcessed.Valid {
		return &lastProcessed.Time, nil
	}

	return nil, nil
}

// Create creates a new obito record (used by seeder)
func (r *ObitoRepository) Create(ctx context.Context, input *models.CreateObitoInput) (*models.ObitoSimulado, error) {
	obito := &models.ObitoSimulado{
		ID:                        uuid.New(),
		HospitalID:                input.HospitalID,
		NomePaciente:              input.NomePaciente,
		DataNascimento:            input.DataNascimento,
		DataObito:                 input.DataObito,
		CausaMortis:               input.CausaMortis,
		Prontuario:                input.Prontuario,
		Setor:                     input.Setor,
		Leito:                     input.Leito,
		IdentificacaoDesconhecida: input.IdentificacaoDesconhecida,
		Processado:                false,
		CreatedAt:                 time.Now(),
	}

	query := `
		INSERT INTO obitos_simulados (
			id, hospital_id, nome_paciente, data_nascimento, data_obito,
			causa_mortis, prontuario, setor, leito, identificacao_desconhecida,
			processado, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		obito.ID,
		obito.HospitalID,
		obito.NomePaciente,
		obito.DataNascimento,
		obito.DataObito,
		obito.CausaMortis,
		obito.Prontuario,
		obito.Setor,
		obito.Leito,
		obito.IdentificacaoDesconhecida,
		obito.Processado,
		obito.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return obito, nil
}
