// Package models defines data structures for the PEP Agent
package models

import (
	"strings"
	"time"
	"unicode"
)

// ObitoEvent represents a death event to be sent to SIDOT central server
// This structure mirrors the ObitoEvent in the main backend for compatibility
type ObitoEvent struct {
	// Unique identifier from the source PEP system
	HospitalIDOrigem string `json:"hospital_id_origem"`

	// Hospital UUID in SIDOT
	HospitalID string `json:"hospital_id"`

	// Detection timestamp (when the agent detected this event)
	TimestampDeteccao string `json:"timestamp_deteccao"`

	// Patient information
	NomePaciente string `json:"nome_paciente"`
	DataObito    string `json:"data_obito"`
	CausaMortis  string `json:"causa_mortis"`

	// Age or birth date
	DataNascimento string `json:"data_nascimento,omitempty"`
	Idade          int    `json:"idade"`

	// Patient identification (LGPD compliant)
	CNS       string `json:"cns,omitempty"`        // Full CNS (health identifier)
	CPFMasked string `json:"cpf_masked,omitempty"` // Masked CPF

	// Location
	Setor string `json:"setor,omitempty"`
	Leito string `json:"leito,omitempty"`

	// Medical record
	Prontuario string `json:"prontuario,omitempty"`

	// Unknown identification flag
	IdentificacaoDesconhecida bool `json:"identificacao_desconhecida"`
}

// PEPRecord represents a raw record from the hospital PEP database
type PEPRecord struct {
	ID                        string
	NomePaciente              string
	DataObito                 time.Time
	CausaMortis               string
	DataNascimento            *time.Time
	Idade                     *int
	CNS                       *string
	CPF                       *string
	Setor                     *string
	Leito                     *string
	Prontuario                *string
	IdentificacaoDesconhecida string // 'S' or 'N'
}

// ToObitoEvent converts a PEPRecord to ObitoEvent with LGPD masking
func (r *PEPRecord) ToObitoEvent(hospitalID string) *ObitoEvent {
	event := &ObitoEvent{
		HospitalIDOrigem:  r.ID,
		HospitalID:        hospitalID,
		TimestampDeteccao: time.Now().Format(time.RFC3339),
		NomePaciente:      r.NomePaciente,
		DataObito:         r.DataObito.Format(time.RFC3339),
		CausaMortis:       r.CausaMortis,
	}

	// Calculate age
	if r.DataNascimento != nil {
		event.DataNascimento = r.DataNascimento.Format("2006-01-02")
		event.Idade = calculateAge(*r.DataNascimento, r.DataObito)
	} else if r.Idade != nil {
		event.Idade = *r.Idade
	}

	// CNS is transmitted in full (health identifier, not personal)
	if r.CNS != nil && *r.CNS != "" {
		event.CNS = *r.CNS
	}

	// CPF must be masked for LGPD compliance
	if r.CPF != nil && *r.CPF != "" {
		event.CPFMasked = MaskCPF(*r.CPF)
	}

	// Optional fields
	if r.Setor != nil {
		event.Setor = *r.Setor
	}
	if r.Leito != nil {
		event.Leito = *r.Leito
	}
	if r.Prontuario != nil {
		event.Prontuario = *r.Prontuario
	}

	// Unknown identification
	event.IdentificacaoDesconhecida = r.IdentificacaoDesconhecida == "S"

	return event
}

// calculateAge computes age at time of death
func calculateAge(birthDate, deathDate time.Time) int {
	years := deathDate.Year() - birthDate.Year()

	// Adjust if birthday hasn't occurred yet in the year of death
	if deathDate.YearDay() < birthDate.YearDay() {
		years--
	}

	return years
}

// MaskCPF masks a CPF number for LGPD compliance
// Example: "123.456.789-10" -> "***.***.***-10"
// Example: "12345678910" -> "***.***.***-10"
func MaskCPF(cpf string) string {
	if cpf == "" {
		return ""
	}

	// Remove formatting
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, cpf)

	if len(cleaned) != 11 {
		// Invalid CPF, mask completely
		return "***.***.***-**"
	}

	// Keep only last 2 digits visible
	return "***.***.***-" + cleaned[9:]
}

// MaskName masks a name for LGPD compliance (for logging purposes)
// Example: "Joao Silva" -> "Jo** Si***"
func MaskName(name string) string {
	if name == "" {
		return ""
	}

	words := strings.Fields(name)
	maskedWords := make([]string, len(words))

	for i, word := range words {
		maskedWords[i] = maskWord(word)
	}

	return strings.Join(maskedWords, " ")
}

// maskWord masks a single word, keeping only the first 2 characters visible
func maskWord(word string) string {
	if word == "" {
		return ""
	}

	runes := []rune(word)
	length := len(runes)

	if length <= 2 {
		if length == 1 {
			return string(runes[0])
		}
		return string(runes[0]) + "*"
	}

	// Keep first 2 characters, mask the rest
	visiblePart := string(runes[:2])
	maskedPart := strings.Repeat("*", length-2)

	return visiblePart + maskedPart
}

// AgentState persists the watermark for resuming after restart
type AgentState struct {
	LastProcessedID string    `json:"last_processed_id"`
	LastProcessedAt time.Time `json:"last_processed_at"`
	TotalProcessed  int64     `json:"total_processed"`
	LastError       string    `json:"last_error,omitempty"`
	LastErrorAt     time.Time `json:"last_error_at,omitempty"`
}
