package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
)

// MapHandler handles map-related HTTP requests
type MapHandler struct {
	hospitalRepo   *repository.HospitalRepository
	occurrenceRepo *repository.OccurrenceRepository
	shiftRepo      *repository.ShiftRepository
}

// NewMapHandler creates a new map handler
func NewMapHandler(
	hospitalRepo *repository.HospitalRepository,
	occurrenceRepo *repository.OccurrenceRepository,
	shiftRepo *repository.ShiftRepository,
) *MapHandler {
	return &MapHandler{
		hospitalRepo:   hospitalRepo,
		occurrenceRepo: occurrenceRepo,
		shiftRepo:      shiftRepo,
	}
}

// GetMapHospitals returns all active hospitals with coordinates and their occurrences for map rendering
// GET /api/v1/map/hospitals
func (h *MapHandler) GetMapHospitals(c *gin.Context) {
	ctx := c.Request.Context()

	// Buscar hospitais ativos com coordenadas
	hospitals, err := h.hospitalRepo.GetActiveHospitalsWithCoordinates(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar hospitais"})
		return
	}

	now := time.Now()
	dayOfWeek := int(now.Weekday())

	// Construir resposta com dados agregados para cada hospital
	var mapHospitals []models.MapHospitalResponse
	for _, hospital := range hospitals {
		hospitalID := hospital.ID.String()

		// Buscar ocorrencias ativas (PENDENTE e EM_ANDAMENTO) para o hospital
		activeOccurrences, err := h.getActiveOccurrencesByHospital(ctx, hospitalID)
		if err != nil {
			continue // Pular hospital em caso de erro
		}

		// Construir resposta do hospital
		hospitalResp := models.MapHospitalResponse{
			ID:               hospital.ID,
			Nome:             hospital.Nome,
			Codigo:           hospital.Codigo,
			Latitude:         *hospital.Latitude,
			Longitude:        *hospital.Longitude,
			Ativo:            hospital.Ativo,
			UrgenciaMaxima:   models.CalculateMaxUrgency(activeOccurrences),
			OcorrenciasCount: len(activeOccurrences),
		}

		// Adicionar ocorrencias formatadas para o mapa
		for _, occ := range activeOccurrences {
			hospitalResp.Ocorrencias = append(hospitalResp.Ocorrencias, occ.ToMapOccurrenceResponse())
		}

		// Buscar operador de plantao atual
		activeShifts, err := h.shiftRepo.GetActiveShifts(ctx, hospital.ID, dayOfWeek, now)
		if err == nil && len(activeShifts) > 0 {
			// Usar o primeiro operador ativo encontrado
			shift := activeShifts[0]
			if shift.User != nil {
				hospitalResp.OperadorPlantao = &models.MapOperatorResponse{
					ID:     shift.User.ID,
					Nome:   shift.User.Nome,
					UserID: shift.UserID,
				}
			}
		}

		mapHospitals = append(mapHospitals, hospitalResp)
	}

	c.JSON(http.StatusOK, models.MapDataResponse{
		Hospitals: mapHospitals,
		Total:     len(mapHospitals),
	})
}

// getActiveOccurrencesByHospital busca ocorrencias ativas (PENDENTE e EM_ANDAMENTO) para um hospital
func (h *MapHandler) getActiveOccurrencesByHospital(ctx context.Context, hospitalID string) ([]models.Occurrence, error) {
	// Filtrar por status PENDENTE e EM_ANDAMENTO
	statusPendente := models.StatusPendente
	statusEmAndamento := models.StatusEmAndamento

	// Buscar ocorrencias pendentes
	filtersPendente := models.OccurrenceListFilters{
		Status:     &statusPendente,
		HospitalID: &hospitalID,
		Page:       1,
		PageSize:   100, // Maximo razoavel por hospital
	}

	pendentes, _, err := h.occurrenceRepo.List(ctx, filtersPendente)
	if err != nil {
		return nil, err
	}

	// Buscar ocorrencias em andamento
	filtersEmAndamento := models.OccurrenceListFilters{
		Status:     &statusEmAndamento,
		HospitalID: &hospitalID,
		Page:       1,
		PageSize:   100,
	}

	emAndamento, _, err := h.occurrenceRepo.List(ctx, filtersEmAndamento)
	if err != nil {
		return nil, err
	}

	// Combinar resultados
	result := append(pendentes, emAndamento...)
	return result, nil
}
