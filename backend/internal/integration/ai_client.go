package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AIServiceClient handles HTTP communication with the Python AI service
type AIServiceClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// AIServiceConfig holds configuration for the AI service client
type AIServiceConfig struct {
	BaseURL       string
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// DefaultAIServiceConfig returns default configuration for the AI service
func DefaultAIServiceConfig() *AIServiceConfig {
	return &AIServiceConfig{
		BaseURL:       getAIServiceURL(),
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}
}

// getAIServiceURL retrieves the AI service URL from environment
func getAIServiceURL() string {
	url := os.Getenv("AI_SERVICE_URL")
	if url == "" {
		url = "http://ai-service:8000"
	}
	return url
}

// NewAIServiceClient creates a new AI service client with the given configuration
func NewAIServiceClient(config *AIServiceConfig) *AIServiceClient {
	if config == nil {
		config = DefaultAIServiceConfig()
	}

	return &AIServiceClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		maxRetries: config.MaxRetries,
	}
}

// AIRequestOptions holds options for making requests to the AI service
type AIRequestOptions struct {
	AuthToken     string
	TenantContext string
	ContentType   string
}

// ChatRequest represents a request to the AI chat endpoint
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// ChatResponse represents a response from the AI chat endpoint
type ChatResponse struct {
	Response             string                 `json:"response"`
	SessionID            string                 `json:"session_id"`
	ToolCalls            []ToolCall             `json:"tool_calls,omitempty"`
	ConfirmationRequired *ConfirmationRequired  `json:"confirmation_required,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool call made by the AI agent
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     interface{}            `json:"result,omitempty"`
}

// ConfirmationRequired represents a pending action requiring user confirmation
type ConfirmationRequired struct {
	ActionID    string                 `json:"action_id"`
	ActionType  string                 `json:"action_type"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ConfirmActionRequest represents a request to confirm/reject a pending action
type ConfirmActionRequest struct {
	Confirmed bool `json:"confirmed"`
}

// ConfirmActionResponse represents a response from the confirm action endpoint
type ConfirmActionResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Result  map[string]interface{} `json:"result,omitempty"`
}

// ConversationListResponse represents a list of conversations
type ConversationListResponse struct {
	Data  []ConversationSummary `json:"data"`
	Total int                   `json:"total"`
}

// ConversationSummary represents a summary of a conversation
type ConversationSummary struct {
	SessionID   string    `json:"session_id"`
	LastMessage string    `json:"last_message"`
	MessageCount int      `json:"message_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConversationMessage represents a single message in a conversation
type ConversationMessage struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// ConversationDetailResponse represents the details of a conversation
type ConversationDetailResponse struct {
	SessionID string                `json:"session_id"`
	Messages  []ConversationMessage `json:"messages"`
}

// HealthResponse represents the health status of the AI service
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// doRequest executes an HTTP request with retry logic
func (c *AIServiceClient) doRequest(ctx context.Context, method, path string, body interface{}, opts *AIRequestOptions) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			// Reset body reader for retry
			if body != nil {
				jsonData, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(jsonData)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		if opts != nil {
			if opts.AuthToken != "" {
				req.Header.Set("Authorization", "Bearer "+opts.AuthToken)
			}
			if opts.TenantContext != "" {
				req.Header.Set("X-Tenant-Context", opts.TenantContext)
			}
			if opts.ContentType != "" {
				req.Header.Set("Content-Type", opts.ContentType)
			} else if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
		} else if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && attempt < c.maxRetries {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Health checks the health of the AI service
func (c *AIServiceClient) Health(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/health", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("health check failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return &healthResp, nil
}

// SendChatMessage sends a chat message to the AI service
func (c *AIServiceClient) SendChatMessage(ctx context.Context, message string, sessionID string, opts *AIRequestOptions) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Message:   message,
		SessionID: sessionID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/ai/chat", reqBody, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}

	return &chatResp, nil
}

// ConfirmAction confirms or rejects a pending AI action
func (c *AIServiceClient) ConfirmAction(ctx context.Context, actionID string, confirmed bool, opts *AIRequestOptions) (*ConfirmActionResponse, error) {
	reqBody := ConfirmActionRequest{
		Confirmed: confirmed,
	}

	path := fmt.Sprintf("/api/v1/ai/confirm/%s", actionID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("confirm action failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var confirmResp ConfirmActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&confirmResp); err != nil {
		return nil, fmt.Errorf("failed to decode confirm response: %w", err)
	}

	return &confirmResp, nil
}

// ListConversations lists all conversations for the authenticated user
func (c *AIServiceClient) ListConversations(ctx context.Context, opts *AIRequestOptions) (*ConversationListResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/ai/conversations", nil, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list conversations failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var listResp ConversationListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode conversations response: %w", err)
	}

	return &listResp, nil
}

// GetConversation retrieves a specific conversation by session ID
func (c *AIServiceClient) GetConversation(ctx context.Context, sessionID string, opts *AIRequestOptions) (*ConversationDetailResponse, error) {
	path := fmt.Sprintf("/api/v1/ai/conversations/%s", sessionID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get conversation failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var convResp ConversationDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&convResp); err != nil {
		return nil, fmt.Errorf("failed to decode conversation response: %w", err)
	}

	return &convResp, nil
}

// DeleteConversation deletes a conversation by session ID
func (c *AIServiceClient) DeleteConversation(ctx context.Context, sessionID string, opts *AIRequestOptions) error {
	path := fmt.Sprintf("/api/v1/ai/conversations/%s", sessionID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete conversation failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetBaseURL returns the base URL of the AI service
func (c *AIServiceClient) GetBaseURL() string {
	return c.baseURL
}
