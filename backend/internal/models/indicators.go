package models

import (
	"time"

	"github.com/google/uuid"
)

// ThresholdStatus represents the visual indicator status
type ThresholdStatus string

const (
	ThresholdGreen  ThresholdStatus = "verde"
	ThresholdYellow ThresholdStatus = "amarelo"
	ThresholdRed    ThresholdStatus = "vermelho"
)

// IndicatorCard represents a single metric card with threshold status
type IndicatorCard struct {
	Valor     float64         `json:"valor"`
	Formatado string          `json:"formatado"`
	Status    ThresholdStatus `json:"status"`
}

// SeriesDataPoint represents a single data point in the time series
type SeriesDataPoint struct {
	Data        string `json:"data"`         // Date in YYYY-MM-DD format
	ObitosTotais int    `json:"obitos_totais"`
	Captados    int    `json:"captados"`
}

// Series30Dias represents the 30-day time series for the chart
type Series30Dias struct {
	Dados []SeriesDataPoint `json:"dados"`
}

// RankingItem represents a hospital in the ranking
type RankingItem struct {
	Posicao    int       `json:"posicao"`
	HospitalID uuid.UUID `json:"hospital_id"`
	Nome       string    `json:"nome"`
	Captacoes  int       `json:"captacoes"`
	Percentual float64   `json:"percentual"` // Relative to first place
}

// RankingHospitais represents the hospital ranking
type RankingHospitais struct {
	Hospitais []RankingItem `json:"hospitais"`
}

// IndicatorsMetrics represents the complete indicators dashboard metrics
type IndicatorsMetrics struct {
	TaxaConversao           IndicatorCard    `json:"taxa_conversao"`
	LatenciaSistema         IndicatorCard    `json:"latencia_sistema"`
	TempoRespostaOperacional IndicatorCard    `json:"tempo_resposta_operacional"`
	Series30Dias            Series30Dias     `json:"series_30_dias"`
	RankingHospitais        RankingHospitais `json:"ranking_hospitais"`
	UltimaAtualizacao       time.Time        `json:"ultima_atualizacao"`
}

// Threshold constants for visual indicators
const (
	// Taxa de Conversao thresholds (percentage)
	TaxaConversaoGreen  = 50.0 // >= 50% is green
	TaxaConversaoYellow = 30.0 // >= 30% is yellow, < 30% is red

	// Latencia Sistema thresholds (seconds)
	LatenciaSistemaGreen  = 30.0  // <= 30s is green
	LatenciaSistemaYellow = 60.0  // <= 60s is yellow, > 60s is red

	// Tempo Resposta Operacional thresholds (minutes)
	TempoRespostaGreen  = 15.0 // <= 15min is green
	TempoRespostaYellow = 30.0 // <= 30min is yellow, > 30min is red
)

// CalculateTaxaConversaoStatus returns the threshold status for conversion rate
func CalculateTaxaConversaoStatus(valor float64) ThresholdStatus {
	if valor >= TaxaConversaoGreen {
		return ThresholdGreen
	}
	if valor >= TaxaConversaoYellow {
		return ThresholdYellow
	}
	return ThresholdRed
}

// CalculateLatenciaSistemaStatus returns the threshold status for system latency
func CalculateLatenciaSistemaStatus(segundos float64) ThresholdStatus {
	if segundos <= LatenciaSistemaGreen {
		return ThresholdGreen
	}
	if segundos <= LatenciaSistemaYellow {
		return ThresholdYellow
	}
	return ThresholdRed
}

// CalculateTempoRespostaStatus returns the threshold status for response time
func CalculateTempoRespostaStatus(minutos float64) ThresholdStatus {
	if minutos <= TempoRespostaGreen {
		return ThresholdGreen
	}
	if minutos <= TempoRespostaYellow {
		return ThresholdYellow
	}
	return ThresholdRed
}

// FormatTaxaConversao formats the conversion rate as a percentage string
func FormatTaxaConversao(valor float64) string {
	return formatPercent(valor)
}

// FormatLatenciaSistema formats the system latency as a readable string
func FormatLatenciaSistema(segundos float64) string {
	if segundos < 60 {
		return formatFloat(segundos) + "s"
	}
	minutos := segundos / 60
	return formatFloat(minutos) + "min"
}

// FormatTempoResposta formats the response time as a readable string
func FormatTempoResposta(minutos float64) string {
	if minutos < 60 {
		return formatFloat(minutos) + "min"
	}
	horas := minutos / 60
	return formatFloat(horas) + "h"
}

// formatPercent formats a float as percentage
func formatPercent(valor float64) string {
	if valor == float64(int(valor)) {
		return formatInt(int(valor)) + "%"
	}
	return formatFloat(valor) + "%"
}

// formatFloat formats a float with one decimal place
func formatFloat(valor float64) string {
	// Round to 1 decimal place
	rounded := float64(int(valor*10)) / 10
	if rounded == float64(int(rounded)) {
		return formatInt(int(rounded))
	}
	return formatFloatWithDecimal(rounded)
}

// formatFloatWithDecimal formats a float with decimal
func formatFloatWithDecimal(valor float64) string {
	intPart := int(valor)
	decPart := int((valor - float64(intPart)) * 10)
	if decPart < 0 {
		decPart = -decPart
	}
	return formatInt(intPart) + "," + formatInt(decPart)
}

// formatInt formats an integer
func formatInt(valor int) string {
	if valor < 0 {
		return "-" + formatPositiveInt(-valor)
	}
	return formatPositiveInt(valor)
}

// formatPositiveInt formats a positive integer
func formatPositiveInt(valor int) string {
	if valor == 0 {
		return "0"
	}

	// Convert to string manually to avoid imports
	result := ""
	for valor > 0 {
		digit := valor % 10
		result = string(rune('0'+digit)) + result
		valor /= 10
	}
	return result
}

// NewIndicatorCard creates a new indicator card with calculated status
func NewIndicatorCard(valor float64, formatter func(float64) string, statusCalc func(float64) ThresholdStatus) IndicatorCard {
	return IndicatorCard{
		Valor:     valor,
		Formatado: formatter(valor),
		Status:    statusCalc(valor),
	}
}
