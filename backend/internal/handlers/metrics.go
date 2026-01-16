package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

var metricsOccurrenceRepo *repository.OccurrenceRepository

// SetMetricsOccurrenceRepository sets the occurrence repository for metrics handler
func SetMetricsOccurrenceRepository(repo *repository.OccurrenceRepository) {
	metricsOccurrenceRepo = repo
}

// GetDashboardMetrics returns dashboard metrics
// GET /api/v1/metrics/dashboard
func GetDashboardMetrics(c *gin.Context) {
	if metricsOccurrenceRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "occurrence repository not configured"})
		return
	}

	ctx := c.Request.Context()

	// Get today's eligible deaths count
	obitosPotenciais, err := metricsOccurrenceRepo.GetTodayEligibleCount(ctx)
	if err != nil {
		obitosPotenciais = 0
	}

	// Get average notification time
	tempoMedioNotificacao, err := metricsOccurrenceRepo.GetAverageNotificationTime(ctx)
	if err != nil {
		tempoMedioNotificacao = 0
	}

	// Get pending occurrences count
	occurrencesPendentes, err := metricsOccurrenceRepo.GetPendingCount(ctx)
	if err != nil {
		occurrencesPendentes = 0
	}

	// Get in-progress occurrences count
	occurrencesEmAndamento, err := metricsOccurrenceRepo.GetEmAndamentoCount(ctx)
	if err != nil {
		occurrencesEmAndamento = 0
	}

	// Calculate potential corneas (eligible deaths * 2)
	corneasPotenciais := obitosPotenciais * 2

	metrics := &models.DashboardMetrics{
		ObitosElegiveisHoje:    obitosPotenciais,
		TempoMedioNotificacao:  tempoMedioNotificacao,
		CorneasPotenciais:      corneasPotenciais,
		OccurrencesPendentes:   occurrencesPendentes,
		OccurrencesEmAndamento: occurrencesEmAndamento,
		UltimaAtualizacao:      time.Now(),
	}

	c.JSON(http.StatusOK, metrics.ToResponse())
}
