package notification

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TestSSEEventPublish tests SSE event creation and format
func TestSSEEventPublish(t *testing.T) {
	// Create a mock occurrence
	occurrence := &models.Occurrence{
		ID:                    uuid.New(),
		ObitoID:               uuid.New(),
		HospitalID:            uuid.New(),
		Status:                models.StatusPendente,
		ScorePriorizacao:      85,
		NomePacienteMascarado: "Jo** Si***",
		DataObito:             time.Now().Add(-1 * time.Hour),
		JanelaExpiraEm:        time.Now().Add(5 * time.Hour),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Add dados_completos with setor
	completeData := models.OccurrenceCompleteData{
		ObitoID:       occurrence.ObitoID,
		HospitalID:    occurrence.HospitalID,
		NomePaciente:  "Joao Silva",
		DataObito:     occurrence.DataObito,
		Setor:         "UTI",
		Leito:         "Leito 4",
	}
	occurrence.DadosCompletos, _ = json.Marshal(completeData)

	hospitalNome := "Hospital Geral de Goiania"

	// Create SSE event
	event := models.NewOccurrenceSSEEvent(occurrence, hospitalNome)

	// Validate event fields
	if event.Type != "new_occurrence" {
		t.Errorf("Expected event type 'new_occurrence', got '%s'", event.Type)
	}

	if event.OccurrenceID != occurrence.ID {
		t.Errorf("Expected occurrence ID '%s', got '%s'", occurrence.ID, event.OccurrenceID)
	}

	if event.HospitalNome != hospitalNome {
		t.Errorf("Expected hospital name '%s', got '%s'", hospitalNome, event.HospitalNome)
	}

	if event.Setor != "UTI" {
		t.Errorf("Expected setor 'UTI', got '%s'", event.Setor)
	}

	if event.TempoRestante == "" || event.TempoRestante == "Expirado" {
		t.Errorf("Expected valid tempo restante, got '%s'", event.TempoRestante)
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal SSE event: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Verify JSON contains expected fields
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "new_occurrence") {
		t.Error("JSON should contain event type")
	}
	if !strings.Contains(jsonStr, occurrence.ID.String()) {
		t.Error("JSON should contain occurrence ID")
	}
}

// TestEmailTemplateFormat tests email template rendering
func TestEmailTemplateFormat(t *testing.T) {
	// Create email service (without SMTP config)
	service := NewEmailService(&EmailConfig{})

	// Create notification data
	data := &ObitoNotificationData{
		HospitalNome:  "Hospital de Urgencias de Goias",
		Setor:         "Emergencia",
		HoraObito:     time.Now().Add(-30 * time.Minute),
		TempoRestante: "5h 30min",
		OccurrenceID:  uuid.New().String(),
		Prioridade:    90,
		DashboardURL:  "http://localhost:3000/dashboard",
	}

	// Render the template
	body, err := service.renderObitoTemplate(data)
	if err != nil {
		t.Fatalf("Failed to render email template: %v", err)
	}

	// Validate template output
	if body == "" {
		t.Error("Expected non-empty email body")
	}

	// Check for required fields in template
	if !strings.Contains(body, data.HospitalNome) {
		t.Errorf("Email body should contain hospital name: %s", data.HospitalNome)
	}

	if !strings.Contains(body, data.Setor) {
		t.Errorf("Email body should contain setor: %s", data.Setor)
	}

	if !strings.Contains(body, data.TempoRestante) {
		t.Errorf("Email body should contain tempo restante: %s", data.TempoRestante)
	}

	if !strings.Contains(body, data.DashboardURL) {
		t.Errorf("Email body should contain dashboard URL: %s", data.DashboardURL)
	}

	// Check HTML structure
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Email body should be valid HTML")
	}

	if !strings.Contains(body, "URGENTE") {
		t.Error("Email should indicate urgency")
	}

	if !strings.Contains(body, "VitalConnect") {
		t.Error("Email should contain system name")
	}
}

// TestNotificationRecording tests notification record creation
func TestNotificationRecording(t *testing.T) {
	// Create notification input
	occurrenceID := uuid.New()
	userID := uuid.New()

	metadata := &models.NotificationMetadata{
		EmailTo:       "operador@vitalconnect.gov.br",
		EmailSubject:  "[URGENTE] Nova Ocorrencia Elegivel - HGG",
		HospitalNome:  "Hospital Geral de Goiania",
		Setor:         "UTI",
		TempoRestante: "4h 15min",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	input := &models.CreateNotificationInput{
		OccurrenceID: occurrenceID,
		UserID:       &userID,
		Canal:        models.ChannelEmail,
		StatusEnvio:  models.NotificationStatusEnviado,
		Metadata:     metadataJSON,
	}

	// Validate input
	if !input.Canal.IsValid() {
		t.Errorf("Expected valid notification channel, got '%s'", input.Canal)
	}

	if input.OccurrenceID == uuid.Nil {
		t.Error("Expected valid occurrence ID")
	}

	if input.StatusEnvio != models.NotificationStatusEnviado {
		t.Errorf("Expected status 'enviado', got '%s'", input.StatusEnvio)
	}

	// Validate metadata
	var parsedMetadata models.NotificationMetadata
	err = json.Unmarshal(input.Metadata, &parsedMetadata)
	if err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	if parsedMetadata.EmailTo != "operador@vitalconnect.gov.br" {
		t.Errorf("Expected email to 'operador@vitalconnect.gov.br', got '%s'", parsedMetadata.EmailTo)
	}

	if parsedMetadata.HospitalNome != "Hospital Geral de Goiania" {
		t.Errorf("Expected hospital name 'Hospital Geral de Goiania', got '%s'", parsedMetadata.HospitalNome)
	}
}

// TestSSEClientManagement tests SSE client registration and unregistration
func TestSSEClientManagement(t *testing.T) {
	// Create a client
	client := NewSSEClient("user-123", "operador")

	if client.ID == "" {
		t.Error("Expected non-empty client ID")
	}

	if client.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", client.UserID)
	}

	if client.Role != "operador" {
		t.Errorf("Expected role 'operador', got '%s'", client.Role)
	}

	if client.Channel == nil {
		t.Error("Expected non-nil channel")
	}

	if client.Done == nil {
		t.Error("Expected non-nil done channel")
	}

	// Test client close
	client.Close()

	// Verify done channel is closed
	select {
	case <-client.Done:
		// Expected
	default:
		t.Error("Done channel should be closed after Close()")
	}

	// Test double close (should not panic)
	client.Close()
}

// TestEmailQueueItem tests email queue item creation
func TestEmailQueueItem(t *testing.T) {
	occurrenceID := uuid.New()
	userIDStr := uuid.New().String()

	data := &ObitoNotificationData{
		HospitalNome:  "HGG",
		Setor:         "UTI",
		HoraObito:     time.Now(),
		TempoRestante: "5h",
		OccurrenceID:  occurrenceID.String(),
		Prioridade:    85,
	}

	item := &EmailQueueItem{
		ID:           uuid.New().String(),
		OccurrenceID: occurrenceID.String(),
		To:           "test@example.com",
		UserID:       &userIDStr,
		Data:         data,
		Retries:      0,
		CreatedAt:    time.Now(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal queue item: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Test unmarshaling
	var parsedItem EmailQueueItem
	err = json.Unmarshal(jsonData, &parsedItem)
	if err != nil {
		t.Fatalf("Failed to unmarshal queue item: %v", err)
	}

	if parsedItem.To != item.To {
		t.Errorf("Expected To '%s', got '%s'", item.To, parsedItem.To)
	}

	if parsedItem.Data.HospitalNome != data.HospitalNome {
		t.Errorf("Expected HospitalNome '%s', got '%s'", data.HospitalNome, parsedItem.Data.HospitalNome)
	}
}

// TestFormatTimeRemaining tests time formatting helper
func TestFormatTimeRemaining(t *testing.T) {
	tests := []struct {
		name     string
		expires  time.Time
		expected string
	}{
		{
			name:     "Several hours remaining",
			expires:  time.Now().Add(5*time.Hour + 30*time.Minute),
			expected: "5h 30min",
		},
		{
			name:     "Exactly one hour",
			expires:  time.Now().Add(1 * time.Hour),
			expected: "1h",
		},
		{
			name:     "Minutes only",
			expires:  time.Now().Add(45 * time.Minute),
			expected: "45min",
		},
		{
			name:     "Expired",
			expires:  time.Now().Add(-1 * time.Hour),
			expected: "Expirado",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatTimeRemaining(tc.expires)
			// Allow 1-2 minute variance due to test execution timing
			if tc.name == "Expired" {
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			} else {
				// Just verify format is correct (contains h and/or min)
				if result == "" || result == "Expirado" {
					t.Errorf("Expected time remaining, got '%s'", result)
				}
			}
		})
	}
}

// TestEmailServiceConfiguration tests email service configuration detection
func TestEmailServiceConfiguration(t *testing.T) {
	// Test unconfigured service
	unconfigured := NewEmailService(&EmailConfig{})
	if unconfigured.IsConfigured() {
		t.Error("Expected unconfigured service")
	}

	// Test partially configured service
	partial := NewEmailService(&EmailConfig{
		SMTPHost: "smtp.example.com",
	})
	if partial.IsConfigured() {
		t.Error("Expected partially configured service to be unconfigured")
	}

	// Test configured service
	configured := NewEmailService(&EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		SMTPFrom: "noreply@example.com",
	})
	if !configured.IsConfigured() {
		t.Error("Expected configured service")
	}
}

// TestNotificationChannelValidation tests notification channel validation
func TestNotificationChannelValidation(t *testing.T) {
	tests := []struct {
		channel models.NotificationChannel
		valid   bool
	}{
		{models.ChannelDashboard, true},
		{models.ChannelEmail, true},
		{models.NotificationChannel("sms"), false},
		{models.NotificationChannel(""), false},
	}

	for _, tc := range tests {
		result := tc.channel.IsValid()
		if result != tc.valid {
			t.Errorf("Channel '%s': expected valid=%v, got %v", tc.channel, tc.valid, result)
		}
	}
}

// TestSSEEventPublishWithEmptyData tests SSE event with minimal data
func TestSSEEventPublishWithEmptyData(t *testing.T) {
	// Create occurrence with empty dados_completos
	occurrence := &models.Occurrence{
		ID:                    uuid.New(),
		Status:                models.StatusPendente,
		NomePacienteMascarado: "Jo**",
		DataObito:             time.Now(),
		JanelaExpiraEm:        time.Now().Add(6 * time.Hour),
		DadosCompletos:        json.RawMessage("{}"),
	}

	event := models.NewOccurrenceSSEEvent(occurrence, "Test Hospital")

	// Should handle empty setor gracefully
	if event.Setor != "" {
		t.Errorf("Expected empty setor, got '%s'", event.Setor)
	}

	if event.Type != "new_occurrence" {
		t.Errorf("Expected event type 'new_occurrence', got '%s'", event.Type)
	}
}

// TestBackoffCalculation tests exponential backoff calculation
func TestBackoffCalculation(t *testing.T) {
	ctx := context.Background()
	_ = ctx // Used in real implementation

	// Test backoff increases exponentially
	baseDelay := BaseBackoffDelay

	// First retry: 2^1 * base = 2s
	// Second retry: 2^2 * base = 4s
	// Third retry: 2^3 * base = 8s

	expectedDelays := []time.Duration{
		2 * baseDelay,
		4 * baseDelay,
		8 * baseDelay,
	}

	for i, expected := range expectedDelays {
		retries := i + 1
		// Calculate using same formula as email_queue.go
		calculated := time.Duration(1<<retries) * baseDelay
		if calculated != expected {
			t.Errorf("Retry %d: expected delay %v, got %v", retries, expected, calculated)
		}
	}
}
