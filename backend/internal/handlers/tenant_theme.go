package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantThemeResponse represents the theme configuration response for a tenant
type TenantThemeResponse struct {
	ThemeConfig interface{} `json:"theme_config"`
	LogoURL     *string     `json:"logo_url,omitempty"`
	FaviconURL  *string     `json:"favicon_url,omitempty"`
}

var tenantThemeDB *sql.DB

// SetTenantThemeDB sets the database connection for tenant theme handlers
func SetTenantThemeDB(db *sql.DB) {
	tenantThemeDB = db
}

// GetCurrentTenantTheme returns the theme configuration for the current user's tenant
// GET /api/v1/tenants/current/theme
func GetCurrentTenantTheme(c *gin.Context) {
	if tenantThemeDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not configured"})
		return
	}

	// Get tenant_id from context (set by TenantContextMiddleware)
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant context not found"})
		return
	}

	tenantID, ok := tenantIDValue.(uuid.UUID)
	if !ok {
		// Try parsing as string
		tenantIDStr, ok := tenantIDValue.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
			return
		}
		var err error
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
			return
		}
	}

	// Query tenant theme config
	query := `
		SELECT theme_config, logo_url, favicon_url
		FROM tenants
		WHERE id = $1
	`

	var themeConfigStr, logoURL, faviconURL sql.NullString
	err := tenantThemeDB.QueryRowContext(c.Request.Context(), query, tenantID).Scan(
		&themeConfigStr,
		&logoURL,
		&faviconURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tenant theme"})
		return
	}

	response := TenantThemeResponse{}

	// Parse theme_config JSON if it exists
	if themeConfigStr.Valid && themeConfigStr.String != "" {
		// Return as raw JSON - the frontend will parse it
		var themeConfig interface{}
		if err := json.Unmarshal([]byte(themeConfigStr.String), &themeConfig); err == nil {
			response.ThemeConfig = themeConfig
		}
	}

	if logoURL.Valid {
		response.LogoURL = &logoURL.String
	}
	if faviconURL.Valid {
		response.FaviconURL = &faviconURL.String
	}

	c.JSON(http.StatusOK, response)
}
