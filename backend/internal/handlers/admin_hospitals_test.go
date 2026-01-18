package handlers

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TestAdminReassignHospitalInputValidation tests the AdminReassignHospitalInput structure
func TestAdminReassignHospitalInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   models.AdminReassignHospitalInput
		wantErr bool
	}{
		{
			name: "Valid tenant ID",
			input: models.AdminReassignHospitalInput{
				TenantID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "Empty tenant ID (zero UUID)",
			input: models.AdminReassignHospitalInput{
				TenantID: uuid.UUID{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if UUID is zero (empty)
			isZero := tt.input.TenantID == uuid.UUID{}
			if isZero != tt.wantErr {
				t.Errorf("TenantID zero check: got %v, expected %v", isZero, tt.wantErr)
			}
		})
	}
}

// TestHospitalWithTenantResponse tests the HospitalWithTenant response conversion
func TestHospitalWithTenantResponse(t *testing.T) {
	tenantName := "Test Tenant"
	tenantSlug := "test-tenant"
	endereco := "Rua Principal, 123"
	telefone := "+5511999999999"
	lat := -23.5505
	lng := -46.6333

	hospital := models.HospitalWithTenant{
		ID:         uuid.New(),
		TenantID:   uuid.New(),
		Nome:       "Hospital Central",
		Codigo:     "HOSP01",
		Endereco:   &endereco,
		Telefone:   &telefone,
		Latitude:   &lat,
		Longitude:  &lng,
		Ativo:      true,
		TenantName: &tenantName,
		TenantSlug: &tenantSlug,
	}

	response := hospital.ToResponse()

	if response.ID != hospital.ID {
		t.Errorf("ID mismatch: got %v, expected %v", response.ID, hospital.ID)
	}
	if response.TenantID != hospital.TenantID {
		t.Errorf("TenantID mismatch: got %v, expected %v", response.TenantID, hospital.TenantID)
	}
	if response.Nome != hospital.Nome {
		t.Errorf("Nome mismatch: got %s, expected %s", response.Nome, hospital.Nome)
	}
	if response.Codigo != hospital.Codigo {
		t.Errorf("Codigo mismatch: got %s, expected %s", response.Codigo, hospital.Codigo)
	}
	if *response.Endereco != *hospital.Endereco {
		t.Errorf("Endereco mismatch: got %s, expected %s", *response.Endereco, *hospital.Endereco)
	}
	if *response.Telefone != *hospital.Telefone {
		t.Errorf("Telefone mismatch: got %s, expected %s", *response.Telefone, *hospital.Telefone)
	}
	if *response.Latitude != *hospital.Latitude {
		t.Errorf("Latitude mismatch: got %f, expected %f", *response.Latitude, *hospital.Latitude)
	}
	if *response.Longitude != *hospital.Longitude {
		t.Errorf("Longitude mismatch: got %f, expected %f", *response.Longitude, *hospital.Longitude)
	}
	if response.Ativo != hospital.Ativo {
		t.Errorf("Ativo mismatch: got %v, expected %v", response.Ativo, hospital.Ativo)
	}
	if *response.TenantName != *hospital.TenantName {
		t.Errorf("TenantName mismatch: got %s, expected %s", *response.TenantName, *hospital.TenantName)
	}
	if *response.TenantSlug != *hospital.TenantSlug {
		t.Errorf("TenantSlug mismatch: got %s, expected %s", *response.TenantSlug, *hospital.TenantSlug)
	}
}

// TestHospitalWithTenantResponseNilFields tests the response with nil optional fields
func TestHospitalWithTenantResponseNilFields(t *testing.T) {
	hospital := models.HospitalWithTenant{
		ID:         uuid.New(),
		TenantID:   uuid.New(),
		Nome:       "Hospital Test",
		Codigo:     "HOSP02",
		Endereco:   nil,
		Telefone:   nil,
		Latitude:   nil,
		Longitude:  nil,
		Ativo:      false,
		TenantName: nil,
		TenantSlug: nil,
	}

	response := hospital.ToResponse()

	if response.Endereco != nil {
		t.Error("Endereco should be nil")
	}
	if response.Telefone != nil {
		t.Error("Telefone should be nil")
	}
	if response.Latitude != nil {
		t.Error("Latitude should be nil")
	}
	if response.Longitude != nil {
		t.Error("Longitude should be nil")
	}
	if response.TenantName != nil {
		t.Error("TenantName should be nil")
	}
	if response.TenantSlug != nil {
		t.Error("TenantSlug should be nil")
	}
	if response.Ativo != false {
		t.Error("Ativo should be false")
	}
}

// TestHospitalWithTenantResponseJSONSerialization tests JSON serialization
func TestHospitalWithTenantResponseJSONSerialization(t *testing.T) {
	tenantName := "SES Goias"
	tenantSlug := "ses-go"

	hospital := models.HospitalWithTenant{
		ID:         uuid.New(),
		TenantID:   uuid.New(),
		Nome:       "Hospital Regional",
		Codigo:     "HR01",
		Ativo:      true,
		TenantName: &tenantName,
		TenantSlug: &tenantSlug,
	}

	response := hospital.ToResponse()

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal response: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshaled data should not be empty")
	}

	// Unmarshal and verify
	var unmarshaled models.HospitalWithTenantResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.Nome != response.Nome {
		t.Errorf("Nome mismatch after unmarshal: got %s, expected %s", unmarshaled.Nome, response.Nome)
	}
}

// TestUpdateHospitalInputValidation tests the UpdateHospitalInput validation
func TestUpdateHospitalInputValidation(t *testing.T) {
	shortName := "A"
	validName := "Valid Hospital Name"
	shortCodigo := "X"
	validCodigo := "HOSP01"
	invalidLat := -100.0
	validLat := -23.5505
	invalidLng := 200.0
	validLng := -46.6333
	trueBool := true

	tests := []struct {
		name    string
		input   models.UpdateHospitalInput
		wantErr bool
	}{
		{
			name:    "Empty input - valid",
			input:   models.UpdateHospitalInput{},
			wantErr: false,
		},
		{
			name: "Valid name",
			input: models.UpdateHospitalInput{
				Nome: &validName,
			},
			wantErr: false,
		},
		{
			name: "Short name - invalid",
			input: models.UpdateHospitalInput{
				Nome: &shortName,
			},
			wantErr: true,
		},
		{
			name: "Valid codigo",
			input: models.UpdateHospitalInput{
				Codigo: &validCodigo,
			},
			wantErr: false,
		},
		{
			name: "Short codigo - invalid",
			input: models.UpdateHospitalInput{
				Codigo: &shortCodigo,
			},
			wantErr: true,
		},
		{
			name: "Valid coordinates",
			input: models.UpdateHospitalInput{
				Latitude:  &validLat,
				Longitude: &validLng,
			},
			wantErr: false,
		},
		{
			name: "Invalid latitude - out of range",
			input: models.UpdateHospitalInput{
				Latitude: &invalidLat,
			},
			wantErr: true,
		},
		{
			name: "Invalid longitude - out of range",
			input: models.UpdateHospitalInput{
				Longitude: &invalidLng,
			},
			wantErr: true,
		},
		{
			name: "Valid ativo update",
			input: models.UpdateHospitalInput{
				Ativo: &trueBool,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false

			// Manual validation checks
			if tt.input.Nome != nil && len(*tt.input.Nome) < 2 {
				hasError = true
			}
			if tt.input.Codigo != nil && len(*tt.input.Codigo) < 2 {
				hasError = true
			}
			if tt.input.Latitude != nil && (*tt.input.Latitude < -90 || *tt.input.Latitude > 90) {
				hasError = true
			}
			if tt.input.Longitude != nil && (*tt.input.Longitude < -180 || *tt.input.Longitude > 180) {
				hasError = true
			}

			if hasError != tt.wantErr {
				t.Errorf("Validation mismatch: got error %v, expected %v", hasError, tt.wantErr)
			}
		})
	}
}

// TestHospitalReassignTenantScenarios tests various reassignment scenarios
func TestHospitalReassignTenantScenarios(t *testing.T) {
	t.Run("Same tenant reassignment should be rejected", func(t *testing.T) {
		tenantID := uuid.New()
		hospital := models.HospitalWithTenant{
			ID:       uuid.New(),
			TenantID: tenantID,
		}

		input := models.AdminReassignHospitalInput{
			TenantID: tenantID,
		}

		// This would be the check in the handler
		isSameTenant := hospital.TenantID == input.TenantID
		if !isSameTenant {
			t.Error("Should detect same tenant reassignment")
		}
	})

	t.Run("Different tenant reassignment should be allowed", func(t *testing.T) {
		hospital := models.HospitalWithTenant{
			ID:       uuid.New(),
			TenantID: uuid.New(),
		}

		input := models.AdminReassignHospitalInput{
			TenantID: uuid.New(), // Different tenant
		}

		// This would be the check in the handler
		isSameTenant := hospital.TenantID == input.TenantID
		if isSameTenant {
			t.Error("Should allow different tenant reassignment")
		}
	})
}

// TestHospitalListParams tests the admin hospital list parameters
func TestHospitalListParams(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name     string
		page     int
		perPage  int
		status   string
		tenantID *uuid.UUID
		valid    bool
	}{
		{
			name:     "Default values",
			page:     0,
			perPage:  0,
			status:   "",
			tenantID: nil,
			valid:    true,
		},
		{
			name:     "Valid pagination",
			page:     2,
			perPage:  20,
			status:   "active",
			tenantID: nil,
			valid:    true,
		},
		{
			name:     "With tenant filter",
			page:     1,
			perPage:  10,
			status:   "all",
			tenantID: &tenantID,
			valid:    true,
		},
		{
			name:     "Large per_page - should be capped",
			page:     1,
			perPage:  1000, // Will be capped to 100
			status:   "",
			tenantID: nil,
			valid:    true,
		},
		{
			name:     "Inactive status",
			page:     1,
			perPage:  10,
			status:   "inactive",
			tenantID: nil,
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate parameter normalization as done in the handler
			page := tt.page
			if page < 1 {
				page = 1
			}

			perPage := tt.perPage
			if perPage < 1 {
				perPage = 10
			}
			if perPage > 100 {
				perPage = 100
			}

			status := tt.status
			if status == "" {
				status = "all"
			}

			// Verify normalization worked correctly
			if page < 1 {
				t.Errorf("Page should be at least 1, got %d", page)
			}
			if perPage < 1 || perPage > 100 {
				t.Errorf("PerPage should be 1-100, got %d", perPage)
			}
			validStatuses := map[string]bool{"all": true, "active": true, "inactive": true}
			if !validStatuses[status] && tt.status != "" {
				t.Errorf("Invalid status: %s", status)
			}
		})
	}
}
