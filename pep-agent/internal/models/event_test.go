package models

import (
	"testing"
	"time"
)

// Test 1: CPF Masking
func TestMaskCPF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CPF with dots and dash",
			input:    "123.456.789-10",
			expected: "***.***.***-10",
		},
		{
			name:     "CPF without formatting",
			input:    "12345678910",
			expected: "***.***.***-10",
		},
		{
			name:     "CPF with different last digits",
			input:    "98765432155",
			expected: "***.***.***-55",
		},
		{
			name:     "Empty CPF",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid CPF (too short)",
			input:    "12345",
			expected: "***.***.***-**",
		},
		{
			name:     "Invalid CPF (too long)",
			input:    "123456789101112",
			expected: "***.***.***-**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskCPF(tt.input)
			if result != tt.expected {
				t.Errorf("MaskCPF(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test 2: Name Masking for LGPD compliance
func TestMaskName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full name",
			input:    "Maria Silva Santos",
			expected: "Ma*** Si*** Sa****",
		},
		{
			name:     "Single word name",
			input:    "Antonio",
			expected: "An*****",
		},
		{
			name:     "Short name",
			input:    "Jo",
			expected: "J*",
		},
		{
			name:     "Empty name",
			input:    "",
			expected: "",
		},
		{
			name:     "Single character",
			input:    "A",
			expected: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskName(tt.input)
			if result != tt.expected {
				t.Errorf("MaskName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test 3: PEP Record to Event conversion with field mapping
func TestPEPRecordToObitoEvent(t *testing.T) {
	hospitalID := "123e4567-e89b-12d3-a456-426614174000"
	birthDate := time.Date(1950, 3, 15, 0, 0, 0, 0, time.UTC)
	deathDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	cns := "700012345678901"
	cpf := "123.456.789-10"
	setor := "UTI CARDIOLOGICA"
	leito := "UC-01"
	prontuario := "PRO-2024-00001"

	record := &PEPRecord{
		ID:                        "12345",
		NomePaciente:              "Maria da Silva Santos",
		DataObito:                 deathDate,
		CausaMortis:               "Infarto agudo do miocardio",
		DataNascimento:            &birthDate,
		CNS:                       &cns,
		CPF:                       &cpf,
		Setor:                     &setor,
		Leito:                     &leito,
		Prontuario:                &prontuario,
		IdentificacaoDesconhecida: "N",
	}

	event := record.ToObitoEvent(hospitalID)

	// Verify required fields
	if event.HospitalIDOrigem != "12345" {
		t.Errorf("HospitalIDOrigem = %q; want %q", event.HospitalIDOrigem, "12345")
	}

	if event.HospitalID != hospitalID {
		t.Errorf("HospitalID = %q; want %q", event.HospitalID, hospitalID)
	}

	if event.NomePaciente != "Maria da Silva Santos" {
		t.Errorf("NomePaciente = %q; want %q", event.NomePaciente, "Maria da Silva Santos")
	}

	if event.CausaMortis != "Infarto agudo do miocardio" {
		t.Errorf("CausaMortis = %q; want %q", event.CausaMortis, "Infarto agudo do miocardio")
	}

	// Verify age calculation
	expectedAge := 73 // 1950 to 2024, birthday passed
	if event.Idade != expectedAge {
		t.Errorf("Idade = %d; want %d", event.Idade, expectedAge)
	}

	// Verify CNS is preserved in full
	if event.CNS != cns {
		t.Errorf("CNS = %q; want %q (should be preserved)", event.CNS, cns)
	}

	// Verify CPF is masked
	expectedMaskedCPF := "***.***.***-10"
	if event.CPFMasked != expectedMaskedCPF {
		t.Errorf("CPFMasked = %q; want %q", event.CPFMasked, expectedMaskedCPF)
	}

	// Verify optional fields
	if event.Setor != setor {
		t.Errorf("Setor = %q; want %q", event.Setor, setor)
	}

	if event.Leito != leito {
		t.Errorf("Leito = %q; want %q", event.Leito, leito)
	}

	if event.Prontuario != prontuario {
		t.Errorf("Prontuario = %q; want %q", event.Prontuario, prontuario)
	}

	if event.IdentificacaoDesconhecida != false {
		t.Errorf("IdentificacaoDesconhecida = %v; want false", event.IdentificacaoDesconhecida)
	}
}

// Test 4: Unknown patient conversion
func TestPEPRecordToObitoEvent_UnknownPatient(t *testing.T) {
	hospitalID := "123e4567-e89b-12d3-a456-426614174000"
	deathDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	age := 45

	record := &PEPRecord{
		ID:                        "99999",
		NomePaciente:              "Paciente Nao Identificado",
		DataObito:                 deathDate,
		CausaMortis:               "Trauma cranioencefalico",
		Idade:                     &age, // Using age instead of birth date
		IdentificacaoDesconhecida: "S",
	}

	event := record.ToObitoEvent(hospitalID)

	// Verify unknown flag
	if !event.IdentificacaoDesconhecida {
		t.Error("IdentificacaoDesconhecida should be true for 'S'")
	}

	// Verify age from direct value
	if event.Idade != 45 {
		t.Errorf("Idade = %d; want %d", event.Idade, 45)
	}

	// Verify CNS and CPF are empty for unknown patient
	if event.CNS != "" {
		t.Errorf("CNS should be empty for unknown patient, got %q", event.CNS)
	}

	if event.CPFMasked != "" {
		t.Errorf("CPFMasked should be empty for unknown patient, got %q", event.CPFMasked)
	}
}

// Test 5: Age calculation edge cases
func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name        string
		birthDate   time.Time
		deathDate   time.Time
		expectedAge int
	}{
		{
			name:        "Birthday already passed in death year",
			birthDate:   time.Date(1950, 3, 15, 0, 0, 0, 0, time.UTC),
			deathDate:   time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC),
			expectedAge: 74,
		},
		{
			name:        "Birthday not yet in death year",
			birthDate:   time.Date(1950, 8, 15, 0, 0, 0, 0, time.UTC),
			deathDate:   time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
			expectedAge: 73,
		},
		{
			name:        "Same day birthday and death",
			birthDate:   time.Date(1950, 6, 15, 0, 0, 0, 0, time.UTC),
			deathDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			expectedAge: 74,
		},
		{
			name:        "Infant death",
			birthDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			deathDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedAge: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAge(tt.birthDate, tt.deathDate)
			if result != tt.expectedAge {
				t.Errorf("calculateAge() = %d; want %d", result, tt.expectedAge)
			}
		})
	}
}
