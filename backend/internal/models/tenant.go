package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidTenantSlug is returned when a tenant slug is invalid
	ErrInvalidTenantSlug = errors.New("invalid tenant slug: must be alphanumeric with hyphens only, 2-100 characters")

	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrTenantSlugExists is returned when a tenant slug already exists
	ErrTenantSlugExists = errors.New("tenant with this slug already exists")

	// ErrTenantInactive is returned when trying to access an inactive tenant
	ErrTenantInactive = errors.New("tenant is inactive")

	// slugRegex validates tenant slugs: alphanumeric with hyphens, no leading/trailing hyphens
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// ThemeColors represents the color configuration for a tenant's UI theme
type ThemeColors struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary,omitempty"`
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
	Muted      string `json:"muted,omitempty"`
	Accent     string `json:"accent,omitempty"`
}

// ThemeFonts represents the font configuration for a tenant's UI theme
type ThemeFonts struct {
	Body    string `json:"body"`
	Heading string `json:"heading,omitempty"`
}

// Theme represents the visual theme configuration
type Theme struct {
	Colors ThemeColors `json:"colors"`
	Fonts  ThemeFonts  `json:"fonts"`
}

// SidebarItem represents a navigation item in the tenant's sidebar
type SidebarItem struct {
	Label string   `json:"label"`
	Icon  string   `json:"icon"`
	Link  string   `json:"link"`
	Roles []string `json:"roles,omitempty"`
	Order int      `json:"order,omitempty"`
}

// TopbarConfig represents the topbar configuration
type TopbarConfig struct {
	ShowUserInfo   bool `json:"show_user_info"`
	ShowTenantLogo bool `json:"show_tenant_logo"`
}

// DashboardWidget represents a configurable dashboard widget
type DashboardWidget struct {
	Type    string                 `json:"type"`
	Visible bool                   `json:"visible"`
	Order   int                    `json:"order"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

// Layout represents the layout configuration for the tenant's UI
type Layout struct {
	Sidebar          []SidebarItem     `json:"sidebar"`
	Topbar           TopbarConfig      `json:"topbar"`
	DashboardWidgets []DashboardWidget `json:"dashboard_widgets"`
}

// ThemeConfig represents the complete UI configuration for a tenant
type ThemeConfig struct {
	Theme  Theme  `json:"theme"`
	Layout Layout `json:"layout"`
}

// DefaultThemeConfig returns the default theme configuration for new tenants
func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Theme: Theme{
			Colors: ThemeColors{
				Primary:    "#2563eb",
				Secondary:  "#64748b",
				Background: "#ffffff",
				Foreground: "#0f172a",
				Muted:      "#f1f5f9",
				Accent:     "#f59e0b",
			},
			Fonts: ThemeFonts{
				Body:    "Inter",
				Heading: "Inter",
			},
		},
		Layout: Layout{
			Sidebar: []SidebarItem{},
			Topbar: TopbarConfig{
				ShowUserInfo:   true,
				ShowTenantLogo: true,
			},
			DashboardWidgets: []DashboardWidget{},
		},
	}
}

// Tenant represents a Central de Transplantes (e.g., SES-GO, SES-PE, SES-SP)
type Tenant struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name" validate:"required,min=2,max=255"`
	Slug        string          `json:"slug" db:"slug" validate:"required,min=2,max=100"`
	ThemeConfig json.RawMessage `json:"theme_config,omitempty" db:"theme_config"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	LogoURL     *string         `json:"logo_url,omitempty" db:"logo_url"`
	FaviconURL  *string         `json:"favicon_url,omitempty" db:"favicon_url"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateTenantInput represents input for creating a tenant
type CreateTenantInput struct {
	Name        string       `json:"name" validate:"required,min=2,max=255"`
	Slug        string       `json:"slug" validate:"required,min=2,max=100"`
	ThemeConfig *ThemeConfig `json:"theme_config,omitempty"`
	LogoURL     *string      `json:"logo_url,omitempty"`
	FaviconURL  *string      `json:"favicon_url,omitempty"`
}

// UpdateTenantInput represents input for updating a tenant
type UpdateTenantInput struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Slug       *string `json:"slug,omitempty" validate:"omitempty,min=2,max=100"`
	IsActive   *bool   `json:"is_active,omitempty"`
	LogoURL    *string `json:"logo_url,omitempty"`
	FaviconURL *string `json:"favicon_url,omitempty"`
}

// UpdateThemeConfigInput represents input for updating tenant theme configuration
type UpdateThemeConfigInput struct {
	ThemeConfig ThemeConfig `json:"theme_config" validate:"required"`
}

// TenantResponse represents the API response for a tenant
type TenantResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	ThemeConfig json.RawMessage `json:"theme_config,omitempty"`
	IsActive    bool            `json:"is_active"`
	LogoURL     *string         `json:"logo_url,omitempty"`
	FaviconURL  *string         `json:"favicon_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TenantWithMetrics represents a tenant with additional metrics for admin views
type TenantWithMetrics struct {
	Tenant
	UserCount       int `json:"user_count" db:"user_count"`
	HospitalCount   int `json:"hospital_count" db:"hospital_count"`
	OccurrenceCount int `json:"occurrence_count" db:"occurrence_count"`
}

// TenantWithMetricsResponse represents the API response for a tenant with metrics
type TenantWithMetricsResponse struct {
	TenantResponse
	UserCount       int `json:"user_count"`
	HospitalCount   int `json:"hospital_count"`
	OccurrenceCount int `json:"occurrence_count"`
}

// ValidateSlug validates that a slug is URL-safe and follows naming conventions
// Valid slugs: lowercase alphanumeric with hyphens, no leading/trailing hyphens
// Examples: "ses-go", "ses-pe", "central-sp"
func ValidateSlug(slug string) error {
	if len(slug) < 2 || len(slug) > 100 {
		return ErrInvalidTenantSlug
	}

	if !slugRegex.MatchString(slug) {
		return ErrInvalidTenantSlug
	}

	return nil
}

// Validate validates the tenant data
func (t *Tenant) Validate() error {
	if t.Name == "" || len(t.Name) < 2 || len(t.Name) > 255 {
		return errors.New("tenant name must be between 2 and 255 characters")
	}

	return ValidateSlug(t.Slug)
}

// ToResponse converts Tenant to TenantResponse
func (t *Tenant) ToResponse() TenantResponse {
	return TenantResponse{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		ThemeConfig: t.ThemeConfig,
		IsActive:    t.IsActive,
		LogoURL:     t.LogoURL,
		FaviconURL:  t.FaviconURL,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// GetThemeConfig parses and returns the theme configuration
func (t *Tenant) GetThemeConfig() (*ThemeConfig, error) {
	if len(t.ThemeConfig) == 0 {
		defaultConfig := DefaultThemeConfig()
		return &defaultConfig, nil
	}

	var config ThemeConfig
	if err := json.Unmarshal(t.ThemeConfig, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetThemeConfig sets the theme configuration from a ThemeConfig struct
func (t *Tenant) SetThemeConfig(config ThemeConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	t.ThemeConfig = data
	return nil
}

// ToWithMetricsResponse converts TenantWithMetrics to TenantWithMetricsResponse
func (t *TenantWithMetrics) ToWithMetricsResponse() TenantWithMetricsResponse {
	return TenantWithMetricsResponse{
		TenantResponse:  t.Tenant.ToResponse(),
		UserCount:       t.UserCount,
		HospitalCount:   t.HospitalCount,
		OccurrenceCount: t.OccurrenceCount,
	}
}

// Validate validates the CreateTenantInput
func (i *CreateTenantInput) Validate() error {
	if i.Name == "" || len(i.Name) < 2 || len(i.Name) > 255 {
		return errors.New("tenant name must be between 2 and 255 characters")
	}

	return ValidateSlug(i.Slug)
}

// Validate validates the UpdateTenantInput
func (i *UpdateTenantInput) Validate() error {
	if i.Name != nil && (len(*i.Name) < 2 || len(*i.Name) > 255) {
		return errors.New("tenant name must be between 2 and 255 characters")
	}

	if i.Slug != nil {
		return ValidateSlug(*i.Slug)
	}

	return nil
}

// ValidateThemeConfig validates the theme configuration structure
func ValidateThemeConfig(config *ThemeConfig) error {
	if config == nil {
		return errors.New("theme_config cannot be nil")
	}

	// Validate colors
	if config.Theme.Colors.Primary == "" {
		return errors.New("theme primary color is required")
	}

	// Validate fonts
	if config.Theme.Fonts.Body == "" {
		return errors.New("theme body font is required")
	}

	// Validate sidebar items
	for i, item := range config.Layout.Sidebar {
		if item.Label == "" {
			return errors.New("sidebar item label is required at index " + string(rune('0'+i)))
		}
		if item.Link == "" {
			return errors.New("sidebar item link is required at index " + string(rune('0'+i)))
		}
	}

	// Validate dashboard widgets
	for i, widget := range config.Layout.DashboardWidgets {
		if widget.Type == "" {
			return errors.New("dashboard widget type is required at index " + string(rune('0'+i)))
		}
	}

	return nil
}
