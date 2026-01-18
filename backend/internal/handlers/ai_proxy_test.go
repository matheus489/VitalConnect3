package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/integration"
	"github.com/vitalconnect/backend/internal/middleware"
)

// setupAIProxyTestRouter creates a test router for AI proxy tests
func setupAIProxyTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// mockAIUserClaims sets up mock user claims in the context
func mockAIUserClaims(c *gin.Context, userID, tenantID, role string, isSuperAdmin bool) {
	claims := &middleware.UserClaims{
		UserID:       userID,
		Email:        "test@example.com",
		Role:         role,
		TenantID:     tenantID,
		IsSuperAdmin: isSuperAdmin,
	}
	c.Set("user_claims", claims)
}

// mockAITenantContext sets up mock tenant context in the context
func mockAITenantContext(c *gin.Context, tenantID string, isSuperAdmin bool) {
	ctx := &middleware.TenantContext{
		TenantID:          tenantID,
		IsSuperAdmin:      isSuperAdmin,
		EffectiveTenantID: tenantID,
	}
	c.Set("tenant_context", ctx)
}

// TestAIProxyForwardsJWTToken verifies that JWT tokens are forwarded correctly
func TestAIProxyForwardsJWTToken(t *testing.T) {
	// Create a mock AI service that captures the incoming request
	var capturedAuthHeader string
	mockAIService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response":   "Test response",
			"session_id": "test-session",
		})
	}))
	defer mockAIService.Close()

	// Configure AI client to use mock server
	config := &integration.AIServiceConfig{
		BaseURL:    mockAIService.URL,
		Timeout:    5000000000, // 5s
		MaxRetries: 1,
	}
	client := integration.NewAIServiceClient(config)
	SetAIServiceClient(client)
	defer SetAIServiceClient(nil)

	// Setup router
	router := setupAIProxyTestRouter()
	router.POST("/api/v1/ai/chat", func(c *gin.Context) {
		// Mock authentication
		mockAIUserClaims(c, "user-123", "tenant-456", "operador", false)
		mockAITenantContext(c, "tenant-456", false)
		c.Request.Header.Set("Authorization", "Bearer test-jwt-token-12345")
		AIChat(c)
	})

	// Make request
	reqBody := `{"message": "Hello AI"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-jwt-token-12345")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify token was forwarded
	if capturedAuthHeader != "Bearer test-jwt-token-12345" {
		t.Errorf("JWT token not forwarded correctly. Expected 'Bearer test-jwt-token-12345', got '%s'", capturedAuthHeader)
	}

	// Verify response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAIProxyForwardsTenantContext verifies that tenant context is forwarded correctly
func TestAIProxyForwardsTenantContext(t *testing.T) {
	// Create a mock AI service that captures the incoming request
	var capturedTenantHeader string
	mockAIService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantHeader = r.Header.Get("X-Tenant-Context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response":   "Test response",
			"session_id": "test-session",
		})
	}))
	defer mockAIService.Close()

	// Configure AI client to use mock server
	config := &integration.AIServiceConfig{
		BaseURL:    mockAIService.URL,
		Timeout:    5000000000, // 5s
		MaxRetries: 1,
	}
	client := integration.NewAIServiceClient(config)
	SetAIServiceClient(client)
	defer SetAIServiceClient(nil)

	// Setup router
	router := setupAIProxyTestRouter()
	router.POST("/api/v1/ai/chat", func(c *gin.Context) {
		// Mock authentication with specific tenant
		tenantID := "550e8400-e29b-41d4-a716-446655440000"
		mockAIUserClaims(c, "user-123", tenantID, "operador", false)
		mockAITenantContext(c, tenantID, false)
		AIChat(c)
	})

	// Make request
	reqBody := `{"message": "Hello AI"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify tenant context was forwarded
	expectedTenantID := "550e8400-e29b-41d4-a716-446655440000"
	if capturedTenantHeader != expectedTenantID {
		t.Errorf("Tenant context not forwarded correctly. Expected '%s', got '%s'", expectedTenantID, capturedTenantHeader)
	}

	// Verify response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAIProxyForwardsRequestCorrectly verifies that the proxy forwards requests to the AI service
func TestAIProxyForwardsRequestCorrectly(t *testing.T) {
	// Create a mock AI service that captures the incoming request
	var capturedMethod string
	var capturedPath string
	var capturedBody map[string]interface{}

	mockAIService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path

		// Parse body
		json.NewDecoder(r.Body).Decode(&capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response":   "AI Response to: " + capturedBody["message"].(string),
			"session_id": "session-abc123",
		})
	}))
	defer mockAIService.Close()

	// Configure AI client to use mock server
	config := &integration.AIServiceConfig{
		BaseURL:    mockAIService.URL,
		Timeout:    5000000000, // 5s
		MaxRetries: 1,
	}
	client := integration.NewAIServiceClient(config)
	SetAIServiceClient(client)
	defer SetAIServiceClient(nil)

	// Setup router
	router := setupAIProxyTestRouter()
	router.POST("/api/v1/ai/chat", func(c *gin.Context) {
		mockAIUserClaims(c, "user-123", "tenant-456", "operador", false)
		mockAITenantContext(c, "tenant-456", false)
		AIChat(c)
	})

	// Make request
	reqBody := `{"message": "What are the pending occurrences?", "session_id": "existing-session"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify request was forwarded
	if capturedMethod != http.MethodPost {
		t.Errorf("Expected POST method, got %s", capturedMethod)
	}

	if capturedPath != "/api/v1/ai/chat" {
		t.Errorf("Expected path /api/v1/ai/chat, got %s", capturedPath)
	}

	if capturedBody["message"] != "What are the pending occurrences?" {
		t.Errorf("Message not forwarded correctly, got %v", capturedBody["message"])
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["session_id"] != "session-abc123" {
		t.Errorf("Session ID not returned correctly, got %v", response["session_id"])
	}
}

// TestAIProxyServiceUnavailable verifies behavior when AI service is not configured
func TestAIProxyServiceUnavailable(t *testing.T) {
	// Clear AI service client
	SetAIServiceClient(nil)

	// Setup router
	router := setupAIProxyTestRouter()
	router.POST("/api/v1/ai/chat", func(c *gin.Context) {
		mockAIUserClaims(c, "user-123", "tenant-456", "operador", false)
		mockAITenantContext(c, "tenant-456", false)
		AIChat(c)
	})

	// Make request
	reqBody := `{"message": "Hello AI"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify service unavailable response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "AI service not configured" {
		t.Errorf("Expected 'AI service not configured' error, got %v", response["error"])
	}
}
