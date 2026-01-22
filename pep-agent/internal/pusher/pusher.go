// Package pusher handles pushing events to the SIDOT central server
package pusher

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sidot/pep-agent/internal/config"
	"github.com/sidot/pep-agent/internal/models"
)

// Pusher sends events to the SIDOT central server
type Pusher struct {
	config *config.AgentConfig
	client *http.Client
}

// PushResult represents the result of a push operation
type PushResult struct {
	Success    bool
	StatusCode int
	Message    string
	EventID    string
}

// RetryConfig defines the backoff strategy for retries
type RetryConfig struct {
	MaxRetries int
	Intervals  []time.Duration
}

// DefaultRetryConfig returns the default retry configuration
// Intervals: 10s, 30s, 1min, 2min, 5min (cap)
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 5,
		Intervals: []time.Duration{
			10 * time.Second,
			30 * time.Second,
			1 * time.Minute,
			2 * time.Minute,
			5 * time.Minute,
		},
	}
}

// NewPusher creates a new Pusher instance
func NewPusher(cfg *config.AgentConfig) *Pusher {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		MaxConnsPerHost:     5,
	}

	// Allow insecure TLS for development (controlled by config)
	if cfg.Central.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.GetTimeout(),
	}

	return &Pusher{
		config: cfg,
		client: client,
	}
}

// Push sends a single event to the central server
func (p *Pusher) Push(ctx context.Context, event *models.ObitoEvent) (*PushResult, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.Central.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.config.Central.APIKey)
	req.Header.Set("User-Agent", "SIDOT-PEP-Agent/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return &PushResult{
			Success: false,
			Message: fmt.Sprintf("request failed: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := &PushResult{
		StatusCode: resp.StatusCode,
		EventID:    event.HospitalIDOrigem,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Message = "event pushed successfully"
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(body))
	}

	return result, nil
}

// PushWithRetry sends an event with exponential backoff retry
func (p *Pusher) PushWithRetry(ctx context.Context, event *models.ObitoEvent, retryConfig *RetryConfig) (*PushResult, error) {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	var lastErr error
	var result *PushResult

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		result, lastErr = p.Push(ctx, event)

		if lastErr == nil && result.Success {
			return result, nil
		}

		// Don't retry on authentication or validation errors
		if result != nil && (result.StatusCode == 401 || result.StatusCode == 400) {
			return result, fmt.Errorf("non-retryable error: %s", result.Message)
		}

		// Calculate backoff
		if attempt < retryConfig.MaxRetries {
			backoff := retryConfig.Intervals[attempt]
			if attempt >= len(retryConfig.Intervals) {
				// Use cap (last interval) for additional retries
				backoff = retryConfig.Intervals[len(retryConfig.Intervals)-1]
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt
			}
		}
	}

	if lastErr != nil {
		return result, fmt.Errorf("all retry attempts failed: %w", lastErr)
	}

	return result, fmt.Errorf("all retry attempts failed: %s", result.Message)
}

// PushBatch sends multiple events with retry
func (p *Pusher) PushBatch(ctx context.Context, events []*models.ObitoEvent, retryConfig *RetryConfig) ([]*PushResult, error) {
	results := make([]*PushResult, 0, len(events))
	var lastErr error

	for _, event := range events {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
			result, err := p.PushWithRetry(ctx, event, retryConfig)
			if err != nil {
				lastErr = err
			}
			results = append(results, result)
		}
	}

	return results, lastErr
}

// HealthCheck verifies connectivity to the central server
func (p *Pusher) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.Central.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("X-API-Key", p.config.Central.APIKey)
	req.Header.Set("User-Agent", "SIDOT-PEP-Agent/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Accept 200 OK or 405 Method Not Allowed (endpoint only supports POST)
	if resp.StatusCode != 200 && resp.StatusCode != 405 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}
