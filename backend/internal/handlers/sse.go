package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sidot/backend/internal/middleware"
	"github.com/sidot/backend/internal/services/auth"
	"github.com/sidot/backend/internal/services/notification"
)

var (
	globalSSEHub *notification.SSEHub
)

// SetGlobalSSEHub sets the global SSE hub instance
func SetGlobalSSEHub(hub *notification.SSEHub) {
	globalSSEHub = hub
}

// GetGlobalSSEHub returns the global SSE hub instance
func GetGlobalSSEHub() *notification.SSEHub {
	return globalSSEHub
}

// NotificationStream handles SSE connections for real-time notifications
// GET /api/v1/notifications/stream
func NotificationStream(c *gin.Context) {
	if globalSSEHub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SSE service not available"})
		return
	}

	// Get user claims from context (set by auth middleware)
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		// Try to authenticate via query param token for SSE
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		// Validate token from query param
		jwtService, jwtOk := middleware.GetJWTService(c)
		if !jwtOk {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service not configured"})
			return
		}

		tokenClaims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token has expired",
					"code":  "TOKEN_EXPIRED",
				})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid token",
					"code":  "INVALID_TOKEN",
				})
			}
			return
		}

		claims = &middleware.UserClaims{
			UserID:     tokenClaims.UserID,
			Email:      tokenClaims.Email,
			Role:       tokenClaims.Role,
			HospitalID: tokenClaims.HospitalID,
		}
	}

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable buffering in nginx

	// Create SSE client
	client := notification.NewSSEClient(claims.UserID, claims.Role)
	globalSSEHub.RegisterClient(client)

	// Ensure cleanup on disconnect
	defer func() {
		globalSSEHub.UnregisterClient(client.ID)
	}()

	// Send initial connection event
	sendSSEEvent(c.Writer, "connected", map[string]interface{}{
		"client_id": client.ID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	c.Writer.Flush()

	// Get request context for cancellation
	ctx := c.Request.Context()

	// Listen for events
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return

		case <-client.Done:
			// Client was closed
			return

		case event := <-client.Channel:
			if event == nil {
				continue
			}

			// Send event to client
			if err := sendSSEEvent(c.Writer, event.Type, event); err != nil {
				// Error writing, client probably disconnected
				return
			}
			c.Writer.Flush()
		}
	}
}

// sendSSEEvent writes an SSE event to the response writer
func sendSSEEvent(w io.Writer, eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// SSE format: event: <type>\ndata: <json>\n\n
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
	return err
}

// SSEHealthResponse represents the health status of SSE
type SSEHealthResponse struct {
	Status           string                 `json:"status"`
	Hub              map[string]interface{} `json:"hub,omitempty"`
	EmailQueue       map[string]interface{} `json:"email_queue,omitempty"`
	Timestamp        string                 `json:"timestamp"`
}

// SSEHealth returns the health status of the SSE service
// GET /api/v1/health/sse
func SSEHealth(c *gin.Context) {
	response := SSEHealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if globalSSEHub != nil {
		response.Hub = globalSSEHub.GetStats()
		if !globalSSEHub.IsRunning() {
			response.Status = "degraded"
		}
	} else {
		response.Status = "degraded"
		response.Hub = map[string]interface{}{
			"running": false,
			"error":   "SSE hub not initialized",
		}
	}

	statusCode := http.StatusOK
	if response.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
