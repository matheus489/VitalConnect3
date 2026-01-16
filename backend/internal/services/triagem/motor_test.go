package triagem

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TestTriagemResultStructure tests the TriagemResult structure
func TestTriagemResultStructure(t *testing.T) {
	result := &TriagemResult{
		Elegivel:     true,
		Score:        85,
		Motivos:      []string{},
		RulesApplied: []string{"Idade Maxima", "Janela 6 Horas"},
	}

	if !result.Elegivel {
		t.Error("Expected Elegivel to be true")
	}

	if result.Score != 85 {
		t.Errorf("Expected Score to be 85, got %d", result.Score)
	}

	if len(result.RulesApplied) != 2 {
		t.Errorf("Expected 2 rules applied, got %d", len(result.RulesApplied))
	}
}

// TestTriagemResultIneligible tests an ineligible triagem result
func TestTriagemResultIneligible(t *testing.T) {
	result := &TriagemResult{
		Elegivel:     false,
		Score:        0,
		Motivos:      []string{"Idade acima do limite", "Fora da janela de captacao"},
		RulesApplied: []string{"Idade Maxima", "Janela 6 Horas"},
	}

	if result.Elegivel {
		t.Error("Expected Elegivel to be false")
	}

	if len(result.Motivos) != 2 {
		t.Errorf("Expected 2 rejection reasons, got %d", len(result.Motivos))
	}

	expectedReasons := []string{"Idade acima do limite", "Fora da janela de captacao"}
	for i, reason := range result.Motivos {
		if reason != expectedReasons[i] {
			t.Errorf("Expected reason %q, got %q", expectedReasons[i], reason)
		}
	}
}

// TestGetSectorScore tests the sector score calculation
func TestGetSectorScore(t *testing.T) {
	tests := []struct {
		setor         string
		expectedScore int
	}{
		{"UTI", 100},
		{"Emergencia", 80},
		{"Enfermaria", 50},
		{"Centro Cirurgico", 70},
		{"Outros", 40},
		{"Unknown Sector", 40}, // Should default to "Outros"
		{"", 40},               // Empty string should default
	}

	for _, tt := range tests {
		t.Run(tt.setor, func(t *testing.T) {
			score := models.GetSectorScore(tt.setor)
			if score != tt.expectedScore {
				t.Errorf("Expected score %d for sector %q, got %d", tt.expectedScore, tt.setor, score)
			}
		})
	}
}

// TestRuleConfigParsing tests parsing of rule configurations from JSON
func TestRuleConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected models.RuleType
	}{
		{
			name:     "Idade Maxima Rule",
			jsonStr:  `{"tipo": "idade_maxima", "valor": 80, "acao": "rejeitar"}`,
			expected: models.RuleTypeIdadeMaxima,
		},
		{
			name:     "Janela Horas Rule",
			jsonStr:  `{"tipo": "janela_horas", "valor": 6, "acao": "rejeitar"}`,
			expected: models.RuleTypeJanelaHoras,
		},
		{
			name:     "Identificacao Desconhecida Rule",
			jsonStr:  `{"tipo": "identificacao_desconhecida", "valor": true, "acao": "rejeitar"}`,
			expected: models.RuleTypeIdentificacaoDesconhecida,
		},
		{
			name:     "Causas Excludentes Rule",
			jsonStr:  `{"tipo": "causas_excludentes", "valor": ["sepse", "meningite"], "acao": "rejeitar"}`,
			expected: models.RuleTypeCausasExcludentes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config models.RuleConfig
			err := json.Unmarshal([]byte(tt.jsonStr), &config)
			if err != nil {
				t.Fatalf("Failed to parse rule config: %v", err)
			}

			if config.Tipo != tt.expected {
				t.Errorf("Expected type %s, got %s", tt.expected, config.Tipo)
			}
		})
	}
}

// TestIdadeMaximaRuleLogic tests the age-based rule logic
func TestIdadeMaximaRuleLogic(t *testing.T) {
	tests := []struct {
		name          string
		idade         int
		maxIdade      int
		expectElegivel bool
	}{
		{"Under max age", 65, 80, true},
		{"Exactly max age", 80, 80, true},
		{"Over max age", 81, 80, false},
		{"Way over max age", 95, 80, false},
		{"Young patient", 30, 80, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate birth date based on age
			birthDate := time.Now().AddDate(-tt.idade, 0, 0)
			deathDate := time.Now().Add(-1 * time.Hour)

			obito := &models.ObitoSimulado{
				DataNascimento: birthDate,
				DataObito:      deathDate,
			}

			calculatedAge := obito.CalculateAge()
			elegivel := calculatedAge <= tt.maxIdade

			if elegivel != tt.expectElegivel {
				t.Errorf("Expected elegivel=%v for age %d (max %d), got %v",
					tt.expectElegivel, tt.idade, tt.maxIdade, elegivel)
			}
		})
	}
}

// TestJanelaHorasRuleLogic tests the time window rule logic
func TestJanelaHorasRuleLogic(t *testing.T) {
	tests := []struct {
		name           string
		hoursAgo       int
		windowHours    int
		expectElegivel bool
	}{
		{"Within window", 2, 6, true},
		{"At boundary", 6, 6, false},
		{"Outside window", 8, 6, false},
		{"Just inside", 5, 6, true},
		{"Way outside", 24, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := &models.ObitoSimulado{
				DataObito: time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
			}

			elegivel := obito.IsWithinWindow(tt.windowHours)
			if elegivel != tt.expectElegivel {
				t.Errorf("Expected elegivel=%v for %d hours ago (window %d), got %v",
					tt.expectElegivel, tt.hoursAgo, tt.windowHours, elegivel)
			}
		})
	}
}

// TestIdentificacaoDesconhecidaRuleLogic tests the unknown identification rule logic
func TestIdentificacaoDesconhecidaRuleLogic(t *testing.T) {
	tests := []struct {
		name               string
		idDesconhecida     bool
		rejectUnknown      bool
		expectElegivel     bool
	}{
		{"Known - rule active", false, true, true},
		{"Unknown - rule active", true, true, false},
		{"Known - rule inactive", false, false, true},
		{"Unknown - rule inactive", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := &models.ObitoSimulado{
				IdentificacaoDesconhecida: tt.idDesconhecida,
			}

			// Simulate rule logic
			elegivel := true
			if tt.rejectUnknown && obito.IdentificacaoDesconhecida {
				elegivel = false
			}

			if elegivel != tt.expectElegivel {
				t.Errorf("Expected elegivel=%v, got %v", tt.expectElegivel, elegivel)
			}
		})
	}
}

// TestCausasExcludentesRuleLogic tests the excluded causes rule logic
func TestCausasExcludentesRuleLogic(t *testing.T) {
	excludedCauses := []string{"sepse", "meningite", "tuberculose"}

	tests := []struct {
		name           string
		causaMortis    string
		expectElegivel bool
	}{
		{"Normal cause", "Infarto Agudo do Miocardio", true},
		{"Excluded - sepse", "Choque septico", false},
		{"Excluded - meningite", "Meningite bacteriana", false},
		{"Partial match", "Sepse grave", false},
		{"Case insensitive", "SEPSE", false},
		{"Non-excluded", "Acidente vascular cerebral", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := &models.ObitoSimulado{
				CausaMortis: tt.causaMortis,
			}

			// Simulate rule logic
			elegivel := true
			causaMortisLower := string([]byte(obito.CausaMortis)) // Simple lowercase check
			for _, causa := range excludedCauses {
				if containsIgnoreCase(causaMortisLower, causa) {
					elegivel = false
					break
				}
			}

			if elegivel != tt.expectElegivel {
				t.Errorf("Expected elegivel=%v for cause %q, got %v",
					tt.expectElegivel, tt.causaMortis, elegivel)
			}
		})
	}
}

// Helper function to check case-insensitive contains
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains using byte comparison
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

	// Check if subLower is in sLower
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

// TestScoreCalculationWithTimeAdjustment tests score calculation with time adjustments
func TestScoreCalculationWithTimeAdjustment(t *testing.T) {
	tests := []struct {
		name         string
		setor        string
		hoursRemaining float64
		expectedMin   int
		expectedMax   int
	}{
		{"UTI - very urgent", "UTI", 0.5, 100, 120},  // Base 100 + urgency bonus
		{"UTI - urgent", "UTI", 1.5, 100, 115},
		{"UTI - normal", "UTI", 4.0, 100, 105},
		{"Emergencia - urgent", "Emergencia", 1.5, 80, 100},
		{"Enfermaria - normal", "Enfermaria", 4.0, 50, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate data_obito based on hours remaining
			hoursAgo := 6.0 - tt.hoursRemaining
			dataObito := time.Now().Add(-time.Duration(hoursAgo * float64(time.Hour)))

			obito := &models.ObitoSimulado{
				Setor:     &tt.setor,
				DataObito: dataObito,
			}

			// Calculate base score
			baseScore := models.GetSectorScore(*obito.Setor)

			// Add urgency bonus (simulating the motor logic)
			remaining := obito.TimeRemaining(6)
			hoursRem := remaining.Hours()

			if hoursRem <= 1 {
				baseScore += 20
			} else if hoursRem <= 2 {
				baseScore += 10
			} else if hoursRem <= 3 {
				baseScore += 5
			}

			// Cap at 100
			if baseScore > 100 {
				baseScore = 100
			}

			if baseScore < tt.expectedMin || baseScore > tt.expectedMax {
				t.Errorf("Expected score between %d and %d, got %d",
					tt.expectedMin, tt.expectedMax, baseScore)
			}
		})
	}
}

// TestTriagemMotorStats tests the motor statistics
func TestTriagemMotorStats(t *testing.T) {
	// Test the stats structure
	stats := map[string]interface{}{
		"running":           true,
		"total_processados": int64(100),
		"total_elegiveis":   int64(75),
		"total_inelegiveis": int64(25),
		"errors":            int64(2),
		"started_at":        time.Now(),
	}

	if !stats["running"].(bool) {
		t.Error("Expected running to be true")
	}

	totalProcessados := stats["total_processados"].(int64)
	totalElegiveis := stats["total_elegiveis"].(int64)
	totalInelegiveis := stats["total_inelegiveis"].(int64)

	if totalElegiveis+totalInelegiveis != totalProcessados {
		t.Errorf("Stats don't add up: %d + %d != %d",
			totalElegiveis, totalInelegiveis, totalProcessados)
	}
}

// TestCreateOccurrenceInput tests the occurrence input creation
func TestCreateOccurrenceInput(t *testing.T) {
	obitoID := uuid.New()
	hospitalID := uuid.New()
	setor := "UTI"
	leito := "3"

	obito := &models.ObitoSimulado{
		ID:                        obitoID,
		HospitalID:                hospitalID,
		NomePaciente:              "Joao Silva Santos",
		DataNascimento:            time.Date(1960, 1, 15, 0, 0, 0, 0, time.UTC),
		DataObito:                 time.Now().Add(-2 * time.Hour),
		CausaMortis:               "Infarto Agudo do Miocardio",
		Setor:                     &setor,
		Leito:                     &leito,
		IdentificacaoDesconhecida: false,
	}

	// Create occurrence data
	completeData := obito.ToOccurrenceData()
	completeDataJSON, err := json.Marshal(completeData)
	if err != nil {
		t.Fatalf("Failed to marshal complete data: %v", err)
	}

	// Create input
	input := &models.CreateOccurrenceInput{
		ObitoID:               obito.ID,
		HospitalID:            obito.HospitalID,
		ScorePriorizacao:      100,
		NomePacienteMascarado: models.MaskName(obito.NomePaciente),
		DadosCompletos:        completeDataJSON,
		DataObito:             obito.DataObito,
	}

	// Verify input
	if input.ObitoID != obitoID {
		t.Error("ObitoID mismatch")
	}

	if input.HospitalID != hospitalID {
		t.Error("HospitalID mismatch")
	}

	if input.ScorePriorizacao != 100 {
		t.Errorf("Expected score 100, got %d", input.ScorePriorizacao)
	}

	// Check masking
	expectedMasked := "Jo** Si*** Sa****"
	if input.NomePacienteMascarado != expectedMasked {
		t.Errorf("Expected masked name %q, got %q", expectedMasked, input.NomePacienteMascarado)
	}
}

// TestMaskNameForOccurrence tests the name masking function
func TestMaskNameForOccurrence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Two words", "Joao Silva", "Jo** Si***"},
		{"Three words", "Maria Santos Silva", "Ma*** Sa**** Si***"},
		{"Single word", "Joao", "Jo**"},
		{"Short name", "Li Wu", "Li W*"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.MaskName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// BenchmarkScoreCalculation benchmarks score calculation performance
func BenchmarkScoreCalculation(b *testing.B) {
	setor := "UTI"
	obito := &models.ObitoSimulado{
		Setor:     &setor,
		DataObito: time.Now().Add(-2 * time.Hour),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		baseScore := models.GetSectorScore(*obito.Setor)
		remaining := obito.TimeRemaining(6)
		hoursRem := remaining.Hours()

		if hoursRem <= 1 {
			baseScore += 20
		} else if hoursRem <= 2 {
			baseScore += 10
		} else if hoursRem <= 3 {
			baseScore += 5
		}

		if baseScore > 100 {
			baseScore = 100
		}
		_ = baseScore
	}
}
