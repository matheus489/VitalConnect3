package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UrgencyLevel representa o nivel de urgencia de uma ocorrencia
// Baseado no tempo restante da janela de isquemia
type UrgencyLevel string

const (
	// UrgencyGreen indica tempo restante > 4 horas (janela folgada)
	UrgencyGreen UrgencyLevel = "green"
	// UrgencyYellow indica tempo restante entre 2 e 4 horas (atencao)
	UrgencyYellow UrgencyLevel = "yellow"
	// UrgencyRed indica tempo restante < 2 horas (critico)
	UrgencyRed UrgencyLevel = "red"
	// UrgencyNone indica que nao ha ocorrencias ativas
	UrgencyNone UrgencyLevel = "none"
)

// MapHospitalResponse representa os dados de um hospital para renderizacao no mapa
type MapHospitalResponse struct {
	ID               uuid.UUID             `json:"id"`
	Nome             string                `json:"nome"`
	Codigo           string                `json:"codigo"`
	Latitude         float64               `json:"latitude"`
	Longitude        float64               `json:"longitude"`
	Ativo            bool                  `json:"ativo"`
	UrgenciaMaxima   UrgencyLevel          `json:"urgencia_maxima"`
	OcorrenciasCount int                   `json:"ocorrencias_count"`
	Ocorrencias      []MapOccurrenceResponse `json:"ocorrencias,omitempty"`
	OperadorPlantao  *MapOperatorResponse  `json:"operador_plantao,omitempty"`
}

// MapOccurrenceResponse representa os dados de uma ocorrencia para o mapa
type MapOccurrenceResponse struct {
	ID                    uuid.UUID        `json:"id"`
	NomePacienteMascarado string           `json:"nome_paciente_mascarado"`
	Setor                 string           `json:"setor"`
	TempoRestante         string           `json:"tempo_restante"`
	TempoRestanteMinutos  int              `json:"tempo_restante_minutos"`
	Status                OccurrenceStatus `json:"status"`
	Urgencia              UrgencyLevel     `json:"urgencia"`
}

// MapOperatorResponse representa os dados do operador de plantao para o mapa
type MapOperatorResponse struct {
	ID     uuid.UUID `json:"id"`
	Nome   string    `json:"nome"`
	UserID uuid.UUID `json:"user_id"`
}

// MapDataResponse representa a resposta completa do endpoint do mapa
type MapDataResponse struct {
	Hospitals []MapHospitalResponse `json:"hospitals"`
	Total     int                   `json:"total"`
}

// CalculateUrgencyLevel calcula o nivel de urgencia baseado no tempo restante em minutos
// >4h (>240min) = verde, 2-4h (120-240min) = amarelo, <2h (<120min) = vermelho
func CalculateUrgencyLevel(tempoRestanteMinutos int) UrgencyLevel {
	if tempoRestanteMinutos <= 0 {
		return UrgencyRed
	}
	if tempoRestanteMinutos < 120 { // < 2 horas
		return UrgencyRed
	}
	if tempoRestanteMinutos < 240 { // 2-4 horas
		return UrgencyYellow
	}
	return UrgencyGreen // > 4 horas
}

// CalculateUrgencyFromExpiration calcula o nivel de urgencia baseado na data de expiracao
func CalculateUrgencyFromExpiration(janelaExpiraEm time.Time) UrgencyLevel {
	remaining := janelaExpiraEm.Sub(time.Now())
	if remaining <= 0 {
		return UrgencyRed
	}
	minutes := int(remaining.Minutes())
	return CalculateUrgencyLevel(minutes)
}

// CalculateMaxUrgency calcula a urgencia maxima de uma lista de ocorrencias
// Retorna o nivel mais critico (vermelho > amarelo > verde > none)
func CalculateMaxUrgency(occurrences []Occurrence) UrgencyLevel {
	if len(occurrences) == 0 {
		return UrgencyNone
	}

	maxUrgency := UrgencyGreen
	for _, o := range occurrences {
		urgency := CalculateUrgencyFromExpiration(o.JanelaExpiraEm)

		// Prioridade: Red > Yellow > Green
		if urgency == UrgencyRed {
			return UrgencyRed // Nao precisa continuar, vermelho e o mais critico
		}
		if urgency == UrgencyYellow && maxUrgency == UrgencyGreen {
			maxUrgency = UrgencyYellow
		}
	}

	return maxUrgency
}

// ToMapOccurrenceResponse converte uma Occurrence para MapOccurrenceResponse
func (o *Occurrence) ToMapOccurrenceResponse() MapOccurrenceResponse {
	// Calcular tempo restante
	remaining := o.TimeRemaining()
	minutesRemaining := int(remaining.Minutes())

	// Extrair setor dos dados completos
	var setor string
	var data OccurrenceCompleteData
	if err := json.Unmarshal(o.DadosCompletos, &data); err == nil {
		setor = data.Setor
	}

	return MapOccurrenceResponse{
		ID:                    o.ID,
		NomePacienteMascarado: o.NomePacienteMascarado,
		Setor:                 setor,
		TempoRestante:         o.FormatTimeRemaining(),
		TempoRestanteMinutos:  minutesRemaining,
		Status:                o.Status,
		Urgencia:              CalculateUrgencyLevel(minutesRemaining),
	}
}

// GetUrgencyLabel retorna o texto em portugues para o nivel de urgencia
func GetUrgencyLabel(level UrgencyLevel) string {
	switch level {
	case UrgencyGreen:
		return "Normal"
	case UrgencyYellow:
		return "Atencao"
	case UrgencyRed:
		return "Critico"
	case UrgencyNone:
		return "Sem ocorrencias"
	default:
		return "Desconhecido"
	}
}

// GetUrgencyColor retorna a cor CSS/hex para o nivel de urgencia
func GetUrgencyColor(level UrgencyLevel) string {
	switch level {
	case UrgencyGreen:
		return "#22c55e" // green-500
	case UrgencyYellow:
		return "#eab308" // yellow-500
	case UrgencyRed:
		return "#ef4444" // red-500
	case UrgencyNone:
		return "#6b7280" // gray-500
	default:
		return "#6b7280"
	}
}
