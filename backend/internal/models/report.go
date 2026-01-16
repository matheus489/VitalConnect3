package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportFilters represents filters for generating reports
type ReportFilters struct {
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	HospitalID *string    `json:"hospital_id,omitempty"`
	Desfechos  []string   `json:"desfechos,omitempty"`
}

// ReportMetrics represents aggregated metrics for reports
type ReportMetrics struct {
	TotalOcorrencias       int                  `json:"total_ocorrencias"`
	OcorrenciasPorDesfecho map[string]int       `json:"ocorrencias_por_desfecho"`
	TaxaPerdaOperacional   float64              `json:"taxa_perda_operacional"` // percentage
	TempoMedioReacaoMin    float64              `json:"tempo_medio_reacao_min"`
	PeriodoInicio          *time.Time           `json:"periodo_inicio,omitempty"`
	PeriodoFim             *time.Time           `json:"periodo_fim,omitempty"`
}

// ReportOccurrenceRow represents a single row in the report
type ReportOccurrenceRow struct {
	HospitalNome           string     `json:"hospital_nome"`
	DataHoraObito          time.Time  `json:"data_hora_obito"`
	IniciaisPaciente       string     `json:"iniciais_paciente"`
	Idade                  int        `json:"idade"`
	StatusFinal            string     `json:"status_final"`
	TempoReacaoMin         *float64   `json:"tempo_reacao_min,omitempty"`
	UsuarioResponsavel     *string    `json:"usuario_responsavel,omitempty"`
	Desfecho               *string    `json:"desfecho,omitempty"`
}

// ReportAuditLog represents an audit log entry for report exports
type ReportAuditLog struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	UserID       uuid.UUID      `json:"user_id" db:"user_id"`
	TipoRelatorio string        `json:"tipo_relatorio" db:"tipo_relatorio"` // CSV or PDF
	Filtros      ReportFilters  `json:"filtros" db:"filtros"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
}

// DesfechoDisplayNames maps outcome types to display names for reports
var DesfechoDisplayNames = map[string]string{
	"sucesso_captacao":       "Captado",
	"familia_recusou":        "Recusa Familiar",
	"contraindicacao_medica": "Contraindicacao Medica",
	"tempo_excedido":         "Expirado",
	"outro":                  "Outro",
}

// DesfechoFromDisplayName maps display names back to outcome types
var DesfechoFromDisplayName = map[string]string{
	"Captado":                 "sucesso_captacao",
	"Recusa Familiar":         "familia_recusou",
	"Contraindicacao Medica":  "contraindicacao_medica",
	"Expirado":                "tempo_excedido",
	"Outro":                   "outro",
}

// ValidDesfechoDisplayNames returns valid desfecho display names for filter validation
var ValidDesfechoDisplayNames = []string{
	"Captado",
	"Recusa Familiar",
	"Contraindicacao Medica",
	"Expirado",
}

// IsValidDesfecho checks if a desfecho display name is valid
func IsValidDesfecho(desfecho string) bool {
	for _, valid := range ValidDesfechoDisplayNames {
		if desfecho == valid {
			return true
		}
	}
	return false
}
