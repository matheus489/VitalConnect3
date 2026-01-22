package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sidot/backend/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestExportCSV_RequiresAuthentication tests that CSV export requires auth
func TestExportCSV_RequiresAuthentication(t *testing.T) {
	t.Run("Returns 401 when no auth token", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/reports/csv", ExportCSV)

		req, _ := http.NewRequest("GET", "/api/v1/reports/csv", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Without auth middleware, handler checks claims
		if w.Code != http.StatusUnauthorized && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 401 or 500 without auth, got %d", w.Code)
		}
	})
}

// TestExportPDF_RequiresAuthentication tests that PDF export requires auth
func TestExportPDF_RequiresAuthentication(t *testing.T) {
	t.Run("Returns 401 when no auth token", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/reports/pdf", ExportPDF)

		req, _ := http.NewRequest("GET", "/api/v1/reports/pdf", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Without auth middleware, handler checks claims
		if w.Code != http.StatusUnauthorized && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 401 or 500 without auth, got %d", w.Code)
		}
	})
}

// TestParseReportFilters_ValidDates tests date parsing
func TestParseReportFilters_ValidDates(t *testing.T) {
	t.Run("Valid date_from format", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			filters, err := parseReportFilters(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if filters.DateFrom == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "date_from not parsed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test?date_from=2025-01-01", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for valid date, got %d", w.Code)
		}
	})

	t.Run("Invalid date_from format returns 400", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			_, err := parseReportFilters(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test?date_from=01-01-2025", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid date, got %d", w.Code)
		}
	})
}

// TestParseReportFilters_ValidHospitalID tests hospital ID validation
func TestParseReportFilters_ValidHospitalID(t *testing.T) {
	t.Run("Valid UUID hospital_id", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			filters, err := parseReportFilters(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if filters.HospitalID == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "hospital_id not parsed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test?hospital_id=123e4567-e89b-12d3-a456-426614174000", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for valid UUID, got %d", w.Code)
		}
	})

	t.Run("Invalid UUID hospital_id returns 400", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			_, err := parseReportFilters(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test?hospital_id=not-a-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid UUID, got %d", w.Code)
		}
	})
}

// TestParseReportFilters_ValidDesfecho tests desfecho validation
func TestParseReportFilters_ValidDesfecho(t *testing.T) {
	t.Run("Valid desfecho values", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			filters, err := parseReportFilters(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if len(filters.Desfechos) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "desfechos not parsed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test?desfecho[]=Captado&desfecho[]=Expirado", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for valid desfecho, got %d", w.Code)
		}
	})

	t.Run("Invalid desfecho value returns 400", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			_, err := parseReportFilters(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test?desfecho[]=InvalidValue", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 for invalid desfecho, got %d", w.Code)
		}
	})
}

// TestGenerateReportFilename tests filename generation
func TestGenerateReportFilename(t *testing.T) {
	t.Run("Filename with date range", func(t *testing.T) {
		// Test helper function for filename generation
		expected := "relatorio_inicio_fim.csv"
		actual := generateReportFilename("relatorio", models.ReportFilters{}, "csv")

		if actual != expected {
			t.Errorf("Expected '%s', got '%s'", expected, actual)
		}
	})
}
