package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/services/auth"
)

// E2E Test Suite for SIDOT
// These tests verify complete user flows from start to finish

// =============================================================================
// Test 1: Complete Authentication Flow
// =============================================================================

func TestE2E_AuthenticationFlow(t *testing.T) {
	// Test data
	testEmail := "e2e_test@sidot.gov.br"
	testPassword := "securePassword123"

	// Step 1: Hash password (simulating user creation)
	passwordHash, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Step 2: Verify password works
	if err := auth.CheckPasswordHash(testPassword, passwordHash); err != nil {
		t.Fatalf("Password verification failed: %v", err)
	}

	// Step 3: Create JWT service
	jwtService, err := auth.NewJWTService(
		"test-access-secret-key-32-chars!",
		"test-refresh-secret-key-32chars!",
		15*time.Minute,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Step 4: Generate tokens
	userID := uuid.New().String()
	accessToken, refreshToken, err := jwtService.GenerateTokenPair(userID, testEmail, "operador", "")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Step 5: Validate access token
	claims, err := jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("Access token validation failed: %v", err)
	}

	if claims.Email != testEmail {
		t.Errorf("Email mismatch: got %s, expected %s", claims.Email, testEmail)
	}

	if claims.Role != "operador" {
		t.Errorf("Role mismatch: got %s, expected operador", claims.Role)
	}

	// Step 6: Validate refresh token
	refreshClaims, err := jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("Refresh token validation failed: %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("UserID mismatch in refresh token")
	}

	// Step 7: Verify tokens are different types
	if _, err := jwtService.ValidateRefreshToken(accessToken); err == nil {
		t.Error("Access token should not be valid as refresh token")
	}

	t.Log("E2E Authentication flow completed successfully")
}

// =============================================================================
// Test 2: Complete Status Transition Flow
// =============================================================================

func TestE2E_StatusTransitionFlow(t *testing.T) {
	// Simulating the complete lifecycle of an occurrence

	// Start state: PENDENTE
	currentStatus := models.StatusPendente

	// Define expected flow:
	// PENDENTE -> EM_ANDAMENTO -> ACEITA -> CONCLUIDA
	expectedFlow := []models.OccurrenceStatus{
		models.StatusEmAndamento,
		models.StatusAceita,
		models.StatusConcluida,
	}

	for _, nextStatus := range expectedFlow {
		// Verify transition is valid
		if !currentStatus.CanTransitionTo(nextStatus) {
			t.Errorf("Transition from %s to %s should be valid", currentStatus, nextStatus)
		}

		// Update current status
		previousStatus := currentStatus
		currentStatus = nextStatus

		t.Logf("Transitioned: %s -> %s", previousStatus, currentStatus)
	}

	// Verify final state
	if currentStatus != models.StatusConcluida {
		t.Errorf("Expected final status CONCLUIDA, got %s", currentStatus)
	}

	// Verify invalid transitions are blocked
	if models.StatusConcluida.CanTransitionTo(models.StatusPendente) {
		t.Error("CONCLUIDA should not be able to transition to PENDENTE")
	}

	// Test CANCELADA path from any state
	testCancelableStates := []models.OccurrenceStatus{
		models.StatusPendente,
		models.StatusEmAndamento,
		models.StatusAceita,
	}

	for _, state := range testCancelableStates {
		if !state.CanTransitionTo(models.StatusCancelada) {
			t.Errorf("%s should be able to transition to CANCELADA", state)
		}
	}

	t.Log("E2E Status transition flow completed successfully")
}

// =============================================================================
// Test 3: Triagem Eligibility Flow
// =============================================================================

func TestE2E_TriagemEligibilityFlow(t *testing.T) {
	// Test complete triagem process for different scenarios

	type TestCase struct {
		name           string
		idade          int
		hoursAgo       int
		causaMortis    string
		idDesconhecida bool
		expectElegivel bool
	}

	cases := []TestCase{
		{
			name:           "Perfect candidate",
			idade:          60,
			hoursAgo:       2,
			causaMortis:    "Infarto Agudo do Miocardio",
			idDesconhecida: false,
			expectElegivel: true,
		},
		{
			name:           "Too old",
			idade:          85,
			hoursAgo:       2,
			causaMortis:    "Infarto Agudo do Miocardio",
			idDesconhecida: false,
			expectElegivel: false,
		},
		{
			name:           "Outside time window",
			idade:          60,
			hoursAgo:       8,
			causaMortis:    "Infarto Agudo do Miocardio",
			idDesconhecida: false,
			expectElegivel: false,
		},
		{
			name:           "Excluded cause - sepse",
			idade:          60,
			hoursAgo:       2,
			causaMortis:    "Choque Septico",
			idDesconhecida: false,
			expectElegivel: false,
		},
		{
			name:           "Unknown identification",
			idade:          60,
			hoursAgo:       2,
			causaMortis:    "Parada Cardiaca",
			idDesconhecida: true,
			expectElegivel: false,
		},
		{
			name:           "At age limit",
			idade:          80,
			hoursAgo:       2,
			causaMortis:    "AVC",
			idDesconhecida: false,
			expectElegivel: true, // 80 is the limit (<=80 is OK)
		},
	}

	excludedCauses := []string{"sepse", "meningite", "tuberculose"}
	maxAge := 80
	windowHours := 6

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create obito
			dataNascimento := time.Now().AddDate(-tc.idade, 0, 0)
			dataObito := time.Now().Add(-time.Duration(tc.hoursAgo) * time.Hour)

			obito := &models.ObitoSimulado{
				ID:                        uuid.New(),
				NomePaciente:              "Paciente Teste",
				DataNascimento:            dataNascimento,
				DataObito:                 dataObito,
				CausaMortis:               tc.causaMortis,
				IdentificacaoDesconhecida: tc.idDesconhecida,
			}

			// Apply rules
			elegivel := true
			motivos := []string{}

			// Rule 1: Age check
			if obito.CalculateAge() > maxAge {
				elegivel = false
				motivos = append(motivos, "Idade acima do limite")
			}

			// Rule 2: Time window check
			if !obito.IsWithinWindow(windowHours) {
				elegivel = false
				motivos = append(motivos, "Fora da janela de captacao")
			}

			// Rule 3: Excluded causes
			for _, causa := range excludedCauses {
				if containsIgnoreCase(obito.CausaMortis, causa) {
					elegivel = false
					motivos = append(motivos, "Causa de obito excludente")
					break
				}
			}

			// Rule 4: Unknown identification
			if obito.IdentificacaoDesconhecida {
				elegivel = false
				motivos = append(motivos, "Identificacao desconhecida")
			}

			// Verify result
			if elegivel != tc.expectElegivel {
				t.Errorf("Expected elegivel=%v, got %v. Motivos: %v",
					tc.expectElegivel, elegivel, motivos)
			}

			if elegivel {
				t.Logf("ELEGIVEL - Age: %d, Window: %dh ago", obito.CalculateAge(), tc.hoursAgo)
			} else {
				t.Logf("INELEGIVEL - Motivos: %v", motivos)
			}
		})
	}

	t.Log("E2E Triagem eligibility flow completed successfully")
}

// Helper function to check case-insensitive contains
func containsIgnoreCase(s, substr string) bool {
	sLower := []byte(s)
	subLower := []byte(substr)

	for i := range sLower {
		if sLower[i] >= 'A' && sLower[i] <= 'Z' {
			sLower[i] = sLower[i] + 32
		}
	}

	for i := range subLower {
		if subLower[i] >= 'A' && subLower[i] <= 'Z' {
			subLower[i] = subLower[i] + 32
		}
	}

	for i := 0; i <= len(sLower)-len(subLower); i++ {
		match := true
		for j := 0; j < len(subLower); j++ {
			if sLower[i+j] != subLower[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

// =============================================================================
// Test 4: Occurrence Creation and Notification Flow
// =============================================================================

func TestE2E_OccurrenceNotificationFlow(t *testing.T) {
	// Simulate the complete flow from obito detection to notification

	// Step 1: Create eligible obito
	hospitalID := uuid.New()
	obitoID := uuid.New()
	setor := "UTI"
	leito := "3"

	obito := &models.ObitoSimulado{
		ID:             obitoID,
		HospitalID:     hospitalID,
		NomePaciente:   "Joao Carlos Silva",
		DataNascimento: time.Date(1960, 3, 15, 0, 0, 0, 0, time.UTC),
		DataObito:      time.Now().Add(-1 * time.Hour),
		CausaMortis:    "Infarto Agudo do Miocardio",
		Setor:          &setor,
		Leito:          &leito,
	}

	// Step 2: Calculate score
	baseScore := models.GetSectorScore(*obito.Setor)
	if baseScore != 100 {
		t.Errorf("Expected UTI score 100, got %d", baseScore)
	}

	// Add urgency bonus based on time remaining
	remaining := obito.TimeRemaining(6)
	hoursRemaining := remaining.Hours()

	finalScore := baseScore
	if hoursRemaining <= 1 {
		finalScore += 20
	} else if hoursRemaining <= 2 {
		finalScore += 10
	} else if hoursRemaining <= 3 {
		finalScore += 5
	}

	// Cap at 100
	if finalScore > 100 {
		finalScore = 100
	}

	t.Logf("Score calculation: base=%d, final=%d (hours remaining: %.1f)",
		baseScore, finalScore, hoursRemaining)

	// Step 3: Create occurrence with masked name
	maskedName := models.MaskName(obito.NomePaciente)
	expectedMasked := "Jo** Ca**** Si***"
	if maskedName != expectedMasked {
		t.Errorf("Name masking failed: got %s, expected %s", maskedName, expectedMasked)
	}

	// Step 4: Create occurrence data
	completeData := obito.ToOccurrenceData()

	if completeData["nome_paciente"] != obito.NomePaciente {
		t.Error("Complete data should contain unmasked name")
	}

	if completeData["setor"] != "UTI" {
		t.Error("Complete data should contain setor")
	}

	// Step 5: Marshal complete data as JSON
	completeDataJSON, err := json.Marshal(completeData)
	if err != nil {
		t.Fatalf("Failed to marshal complete data: %v", err)
	}

	// Step 6: Create occurrence input
	input := &models.CreateOccurrenceInput{
		ObitoID:               obitoID,
		HospitalID:            hospitalID,
		ScorePriorizacao:      finalScore,
		NomePacienteMascarado: maskedName,
		DadosCompletos:        completeDataJSON,
		DataObito:             obito.DataObito,
	}

	// Verify input
	if input.ScorePriorizacao != finalScore {
		t.Errorf("Score mismatch in input")
	}

	// Step 7: Simulate SSE event creation
	occurrence := &models.Occurrence{
		ID:                    uuid.New(),
		ObitoID:               input.ObitoID,
		HospitalID:            input.HospitalID,
		Status:                models.StatusPendente,
		ScorePriorizacao:      input.ScorePriorizacao,
		NomePacienteMascarado: input.NomePacienteMascarado,
		DadosCompletos:        input.DadosCompletos,
		DataObito:             input.DataObito,
		JanelaExpiraEm:        input.DataObito.Add(6 * time.Hour),
		CreatedAt:             time.Now(),
	}

	// Step 8: Create SSE event
	hospitalNome := "Hospital Geral de Goiania"
	sseEvent := models.NewOccurrenceSSEEvent(occurrence, hospitalNome)

	if sseEvent.Type != "new_occurrence" {
		t.Errorf("Expected SSE event type 'new_occurrence', got '%s'", sseEvent.Type)
	}

	if sseEvent.HospitalNome != hospitalNome {
		t.Error("SSE event should contain hospital name")
	}

	if sseEvent.TempoRestante == "" || sseEvent.TempoRestante == "Expirado" {
		t.Errorf("SSE event should have valid tempo restante, got '%s'", sseEvent.TempoRestante)
	}

	t.Log("E2E Occurrence notification flow completed successfully")
}

// =============================================================================
// Test 5: Complete User Role Permission Flow
// =============================================================================

func TestE2E_UserRolePermissionFlow(t *testing.T) {
	// Test that all role permissions work correctly

	// Create users with different roles
	adminUser := &models.User{Role: models.RoleAdmin}
	gestorUser := &models.User{Role: models.RoleGestor}
	operadorUser := &models.User{Role: models.RoleOperador}

	// Test permission matrix
	tests := []struct {
		user       *models.User
		permission string
		checkFunc  func(*models.User) bool
		expected   bool
	}{
		// Admin permissions
		{adminUser, "manage_users", (*models.User).CanManageUsers, true},
		{adminUser, "manage_hospitals", (*models.User).CanManageHospitals, true},
		{adminUser, "manage_triagem_rules", (*models.User).CanManageTriagemRules, true},
		{adminUser, "view_metrics", (*models.User).CanViewMetrics, true},
		{adminUser, "operate_occurrences", (*models.User).CanOperateOccurrences, true},

		// Gestor permissions
		{gestorUser, "manage_users", (*models.User).CanManageUsers, false},
		{gestorUser, "manage_hospitals", (*models.User).CanManageHospitals, false},
		{gestorUser, "manage_triagem_rules", (*models.User).CanManageTriagemRules, true},
		{gestorUser, "view_metrics", (*models.User).CanViewMetrics, true},
		{gestorUser, "operate_occurrences", (*models.User).CanOperateOccurrences, true},

		// Operador permissions
		{operadorUser, "manage_users", (*models.User).CanManageUsers, false},
		{operadorUser, "manage_hospitals", (*models.User).CanManageHospitals, false},
		{operadorUser, "manage_triagem_rules", (*models.User).CanManageTriagemRules, false},
		{operadorUser, "view_metrics", (*models.User).CanViewMetrics, false},
		{operadorUser, "operate_occurrences", (*models.User).CanOperateOccurrences, true},
	}

	for _, tc := range tests {
		t.Run(string(tc.user.Role)+"_"+tc.permission, func(t *testing.T) {
			result := tc.checkFunc(tc.user)
			if result != tc.expected {
				t.Errorf("%s should have %s=%v, got %v",
					tc.user.Role, tc.permission, tc.expected, result)
			}
		})
	}

	t.Log("E2E User role permission flow completed successfully")
}

// =============================================================================
// Test 6: LGPD Data Masking Flow
// =============================================================================

func TestE2E_LGPDMaskingFlow(t *testing.T) {
	// Test that LGPD masking is correctly applied throughout the system

	testCases := []struct {
		fullName       string
		expectedMasked string
		description    string
	}{
		{"Joao Silva", "Jo** Si***", "Two word name"},
		{"Maria Santos Oliveira", "Ma*** Sa**** Ol******", "Three word name"},
		{"Ana", "An*", "Single name"},
		{"Li", "Li", "Very short name"},
		{"Jose Carlos da Silva Junior", "Jo** Ca**** da Si*** Ju****", "Long name with particles"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			masked := models.MaskName(tc.fullName)
			if masked != tc.expectedMasked {
				t.Errorf("Masking failed for '%s': got '%s', expected '%s'",
					tc.fullName, masked, tc.expectedMasked)
			}

			// Verify masked name doesn't contain full name
			if len(tc.fullName) > 3 && masked == tc.fullName {
				t.Error("Masked name should not equal full name")
			}

			// Verify some characters are revealed (first 2)
			if len(tc.fullName) > 2 && masked[:2] != tc.fullName[:2] {
				t.Error("First 2 characters should be preserved")
			}
		})
	}

	// Test that complete data in occurrence preserves full name
	obito := &models.ObitoSimulado{
		NomePaciente: "Joao Carlos Silva",
	}

	completeData := obito.ToOccurrenceData()
	if completeData["nome_paciente"] != obito.NomePaciente {
		t.Error("Complete data should contain full name for authorized access")
	}

	t.Log("E2E LGPD masking flow completed successfully")
}

// =============================================================================
// Test 7: Time Window Calculation Flow
// =============================================================================

func TestE2E_TimeWindowFlow(t *testing.T) {
	windowHours := 6

	// Test different scenarios
	scenarios := []struct {
		name            string
		hoursAgo        float64
		expectWithin    bool
		expectRemaining bool
	}{
		{"Just died", 0.1, true, true},
		{"1 hour ago", 1.0, true, true},
		{"5 hours ago", 5.0, true, true},
		{"5.9 hours ago", 5.9, true, true},
		{"Exactly 6 hours", 6.0, false, false},
		{"7 hours ago", 7.0, false, false},
		{"24 hours ago", 24.0, false, false},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			obito := &models.ObitoSimulado{
				DataObito: time.Now().Add(-time.Duration(s.hoursAgo * float64(time.Hour))),
			}

			isWithin := obito.IsWithinWindow(windowHours)
			if isWithin != s.expectWithin {
				t.Errorf("IsWithinWindow: expected %v, got %v", s.expectWithin, isWithin)
			}

			remaining := obito.TimeRemaining(windowHours)
			hasRemaining := remaining > 0
			if hasRemaining != s.expectRemaining {
				t.Errorf("TimeRemaining: expected hasRemaining=%v, got %v (remaining: %v)",
					s.expectRemaining, hasRemaining, remaining)
			}

			// For valid windows, verify remaining time is approximately correct
			if s.expectRemaining {
				expectedRemaining := time.Duration(float64(windowHours)-s.hoursAgo) * time.Hour
				tolerance := 5 * time.Minute

				diff := remaining - expectedRemaining
				if diff < 0 {
					diff = -diff
				}

				if diff > tolerance {
					t.Errorf("Remaining time calculation off by more than tolerance: got %v, expected ~%v",
						remaining, expectedRemaining)
				}
			}
		})
	}

	t.Log("E2E Time window flow completed successfully")
}

// =============================================================================
// Test 8: Sector Prioritization Flow
// =============================================================================

func TestE2E_SectorPrioritizationFlow(t *testing.T) {
	// Verify sector scores are correctly assigned

	expectedScores := map[string]int{
		"UTI":              100,
		"Emergencia":       80,
		"Centro Cirurgico": 70,
		"Enfermaria":       50,
		"Outros":           40,
		"Ambulatorio":      40, // Unknown should default to Outros
	}

	for setor, expectedScore := range expectedScores {
		t.Run(setor, func(t *testing.T) {
			score := models.GetSectorScore(setor)
			if score != expectedScore {
				t.Errorf("Score for %s: expected %d, got %d", setor, expectedScore, score)
			}
		})
	}

	// Verify ordering is correct (UTI > Emergencia > Centro Cirurgico > Enfermaria > Outros)
	sectors := []string{"UTI", "Emergencia", "Centro Cirurgico", "Enfermaria", "Outros"}
	for i := 0; i < len(sectors)-1; i++ {
		score1 := models.GetSectorScore(sectors[i])
		score2 := models.GetSectorScore(sectors[i+1])
		if score1 <= score2 {
			t.Errorf("Priority order violated: %s (%d) should be > %s (%d)",
				sectors[i], score1, sectors[i+1], score2)
		}
	}

	t.Log("E2E Sector prioritization flow completed successfully")
}

// =============================================================================
// Benchmark: Full Triagem Processing
// =============================================================================

func BenchmarkE2E_FullTriagemProcess(b *testing.B) {
	// Benchmark the complete triagem process

	excludedCauses := []string{"sepse", "meningite", "tuberculose"}
	maxAge := 80
	windowHours := 6

	obito := &models.ObitoSimulado{
		ID:                        uuid.New(),
		HospitalID:                uuid.New(),
		NomePaciente:              "Benchmark Paciente",
		DataNascimento:            time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC),
		DataObito:                 time.Now().Add(-2 * time.Hour),
		CausaMortis:               "Infarto Agudo",
		IdentificacaoDesconhecida: false,
	}
	setor := "UTI"
	obito.Setor = &setor

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Calculate eligibility
		elegivel := true

		if obito.CalculateAge() > maxAge {
			elegivel = false
		}

		if !obito.IsWithinWindow(windowHours) {
			elegivel = false
		}

		for _, causa := range excludedCauses {
			if containsIgnoreCase(obito.CausaMortis, causa) {
				elegivel = false
				break
			}
		}

		if obito.IdentificacaoDesconhecida {
			elegivel = false
		}

		// Calculate score
		score := models.GetSectorScore(*obito.Setor)
		remaining := obito.TimeRemaining(windowHours)
		if remaining.Hours() <= 1 {
			score += 20
		}

		// Mask name
		_ = models.MaskName(obito.NomePaciente)

		// Create occurrence data
		_ = obito.ToOccurrenceData()

		_ = elegivel
		_ = score
	}
}

// Helper function for context
func contextWithTimeout() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ctx
}
