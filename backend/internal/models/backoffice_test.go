package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test 1: Test Tenant model with theme_config JSONB field
func TestTenant_ThemeConfig(t *testing.T) {
	t.Run("get default theme config when empty", func(t *testing.T) {
		tenant := &Tenant{
			ID:          uuid.New(),
			Name:        "Test Tenant",
			Slug:        "test-tenant",
			ThemeConfig: nil,
			IsActive:    true,
		}

		config, err := tenant.GetThemeConfig()
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "#2563eb", config.Theme.Colors.Primary)
		assert.Equal(t, "Inter", config.Theme.Fonts.Body)
	})

	t.Run("set and get theme config", func(t *testing.T) {
		tenant := &Tenant{
			ID:       uuid.New(),
			Name:     "Test Tenant",
			Slug:     "test-tenant",
			IsActive: true,
		}

		customConfig := ThemeConfig{
			Theme: Theme{
				Colors: ThemeColors{
					Primary:    "#ff0000",
					Secondary:  "#00ff00",
					Background: "#ffffff",
				},
				Fonts: ThemeFonts{
					Body:    "Roboto",
					Heading: "Roboto",
				},
			},
			Layout: Layout{
				Sidebar: []SidebarItem{
					{Label: "Dashboard", Icon: "Home", Link: "/dashboard"},
				},
				Topbar: TopbarConfig{ShowUserInfo: true, ShowTenantLogo: true},
				DashboardWidgets: []DashboardWidget{
					{Type: "stats_card", Visible: true, Order: 1},
				},
			},
		}

		err := tenant.SetThemeConfig(customConfig)
		assert.NoError(t, err)
		assert.NotNil(t, tenant.ThemeConfig)

		retrievedConfig, err := tenant.GetThemeConfig()
		assert.NoError(t, err)
		assert.Equal(t, "#ff0000", retrievedConfig.Theme.Colors.Primary)
		assert.Equal(t, "Roboto", retrievedConfig.Theme.Fonts.Body)
		assert.Len(t, retrievedConfig.Layout.Sidebar, 1)
		assert.Len(t, retrievedConfig.Layout.DashboardWidgets, 1)
	})

	t.Run("tenant to response includes new fields", func(t *testing.T) {
		logoURL := "https://example.com/logo.png"
		faviconURL := "https://example.com/favicon.ico"
		tenant := &Tenant{
			ID:          uuid.New(),
			Name:        "Test Tenant",
			Slug:        "test-tenant",
			ThemeConfig: json.RawMessage(`{"theme":{"colors":{"primary":"#ff0000"}}}`),
			IsActive:    true,
			LogoURL:     &logoURL,
			FaviconURL:  &faviconURL,
		}

		response := tenant.ToResponse()
		assert.Equal(t, tenant.ID, response.ID)
		assert.True(t, response.IsActive)
		assert.Equal(t, &logoURL, response.LogoURL)
		assert.Equal(t, &faviconURL, response.FaviconURL)
		assert.NotNil(t, response.ThemeConfig)
	})
}

// Test 2: Test SystemSetting model CRUD operations
func TestSystemSetting_CRUD(t *testing.T) {
	t.Run("valid system setting creation", func(t *testing.T) {
		input := &CreateSystemSettingInput{
			Key:   "smtp_config",
			Value: json.RawMessage(`{"host":"smtp.example.com","port":587}`),
		}

		err := input.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid key - too short", func(t *testing.T) {
		input := &CreateSystemSettingInput{
			Key:   "a",
			Value: json.RawMessage(`{}`),
		}

		err := input.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidSettingKey, err)
	})

	t.Run("invalid value - not JSON", func(t *testing.T) {
		input := &CreateSystemSettingInput{
			Key:   "test_key",
			Value: json.RawMessage(`not valid json`),
		}

		err := input.Validate()
		assert.Error(t, err)
	})

	t.Run("system setting to response", func(t *testing.T) {
		description := "SMTP configuration"
		setting := &SystemSetting{
			ID:          uuid.New(),
			Key:         "smtp_config",
			Value:       json.RawMessage(`{"host":"smtp.example.com"}`),
			Description: &description,
			IsEncrypted: true,
		}

		response := setting.ToResponse()
		assert.Equal(t, setting.ID, response.ID)
		assert.Equal(t, "smtp_config", response.Key)
		assert.True(t, response.IsEncrypted)
	})

	t.Run("masked response for encrypted setting", func(t *testing.T) {
		setting := &SystemSetting{
			ID:          uuid.New(),
			Key:         "smtp_password",
			Value:       json.RawMessage(`"secret_password"`),
			IsEncrypted: true,
		}

		response := setting.ToMaskedResponse()
		assert.Equal(t, "********", response.Value)
	})

	t.Run("non-masked response for non-encrypted setting", func(t *testing.T) {
		setting := &SystemSetting{
			ID:          uuid.New(),
			Key:         "app_name",
			Value:       json.RawMessage(`"SIDOT"`),
			IsEncrypted: false,
		}

		response := setting.ToMaskedResponse()
		assert.Equal(t, `"SIDOT"`, response.Value)
	})
}

// Test 3: Test TriagemRuleTemplate model CRUD operations
func TestTriagemRuleTemplate_CRUD(t *testing.T) {
	t.Run("valid template creation", func(t *testing.T) {
		input := &CreateTriagemRuleTemplateInput{
			Nome:     "Idade Maxima 75",
			Tipo:     "idade_maxima",
			Condicao: json.RawMessage(`{"valor":75,"acao":"rejeitar"}`),
		}

		err := input.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid template type", func(t *testing.T) {
		input := &CreateTriagemRuleTemplateInput{
			Nome:     "Invalid Type",
			Tipo:     "invalid_type",
			Condicao: json.RawMessage(`{"valor":75}`),
		}

		err := input.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTriagemRuleTemplateType, err)
	})

	t.Run("template type validation", func(t *testing.T) {
		validTypes := []TriagemRuleTemplateType{
			TemplateTypeIdadeMaxima,
			TemplateTypeCausasExcludentes,
			TemplateTypeJanelaHoras,
			TemplateTypeIdentificacaoDesconhecida,
			TemplateTypeSetorPriorizacao,
		}

		for _, tt := range validTypes {
			assert.True(t, tt.IsValid(), "Expected %s to be valid", tt)
		}

		invalidType := TriagemRuleTemplateType("invalid")
		assert.False(t, invalidType.IsValid())
	})

	t.Run("template to response", func(t *testing.T) {
		description := "Test template description"
		template := &TriagemRuleTemplate{
			ID:         uuid.New(),
			Nome:       "Test Template",
			Tipo:       TemplateTypeIdadeMaxima,
			Condicao:   json.RawMessage(`{"valor":75,"acao":"rejeitar"}`),
			Descricao:  &description,
			Ativo:      true,
			Prioridade: 1,
		}

		response := template.ToResponse()
		assert.Equal(t, template.ID, response.ID)
		assert.Equal(t, "Test Template", response.Nome)
		assert.Equal(t, "idade_maxima", response.Tipo)
		assert.True(t, response.Ativo)
	})
}

// Test 4: Test default theme_config structure validation
func TestThemeConfig_Validation(t *testing.T) {
	t.Run("valid theme config", func(t *testing.T) {
		config := &ThemeConfig{
			Theme: Theme{
				Colors: ThemeColors{Primary: "#2563eb"},
				Fonts:  ThemeFonts{Body: "Inter"},
			},
			Layout: Layout{
				Sidebar: []SidebarItem{
					{Label: "Home", Icon: "Home", Link: "/"},
				},
				Topbar:           TopbarConfig{ShowUserInfo: true},
				DashboardWidgets: []DashboardWidget{},
			},
		}

		err := ValidateThemeConfig(config)
		assert.NoError(t, err)
	})

	t.Run("missing primary color", func(t *testing.T) {
		config := &ThemeConfig{
			Theme: Theme{
				Colors: ThemeColors{Primary: ""},
				Fonts:  ThemeFonts{Body: "Inter"},
			},
		}

		err := ValidateThemeConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "primary color")
	})

	t.Run("missing body font", func(t *testing.T) {
		config := &ThemeConfig{
			Theme: Theme{
				Colors: ThemeColors{Primary: "#2563eb"},
				Fonts:  ThemeFonts{Body: ""},
			},
		}

		err := ValidateThemeConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "body font")
	})

	t.Run("default theme config is valid", func(t *testing.T) {
		config := DefaultThemeConfig()
		err := ValidateThemeConfig(&config)
		assert.NoError(t, err)
	})
}

// Test 5: Test TenantWithMetrics response
func TestTenantWithMetrics_Response(t *testing.T) {
	t.Run("tenant with metrics to response", func(t *testing.T) {
		tenantWithMetrics := &TenantWithMetrics{
			Tenant: Tenant{
				ID:       uuid.New(),
				Name:     "Test Tenant",
				Slug:     "test-tenant",
				IsActive: true,
			},
			UserCount:       10,
			HospitalCount:   5,
			OccurrenceCount: 100,
		}

		response := tenantWithMetrics.ToWithMetricsResponse()
		assert.Equal(t, "Test Tenant", response.Name)
		assert.Equal(t, 10, response.UserCount)
		assert.Equal(t, 5, response.HospitalCount)
		assert.Equal(t, 100, response.OccurrenceCount)
	})
}

// Test 6: Test CloneTriagemRuleTemplateInput validation
func TestCloneTriagemRuleTemplate_Validation(t *testing.T) {
	t.Run("valid clone input", func(t *testing.T) {
		input := &CloneTriagemRuleTemplateInput{
			TenantIDs: []uuid.UUID{uuid.New(), uuid.New()},
		}

		err := input.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty tenant IDs", func(t *testing.T) {
		input := &CloneTriagemRuleTemplateInput{
			TenantIDs: []uuid.UUID{},
		}

		err := input.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one tenant_id")
	})

	t.Run("duplicate tenant IDs", func(t *testing.T) {
		duplicateID := uuid.New()
		input := &CloneTriagemRuleTemplateInput{
			TenantIDs: []uuid.UUID{duplicateID, duplicateID},
		}

		err := input.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})

	t.Run("template to triagem rule conversion", func(t *testing.T) {
		template := &TriagemRuleTemplate{
			ID:         uuid.New(),
			Nome:       "Test Template",
			Tipo:       TemplateTypeIdadeMaxima,
			Condicao:   json.RawMessage(`{"valor":75,"acao":"rejeitar"}`),
			Ativo:      true,
			Prioridade: 1,
		}

		tenantID := uuid.New()
		rule := template.ToTriagemRule(tenantID)

		assert.NotEqual(t, template.ID, rule.ID) // New ID for the rule
		assert.Equal(t, template.Nome, rule.Nome)
		assert.Equal(t, template.Ativo, rule.Ativo)
		assert.Equal(t, template.Prioridade, rule.Prioridade)
	})
}
