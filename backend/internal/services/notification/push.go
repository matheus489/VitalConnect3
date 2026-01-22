package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sidot/backend/internal/models"
)

// PushConfig holds Firebase Cloud Messaging configuration
type PushConfig struct {
	// FCM Server Key (Legacy) or Service Account JSON path
	// Get from Firebase Console > Project Settings > Cloud Messaging
	ServerKey string

	// Optional: Path to Firebase service account JSON for FCM v1 API
	ServiceAccountPath string

	// FCM API endpoint
	FCMURL string
}

// PushService handles sending push notifications via FCM
type PushService struct {
	config     *PushConfig
	httpClient *http.Client
}

// PushPayload represents the notification payload
type PushPayload struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Icon     string            `json:"icon,omitempty"`
	Badge    string            `json:"badge,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	ClickURL string            `json:"click_url,omitempty"`
}

// FCMMessage represents the FCM message format
type FCMMessage struct {
	To           string                 `json:"to,omitempty"`
	Notification *FCMNotification       `json:"notification,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	WebPush      *FCMWebPush            `json:"webpush,omitempty"`
	Android      *FCMAndroid            `json:"android,omitempty"`
}

type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Badge string `json:"badge,omitempty"`
}

type FCMWebPush struct {
	Headers      map[string]string `json:"headers,omitempty"`
	Notification *FCMNotification  `json:"notification,omitempty"`
	FCMOptions   *FCMWebPushOptions `json:"fcm_options,omitempty"`
}

type FCMWebPushOptions struct {
	Link string `json:"link,omitempty"`
}

type FCMAndroid struct {
	Priority     string           `json:"priority,omitempty"`
	Notification *FCMNotification `json:"notification,omitempty"`
}

// FCMResponse represents the FCM API response
type FCMResponse struct {
	MulticastID  int64        `json:"multicast_id"`
	Success      int          `json:"success"`
	Failure      int          `json:"failure"`
	Results      []FCMResult  `json:"results"`
}

type FCMResult struct {
	MessageID      string `json:"message_id,omitempty"`
	Error          string `json:"error,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
}

// NewPushService creates a new push notification service
func NewPushService(config *PushConfig) *PushService {
	if config.FCMURL == "" {
		config.FCMURL = "https://fcm.googleapis.com/fcm/send"
	}

	return &PushService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if the push service is properly configured
func (s *PushService) IsConfigured() bool {
	return s.config != nil && s.config.ServerKey != ""
}

// SendToToken sends a push notification to a specific FCM token
func (s *PushService) SendToToken(ctx context.Context, token string, payload *PushPayload) error {
	if !s.IsConfigured() {
		return fmt.Errorf("push service not configured: FCM server key required")
	}

	message := &FCMMessage{
		To: token,
		Notification: &FCMNotification{
			Title: payload.Title,
			Body:  payload.Body,
			Icon:  payload.Icon,
			Badge: payload.Badge,
		},
		Data: payload.Data,
		WebPush: &FCMWebPush{
			FCMOptions: &FCMWebPushOptions{
				Link: payload.ClickURL,
			},
		},
	}

	return s.sendFCMMessage(ctx, message)
}

// SendToUser sends a push notification to all devices of a user
func (s *PushService) SendToUser(ctx context.Context, subscriptions []models.PushSubscription, payload *PushPayload) error {
	if !s.IsConfigured() {
		return fmt.Errorf("push service not configured")
	}

	var lastErr error
	successCount := 0

	for _, sub := range subscriptions {
		err := s.SendToToken(ctx, sub.Token, payload)
		if err != nil {
			log.Printf("[PushService] Failed to send to token %s: %v", sub.Token[:20]+"...", err)
			lastErr = err
		} else {
			successCount++
		}
	}

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("failed to send to any device: %w", lastErr)
	}

	log.Printf("[PushService] Sent notification to %d/%d devices", successCount, len(subscriptions))
	return nil
}

// sendFCMMessage sends the actual HTTP request to FCM
func (s *PushService) sendFCMMessage(ctx context.Context, message *FCMMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.FCMURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.config.ServerKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send FCM request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM returned status %d", resp.StatusCode)
	}

	var fcmResp FCMResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		return fmt.Errorf("failed to decode FCM response: %w", err)
	}

	if fcmResp.Failure > 0 && len(fcmResp.Results) > 0 {
		for _, result := range fcmResp.Results {
			if result.Error != "" {
				return fmt.Errorf("FCM error: %s", result.Error)
			}
		}
	}

	return nil
}

// NewOccurrenceNotificationPayload creates a push payload for new occurrences
func NewOccurrenceNotificationPayload(hospitalNome, setor string, tempoRestante int, occurrenceID string, dashboardURL string) *PushPayload {
	return &PushPayload{
		Title: fmt.Sprintf("Nova Ocorrencia - %s", hospitalNome),
		Body:  fmt.Sprintf("Setor: %s | Tempo restante: %d min", setor, tempoRestante),
		Icon:  "/icons/icon-192x192.png",
		Badge: "/icons/badge-72x72.png",
		Data: map[string]string{
			"type":          "new_occurrence",
			"occurrence_id": occurrenceID,
			"hospital":      hospitalNome,
			"setor":         setor,
		},
		ClickURL: fmt.Sprintf("%s/dashboard/occurrences?id=%s", dashboardURL, occurrenceID),
	}
}
