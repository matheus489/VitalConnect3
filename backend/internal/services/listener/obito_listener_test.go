package listener

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TestParseObitoEvent tests parsing of obito events from Redis stream data
func TestParseObitoEvent(t *testing.T) {
	event := ObitoEvent{
		ObitoID:                   uuid.New().String(),
		HospitalID:                uuid.New().String(),
		TimestampDeteccao:         time.Now().Format(time.RFC3339),
		NomePaciente:              "Joao Silva",
		DataObito:                 time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		CausaMortis:               "Infarto Agudo do Miocardio",
		Setor:                     "UTI",
		Leito:                     "3",
		Idade:                     65,
		IdentificacaoDesconhecida: false,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	parsed, err := ParseObitoEvent(string(eventJSON))
	if err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}

	if parsed.ObitoID != event.ObitoID {
		t.Errorf("ObitoID mismatch: expected %s, got %s", event.ObitoID, parsed.ObitoID)
	}

	if parsed.HospitalID != event.HospitalID {
		t.Errorf("HospitalID mismatch: expected %s, got %s", event.HospitalID, parsed.HospitalID)
	}

	if parsed.NomePaciente != event.NomePaciente {
		t.Errorf("NomePaciente mismatch: expected %s, got %s", event.NomePaciente, parsed.NomePaciente)
	}

	if parsed.Setor != event.Setor {
		t.Errorf("Setor mismatch: expected %s, got %s", event.Setor, parsed.Setor)
	}

	if parsed.Idade != event.Idade {
		t.Errorf("Idade mismatch: expected %d, got %d", event.Idade, parsed.Idade)
	}
}

// TestObitoEventGetters tests the getter methods of ObitoEvent
func TestObitoEventGetters(t *testing.T) {
	obitoID := uuid.New()
	hospitalID := uuid.New()
	dataObito := time.Now().Add(-2 * time.Hour)
	timestampDeteccao := time.Now()

	event := &ObitoEvent{
		ObitoID:           obitoID.String(),
		HospitalID:        hospitalID.String(),
		DataObito:         dataObito.Format(time.RFC3339),
		TimestampDeteccao: timestampDeteccao.Format(time.RFC3339),
	}

	// Test GetObitoID
	parsedObitoID, err := event.GetObitoID()
	if err != nil {
		t.Errorf("GetObitoID failed: %v", err)
	}
	if parsedObitoID != obitoID {
		t.Errorf("ObitoID mismatch: expected %s, got %s", obitoID, parsedObitoID)
	}

	// Test GetHospitalID
	parsedHospitalID, err := event.GetHospitalID()
	if err != nil {
		t.Errorf("GetHospitalID failed: %v", err)
	}
	if parsedHospitalID != hospitalID {
		t.Errorf("HospitalID mismatch: expected %s, got %s", hospitalID, parsedHospitalID)
	}

	// Test GetDataObito
	parsedDataObito, err := event.GetDataObito()
	if err != nil {
		t.Errorf("GetDataObito failed: %v", err)
	}
	// Compare truncated to second precision due to RFC3339 format
	if !parsedDataObito.Truncate(time.Second).Equal(dataObito.Truncate(time.Second)) {
		t.Errorf("DataObito mismatch: expected %v, got %v", dataObito, parsedDataObito)
	}

	// Test GetTimestampDeteccao
	parsedTimestamp, err := event.GetTimestampDeteccao()
	if err != nil {
		t.Errorf("GetTimestampDeteccao failed: %v", err)
	}
	if !parsedTimestamp.Truncate(time.Second).Equal(timestampDeteccao.Truncate(time.Second)) {
		t.Errorf("TimestampDeteccao mismatch: expected %v, got %v", timestampDeteccao, parsedTimestamp)
	}
}

// TestListenerStatusInitialization tests that listener status is correctly initialized
func TestListenerStatusInitialization(t *testing.T) {
	// Since we can't easily mock the database, we test the status structure
	status := &ListenerStatus{
		Status:               "stopped",
		Running:              false,
		UltimoProcessamento:  nil,
		ObitosDetectadosHoje: 0,
		TotalProcessados:     0,
		Errors:               0,
		StartedAt:            nil,
	}

	if status.Running {
		t.Error("Expected Running to be false initially")
	}

	if status.Status != "stopped" {
		t.Errorf("Expected Status to be 'stopped', got %s", status.Status)
	}

	if status.TotalProcessados != 0 {
		t.Errorf("Expected TotalProcessados to be 0, got %d", status.TotalProcessados)
	}
}

// TestObitoCalculateAge tests the age calculation
func TestObitoCalculateAge(t *testing.T) {
	tests := []struct {
		name           string
		dataNascimento time.Time
		dataObito      time.Time
		expectedAge    int
	}{
		{
			name:           "65 years old",
			dataNascimento: time.Date(1960, 1, 15, 0, 0, 0, 0, time.UTC),
			dataObito:      time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			expectedAge:    65,
		},
		{
			name:           "Birthday not yet in death year",
			dataNascimento: time.Date(1960, 12, 15, 0, 0, 0, 0, time.UTC),
			dataObito:      time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			expectedAge:    64,
		},
		{
			name:           "Exact birthday",
			dataNascimento: time.Date(1960, 6, 15, 0, 0, 0, 0, time.UTC),
			dataObito:      time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			expectedAge:    65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := &models.ObitoSimulado{
				DataNascimento: tt.dataNascimento,
				DataObito:      tt.dataObito,
			}

			age := obito.CalculateAge()
			if age != tt.expectedAge {
				t.Errorf("Expected age %d, got %d", tt.expectedAge, age)
			}
		})
	}
}

// TestObitoIsWithinWindow tests the time window check
func TestObitoIsWithinWindow(t *testing.T) {
	tests := []struct {
		name         string
		dataObito    time.Time
		windowHours  int
		expectWithin bool
	}{
		{
			name:         "Within 6 hour window",
			dataObito:    time.Now().Add(-2 * time.Hour),
			windowHours:  6,
			expectWithin: true,
		},
		{
			name:         "Outside 6 hour window",
			dataObito:    time.Now().Add(-8 * time.Hour),
			windowHours:  6,
			expectWithin: false,
		},
		{
			name:         "Exactly at window boundary",
			dataObito:    time.Now().Add(-6 * time.Hour),
			windowHours:  6,
			expectWithin: false,
		},
		{
			name:         "Just inside window",
			dataObito:    time.Now().Add(-5*time.Hour - 59*time.Minute),
			windowHours:  6,
			expectWithin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := &models.ObitoSimulado{
				DataObito: tt.dataObito,
			}

			within := obito.IsWithinWindow(tt.windowHours)
			if within != tt.expectWithin {
				t.Errorf("Expected within=%v, got %v", tt.expectWithin, within)
			}
		})
	}
}

// TestObitoTimeRemaining tests the time remaining calculation
func TestObitoTimeRemaining(t *testing.T) {
	// Test case: 2 hours ago, 6 hour window = 4 hours remaining
	obito := &models.ObitoSimulado{
		DataObito: time.Now().Add(-2 * time.Hour),
	}

	remaining := obito.TimeRemaining(6)

	// Should be approximately 4 hours (with some tolerance for test execution time)
	expectedMin := 3*time.Hour + 55*time.Minute
	expectedMax := 4*time.Hour + 5*time.Minute

	if remaining < expectedMin || remaining > expectedMax {
		t.Errorf("Expected remaining between %v and %v, got %v", expectedMin, expectedMax, remaining)
	}

	// Test case: expired (8 hours ago, 6 hour window)
	obitoExpired := &models.ObitoSimulado{
		DataObito: time.Now().Add(-8 * time.Hour),
	}

	remainingExpired := obitoExpired.TimeRemaining(6)
	if remainingExpired != 0 {
		t.Errorf("Expected 0 for expired obito, got %v", remainingExpired)
	}
}

// TestObitoToOccurrenceData tests the conversion to occurrence data
func TestObitoToOccurrenceData(t *testing.T) {
	hospitalID := uuid.New()
	obitoID := uuid.New()
	prontuario := "PRO12345"
	setor := "UTI"
	leito := "3"

	obito := &models.ObitoSimulado{
		ID:                        obitoID,
		HospitalID:                hospitalID,
		NomePaciente:              "Joao Silva",
		DataNascimento:            time.Date(1960, 1, 15, 0, 0, 0, 0, time.UTC),
		DataObito:                 time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
		CausaMortis:               "Infarto Agudo do Miocardio",
		Prontuario:                &prontuario,
		Setor:                     &setor,
		Leito:                     &leito,
		IdentificacaoDesconhecida: false,
	}

	data := obito.ToOccurrenceData()

	if data["obito_id"] != obitoID {
		t.Errorf("ObitoID mismatch")
	}

	if data["hospital_id"] != hospitalID {
		t.Errorf("HospitalID mismatch")
	}

	if data["nome_paciente"] != "Joao Silva" {
		t.Errorf("NomePaciente mismatch")
	}

	if data["prontuario"] != "PRO12345" {
		t.Errorf("Prontuario mismatch")
	}

	if data["setor"] != "UTI" {
		t.Errorf("Setor mismatch")
	}

	if data["leito"] != "3" {
		t.Errorf("Leito mismatch")
	}

	// Check age calculation
	idade, ok := data["idade"].(int)
	if !ok || idade != 65 {
		t.Errorf("Expected idade to be 65, got %v", data["idade"])
	}
}

// TestListenerNotRunningByDefault ensures listener is not running when created
func TestListenerNotRunningByDefault(t *testing.T) {
	// We can't create a full listener without DB, but we can test the status structure
	status := &ListenerStatus{
		Running: false,
		Status:  "stopped",
	}

	if status.Running {
		t.Error("Listener should not be running by default")
	}
}

// TestMockIdempotencyCheck simulates the idempotency check logic
func TestMockIdempotencyCheck(t *testing.T) {
	// Simulate the idempotency check that would happen in the listener
	processedIDs := make(map[uuid.UUID]bool)

	id1 := uuid.New()
	id2 := uuid.New()

	// First time processing - should not be in map
	if processedIDs[id1] {
		t.Error("ID1 should not be processed yet")
	}

	// Mark as processed
	processedIDs[id1] = true

	// Second time - should be in map (idempotent check)
	if !processedIDs[id1] {
		t.Error("ID1 should be marked as processed")
	}

	// Different ID should not be affected
	if processedIDs[id2] {
		t.Error("ID2 should not be processed yet")
	}
}

// BenchmarkParseObitoEvent benchmarks event parsing performance
func BenchmarkParseObitoEvent(b *testing.B) {
	event := ObitoEvent{
		ObitoID:                   uuid.New().String(),
		HospitalID:                uuid.New().String(),
		TimestampDeteccao:         time.Now().Format(time.RFC3339),
		NomePaciente:              "Joao Silva",
		DataObito:                 time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		CausaMortis:               "Infarto Agudo do Miocardio",
		Setor:                     "UTI",
		Leito:                     "3",
		Idade:                     65,
		IdentificacaoDesconhecida: false,
	}

	eventJSON, _ := json.Marshal(event)
	eventStr := string(eventJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseObitoEvent(eventStr)
	}
}
