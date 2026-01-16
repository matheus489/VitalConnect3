package models

import (
	"time"

	"github.com/google/uuid"
)

// ObitoSimulado represents a simulated death record from hospital systems
type ObitoSimulado struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	HospitalID              uuid.UUID  `json:"hospital_id" db:"hospital_id" validate:"required"`
	NomePaciente            string     `json:"nome_paciente" db:"nome_paciente" validate:"required,min=2,max=255"`
	DataNascimento          time.Time  `json:"data_nascimento" db:"data_nascimento" validate:"required"`
	DataObito               time.Time  `json:"data_obito" db:"data_obito" validate:"required"`
	CausaMortis             string     `json:"causa_mortis" db:"causa_mortis" validate:"required,min=2,max=500"`
	Prontuario              *string    `json:"prontuario,omitempty" db:"prontuario"`
	Setor                   *string    `json:"setor,omitempty" db:"setor"`
	Leito                   *string    `json:"leito,omitempty" db:"leito"`
	IdentificacaoDesconhecida bool     `json:"identificacao_desconhecida" db:"identificacao_desconhecida"`
	Processado              bool       `json:"processado" db:"processado"`
	ProcessadoEm            *time.Time `json:"processado_em,omitempty" db:"processado_em"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`

	// Related data (populated by queries)
	Hospital *Hospital `json:"hospital,omitempty" db:"-"`
}

// CreateObitoInput represents input for creating a simulated death record
type CreateObitoInput struct {
	HospitalID                uuid.UUID `json:"hospital_id" validate:"required"`
	NomePaciente              string    `json:"nome_paciente" validate:"required,min=2,max=255"`
	DataNascimento            time.Time `json:"data_nascimento" validate:"required"`
	DataObito                 time.Time `json:"data_obito" validate:"required"`
	CausaMortis               string    `json:"causa_mortis" validate:"required,min=2,max=500"`
	Prontuario                *string   `json:"prontuario,omitempty" validate:"omitempty,max=50"`
	Setor                     *string   `json:"setor,omitempty" validate:"omitempty,max=100"`
	Leito                     *string   `json:"leito,omitempty" validate:"omitempty,max=50"`
	IdentificacaoDesconhecida bool      `json:"identificacao_desconhecida"`
}

// CalculateAge returns the age of the patient at the time of death
func (o *ObitoSimulado) CalculateAge() int {
	years := o.DataObito.Year() - o.DataNascimento.Year()

	// Adjust if birthday hasn't occurred yet in the year of death
	if o.DataObito.YearDay() < o.DataNascimento.YearDay() {
		years--
	}

	return years
}

// IsWithinWindow checks if the death is within the 6-hour capture window
func (o *ObitoSimulado) IsWithinWindow(windowHours int) bool {
	deadline := o.DataObito.Add(time.Duration(windowHours) * time.Hour)
	return time.Now().Before(deadline)
}

// TimeRemaining returns the time remaining in the capture window
func (o *ObitoSimulado) TimeRemaining(windowHours int) time.Duration {
	deadline := o.DataObito.Add(time.Duration(windowHours) * time.Hour)
	remaining := deadline.Sub(time.Now())

	if remaining < 0 {
		return 0
	}

	return remaining
}

// ToOccurrenceData converts ObitoSimulado to data suitable for occurrence creation
func (o *ObitoSimulado) ToOccurrenceData() map[string]interface{} {
	data := map[string]interface{}{
		"obito_id":         o.ID,
		"hospital_id":      o.HospitalID,
		"nome_paciente":    o.NomePaciente,
		"data_nascimento":  o.DataNascimento,
		"data_obito":       o.DataObito,
		"causa_mortis":     o.CausaMortis,
		"idade":            o.CalculateAge(),
		"identificacao_desconhecida": o.IdentificacaoDesconhecida,
	}

	if o.Prontuario != nil {
		data["prontuario"] = *o.Prontuario
	}
	if o.Setor != nil {
		data["setor"] = *o.Setor
	}
	if o.Leito != nil {
		data["leito"] = *o.Leito
	}

	return data
}
