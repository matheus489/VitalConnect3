package report

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

// ReportService handles report generation
type ReportService struct {
	db *sql.DB
}

// NewReportService creates a new report service
func NewReportService(db *sql.DB) *ReportService {
	return &ReportService{db: db}
}

// GenerateCSV generates a CSV report with streaming
func (s *ReportService) GenerateCSV(ctx context.Context, filters models.ReportFilters, writer io.Writer) error {
	// Write UTF-8 BOM for Excel compatibility
	if _, err := writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("failed to write BOM: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Hospital",
		"Data/Hora Obito",
		"Iniciais Paciente",
		"Idade",
		"Status Final",
		"Tempo de Reacao (min)",
		"Usuario Responsavel",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Stream rows from database
	rows, err := s.queryReportData(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to query report data: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		row, err := s.scanReportRow(rows)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		tempoReacao := ""
		if row.TempoReacaoMin != nil {
			tempoReacao = fmt.Sprintf("%.1f", *row.TempoReacaoMin)
		}

		usuarioResponsavel := ""
		if row.UsuarioResponsavel != nil {
			usuarioResponsavel = *row.UsuarioResponsavel
		}

		record := []string{
			row.HospitalNome,
			row.DataHoraObito.Format("02/01/2006 15:04"),
			row.IniciaisPaciente,
			fmt.Sprintf("%d", row.Idade),
			row.StatusFinal,
			tempoReacao,
			usuarioResponsavel,
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}

// GeneratePDF generates a PDF report and returns the bytes
func (s *ReportService) GeneratePDF(ctx context.Context, filters models.ReportFilters) ([]byte, error) {
	// Get metrics for the report header
	metrics, err := s.CalculateMetrics(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate metrics: %w", err)
	}

	// Get all report data
	reportRows, err := s.getReportRows(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get report rows: %w", err)
	}

	// Generate PDF using our simple PDF generator
	pdfGen := NewPDFGenerator()
	pdfBytes, err := pdfGen.GenerateReport(filters, metrics, reportRows)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBytes, nil
}

// CalculateMetrics calculates aggregated metrics for the report
func (s *ReportService) CalculateMetrics(ctx context.Context, filters models.ReportFilters) (*models.ReportMetrics, error) {
	metrics := &models.ReportMetrics{
		OcorrenciasPorDesfecho: make(map[string]int),
		PeriodoInicio:          filters.DateFrom,
		PeriodoFim:             filters.DateTo,
	}

	// Build WHERE clause
	where, args := s.buildWhereClause(filters)

	// Total occurrences
	totalQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		%s
	`, where)
	if err := s.db.QueryRowContext(ctx, totalQuery, args...).Scan(&metrics.TotalOcorrencias); err != nil {
		return nil, fmt.Errorf("failed to count total: %w", err)
	}

	// Occurrences by desfecho
	desfechoQuery := fmt.Sprintf(`
		SELECT COALESCE(oh.desfecho::text, 'Pendente') as desfecho, COUNT(*) as count
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		LEFT JOIN (
			SELECT DISTINCT ON (occurrence_id) occurrence_id, desfecho
			FROM occurrence_history
			WHERE desfecho IS NOT NULL
			ORDER BY occurrence_id, created_at DESC
		) oh ON o.id = oh.occurrence_id
		%s
		GROUP BY COALESCE(oh.desfecho::text, 'Pendente')
	`, where)
	desfechoRows, err := s.db.QueryContext(ctx, desfechoQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query desfecho counts: %w", err)
	}
	defer desfechoRows.Close()

	for desfechoRows.Next() {
		var desfecho string
		var count int
		if err := desfechoRows.Scan(&desfecho, &count); err != nil {
			return nil, fmt.Errorf("failed to scan desfecho row: %w", err)
		}
		// Map to display name
		displayName := models.DesfechoDisplayNames[desfecho]
		if displayName == "" {
			displayName = desfecho
		}
		metrics.OcorrenciasPorDesfecho[displayName] = count
	}

	// Taxa de Perda Operacional: % of notifications that expired after 6h without action (status CANCELADA)
	perdaQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) FILTER (WHERE o.status = 'CANCELADA' AND o.janela_expira_em < NOW()) as expirados,
			COUNT(*) as total
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		%s
	`, where)
	var expirados, total int
	if err := s.db.QueryRowContext(ctx, perdaQuery, args...).Scan(&expirados, &total); err != nil {
		return nil, fmt.Errorf("failed to calculate loss rate: %w", err)
	}
	if total > 0 {
		metrics.TaxaPerdaOperacional = float64(expirados) / float64(total) * 100
	}

	// Tempo medio de reacao (difference between created_at and first status change)
	tempoQuery := fmt.Sprintf(`
		SELECT COALESCE(AVG(
			EXTRACT(EPOCH FROM (first_action.first_change - o.created_at)) / 60
		), 0) as avg_reaction_time
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		LEFT JOIN (
			SELECT occurrence_id, MIN(created_at) as first_change
			FROM occurrence_history
			WHERE status_novo IS NOT NULL AND status_novo != 'PENDENTE'
			GROUP BY occurrence_id
		) first_action ON o.id = first_action.occurrence_id
		%s
		AND first_action.first_change IS NOT NULL
	`, where)
	if err := s.db.QueryRowContext(ctx, tempoQuery, args...).Scan(&metrics.TempoMedioReacaoMin); err != nil {
		return nil, fmt.Errorf("failed to calculate avg reaction time: %w", err)
	}

	return metrics, nil
}

// LogExport creates an audit log entry for the export
func (s *ReportService) LogExport(ctx context.Context, userID uuid.UUID, tipoRelatorio string, filters models.ReportFilters) error {
	filtrosJSON, err := json.Marshal(filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	query := `
		INSERT INTO report_audit_logs (id, user_id, tipo_relatorio, filtros, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = s.db.ExecContext(ctx, query,
		uuid.New(),
		userID,
		tipoRelatorio,
		filtrosJSON,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to log export: %w", err)
	}

	return nil
}

// buildWhereClause builds the WHERE clause for report queries
func (s *ReportService) buildWhereClause(filters models.ReportFilters) (string, []interface{}) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND o.created_at >= $%d", argIndex)
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND o.created_at <= $%d", argIndex)
		args = append(args, *filters.DateTo)
		argIndex++
	}

	if filters.HospitalID != nil && *filters.HospitalID != "" {
		where += fmt.Sprintf(" AND o.hospital_id = $%d", argIndex)
		args = append(args, *filters.HospitalID)
		argIndex++
	}

	if len(filters.Desfechos) > 0 {
		// Convert display names to database values
		dbDesfechos := make([]string, 0, len(filters.Desfechos))
		for _, d := range filters.Desfechos {
			if dbVal, ok := models.DesfechoFromDisplayName[d]; ok {
				dbDesfechos = append(dbDesfechos, dbVal)
			}
		}
		if len(dbDesfechos) > 0 {
			// Need subquery to filter by desfecho
			placeholders := ""
			for i, d := range dbDesfechos {
				if i > 0 {
					placeholders += ", "
				}
				placeholders += fmt.Sprintf("$%d", argIndex)
				args = append(args, d)
				argIndex++
			}
			where += fmt.Sprintf(` AND EXISTS (
				SELECT 1 FROM occurrence_history oh
				WHERE oh.occurrence_id = o.id
				AND oh.desfecho IN (%s)
			)`, placeholders)
		}
	}

	return where, args
}

// queryReportData returns a rows cursor for streaming
func (s *ReportService) queryReportData(ctx context.Context, filters models.ReportFilters) (*sql.Rows, error) {
	where, args := s.buildWhereClause(filters)

	query := fmt.Sprintf(`
		SELECT
			h.nome as hospital_nome,
			o.data_obito,
			o.nome_paciente_mascarado,
			o.dados_completos,
			o.status,
			o.created_at,
			first_action.first_change,
			u.nome as usuario_nome,
			latest_outcome.desfecho
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		LEFT JOIN (
			SELECT occurrence_id, MIN(created_at) as first_change
			FROM occurrence_history
			WHERE status_novo IS NOT NULL AND status_novo != 'PENDENTE'
			GROUP BY occurrence_id
		) first_action ON o.id = first_action.occurrence_id
		LEFT JOIN (
			SELECT DISTINCT ON (occurrence_id) occurrence_id, user_id
			FROM occurrence_history
			WHERE status_novo = 'EM_ANDAMENTO'
			ORDER BY occurrence_id, created_at ASC
		) assigned ON o.id = assigned.occurrence_id
		LEFT JOIN users u ON assigned.user_id = u.id
		LEFT JOIN (
			SELECT DISTINCT ON (occurrence_id) occurrence_id, desfecho
			FROM occurrence_history
			WHERE desfecho IS NOT NULL
			ORDER BY occurrence_id, created_at DESC
		) latest_outcome ON o.id = latest_outcome.occurrence_id
		%s
		ORDER BY o.created_at DESC
	`, where)

	return s.db.QueryContext(ctx, query, args...)
}

// scanReportRow scans a single row from the report query
func (s *ReportService) scanReportRow(rows *sql.Rows) (*models.ReportOccurrenceRow, error) {
	var hospitalNome, nomeMascarado, status string
	var dadosCompletos string
	var dataObito, createdAt time.Time
	var firstChange sql.NullTime
	var usuarioNome, desfecho sql.NullString

	err := rows.Scan(
		&hospitalNome,
		&dataObito,
		&nomeMascarado,
		&dadosCompletos,
		&status,
		&createdAt,
		&firstChange,
		&usuarioNome,
		&desfecho,
	)
	if err != nil {
		return nil, err
	}

	// Parse dados_completos to get age
	var dados models.OccurrenceCompleteData
	idade := 0
	if err := json.Unmarshal([]byte(dadosCompletos), &dados); err == nil {
		idade = dados.Idade
	}

	row := &models.ReportOccurrenceRow{
		HospitalNome:     hospitalNome,
		DataHoraObito:    dataObito,
		IniciaisPaciente: nomeMascarado,
		Idade:            idade,
		StatusFinal:      status,
	}

	// Calculate tempo de reacao
	if firstChange.Valid {
		tempoMin := firstChange.Time.Sub(createdAt).Minutes()
		row.TempoReacaoMin = &tempoMin
	}

	if usuarioNome.Valid {
		row.UsuarioResponsavel = &usuarioNome.String
	}

	if desfecho.Valid {
		displayName := models.DesfechoDisplayNames[desfecho.String]
		if displayName == "" {
			displayName = desfecho.String
		}
		row.Desfecho = &displayName
	}

	return row, nil
}

// getReportRows gets all report rows (for PDF generation)
func (s *ReportService) getReportRows(ctx context.Context, filters models.ReportFilters) ([]models.ReportOccurrenceRow, error) {
	rows, err := s.queryReportData(ctx, filters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ReportOccurrenceRow
	for rows.Next() {
		row, err := s.scanReportRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
