package models

import (
	"fmt"
	"time"
)

// DashboardMetrics represents the metrics displayed on the dashboard
type DashboardMetrics struct {
	ObitosElegiveisHoje    int       `json:"obitos_elegiveis_hoje"`
	TempoMedioNotificacao  float64   `json:"tempo_medio_notificacao"` // in seconds
	CorneasPotenciais      int       `json:"corneas_potenciais"`      // obitos_elegiveis * 2
	OccurrencesPendentes   int       `json:"occurrences_pendentes"`
	OccurrencesEmAndamento int       `json:"occurrences_em_andamento"`
	UltimaAtualizacao      time.Time `json:"ultima_atualizacao"`
}

// FormatTempoMedioNotificacao returns a human-readable string for the average notification time
func (m *DashboardMetrics) FormatTempoMedioNotificacao() string {
	if m.TempoMedioNotificacao <= 0 {
		return "N/A"
	}

	seconds := int(m.TempoMedioNotificacao)

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	remainingSeconds := seconds % 60

	if minutes < 60 {
		if remainingSeconds > 0 {
			return fmt.Sprintf("%dmin %ds", minutes, remainingSeconds)
		}
		return fmt.Sprintf("%dmin", minutes)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if remainingMinutes > 0 {
		return fmt.Sprintf("%dh %dmin", hours, remainingMinutes)
	}
	return fmt.Sprintf("%dh", hours)
}

// MetricsResponse represents the API response for dashboard metrics
type MetricsResponse struct {
	ObitosElegiveisHoje            int       `json:"obitos_elegiveis_hoje"`
	TempoMedioNotificacao          float64   `json:"tempo_medio_notificacao"`
	TempoMedioNotificacaoFormatado string    `json:"tempo_medio_notificacao_formatado"`
	CorneasPotenciais              int       `json:"corneas_potenciais"`
	OccurrencesPendentes           int       `json:"occurrences_pendentes"`
	OccurrencesEmAndamento         int       `json:"occurrences_em_andamento"`
	UltimaAtualizacao              time.Time `json:"ultima_atualizacao"`
}

// ToResponse converts DashboardMetrics to MetricsResponse
func (m *DashboardMetrics) ToResponse() MetricsResponse {
	return MetricsResponse{
		ObitosElegiveisHoje:            m.ObitosElegiveisHoje,
		TempoMedioNotificacao:          m.TempoMedioNotificacao,
		TempoMedioNotificacaoFormatado: m.FormatTempoMedioNotificacao(),
		CorneasPotenciais:              m.CorneasPotenciais,
		OccurrencesPendentes:           m.OccurrencesPendentes,
		OccurrencesEmAndamento:         m.OccurrencesEmAndamento,
		UltimaAtualizacao:              m.UltimaAtualizacao,
	}
}

// OccurrenceListFilters represents filters for listing occurrences
type OccurrenceListFilters struct {
	Status     *OccurrenceStatus `json:"status,omitempty"`
	HospitalID *string           `json:"hospital_id,omitempty"`
	DateFrom   *time.Time        `json:"date_from,omitempty"`
	DateTo     *time.Time        `json:"date_to,omitempty"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	SortBy     string            `json:"sort_by"`
	SortOrder  string            `json:"sort_order"`
}

// DefaultFilters returns default filter values
func DefaultFilters() OccurrenceListFilters {
	return OccurrenceListFilters{
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, page, pageSize, totalItems int) PaginatedResponse {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
