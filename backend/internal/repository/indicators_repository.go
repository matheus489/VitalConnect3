package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// IndicatorsRepository handles indicators data access
type IndicatorsRepository struct {
	db *sql.DB
}

// NewIndicatorsRepository creates a new indicators repository
func NewIndicatorsRepository(db *sql.DB) *IndicatorsRepository {
	return &IndicatorsRepository{db: db}
}

// CalculateTaxaConversao calculates the conversion rate
// Formula: (Captacoes Realizadas / Notificacoes Validas) * 100
func (r *IndicatorsRepository) CalculateTaxaConversao(ctx context.Context, hospitalID *uuid.UUID) (float64, error) {
	// Build query based on whether hospital filter is provided
	args := []interface{}{}
	whereHospital := ""
	argIndex := 1

	if hospitalID != nil {
		whereHospital = " AND hospital_id = $" + itoa(argIndex)
		args = append(args, *hospitalID)
		argIndex++
	}

	// Count total completed occurrences (valid notifications)
	totalQuery := `
		SELECT COUNT(*) FROM occurrences
		WHERE status = 'CONCLUIDA'` + whereHospital

	var totalNotificacoes int
	err := r.db.QueryRowContext(ctx, totalQuery, args...).Scan(&totalNotificacoes)
	if err != nil {
		return 0, err
	}

	if totalNotificacoes == 0 {
		return 0, nil
	}

	// Count successful captures
	captacoesQuery := `
		SELECT COUNT(DISTINCT o.id) FROM occurrences o
		INNER JOIN occurrence_history oh ON o.id = oh.occurrence_id
		WHERE o.status = 'CONCLUIDA'
		AND oh.desfecho = 'sucesso_captacao'` + whereHospital

	var captacoes int
	err = r.db.QueryRowContext(ctx, captacoesQuery, args...).Scan(&captacoes)
	if err != nil {
		return 0, err
	}

	// Calculate rate
	taxa := (float64(captacoes) / float64(totalNotificacoes)) * 100
	return taxa, nil
}

// CalculateLatenciaSistema calculates the average system latency
// Time between detection (created_at) and notification (notificado_em) in seconds
func (r *IndicatorsRepository) CalculateLatenciaSistema(ctx context.Context, hospitalID *uuid.UUID) (float64, error) {
	args := []interface{}{}
	whereHospital := ""
	argIndex := 1

	if hospitalID != nil {
		whereHospital = " AND hospital_id = $" + itoa(argIndex)
		args = append(args, *hospitalID)
		argIndex++
	}

	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (notificado_em - created_at))), 0)
		FROM occurrences
		WHERE notificado_em IS NOT NULL
		AND created_at >= NOW() - INTERVAL '30 days'` + whereHospital

	var avgLatencia float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&avgLatencia)
	if err != nil {
		return 0, err
	}

	return avgLatencia, nil
}

// CalculateTempoRespostaOperacional calculates the average operational response time
// Time between notification (notificado_em) and acceptance (status change to EM_ANDAMENTO) in minutes
func (r *IndicatorsRepository) CalculateTempoRespostaOperacional(ctx context.Context, hospitalID *uuid.UUID) (float64, error) {
	args := []interface{}{}
	whereHospital := ""
	argIndex := 1

	if hospitalID != nil {
		whereHospital = " AND o.hospital_id = $" + itoa(argIndex)
		args = append(args, *hospitalID)
		argIndex++
	}

	// Get average time between notification and first status change to EM_ANDAMENTO
	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (oh.created_at - o.notificado_em)) / 60), 0)
		FROM occurrences o
		INNER JOIN occurrence_history oh ON o.id = oh.occurrence_id
		WHERE o.notificado_em IS NOT NULL
		AND oh.status_novo = 'EM_ANDAMENTO'
		AND o.created_at >= NOW() - INTERVAL '30 days'` + whereHospital

	var avgTempo float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&avgTempo)
	if err != nil {
		return 0, err
	}

	return avgTempo, nil
}

// SeriesDataPoint represents a daily data point
type SeriesDataPoint struct {
	Data         time.Time
	ObitosTotais int
	Captados     int
}

// GetSeries30Dias returns the 30-day time series data
func (r *IndicatorsRepository) GetSeries30Dias(ctx context.Context, hospitalID *uuid.UUID) ([]SeriesDataPoint, error) {
	args := []interface{}{}
	whereHospital := ""
	argIndex := 1

	if hospitalID != nil {
		whereHospital = " AND o.hospital_id = $" + itoa(argIndex)
		args = append(args, *hospitalID)
		argIndex++
	}

	// Generate series of last 30 days with counts
	query := `
		WITH date_series AS (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '29 days',
				CURRENT_DATE,
				'1 day'::interval
			)::date AS data
		),
		obitos_por_dia AS (
			SELECT
				DATE(o.created_at) as data,
				COUNT(*) as obitos_totais,
				COUNT(CASE WHEN oh.desfecho = 'sucesso_captacao' THEN 1 END) as captados
			FROM occurrences o
			LEFT JOIN occurrence_history oh ON o.id = oh.occurrence_id AND oh.desfecho IS NOT NULL
			WHERE o.created_at >= CURRENT_DATE - INTERVAL '29 days'` + whereHospital + `
			GROUP BY DATE(o.created_at)
		)
		SELECT
			ds.data,
			COALESCE(opd.obitos_totais, 0) as obitos_totais,
			COALESCE(opd.captados, 0) as captados
		FROM date_series ds
		LEFT JOIN obitos_por_dia opd ON ds.data = opd.data
		ORDER BY ds.data ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var series []SeriesDataPoint
	for rows.Next() {
		var dp SeriesDataPoint
		err := rows.Scan(&dp.Data, &dp.ObitosTotais, &dp.Captados)
		if err != nil {
			return nil, err
		}
		series = append(series, dp)
	}

	return series, rows.Err()
}

// RankingItem represents a hospital in the ranking
type RankingItem struct {
	HospitalID uuid.UUID
	Nome       string
	Captacoes  int
}

// GetRankingHospitais returns the top N hospitals by captures
func (r *IndicatorsRepository) GetRankingHospitais(ctx context.Context, limit int) ([]RankingItem, error) {
	query := `
		SELECT
			h.id,
			h.nome,
			COUNT(CASE WHEN oh.desfecho = 'sucesso_captacao' THEN 1 END) as captacoes
		FROM hospitals h
		LEFT JOIN occurrences o ON h.id = o.hospital_id AND o.created_at >= CURRENT_DATE - INTERVAL '30 days'
		LEFT JOIN occurrence_history oh ON o.id = oh.occurrence_id AND oh.desfecho IS NOT NULL
		WHERE h.ativo = true AND h.deleted_at IS NULL
		GROUP BY h.id, h.nome
		ORDER BY captacoes DESC, h.nome ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranking []RankingItem
	for rows.Next() {
		var item RankingItem
		err := rows.Scan(&item.HospitalID, &item.Nome, &item.Captacoes)
		if err != nil {
			return nil, err
		}
		ranking = append(ranking, item)
	}

	return ranking, rows.Err()
}

// GetAllIndicators fetches all indicators data in optimized queries
func (r *IndicatorsRepository) GetAllIndicators(ctx context.Context, hospitalID *uuid.UUID) (*models.IndicatorsMetrics, error) {
	// Calculate all metrics
	taxaConversao, err := r.CalculateTaxaConversao(ctx, hospitalID)
	if err != nil {
		taxaConversao = 0
	}

	latenciaSistema, err := r.CalculateLatenciaSistema(ctx, hospitalID)
	if err != nil {
		latenciaSistema = 0
	}

	tempoResposta, err := r.CalculateTempoRespostaOperacional(ctx, hospitalID)
	if err != nil {
		tempoResposta = 0
	}

	seriesData, err := r.GetSeries30Dias(ctx, hospitalID)
	if err != nil {
		seriesData = []SeriesDataPoint{}
	}

	rankingData, err := r.GetRankingHospitais(ctx, 5)
	if err != nil {
		rankingData = []RankingItem{}
	}

	// Build response
	metrics := &models.IndicatorsMetrics{
		TaxaConversao: models.NewIndicatorCard(
			taxaConversao,
			models.FormatTaxaConversao,
			models.CalculateTaxaConversaoStatus,
		),
		LatenciaSistema: models.NewIndicatorCard(
			latenciaSistema,
			models.FormatLatenciaSistema,
			models.CalculateLatenciaSistemaStatus,
		),
		TempoRespostaOperacional: models.NewIndicatorCard(
			tempoResposta,
			models.FormatTempoResposta,
			models.CalculateTempoRespostaStatus,
		),
		UltimaAtualizacao: time.Now(),
	}

	// Convert series data
	seriesPoints := make([]models.SeriesDataPoint, len(seriesData))
	for i, dp := range seriesData {
		seriesPoints[i] = models.SeriesDataPoint{
			Data:         dp.Data.Format("2006-01-02"),
			ObitosTotais: dp.ObitosTotais,
			Captados:     dp.Captados,
		}
	}
	metrics.Series30Dias = models.Series30Dias{Dados: seriesPoints}

	// Convert ranking data
	maxCaptacoes := 0
	if len(rankingData) > 0 {
		maxCaptacoes = rankingData[0].Captacoes
	}

	rankingItems := make([]models.RankingItem, len(rankingData))
	for i, item := range rankingData {
		percentual := 0.0
		if maxCaptacoes > 0 {
			percentual = (float64(item.Captacoes) / float64(maxCaptacoes)) * 100
		}
		rankingItems[i] = models.RankingItem{
			Posicao:    i + 1,
			HospitalID: item.HospitalID,
			Nome:       item.Nome,
			Captacoes:  item.Captacoes,
			Percentual: percentual,
		}
	}
	metrics.RankingHospitais = models.RankingHospitais{Hospitais: rankingItems}

	return metrics, nil
}

// itoa converts int to string (simple implementation to avoid strconv import)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n /= 10
	}
	return result
}
