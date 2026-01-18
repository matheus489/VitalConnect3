package handlers

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TestCreateTenantInputValidation tests the CreateTenantInput validation
func TestCreateTenantInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   models.CreateTenantInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid input",
			input: models.CreateTenantInput{
				Name: "Test Tenant",
				Slug: "test-tenant",
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			input: models.CreateTenantInput{
				Name: "",
				Slug: "test-tenant",
			},
			wantErr: true,
		},
		{
			name: "Short name",
			input: models.CreateTenantInput{
				Name: "T",
				Slug: "test-tenant",
			},
			wantErr: true,
		},
		{
			name: "Invalid slug - uppercase",
			input: models.CreateTenantInput{
				Name: "Test Tenant",
				Slug: "Test-Tenant",
			},
			wantErr: true,
		},
		{
			name: "Invalid slug - spaces",
			input: models.CreateTenantInput{
				Name: "Test Tenant",
				Slug: "test tenant",
			},
			wantErr: true,
		},
		{
			name: "Invalid slug - special characters",
			input: models.CreateTenantInput{
				Name: "Test Tenant",
				Slug: "test@tenant",
			},
			wantErr: true,
		},
		{
			name: "Valid slug with hyphens",
			input: models.CreateTenantInput{
				Name: "SES Goias",
				Slug: "ses-go",
			},
			wantErr: false,
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

// TestUpdateTenantInputValidation tests the UpdateTenantInput validation
func TestUpdateTenantInputValidation(t *testing.T) {
	name := "Updated Name"
	shortName := "X"
	validSlug := "updated-slug"
	invalidSlug := "Invalid Slug"

	tests := []struct {
		name    string
		input   models.UpdateTenantInput
		wantErr bool
	}{
		{
			name:    "Empty update - valid",
			input:   models.UpdateTenantInput{},
			wantErr: false,
		},
		{
			name: "Valid name update",
			input: models.UpdateTenantInput{
				Name: &name,
			},
			wantErr: false,
		},
		{
			name: "Short name update - invalid",
			input: models.UpdateTenantInput{
				Name: &shortName,
			},
			wantErr: true,
		},
		{
			name: "Valid slug update",
			input: models.UpdateTenantInput{
				Slug: &validSlug,
			},
			wantErr: false,
		},
		{
			name: "Invalid slug update",
			input: models.UpdateTenantInput{
				Slug: &invalidSlug,
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

// TestThemeConfigValidation tests the theme config validation
func TestThemeConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.ThemeConfig
		wantErr bool
	}{
		{
			name:    "Nil config - invalid",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Empty primary color - invalid",
			config: &models.ThemeConfig{
				Theme: models.Theme{
					Colors: models.ThemeColors{
						Primary: "",
					},
					Fonts: models.ThemeFonts{
						Body: "Inter",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Empty body font - invalid",
			config: &models.ThemeConfig{
				Theme: models.Theme{
					Colors: models.ThemeColors{
						Primary: "#2563eb",
					},
					Fonts: models.ThemeFonts{
						Body: "",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid theme config",
			config: &models.ThemeConfig{
				Theme: models.Theme{
					Colors: models.ThemeColors{
						Primary:    "#2563eb",
						Background: "#ffffff",
					},
					Fonts: models.ThemeFonts{
						Body: "Inter",
					},
				},
				Layout: models.Layout{
					Sidebar: []models.SidebarItem{},
					Topbar: models.TopbarConfig{
						ShowUserInfo:   true,
						ShowTenantLogo: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid sidebar item - missing label",
			config: &models.ThemeConfig{
				Theme: models.Theme{
					Colors: models.ThemeColors{
						Primary: "#2563eb",
					},
					Fonts: models.ThemeFonts{
						Body: "Inter",
					},
				},
				Layout: models.Layout{
					Sidebar: []models.SidebarItem{
						{Label: "", Icon: "Home", Link: "/dashboard"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid sidebar item - missing link",
			config: &models.ThemeConfig{
				Theme: models.Theme{
					Colors: models.ThemeColors{
						Primary: "#2563eb",
					},
					Fonts: models.ThemeFonts{
						Body: "Inter",
					},
				},
				Layout: models.Layout{
					Sidebar: []models.SidebarItem{
						{Label: "Dashboard", Icon: "Home", Link: ""},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid dashboard widget - missing type",
			config: &models.ThemeConfig{
				Theme: models.Theme{
					Colors: models.ThemeColors{
						Primary: "#2563eb",
					},
					Fonts: models.ThemeFonts{
						Body: "Inter",
					},
				},
				Layout: models.Layout{
					DashboardWidgets: []models.DashboardWidget{
						{Type: "", Visible: true, Order: 1},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateThemeConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateThemeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultThemeConfig tests the default theme config generation
func TestDefaultThemeConfig(t *testing.T) {
	config := models.DefaultThemeConfig()

	// Check theme colors
	if config.Theme.Colors.Primary == "" {
		t.Error("Default primary color should not be empty")
	}
	if config.Theme.Colors.Primary != "#2563eb" {
		t.Errorf("Default primary color = %s, expected #2563eb", config.Theme.Colors.Primary)
	}

	// Check theme fonts
	if config.Theme.Fonts.Body == "" {
		t.Error("Default body font should not be empty")
	}
	if config.Theme.Fonts.Body != "Inter" {
		t.Errorf("Default body font = %s, expected Inter", config.Theme.Fonts.Body)
	}

	// Check layout defaults
	if config.Layout.Sidebar == nil {
		t.Error("Sidebar should not be nil")
	}
	if !config.Layout.Topbar.ShowUserInfo {
		t.Error("ShowUserInfo should be true by default")
	}
	if !config.Layout.Topbar.ShowTenantLogo {
		t.Error("ShowTenantLogo should be true by default")
	}

	// Verify it's valid JSON
	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal default config to JSON: %v", err)
	}
	if len(data) == 0 {
		t.Error("Marshaled config should not be empty")
	}
}

// TestTenantWithMetricsResponse tests the conversion from model to response
func TestTenantWithMetricsResponse(t *testing.T) {
	tenant := models.TenantWithMetrics{
		Tenant: models.Tenant{
			ID:       uuid.New(),
			Name:     "Test Tenant",
			Slug:     "test-tenant",
			IsActive: true,
		},
		UserCount:       10,
		HospitalCount:   5,
		OccurrenceCount: 100,
	}

	response := tenant.ToWithMetricsResponse()

	if response.ID != tenant.ID {
		t.Errorf("ID mismatch: got %v, expected %v", response.ID, tenant.ID)
	}
	if response.Name != tenant.Name {
		t.Errorf("Name mismatch: got %s, expected %s", response.Name, tenant.Name)
	}
	if response.Slug != tenant.Slug {
		t.Errorf("Slug mismatch: got %s, expected %s", response.Slug, tenant.Slug)
	}
	if response.IsActive != tenant.IsActive {
		t.Errorf("IsActive mismatch: got %v, expected %v", response.IsActive, tenant.IsActive)
	}
	if response.UserCount != tenant.UserCount {
		t.Errorf("UserCount mismatch: got %d, expected %d", response.UserCount, tenant.UserCount)
	}
	if response.HospitalCount != tenant.HospitalCount {
		t.Errorf("HospitalCount mismatch: got %d, expected %d", response.HospitalCount, tenant.HospitalCount)
	}
	if response.OccurrenceCount != tenant.OccurrenceCount {
		t.Errorf("OccurrenceCount mismatch: got %d, expected %d", response.OccurrenceCount, tenant.OccurrenceCount)
	}
}

// TestSlugValidation tests the slug validation function
func TestSlugValidation(t *testing.T) {
	tests := []struct {
		slug    string
		wantErr bool
	}{
		{"ses-go", false},          // Valid: lowercase with hyphen
		{"ses-pe", false},          // Valid: lowercase with hyphen
		{"central-sp", false},      // Valid: lowercase with hyphen
		{"tenant123", false},       // Valid: alphanumeric
		{"my-tenant-name", false},  // Valid: multiple hyphens
		{"ab", false},              // Valid: minimum length
		{"a", true},                // Invalid: too short
		{"SES-GO", true},           // Invalid: uppercase
		{"ses_go", true},           // Invalid: underscore
		{"ses go", true},           // Invalid: space
		{"ses-", true},             // Invalid: trailing hyphen
		{"-ses", true},             // Invalid: leading hyphen
		{"ses--go", true},          // Invalid: double hyphen
		{"", true},                 // Invalid: empty
		{"@#$%", true},             // Invalid: special chars
		{"ses.go", true},           // Invalid: dot
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			err := models.ValidateSlug(tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSlug(%q) error = %v, wantErr %v", tt.slug, err, tt.wantErr)
			}
		})
	}
}

// TestGetThemeConfig tests the GetThemeConfig method on Tenant
func TestGetThemeConfig(t *testing.T) {
	t.Run("Empty theme config returns default", func(t *testing.T) {
		tenant := &models.Tenant{
			ID:          uuid.New(),
			Name:        "Test",
			Slug:        "test",
			ThemeConfig: nil,
		}

		config, err := tenant.GetThemeConfig()
		if err != nil {
			t.Errorf("GetThemeConfig() error = %v", err)
		}
		if config == nil {
			t.Error("GetThemeConfig() should return default config, not nil")
		}
		if config.Theme.Colors.Primary != "#2563eb" {
			t.Errorf("Expected default primary color #2563eb, got %s", config.Theme.Colors.Primary)
		}
	})

	t.Run("Valid JSON theme config is parsed", func(t *testing.T) {
		customConfig := `{
			"theme": {
				"colors": {"primary": "#ff0000"},
				"fonts": {"body": "Roboto"}
			},
			"layout": {
				"sidebar": [],
				"topbar": {"show_user_info": true, "show_tenant_logo": false},
				"dashboard_widgets": []
			}
		}`

		tenant := &models.Tenant{
			ID:          uuid.New(),
			Name:        "Test",
			Slug:        "test",
			ThemeConfig: []byte(customConfig),
		}

		config, err := tenant.GetThemeConfig()
		if err != nil {
			t.Errorf("GetThemeConfig() error = %v", err)
		}
		if config.Theme.Colors.Primary != "#ff0000" {
			t.Errorf("Expected custom primary color #ff0000, got %s", config.Theme.Colors.Primary)
		}
		if config.Theme.Fonts.Body != "Roboto" {
			t.Errorf("Expected custom font Roboto, got %s", config.Theme.Fonts.Body)
		}
	})

	t.Run("Invalid JSON returns error", func(t *testing.T) {
		tenant := &models.Tenant{
			ID:          uuid.New(),
			Name:        "Test",
			Slug:        "test",
			ThemeConfig: []byte("invalid json"),
		}

		_, err := tenant.GetThemeConfig()
		if err == nil {
			t.Error("GetThemeConfig() should return error for invalid JSON")
		}
	})
}

// TestSetThemeConfig tests the SetThemeConfig method on Tenant
func TestSetThemeConfig(t *testing.T) {
	tenant := &models.Tenant{
		ID:   uuid.New(),
		Name: "Test",
		Slug: "test",
	}

	config := models.ThemeConfig{
		Theme: models.Theme{
			Colors: models.ThemeColors{
				Primary: "#00ff00",
			},
			Fonts: models.ThemeFonts{
				Body: "Arial",
			},
		},
	}

	err := tenant.SetThemeConfig(config)
	if err != nil {
		t.Errorf("SetThemeConfig() error = %v", err)
	}

	if tenant.ThemeConfig == nil {
		t.Error("ThemeConfig should not be nil after SetThemeConfig")
	}

	// Verify the config was set correctly by reading it back
	readConfig, err := tenant.GetThemeConfig()
	if err != nil {
		t.Errorf("GetThemeConfig() error after SetThemeConfig = %v", err)
	}
	if readConfig.Theme.Colors.Primary != "#00ff00" {
		t.Errorf("Primary color mismatch: got %s, expected #00ff00", readConfig.Theme.Colors.Primary)
	}
}
