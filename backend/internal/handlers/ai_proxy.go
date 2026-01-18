package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/integration"
	"github.com/vitalconnect/backend/internal/middleware"
)

var (
	aiServiceClient *integration.AIServiceClient
)

// SetAIServiceClient sets the global AI service client
func SetAIServiceClient(client *integration.AIServiceClient) {
	aiServiceClient = client
}

// GetAIServiceClient returns the global AI service client
func GetAIServiceClient() *integration.AIServiceClient {
	return aiServiceClient
}

// getAIServiceURL retrieves the AI service URL from environment
func getAIServiceURL() string {
	url := os.Getenv("AI_SERVICE_URL")
	if url == "" {
		url = "http://ai-service:8000"
	}
	return url
}

// extractAuthAndTenant extracts auth token and tenant context from the Gin context
func extractAuthAndTenant(c *gin.Context) *integration.AIRequestOptions {
	opts := &integration.AIRequestOptions{}

	// Get Authorization header - forward the full token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Remove "Bearer " prefix if present to store just the token
		if strings.HasPrefix(authHeader, "Bearer ") {
			opts.AuthToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			opts.AuthToken = authHeader
		}
	}

	// Only forward X-Tenant-Context header for super admins doing context switching
	// The AI service will use the tenant_id from JWT for regular users
	if tenantCtx, ok := middleware.GetTenantContext(c); ok {
		// Only send X-Tenant-Context if user is super admin AND switched context
		if tenantCtx.IsSuperAdmin && tenantCtx.EffectiveTenantID != tenantCtx.TenantID {
			opts.TenantContext = tenantCtx.EffectiveTenantID
		}
	}

	return opts
}

// AIChatRequest represents the request body for AI chat
type AIChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
}

// AIChat handles POST /api/v1/ai/chat
// Proxies chat messages to the Python AI service
func AIChat(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	opts := extractAuthAndTenant(c)

	// Forward the request to the AI service
	resp, err := aiServiceClient.SendChatMessage(c.Request.Context(), req.Message, req.SessionID, opts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to communicate with AI service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AIChatStream handles POST /api/v1/ai/chat/stream
// Proxies streaming chat responses from the Python AI service using SSE
func AIChatStream(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	opts := extractAuthAndTenant(c)

	// Build the upstream request to Python AI service
	aiServiceURL := aiServiceClient.GetBaseURL()
	upstreamURL := aiServiceURL + "/api/v1/ai/chat/stream"

	reqBody, _ := json.Marshal(req)
	upstreamReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, strings.NewReader(string(reqBody)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create upstream request",
		})
		return
	}

	// Set headers
	upstreamReq.Header.Set("Content-Type", "application/json")
	if opts.AuthToken != "" {
		upstreamReq.Header.Set("Authorization", "Bearer "+opts.AuthToken)
	}
	if opts.TenantContext != "" {
		upstreamReq.Header.Set("X-Tenant-Context", opts.TenantContext)
	}

	// Create a client with no timeout for streaming
	streamClient := &http.Client{
		Timeout: 0, // No timeout for streaming
	}

	// Make the request to the AI service
	upstreamResp, err := streamClient.Do(upstreamReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to connect to AI service",
			"details": err.Error(),
		})
		return
	}
	defer upstreamResp.Body.Close()

	// Check if the upstream response is an error
	if upstreamResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(upstreamResp.Body)
		c.JSON(upstreamResp.StatusCode, gin.H{
			"error":   "AI service returned an error",
			"details": string(body),
		})
		return
	}

	// Set SSE headers for the response
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Proxy the SSE stream from the AI service to the client
	reader := bufio.NewReader(upstreamResp.Body)
	for {
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			// Log error but continue - could be temporary
			continue
		}

		// Write the line to the client
		_, writeErr := c.Writer.Write(line)
		if writeErr != nil {
			return
		}
		c.Writer.Flush()
	}
}

// AIConfirmActionRequest represents the request body for confirming an AI action
type AIConfirmActionRequest struct {
	Confirmed bool `json:"confirmed"`
}

// AIConfirmAction handles POST /api/v1/ai/confirm/:action_id
// Confirms or rejects a pending AI action
func AIConfirmAction(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	actionID := c.Param("action_id")
	if actionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action_id is required",
		})
		return
	}

	var req AIConfirmActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	opts := extractAuthAndTenant(c)

	resp, err := aiServiceClient.ConfirmAction(c.Request.Context(), actionID, req.Confirmed, opts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to communicate with AI service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AIListConversations handles GET /api/v1/ai/conversations
// Lists all conversations for the authenticated user
func AIListConversations(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	opts := extractAuthAndTenant(c)

	resp, err := aiServiceClient.ListConversations(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to communicate with AI service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AIGetConversation handles GET /api/v1/ai/conversations/:session_id
// Gets a specific conversation by session ID
func AIGetConversation(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	opts := extractAuthAndTenant(c)

	resp, err := aiServiceClient.GetConversation(c.Request.Context(), sessionID, opts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to communicate with AI service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AIDeleteConversation handles DELETE /api/v1/ai/conversations/:session_id
// Deletes a conversation by session ID
func AIDeleteConversation(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id is required",
		})
		return
	}

	opts := extractAuthAndTenant(c)

	err := aiServiceClient.DeleteConversation(c.Request.Context(), sessionID, opts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to communicate with AI service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "conversation deleted successfully",
	})
}

// AIHealth handles GET /api/v1/ai/health
// Returns the health status of the AI service
func AIHealth(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unavailable",
			"error":     "AI service not configured",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := aiServiceClient.Health(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"error":     fmt.Sprintf("AI service health check failed: %v", err),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     resp.Status,
		"ai_service": resp,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// AIGenericProxy handles any AI-related request that doesn't have a specific handler
// This is useful for proxying document management endpoints or future AI endpoints
func AIGenericProxy(c *gin.Context) {
	if aiServiceClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI service not configured",
		})
		return
	}

	aiServiceURL := aiServiceClient.GetBaseURL()

	// Build the target URL - preserve the path after /api/v1/ai
	targetPath := c.Request.URL.Path
	targetURL := aiServiceURL + targetPath

	// Add query string if present
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Read the request body
	var bodyReader io.Reader
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil && len(bodyBytes) > 0 {
			bodyReader = strings.NewReader(string(bodyBytes))
		}
	}

	// Create the upstream request
	upstreamReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, bodyReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create upstream request",
		})
		return
	}

	// Forward relevant headers
	opts := extractAuthAndTenant(c)
	if opts.AuthToken != "" {
		upstreamReq.Header.Set("Authorization", "Bearer "+opts.AuthToken)
	}
	if opts.TenantContext != "" {
		upstreamReq.Header.Set("X-Tenant-Context", opts.TenantContext)
	}

	// Copy content-type header
	if contentType := c.GetHeader("Content-Type"); contentType != "" {
		upstreamReq.Header.Set("Content-Type", contentType)
	}

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to communicate with AI service",
			"details": err.Error(),
		})
		return
	}
	defer upstreamResp.Body.Close()

	// Copy response headers
	for key, values := range upstreamResp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// Write status code and body
	c.Writer.WriteHeader(upstreamResp.StatusCode)
	io.Copy(c.Writer, upstreamResp.Body)
}
