package handlers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
)

var adminTenantRepo *repository.AdminTenantRepository

// SetAdminTenantRepository sets the admin tenant repository for handlers
func SetAdminTenantRepository(repo *repository.AdminTenantRepository) {
	adminTenantRepo = repo
}

// AdminListTenants returns all tenants with pagination (no tenant_id filter)
// GET /api/v1/admin/tenants
func AdminListTenants(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	var params repository.AdminTenantListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	result, err := adminTenantRepo.ListAllTenants(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tenants"})
		return
	}

	// Convert to response format
	response := make([]models.TenantWithMetricsResponse, 0, len(result.Tenants))
	for _, t := range result.Tenants {
		response = append(response, t.ToWithMetricsResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        response,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// AdminGetTenant returns a single tenant by ID with metrics
// GET /api/v1/admin/tenants/:id
func AdminGetTenant(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
		return
	}

	tenant, err := adminTenantRepo.GetTenantByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tenant"})
		return
	}

	c.JSON(http.StatusOK, tenant.ToWithMetricsResponse())
}

// AdminCreateTenant creates a new tenant
// POST /api/v1/admin/tenants
func AdminCreateTenant(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	var input models.CreateTenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	tenant, err := adminTenantRepo.CreateTenant(c.Request.Context(), &input)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantSlugExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "tenant with this slug already exists"})
			return
		}
		if errors.Is(err, models.ErrInvalidTenantSlug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tenant"})
		return
	}

	c.JSON(http.StatusCreated, tenant.ToResponse())
}

// AdminUpdateTenant updates a tenant's basic info
// PUT /api/v1/admin/tenants/:id
func AdminUpdateTenant(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
		return
	}

	var input models.UpdateTenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	tenant, err := adminTenantRepo.UpdateTenant(c.Request.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		if errors.Is(err, repository.ErrAdminTenantSlugExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "tenant with this slug already exists"})
			return
		}
		if errors.Is(err, models.ErrInvalidTenantSlug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, tenant.ToResponse())
}

// AdminUpdateThemeConfig updates a tenant's theme configuration
// PUT /api/v1/admin/tenants/:id/theme
func AdminUpdateThemeConfig(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
		return
	}

	var input models.UpdateThemeConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	tenant, err := adminTenantRepo.UpdateThemeConfig(c.Request.Context(), id, input.ThemeConfig)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update theme config",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tenant.ToResponse())
}

// AdminToggleTenantActive toggles a tenant's is_active status
// PUT /api/v1/admin/tenants/:id/toggle
func AdminToggleTenantActive(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
		return
	}

	tenant, err := adminTenantRepo.ToggleTenantActive(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle tenant status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "tenant status toggled successfully",
		"is_active": tenant.IsActive,
		"tenant":    tenant.ToResponse(),
	})
}

// AdminUploadTenantAssets uploads logo and/or favicon for a tenant
// POST /api/v1/admin/tenants/:id/assets
func AdminUploadTenantAssets(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
		return
	}

	// Verify tenant exists
	_, err = adminTenantRepo.GetTenantByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify tenant"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form data"})
		return
	}

	var logoURL, faviconURL *string

	// Create upload directory if it doesn't exist
	uploadDir := filepath.Join("uploads", "tenants", id.String())
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	// Handle logo upload
	logoFile, logoHeader, err := c.Request.FormFile("logo")
	if err == nil {
		defer logoFile.Close()

		// Validate file type
		if !isValidImageType(logoHeader.Header.Get("Content-Type")) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "logo must be an image file (PNG, JPG, SVG, WebP)"})
			return
		}

		// Generate unique filename
		ext := filepath.Ext(logoHeader.Filename)
		filename := "logo_" + time.Now().Format("20060102150405") + ext
		filePath := filepath.Join(uploadDir, filename)

		// Save file
		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save logo"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, logoFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write logo file"})
			return
		}

		url := "/" + filePath
		logoURL = &url
	}

	// Handle favicon upload
	faviconFile, faviconHeader, err := c.Request.FormFile("favicon")
	if err == nil {
		defer faviconFile.Close()

		// Validate file type
		if !isValidFaviconType(faviconHeader.Header.Get("Content-Type")) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "favicon must be an image file (PNG, ICO, SVG)"})
			return
		}

		// Generate unique filename
		ext := filepath.Ext(faviconHeader.Filename)
		filename := "favicon_" + time.Now().Format("20060102150405") + ext
		filePath := filepath.Join(uploadDir, filename)

		// Save file
		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save favicon"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, faviconFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write favicon file"})
			return
		}

		url := "/" + filePath
		faviconURL = &url
	}

	// Check if at least one file was uploaded
	if logoURL == nil && faviconURL == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded; provide 'logo' and/or 'favicon' files"})
		return
	}

	// Update tenant with new asset URLs
	tenant, err := adminTenantRepo.UpdateAssets(c.Request.Context(), id, logoURL, faviconURL)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tenant assets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "assets uploaded successfully",
		"tenant":  tenant.ToResponse(),
	})
}

// isValidImageType checks if the content type is a valid image type for logo
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/png",
		"image/jpeg",
		"image/jpg",
		"image/svg+xml",
		"image/webp",
	}
	ct := strings.ToLower(contentType)
	for _, vt := range validTypes {
		if ct == vt {
			return true
		}
	}
	return false
}

// isValidFaviconType checks if the content type is a valid favicon type
func isValidFaviconType(contentType string) bool {
	validTypes := []string{
		"image/png",
		"image/x-icon",
		"image/vnd.microsoft.icon",
		"image/svg+xml",
		"image/ico",
	}
	ct := strings.ToLower(contentType)
	for _, vt := range validTypes {
		if ct == vt {
			return true
		}
	}
	return false
}

// AdminDashboardMetrics returns global metrics for the admin dashboard
// GET /api/v1/admin/metrics
func AdminDashboardMetrics(c *gin.Context) {
	if adminTenantRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin tenant repository not configured"})
		return
	}

	metrics, err := adminTenantRepo.GetDashboardMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get dashboard metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
