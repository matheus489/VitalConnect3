package notification

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test 1: SMSService IsConfigured returns false without credentials
func TestSMSService_IsConfigured_WithoutCredentials(t *testing.T) {
	svc := NewSMSService(nil)
	if svc.IsConfigured() {
		t.Error("Expected IsConfigured() to return false with nil config")
	}

	svc = NewSMSService(&SMSConfig{})
	if svc.IsConfigured() {
		t.Error("Expected IsConfigured() to return false with empty config")
	}

	svc = NewSMSService(&SMSConfig{
		AccountSID: "test",
	})
	if svc.IsConfigured() {
		t.Error("Expected IsConfigured() to return false with partial config")
	}
}

// Test 2: SMSService IsConfigured returns true with valid credentials
func TestSMSService_IsConfigured_WithCredentials(t *testing.T) {
	svc := NewSMSService(&SMSConfig{
		AccountSID:      "ACtest123",
		AuthToken:       "token123",
		FromPhoneNumber: "+15551234567",
	})
	if !svc.IsConfigured() {
		t.Error("Expected IsConfigured() to return true with complete config")
	}
}

// Test 3: BuildSMSMessage generates correct message with 160 char limit
func TestBuildSMSMessage_WithinLimit(t *testing.T) {
	data := &SMSNotificationData{
		HospitalNome:  "Hospital Test",
		Idade:         65,
		HorasRestante: 4,
		OccurrenceID:  uuid.New(),
		BaseURL:       "",
	}

	message := BuildSMSMessage(data)

	if len(message) > 160 {
		t.Errorf("Expected message length <= 160, got %d", len(message))
	}

	// Check message contains required parts
	if !containsAll(message, []string{"[SIDOT]", "ALERTA CRITICO", "Hosp:", "Idade:", "Janela:", "Acao:"}) {
		t.Error("Message missing required parts")
	}
}

// Test 4: BuildSMSMessage truncates long hospital names
func TestBuildSMSMessage_TruncatesLongHospitalName(t *testing.T) {
	data := &SMSNotificationData{
		HospitalNome:  "Hospital Municipal de Atendimento de Emergencia e Urgencia do Centro Historico de Sao Paulo",
		Idade:         75,
		HorasRestante: 2,
		OccurrenceID:  uuid.New(),
		BaseURL:       "http://localhost:3000",
	}

	message := BuildSMSMessage(data)

	if len(message) > 160 {
		t.Errorf("Expected message length <= 160 with truncation, got %d", len(message))
	}
}

// Test 5: CalculateBackoffDelay returns correct delays
func TestCalculateBackoffDelay(t *testing.T) {
	tests := []struct {
		retries  int
		expected time.Duration
	}{
		{0, 1 * time.Second},  // Base case
		{1, 1 * time.Second},  // 2^0 = 1
		{2, 2 * time.Second},  // 2^1 = 2
		{3, 4 * time.Second},  // 2^2 = 4
		{4, 8 * time.Second},  // 2^3 = 8
		{5, 16 * time.Second}, // 2^4 = 16
	}

	for _, tc := range tests {
		result := CalculateBackoffDelay(tc.retries)
		if result != tc.expected {
			t.Errorf("CalculateBackoffDelay(%d) = %v, expected %v", tc.retries, result, tc.expected)
		}
	}
}

// Test 6: MaskPhoneForLog masks correctly
func TestMaskPhoneForLog(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+5511999999999", "+5511****9999"},
		{"+15551234567", "+1555***4567"},
		{"", ""},
	}

	for _, tc := range tests {
		result := MaskPhoneForLog(tc.input)
		if result != tc.expected {
			t.Errorf("MaskPhoneForLog(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// Helper function
func containsAll(s string, substrs []string) bool {
	for _, sub := range substrs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
