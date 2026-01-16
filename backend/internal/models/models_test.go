package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test 1: Validacao de campos obrigatorios em User
func TestUserRoleValidation(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected bool
	}{
		{"valid operador role", RoleOperador, true},
		{"valid gestor role", RoleGestor, true},
		{"valid admin role", RoleAdmin, true},
		{"invalid empty role", UserRole(""), false},
		{"invalid unknown role", UserRole("superadmin"), false},
		{"invalid capitalized role", UserRole("ADMIN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v for role %q", result, tt.expected, tt.role)
			}
		})
	}
}

// Test 2: Testar enum de roles (operador, gestor, admin)
func TestUserRolePermissions(t *testing.T) {
	adminUser := User{Role: RoleAdmin}
	gestorUser := User{Role: RoleGestor}
	operadorUser := User{Role: RoleOperador}

	// Admin permissions
	if !adminUser.CanManageUsers() {
		t.Error("Admin should be able to manage users")
	}
	if !adminUser.CanManageHospitals() {
		t.Error("Admin should be able to manage hospitals")
	}
	if !adminUser.CanManageTriagemRules() {
		t.Error("Admin should be able to manage triagem rules")
	}

	// Gestor permissions
	if gestorUser.CanManageUsers() {
		t.Error("Gestor should not be able to manage users")
	}
	if gestorUser.CanManageHospitals() {
		t.Error("Gestor should not be able to manage hospitals")
	}
	if !gestorUser.CanManageTriagemRules() {
		t.Error("Gestor should be able to manage triagem rules")
	}
	if !gestorUser.CanViewMetrics() {
		t.Error("Gestor should be able to view metrics")
	}

	// Operador permissions
	if operadorUser.CanManageUsers() {
		t.Error("Operador should not be able to manage users")
	}
	if operadorUser.CanManageTriagemRules() {
		t.Error("Operador should not be able to manage triagem rules")
	}
	if !operadorUser.CanOperateOccurrences() {
		t.Error("Operador should be able to operate occurrences")
	}
}

// Test 3: Testar validacao de status de ocorrencia
func TestOccurrenceStatusValidation(t *testing.T) {
	tests := []struct {
		name     string
		status   OccurrenceStatus
		expected bool
	}{
		{"valid PENDENTE", StatusPendente, true},
		{"valid EM_ANDAMENTO", StatusEmAndamento, true},
		{"valid ACEITA", StatusAceita, true},
		{"valid RECUSADA", StatusRecusada, true},
		{"valid CANCELADA", StatusCancelada, true},
		{"valid CONCLUIDA", StatusConcluida, true},
		{"invalid empty status", OccurrenceStatus(""), false},
		{"invalid lowercase status", OccurrenceStatus("pendente"), false},
		{"invalid unknown status", OccurrenceStatus("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v for status %q", result, tt.expected, tt.status)
			}
		})
	}
}

// Test 4: Testar transicoes de status validas
func TestOccurrenceStatusTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     OccurrenceStatus
		to       OccurrenceStatus
		expected bool
	}{
		// Valid transitions from PENDENTE
		{"PENDENTE to EM_ANDAMENTO", StatusPendente, StatusEmAndamento, true},
		{"PENDENTE to CANCELADA", StatusPendente, StatusCancelada, true},
		{"PENDENTE to ACEITA (invalid)", StatusPendente, StatusAceita, false},

		// Valid transitions from EM_ANDAMENTO
		{"EM_ANDAMENTO to ACEITA", StatusEmAndamento, StatusAceita, true},
		{"EM_ANDAMENTO to RECUSADA", StatusEmAndamento, StatusRecusada, true},
		{"EM_ANDAMENTO to CANCELADA", StatusEmAndamento, StatusCancelada, true},
		{"EM_ANDAMENTO to PENDENTE (invalid)", StatusEmAndamento, StatusPendente, false},

		// Valid transitions from ACEITA
		{"ACEITA to CONCLUIDA", StatusAceita, StatusConcluida, true},
		{"ACEITA to CANCELADA", StatusAceita, StatusCancelada, true},
		{"ACEITA to PENDENTE (invalid)", StatusAceita, StatusPendente, false},

		// Terminal states - no transitions allowed
		{"CONCLUIDA to any (invalid)", StatusConcluida, StatusPendente, false},
		{"CANCELADA to any (invalid)", StatusCancelada, StatusPendente, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%s) from %s = %v, expected %v",
					tt.to, tt.from, result, tt.expected)
			}
		})
	}
}

// Test 5: Testar relacionamentos hospital-ocorrencia (janela de 6 horas)
func TestOccurrenceTimeWindow(t *testing.T) {
	windowHours := 6

	tests := []struct {
		name           string
		dataObito      time.Time
		expectedExpired bool
	}{
		{
			"obito 1 hora atras (dentro da janela)",
			time.Now().Add(-1 * time.Hour),
			false,
		},
		{
			"obito 5 horas atras (dentro da janela)",
			time.Now().Add(-5 * time.Hour),
			false,
		},
		{
			"obito 7 horas atras (fora da janela)",
			time.Now().Add(-7 * time.Hour),
			true,
		},
		{
			"obito 24 horas atras (fora da janela)",
			time.Now().Add(-24 * time.Hour),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := ObitoSimulado{
				DataObito: tt.dataObito,
			}

			isWithinWindow := obito.IsWithinWindow(windowHours)
			if isWithinWindow == tt.expectedExpired {
				t.Errorf("IsWithinWindow() = %v, expected %v for obito %s",
					isWithinWindow, !tt.expectedExpired, tt.name)
			}

			remaining := obito.TimeRemaining(windowHours)
			if tt.expectedExpired && remaining > 0 {
				t.Errorf("TimeRemaining() = %v, expected 0 for expired obito", remaining)
			}
			if !tt.expectedExpired && remaining <= 0 {
				t.Errorf("TimeRemaining() = %v, expected > 0 for valid obito", remaining)
			}
		})
	}
}

// Test 6: Testar mascaramento LGPD
func TestLGPDMasking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full name", "Joao Silva", "Jo** Si***"},
		{"single name", "Maria", "Ma***"},
		{"short name", "Ana", "An*"},
		{"very short name", "Li", "Li"},
		{"single char", "A", "A"},
		{"empty string", "", ""},
		{"three names", "Jose Da Silva", "Jo** Da Si***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskName(tt.input)
			if result != tt.expected {
				t.Errorf("MaskName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test adicional: Testar calculo de idade do paciente
func TestObitoCalculateAge(t *testing.T) {
	tests := []struct {
		name           string
		dataNascimento time.Time
		dataObito      time.Time
		expectedAge    int
	}{
		{
			"60 anos completos",
			time.Date(1960, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 15, 10, 0, 0, 0, time.UTC),
			60,
		},
		{
			"59 anos (aniversario ainda nao chegou)",
			time.Date(1960, 6, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 15, 10, 0, 0, 0, time.UTC),
			59,
		},
		{
			"80 anos (limite de elegibilidade)",
			time.Date(1940, 3, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 5, 20, 10, 0, 0, 0, time.UTC),
			80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obito := ObitoSimulado{
				DataNascimento: tt.dataNascimento,
				DataObito:      tt.dataObito,
			}

			age := obito.CalculateAge()
			if age != tt.expectedAge {
				t.Errorf("CalculateAge() = %d, expected %d", age, tt.expectedAge)
			}
		})
	}
}

// Test adicional: Testar sector scores
func TestSectorScores(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.setor, func(t *testing.T) {
			score := GetSectorScore(tt.setor)
			if score != tt.expectedScore {
				t.Errorf("GetSectorScore(%q) = %d, expected %d", tt.setor, score, tt.expectedScore)
			}
		})
	}
}

// Test adicional: Testar conversao de Hospital para response
func TestHospitalToResponse(t *testing.T) {
	endereco := "Rua Principal, 123"
	hospital := Hospital{
		ID:        uuid.New(),
		Nome:      "Hospital Geral de Goiania",
		Codigo:    "HGG",
		Endereco:  &endereco,
		Ativo:     true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	response := hospital.ToResponse()

	if response.ID != hospital.ID {
		t.Errorf("ID mismatch: got %v, expected %v", response.ID, hospital.ID)
	}
	if response.Nome != hospital.Nome {
		t.Errorf("Nome mismatch: got %v, expected %v", response.Nome, hospital.Nome)
	}
	if response.Codigo != hospital.Codigo {
		t.Errorf("Codigo mismatch: got %v, expected %v", response.Codigo, hospital.Codigo)
	}
	if *response.Endereco != *hospital.Endereco {
		t.Errorf("Endereco mismatch: got %v, expected %v", *response.Endereco, *hospital.Endereco)
	}
	if response.Ativo != hospital.Ativo {
		t.Errorf("Ativo mismatch: got %v, expected %v", response.Ativo, hospital.Ativo)
	}
}
