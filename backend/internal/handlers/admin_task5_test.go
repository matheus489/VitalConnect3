package handlers

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/services"
)

// TestTriagemRuleTemplateInputValidation tests the CreateTriagemRuleTemplateInput validation
func TestTriagemRuleTemplateInputValidation(t *testing.T) {
	validCondicao := json.RawMessage(`{"tipo":"idade_maxima","valor":80,"acao":"rejeitar"}`)

	tests := []struct {
		name    string
		input   models.CreateTriagemRuleTemplateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid input - idade_maxima",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "Idade Maxima Test",
				Tipo:     "idade_maxima",
				Condicao: validCondicao,
			},
			wantErr: false,
		},
		{
			name: "Valid input - causas_excludentes",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "Causas Excludentes Test",
				Tipo:     "causas_excludentes",
				Condicao: json.RawMessage(`{"tipo":"causas_excludentes","valor":["HIV"],"acao":"rejeitar"}`),
			},
			wantErr: false,
		},
		{
			name: "Empty name - invalid",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "",
				Tipo:     "idade_maxima",
				Condicao: validCondicao,
			},
			wantErr: true,
		},
		{
			name: "Short name - invalid",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "X",
				Tipo:     "idade_maxima",
				Condicao: validCondicao,
			},
			wantErr: true,
		},
		{
			name: "Invalid tipo",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "Test Template",
				Tipo:     "invalid_type",
				Condicao: validCondicao,
			},
			wantErr: true,
		},
		{
			name: "Empty condicao - invalid",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "Test Template",
				Tipo:     "idade_maxima",
				Condicao: nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid JSON condicao",
			input: models.CreateTriagemRuleTemplateInput{
				Nome:     "Test Template",
				Tipo:     "idade_maxima",
				Condicao: json.RawMessage(`invalid json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUpdateTriagemRuleTemplateInputValidation tests partial update validation
func TestUpdateTriagemRuleTemplateInputValidation(t *testing.T) {
	validNome := "Updated Template"
	shortNome := "X"
	validTipo := "janela_horas"
	invalidTipo := "invalid_type"

	tests := []struct {
		name    string
		input   models.UpdateTriagemRuleTemplateInput
		wantErr bool
	}{
		{
			name:    "Empty update - valid",
			input:   models.UpdateTriagemRuleTemplateInput{},
			wantErr: false,
		},
		{
			name: "Valid nome update",
			input: models.UpdateTriagemRuleTemplateInput{
				Nome: &validNome,
			},
			wantErr: false,
		},
		{
			name: "Short nome - invalid",
			input: models.UpdateTriagemRuleTemplateInput{
				Nome: &shortNome,
			},
			wantErr: true,
		},
		{
			name: "Valid tipo update",
			input: models.UpdateTriagemRuleTemplateInput{
				Tipo: &validTipo,
			},
			wantErr: false,
		},
		{
			name: "Invalid tipo update",
			input: models.UpdateTriagemRuleTemplateInput{
				Tipo: &invalidTipo,
			},
			wantErr: true,
		},
		{
			name: "Valid condicao update",
			input: models.UpdateTriagemRuleTemplateInput{
				Condicao: json.RawMessage(`{"tipo":"test","valor":10}`),
			},
			wantErr: false,
		},
		{
			name: "Invalid condicao update",
			input: models.UpdateTriagemRuleTemplateInput{
				Condicao: json.RawMessage(`not valid json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCloneTriagemRuleTemplateInputValidation tests clone input validation
func TestCloneTriagemRuleTemplateInputValidation(t *testing.T) {
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()

	tests := []struct {
		name    string
		input   models.CloneTriagemRuleTemplateInput
		wantErr bool
	}{
		{
			name: "Valid - single tenant",
			input: models.CloneTriagemRuleTemplateInput{
				TenantIDs: []uuid.UUID{tenantID1},
			},
			wantErr: false,
		},
		{
			name: "Valid - multiple tenants",
			input: models.CloneTriagemRuleTemplateInput{
				TenantIDs: []uuid.UUID{tenantID1, tenantID2},
			},
			wantErr: false,
		},
		{
			name: "Empty tenant list - invalid",
			input: models.CloneTriagemRuleTemplateInput{
				TenantIDs: []uuid.UUID{},
			},
			wantErr: true,
		},
		{
			name: "Duplicate tenant IDs - invalid",
			input: models.CloneTriagemRuleTemplateInput{
				TenantIDs: []uuid.UUID{tenantID1, tenantID1},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSystemSettingInputValidation tests system setting input validation
func TestSystemSettingInputValidation(t *testing.T) {
	validValue := json.RawMessage(`{"host":"smtp.example.com","port":587}`)

	tests := []struct {
		name    string
		input   models.CreateSystemSettingInput
		wantErr bool
	}{
		{
			name: "Valid input",
			input: models.CreateSystemSettingInput{
				Key:   "smtp_config",
				Value: validValue,
			},
			wantErr: false,
		},
		{
			name: "Valid input with encryption",
			input: models.CreateSystemSettingInput{
				Key:         "twilio_config",
				Value:       json.RawMessage(`{"account_sid":"xxx","auth_token":"yyy"}`),
				IsEncrypted: true,
			},
			wantErr: false,
		},
		{
			name: "Empty key - invalid",
			input: models.CreateSystemSettingInput{
				Key:   "",
				Value: validValue,
			},
			wantErr: true,
		},
		{
			name: "Short key - invalid",
			input: models.CreateSystemSettingInput{
				Key:   "x",
				Value: validValue,
			},
			wantErr: true,
		},
		{
			name: "Empty value - invalid",
			input: models.CreateSystemSettingInput{
				Key:   "test_key",
				Value: nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid JSON value",
			input: models.CreateSystemSettingInput{
				Key:   "test_key",
				Value: json.RawMessage(`not valid json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEncryptionService tests the encryption service
func TestEncryptionService(t *testing.T) {
	// Generate a test key (32 bytes)
	testKey := make([]byte, 32)
	for i := range testKey {
		testKey[i] = byte(i)
	}

	t.Run("EncryptDecrypt - success", func(t *testing.T) {
		svc, err := services.NewEncryptionServiceWithKey(testKey)
		if err != nil {
			t.Fatalf("Failed to create encryption service: %v", err)
		}

		plaintext := "This is a secret message"

		encrypted, err := svc.EncryptValue(plaintext)
		if err != nil {
			t.Fatalf("EncryptValue failed: %v", err)
		}

		if encrypted == plaintext {
			t.Error("Encrypted value should not equal plaintext")
		}

		decrypted, err := svc.DecryptValue(encrypted)
		if err != nil {
			t.Fatalf("DecryptValue failed: %v", err)
		}

		if decrypted != plaintext {
			t.Errorf("Decrypted value = %s, expected %s", decrypted, plaintext)
		}
	})

	t.Run("EncryptDecrypt - JSON data", func(t *testing.T) {
		svc, err := services.NewEncryptionServiceWithKey(testKey)
		if err != nil {
			t.Fatalf("Failed to create encryption service: %v", err)
		}

		config := map[string]interface{}{
			"host":     "smtp.example.com",
			"port":     587,
			"password": "super_secret_password",
		}

		configJSON, _ := json.Marshal(config)
		plaintext := string(configJSON)

		encrypted, err := svc.EncryptValue(plaintext)
		if err != nil {
			t.Fatalf("EncryptValue failed: %v", err)
		}

		decrypted, err := svc.DecryptValue(encrypted)
		if err != nil {
			t.Fatalf("DecryptValue failed: %v", err)
		}

		if decrypted != plaintext {
			t.Error("Decrypted JSON should match original")
		}

		// Verify we can unmarshal the decrypted value
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(decrypted), &result); err != nil {
			t.Errorf("Failed to unmarshal decrypted JSON: %v", err)
		}
	})

	t.Run("Invalid key length", func(t *testing.T) {
		shortKey := make([]byte, 16) // 16 bytes instead of 32
		_, err := services.NewEncryptionServiceWithKey(shortKey)
		if err == nil {
			t.Error("Expected error for invalid key length")
		}
	})

	t.Run("Decryption with wrong key fails", func(t *testing.T) {
		svc1, _ := services.NewEncryptionServiceWithKey(testKey)

		// Create different key
		otherKey := make([]byte, 32)
		for i := range otherKey {
			otherKey[i] = byte(i + 1)
		}
		svc2, _ := services.NewEncryptionServiceWithKey(otherKey)

		encrypted, _ := svc1.EncryptValue("secret")

		_, err := svc2.DecryptValue(encrypted)
		if err == nil {
			t.Error("Expected decryption to fail with wrong key")
		}
	})

	t.Run("GenerateRandomKey", func(t *testing.T) {
		key1, err := services.GenerateRandomKey()
		if err != nil {
			t.Fatalf("GenerateRandomKey failed: %v", err)
		}

		key2, err := services.GenerateRandomKey()
		if err != nil {
			t.Fatalf("GenerateRandomKey failed: %v", err)
		}

		if key1 == key2 {
			t.Error("Generated keys should be unique")
		}

		// Verify key is valid base64 and correct length
		decoded, err := base64.StdEncoding.DecodeString(key1)
		if err != nil {
			t.Errorf("Generated key is not valid base64: %v", err)
		}
		if len(decoded) != 32 {
			t.Errorf("Decoded key length = %d, expected 32", len(decoded))
		}
	})
}

// TestSystemSettingMaskedResponse tests the masked response for encrypted settings
func TestSystemSettingMaskedResponse(t *testing.T) {
	t.Run("Non-encrypted setting shows value", func(t *testing.T) {
		setting := models.SystemSetting{
			ID:          uuid.New(),
			Key:         "test_key",
			Value:       json.RawMessage(`{"host":"example.com"}`),
			IsEncrypted: false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		response := setting.ToMaskedResponse()

		if response.Value == "********" {
			t.Error("Non-encrypted value should not be masked")
		}
		if response.Value != `{"host":"example.com"}` {
			t.Errorf("Value = %s, expected raw JSON", response.Value)
		}
	})

	t.Run("Encrypted setting is masked", func(t *testing.T) {
		setting := models.SystemSetting{
			ID:          uuid.New(),
			Key:         "smtp_password",
			Value:       json.RawMessage(`"encrypted_data_here"`),
			IsEncrypted: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		response := setting.ToMaskedResponse()

		if response.Value != "********" {
			t.Errorf("Encrypted value should be masked, got %s", response.Value)
		}
		if !response.IsEncrypted {
			t.Error("IsEncrypted should be true")
		}
	})
}

// TestTriagemRuleTemplateTypes tests valid template types
func TestTriagemRuleTemplateTypes(t *testing.T) {
	validTypes := []models.TriagemRuleTemplateType{
		models.TemplateTypeIdadeMaxima,
		models.TemplateTypeCausasExcludentes,
		models.TemplateTypeJanelaHoras,
		models.TemplateTypeIdentificacaoDesconhecida,
		models.TemplateTypeSetorPriorizacao,
	}

	for _, typ := range validTypes {
		if !typ.IsValid() {
			t.Errorf("Type %s should be valid", typ)
		}
	}

	invalidType := models.TriagemRuleTemplateType("invalid_type")
	if invalidType.IsValid() {
		t.Error("invalid_type should not be valid")
	}
}

// TestTriagemRuleTemplateToResponse tests the ToResponse conversion
func TestTriagemRuleTemplateToResponse(t *testing.T) {
	descricao := "Test description"
	template := models.TriagemRuleTemplate{
		ID:         uuid.New(),
		Nome:       "Test Template",
		Tipo:       models.TemplateTypeIdadeMaxima,
		Condicao:   json.RawMessage(`{"valor":80}`),
		Descricao:  &descricao,
		Ativo:      true,
		Prioridade: 100,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	response := template.ToResponse()

	if response.ID != template.ID {
		t.Errorf("ID mismatch: got %v, expected %v", response.ID, template.ID)
	}
	if response.Nome != template.Nome {
		t.Errorf("Nome mismatch: got %s, expected %s", response.Nome, template.Nome)
	}
	if response.Tipo != string(template.Tipo) {
		t.Errorf("Tipo mismatch: got %s, expected %s", response.Tipo, template.Tipo)
	}
	if response.Ativo != template.Ativo {
		t.Errorf("Ativo mismatch: got %v, expected %v", response.Ativo, template.Ativo)
	}
	if response.Prioridade != template.Prioridade {
		t.Errorf("Prioridade mismatch: got %d, expected %d", response.Prioridade, template.Prioridade)
	}
}

// TestTriagemRuleTemplateWithUsageResponse tests the ToWithUsageResponse conversion
func TestTriagemRuleTemplateWithUsageResponse(t *testing.T) {
	template := models.TriagemRuleTemplateWithUsage{
		TriagemRuleTemplate: models.TriagemRuleTemplate{
			ID:         uuid.New(),
			Nome:       "Test Template",
			Tipo:       models.TemplateTypeIdadeMaxima,
			Condicao:   json.RawMessage(`{"valor":80}`),
			Ativo:      true,
			Prioridade: 100,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		TenantCount: 5,
	}

	response := template.ToWithUsageResponse()

	if response.TenantCount != 5 {
		t.Errorf("TenantCount mismatch: got %d, expected 5", response.TenantCount)
	}
	if response.Nome != template.Nome {
		t.Errorf("Nome mismatch: got %s, expected %s", response.Nome, template.Nome)
	}
}
