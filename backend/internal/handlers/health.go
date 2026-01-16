package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/services/listener"
	"github.com/vitalconnect/backend/internal/services/triagem"
)

var (
	globalListener     *listener.ObitoListener
	globalTriagemMotor *triagem.TriagemMotor
)

// SetGlobalListener sets the global listener instance for health checks
func SetGlobalListener(l *listener.ObitoListener) {
	globalListener = l
}

// SetGlobalTriagemMotor sets the global triagem motor instance for health checks
func SetGlobalTriagemMotor(m *triagem.TriagemMotor) {
	globalTriagemMotor = m
}

// GetGlobalListener returns the global listener instance
func GetGlobalListener() *listener.ObitoListener {
	return globalListener
}

// GetGlobalTriagemMotor returns the global triagem motor instance
func GetGlobalTriagemMotor() *triagem.TriagemMotor {
	return globalTriagemMotor
}

// ListenerHealthResponse represents the response for listener health check
type ListenerHealthResponse struct {
	Status              string                 `json:"status"`
	Listener            *ListenerDetails       `json:"listener,omitempty"`
	TriagemMotor        *TriagemMotorDetails   `json:"triagem_motor,omitempty"`
	Timestamp           string                 `json:"timestamp"`
}

// ListenerDetails represents the listener details in health response
type ListenerDetails struct {
	Status              string     `json:"status"`
	Running             bool       `json:"running"`
	UltimoProcessamento *time.Time `json:"ultimo_processamento,omitempty"`
	ObitosDetectadosHoje int       `json:"obitos_detectados_hoje"`
	TotalProcessados    int64      `json:"total_processados"`
	Errors              int64      `json:"errors"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
}

// TriagemMotorDetails represents the triagem motor details in health response
type TriagemMotorDetails struct {
	Status            string `json:"status"`
	Running           bool   `json:"running"`
	TotalProcessados  int64  `json:"total_processados"`
	TotalElegiveis    int64  `json:"total_elegiveis"`
	TotalInelegiveis  int64  `json:"total_inelegiveis"`
	Errors            int64  `json:"errors"`
}

// ListenerHealth returns the health status of the obito listener
func ListenerHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response := ListenerHealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check listener status
	if globalListener != nil {
		listenerStatus := globalListener.GetStatus(ctx)
		response.Listener = &ListenerDetails{
			Status:               listenerStatus.Status,
			Running:              listenerStatus.Running,
			UltimoProcessamento:  listenerStatus.UltimoProcessamento,
			ObitosDetectadosHoje: listenerStatus.ObitosDetectadosHoje,
			TotalProcessados:     listenerStatus.TotalProcessados,
			Errors:               listenerStatus.Errors,
			StartedAt:            listenerStatus.StartedAt,
		}

		if !listenerStatus.Running {
			response.Status = "degraded"
		}
	} else {
		response.Status = "degraded"
		response.Listener = &ListenerDetails{
			Status:  "not_initialized",
			Running: false,
		}
	}

	// Check triagem motor status
	if globalTriagemMotor != nil {
		stats := globalTriagemMotor.GetStats()
		response.TriagemMotor = &TriagemMotorDetails{
			Status:           "running",
			Running:          stats["running"].(bool),
			TotalProcessados: stats["total_processados"].(int64),
			TotalElegiveis:   stats["total_elegiveis"].(int64),
			TotalInelegiveis: stats["total_inelegiveis"].(int64),
			Errors:           stats["errors"].(int64),
		}

		if !stats["running"].(bool) {
			response.Status = "degraded"
			response.TriagemMotor.Status = "stopped"
		}
	} else {
		response.TriagemMotor = &TriagemMotorDetails{
			Status:  "not_initialized",
			Running: false,
		}
	}

	statusCode := http.StatusOK
	if response.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
