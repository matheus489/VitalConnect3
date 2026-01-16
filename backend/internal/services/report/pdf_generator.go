package report

import (
	"bytes"
	"fmt"
	"time"

	"github.com/vitalconnect/backend/internal/models"
)

// PDFGenerator generates PDF reports
// Note: Using a simple text-based PDF approach for the MVP
// Full PDF generation with gofpdf can be added later
type PDFGenerator struct{}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

// GenerateReport generates a PDF report
func (g *PDFGenerator) GenerateReport(filters models.ReportFilters, metrics *models.ReportMetrics, rows []models.ReportOccurrenceRow) ([]byte, error) {
	var buf bytes.Buffer

	// PDF Header
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("1 0 obj\n")
	buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	buf.WriteString("endobj\n")

	// Pages object
	buf.WriteString("2 0 obj\n")
	buf.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	buf.WriteString("endobj\n")

	// Page object
	buf.WriteString("3 0 obj\n")
	buf.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	buf.WriteString("endobj\n")

	// Font object
	buf.WriteString("5 0 obj\n")
	buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	buf.WriteString("endobj\n")

	// Build content stream
	var content bytes.Buffer
	content.WriteString("BT\n")

	// Header - VitalConnect logo placeholder and SES text
	yPos := 750
	content.WriteString(fmt.Sprintf("/F1 16 Tf\n50 %d Td\n(VitalConnect) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n350 %d Td\n(Governo do Estado de Goias - SES) Tj\n", yPos))

	// Title
	yPos -= 40
	content.WriteString(fmt.Sprintf("/F1 14 Tf\n50 %d Td\n(Relatorio de Ocorrencias) Tj\n", yPos))

	// Period
	yPos -= 25
	periodoText := "Periodo: "
	if filters.DateFrom != nil {
		periodoText += filters.DateFrom.Format("02/01/2006")
	} else {
		periodoText += "Inicio"
	}
	periodoText += " a "
	if filters.DateTo != nil {
		periodoText += filters.DateTo.Format("02/01/2006")
	} else {
		periodoText += "Fim"
	}
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n50 %d Td\n(%s) Tj\n", yPos, escapeString(periodoText)))

	// Metrics section
	yPos -= 30
	content.WriteString(fmt.Sprintf("/F1 12 Tf\n50 %d Td\n(Metricas Agregadas) Tj\n", yPos))

	yPos -= 20
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n50 %d Td\n(Total de Ocorrencias: %d) Tj\n", yPos, metrics.TotalOcorrencias))

	yPos -= 15
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n50 %d Td\n(Taxa de Perda Operacional: %.1f%%) Tj\n", yPos, metrics.TaxaPerdaOperacional))

	yPos -= 15
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n50 %d Td\n(Tempo Medio de Reacao: %.1f min) Tj\n", yPos, metrics.TempoMedioReacaoMin))

	// Desfechos breakdown
	yPos -= 20
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n50 %d Td\n(Ocorrencias por Desfecho:) Tj\n", yPos))
	for desfecho, count := range metrics.OcorrenciasPorDesfecho {
		yPos -= 15
		content.WriteString(fmt.Sprintf("/F1 9 Tf\n60 %d Td\n(- %s: %d) Tj\n", yPos, escapeString(desfecho), count))
	}

	// Table header
	yPos -= 30
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n50 %d Td\n(Hospital) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("150 %d Td\n(Data Obito) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("230 %d Td\n(Paciente) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("290 %d Td\n(Idade) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("330 %d Td\n(Status) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("420 %d Td\n(Tempo) Tj\n", yPos))
	content.WriteString(fmt.Sprintf("470 %d Td\n(Responsavel) Tj\n", yPos))

	// Separator line
	yPos -= 5
	content.WriteString("ET\n")
	content.WriteString(fmt.Sprintf("50 %d m 560 %d l S\n", yPos, yPos))
	content.WriteString("BT\n")

	// Table rows (zebra striped)
	yPos -= 15
	for i, row := range rows {
		if yPos < 80 {
			// Would need pagination for more rows - skip for MVP
			break
		}

		// Zebra stripe background (not implemented in basic PDF)
		_ = i

		// Truncate long text
		hospitalNome := truncateString(row.HospitalNome, 15)
		iniciaisPaciente := truncateString(row.IniciaisPaciente, 10)
		statusFinal := truncateString(row.StatusFinal, 12)

		tempoReacao := ""
		if row.TempoReacaoMin != nil {
			tempoReacao = fmt.Sprintf("%.0f", *row.TempoReacaoMin)
		}

		usuarioResp := ""
		if row.UsuarioResponsavel != nil {
			usuarioResp = truncateString(*row.UsuarioResponsavel, 12)
		}

		content.WriteString(fmt.Sprintf("/F1 8 Tf\n50 %d Td\n(%s) Tj\n", yPos, escapeString(hospitalNome)))
		content.WriteString(fmt.Sprintf("150 %d Td\n(%s) Tj\n", yPos, row.DataHoraObito.Format("02/01/06 15:04")))
		content.WriteString(fmt.Sprintf("230 %d Td\n(%s) Tj\n", yPos, escapeString(iniciaisPaciente)))
		content.WriteString(fmt.Sprintf("290 %d Td\n(%d) Tj\n", yPos, row.Idade))
		content.WriteString(fmt.Sprintf("330 %d Td\n(%s) Tj\n", yPos, escapeString(statusFinal)))
		content.WriteString(fmt.Sprintf("420 %d Td\n(%s) Tj\n", yPos, tempoReacao))
		content.WriteString(fmt.Sprintf("470 %d Td\n(%s) Tj\n", yPos, escapeString(usuarioResp)))

		yPos -= 12
	}

	// Footer
	content.WriteString("ET\n")
	content.WriteString("BT\n")
	content.WriteString(fmt.Sprintf("/F1 8 Tf\n50 30 Td\n(Gerado automaticamente por VitalConnect em %s) Tj\n", time.Now().Format("02/01/2006 15:04")))
	content.WriteString(fmt.Sprintf("500 30 Td\n(Pagina 1) Tj\n"))
	content.WriteString("ET\n")

	// Content stream object
	contentBytes := content.Bytes()
	buf.WriteString("4 0 obj\n")
	buf.WriteString(fmt.Sprintf("<< /Length %d >>\n", len(contentBytes)))
	buf.WriteString("stream\n")
	buf.Write(contentBytes)
	buf.WriteString("\nendstream\n")
	buf.WriteString("endobj\n")

	// Cross-reference table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString("0000000009 00000 n \n")
	buf.WriteString("0000000058 00000 n \n")
	buf.WriteString("0000000115 00000 n \n")
	buf.WriteString("0000000270 00000 n \n")
	buf.WriteString("0000000214 00000 n \n")

	// Trailer
	buf.WriteString("trailer\n")
	buf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	buf.WriteString("%%EOF\n")

	return buf.Bytes(), nil
}

// escapeString escapes special characters for PDF strings
func escapeString(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '(':
			result += "\\("
		case ')':
			result += "\\)"
		case '\\':
			result += "\\\\"
		default:
			result += string(c)
		}
	}
	return result
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
