package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/services/report"
)

var reportService *report.ReportService

// SetReportService sets the report service for handlers
func SetReportService(service *report.ReportService) {
	reportService = service
}

// ExportCSV handles GET /api/v1/reports/csv
// Generates and returns a CSV report with streaming
func ExportCSV(c *gin.Context) {
	if reportService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report service not configured"})
		return
	}

	// Get user claims for audit logging
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Parse and validate filters
	filters, err := parseReportFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// Generate filename with period
	filename := generateReportFilename("relatorio", filters, "csv")

	// Set response headers for CSV download
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Stream CSV directly to response
	c.Status(http.StatusOK)
	if err := reportService.GenerateCSV(c.Request.Context(), filters, c.Writer); err != nil {
		// Since we already started writing, we can't return a JSON error
		// Log the error instead
		_ = err
		return
	}

	// Log the export for LGPD compliance
	userID, err := uuid.Parse(claims.UserID)
	if err == nil {
		_ = reportService.LogExport(c.Request.Context(), userID, "CSV", filters)
	}
}

// ExportPDF handles GET /api/v1/reports/pdf
// Generates and returns a PDF report
func ExportPDF(c *gin.Context) {
	if reportService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report service not configured"})
		return
	}

	// Get user claims for audit logging
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Parse and validate filters
	filters, err := parseReportFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// Generate PDF
	pdfBytes, err := reportService.GeneratePDF(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to generate PDF report",
			"details": err.Error(),
		})
		return
	}

	// Generate filename with period
	filename := generateReportFilename("relatorio", filters, "pdf")

	// Set response headers for PDF download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Write PDF to response
	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	// Log the export for LGPD compliance
	userID, err := uuid.Parse(claims.UserID)
	if err == nil {
		_ = reportService.LogExport(c.Request.Context(), userID, "PDF", filters)
	}
}

// parseReportFilters parses and validates query parameters for report filters
func parseReportFilters(c *gin.Context) (models.ReportFilters, error) {
	var filters models.ReportFilters

	// Parse date_from
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			return filters, fmt.Errorf("date_from must be in YYYY-MM-DD format")
		}
		filters.DateFrom = &t
	}

	// Parse date_to
	if dateTo := c.Query("date_to"); dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			return filters, fmt.Errorf("date_to must be in YYYY-MM-DD format")
		}
		// Set to end of day
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.DateTo = &t
	}

	// Parse hospital_id
	if hospitalID := c.Query("hospital_id"); hospitalID != "" {
		// Validate UUID format
		if _, err := uuid.Parse(hospitalID); err != nil {
			return filters, fmt.Errorf("hospital_id must be a valid UUID")
		}
		filters.HospitalID = &hospitalID
	}

	// Parse desfecho[] (array parameter)
	desfechos := c.QueryArray("desfecho[]")
	if len(desfechos) == 0 {
		// Try without brackets
		desfechos = c.QueryArray("desfecho")
	}
	if len(desfechos) > 0 {
		for _, d := range desfechos {
			if !models.IsValidDesfecho(d) {
				validOptions := strings.Join(models.ValidDesfechoDisplayNames, ", ")
				return filters, fmt.Errorf("invalid desfecho value '%s'. Valid values: %s", d, validOptions)
			}
		}
		filters.Desfechos = desfechos
	}

	return filters, nil
}

// generateReportFilename generates a filename including the filter period
func generateReportFilename(prefix string, filters models.ReportFilters, extension string) string {
	dateFromStr := "inicio"
	dateToStr := "fim"

	if filters.DateFrom != nil {
		dateFromStr = filters.DateFrom.Format("2006-01-02")
	}
	if filters.DateTo != nil {
		dateToStr = filters.DateTo.Format("2006-01-02")
	}

	return fmt.Sprintf("%s_%s_%s.%s", prefix, dateFromStr, dateToStr, extension)
}
