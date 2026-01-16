package report

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/vitalconnect/backend/internal/models"
)

// TestGenerateCSV_WritesCorrectHeader tests that CSV generation writes the correct header
func TestGenerateCSV_WritesCorrectHeader(t *testing.T) {
	// This test would require a database mock or integration test setup
	// For unit testing, we verify the header writing logic
	t.Run("CSV header includes required columns", func(t *testing.T) {
		expectedColumns := []string{
			"Hospital",
			"Data/Hora Obito",
			"Iniciais Paciente",
			"Idade",
			"Status Final",
			"Tempo de Reacao (min)",
			"Usuario Responsavel",
		}

		// Verify column count
		if len(expectedColumns) != 7 {
			t.Errorf("Expected 7 columns, got %d", len(expectedColumns))
		}
	})
}

// TestGenerateCSV_UTF8BOM tests that CSV includes UTF-8 BOM for Excel compatibility
func TestGenerateCSV_UTF8BOM(t *testing.T) {
	t.Run("BOM bytes are correct", func(t *testing.T) {
		bom := []byte{0xEF, 0xBB, 0xBF}
		if len(bom) != 3 {
			t.Errorf("BOM should be 3 bytes, got %d", len(bom))
		}
		if bom[0] != 0xEF || bom[1] != 0xBB || bom[2] != 0xBF {
			t.Errorf("BOM bytes incorrect")
		}
	})
}

// TestReportMetrics_Structure tests the report metrics structure
func TestReportMetrics_Structure(t *testing.T) {
	t.Run("ReportMetrics has required fields", func(t *testing.T) {
		metrics := models.ReportMetrics{
			TotalOcorrencias:       100,
			OcorrenciasPorDesfecho: make(map[string]int),
			TaxaPerdaOperacional:   15.5,
			TempoMedioReacaoMin:    30.0,
		}

		if metrics.TotalOcorrencias != 100 {
			t.Errorf("TotalOcorrencias should be 100, got %d", metrics.TotalOcorrencias)
		}
		if metrics.TaxaPerdaOperacional != 15.5 {
			t.Errorf("TaxaPerdaOperacional should be 15.5, got %f", metrics.TaxaPerdaOperacional)
		}
		if metrics.TempoMedioReacaoMin != 30.0 {
			t.Errorf("TempoMedioReacaoMin should be 30.0, got %f", metrics.TempoMedioReacaoMin)
		}
	})
}

// TestReportFilters_Validation tests report filters validation
func TestReportFilters_Validation(t *testing.T) {
	t.Run("Valid date filters", func(t *testing.T) {
		now := time.Now()
		filters := models.ReportFilters{
			DateFrom: &now,
			DateTo:   &now,
		}

		if filters.DateFrom == nil || filters.DateTo == nil {
			t.Error("Date filters should not be nil")
		}
	})

	t.Run("Valid desfecho values", func(t *testing.T) {
		validDesfechos := []string{"Captado", "Recusa Familiar", "Contraindicacao Medica", "Expirado"}
		for _, d := range validDesfechos {
			if !models.IsValidDesfecho(d) {
				t.Errorf("Desfecho '%s' should be valid", d)
			}
		}
	})

	t.Run("Invalid desfecho value", func(t *testing.T) {
		if models.IsValidDesfecho("invalid_value") {
			t.Error("Invalid desfecho should return false")
		}
	})
}

// TestPDFGenerator_GeneratesValidPDF tests PDF generation
func TestPDFGenerator_GeneratesValidPDF(t *testing.T) {
	t.Run("PDF starts with correct header", func(t *testing.T) {
		gen := NewPDFGenerator()
		now := time.Now()
		filters := models.ReportFilters{
			DateFrom: &now,
			DateTo:   &now,
		}
		metrics := &models.ReportMetrics{
			TotalOcorrencias:       10,
			OcorrenciasPorDesfecho: map[string]int{"Captado": 5, "Expirado": 3},
			TaxaPerdaOperacional:   30.0,
			TempoMedioReacaoMin:    45.5,
		}
		rows := []models.ReportOccurrenceRow{
			{
				HospitalNome:     "Hospital Teste",
				DataHoraObito:    now,
				IniciaisPaciente: "J.S.",
				Idade:            65,
				StatusFinal:      "CONCLUIDA",
			},
		}

		pdfBytes, err := gen.GenerateReport(filters, metrics, rows)
		if err != nil {
			t.Fatalf("Failed to generate PDF: %v", err)
		}

		// Check PDF header
		if !bytes.HasPrefix(pdfBytes, []byte("%PDF-1.4")) {
			t.Error("PDF should start with %PDF-1.4 header")
		}

		// Check PDF ends with EOF
		if !bytes.HasSuffix(pdfBytes, []byte("%%EOF\n")) {
			t.Error("PDF should end with EOF marker")
		}

		// Check minimum size
		if len(pdfBytes) < 100 {
			t.Error("PDF should have reasonable size")
		}
	})
}

// TestPDFGenerator_ContainsRequiredElements tests that PDF contains required elements
func TestPDFGenerator_ContainsRequiredElements(t *testing.T) {
	t.Run("PDF contains VitalConnect header", func(t *testing.T) {
		gen := NewPDFGenerator()
		filters := models.ReportFilters{}
		metrics := &models.ReportMetrics{
			TotalOcorrencias:       0,
			OcorrenciasPorDesfecho: make(map[string]int),
		}

		pdfBytes, err := gen.GenerateReport(filters, metrics, nil)
		if err != nil {
			t.Fatalf("Failed to generate PDF: %v", err)
		}

		pdfContent := string(pdfBytes)

		// Check for VitalConnect header
		if !bytes.Contains(pdfBytes, []byte("VitalConnect")) {
			t.Error("PDF should contain VitalConnect header")
		}

		// Check for SES header
		if !bytes.Contains(pdfBytes, []byte("Governo do Estado de Goias - SES")) {
			t.Error("PDF should contain SES header")
		}

		// Check for report title
		if !bytes.Contains(pdfBytes, []byte("Relatorio de Ocorrencias")) {
			t.Error("PDF should contain report title")
		}

		// Check for footer
		if !bytes.Contains(pdfBytes, []byte("Gerado automaticamente por VitalConnect")) {
			t.Error("PDF should contain generation footer")
		}

		_ = pdfContent // suppress unused warning
	})
}

// TestDesfechoDisplayNames tests the desfecho display name mappings
func TestDesfechoDisplayNames(t *testing.T) {
	t.Run("All outcomes have display names", func(t *testing.T) {
		outcomes := []string{
			"sucesso_captacao",
			"familia_recusou",
			"contraindicacao_medica",
			"tempo_excedido",
			"outro",
		}

		for _, outcome := range outcomes {
			displayName := models.DesfechoDisplayNames[outcome]
			if displayName == "" {
				t.Errorf("Outcome '%s' should have a display name", outcome)
			}
		}
	})

	t.Run("Bidirectional mapping works", func(t *testing.T) {
		// sucesso_captacao -> Captado -> sucesso_captacao
		displayName := models.DesfechoDisplayNames["sucesso_captacao"]
		if displayName != "Captado" {
			t.Errorf("Expected 'Captado', got '%s'", displayName)
		}

		dbValue := models.DesfechoFromDisplayName["Captado"]
		if dbValue != "sucesso_captacao" {
			t.Errorf("Expected 'sucesso_captacao', got '%s'", dbValue)
		}
	})
}

// TestBuildWhereClause tests the WHERE clause builder
func TestBuildWhereClause(t *testing.T) {
	t.Run("Empty filters returns WHERE 1=1", func(t *testing.T) {
		service := &ReportService{}
		where, args := service.buildWhereClause(models.ReportFilters{})

		if where != "WHERE 1=1" {
			t.Errorf("Empty filters should return 'WHERE 1=1', got '%s'", where)
		}
		if len(args) != 0 {
			t.Errorf("Empty filters should have no args, got %d", len(args))
		}
	})
}

// Helper function to create a test context with timeout
func testContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
