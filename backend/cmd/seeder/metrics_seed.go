package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// MetricsSeedConfig holds configuration for metrics seeding
type MetricsSeedConfig struct {
	DaysToGenerate int
	MinObitosPerDay int
	MaxObitosPerDay int
}

// DefaultMetricsSeedConfig returns the default configuration
func DefaultMetricsSeedConfig() MetricsSeedConfig {
	return MetricsSeedConfig{
		DaysToGenerate:  30,
		MinObitosPerDay: 3,
		MaxObitosPerDay: 8,
	}
}

// SeedMetricsData seeds 30 days of metrics data for demonstration
func SeedMetricsData(ctx context.Context, db *sql.DB, config MetricsSeedConfig) error {
	log.Println("Starting metrics seed...")

	// Get existing hospitals
	hospitals, err := getActiveHospitals(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get hospitals: %w", err)
	}

	if len(hospitals) == 0 {
		return fmt.Errorf("no active hospitals found - please run the main seeder first")
	}

	log.Printf("Found %d active hospitals", len(hospitals))

	// Generate seed for reproducibility (optional)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate data for the last 30 days
	now := time.Now()
	totalObitos := 0
	totalOccurrences := 0

	for dayOffset := config.DaysToGenerate; dayOffset >= 0; dayOffset-- {
		date := now.AddDate(0, 0, -dayOffset)
		obitosCount := rng.Intn(config.MaxObitosPerDay-config.MinObitosPerDay+1) + config.MinObitosPerDay

		for i := 0; i < obitosCount; i++ {
			// Select random hospital
			hospital := hospitals[rng.Intn(len(hospitals))]

			// Generate random time within the day
			hour := rng.Intn(24)
			minute := rng.Intn(60)
			obitoTime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, time.Local)

			// Skip if time is in the future
			if obitoTime.After(now) {
				continue
			}

			// Create obito and occurrence
			obitoID := uuid.New()
			occurrenceID := uuid.New()

			// Determine outcome - realistic distribution
			outcome := generateOutcome(rng)
			status := "CONCLUIDA"

			// Create obito simulado
			err := createObitoSimulado(ctx, db, obitoID, hospital.ID, obitoTime, rng)
			if err != nil {
				log.Printf("Warning: failed to create obito: %v", err)
				continue
			}
			totalObitos++

			// Create occurrence with realistic timestamps
			notificadoEm := obitoTime.Add(time.Duration(rng.Intn(60)+5) * time.Second) // 5-65 seconds after detection
			aceiteEm := notificadoEm.Add(time.Duration(rng.Intn(30)+2) * time.Minute)  // 2-32 minutes after notification
			conclusaoEm := aceiteEm.Add(time.Duration(rng.Intn(120)+30) * time.Minute) // 30-150 minutes after aceite

			err = createOccurrenceWithHistory(ctx, db, occurrenceID, obitoID, hospital.ID, hospital.Nome, obitoTime, notificadoEm, aceiteEm, conclusaoEm, status, outcome, rng)
			if err != nil {
				log.Printf("Warning: failed to create occurrence: %v", err)
				continue
			}
			totalOccurrences++
		}

		if dayOffset%7 == 0 {
			log.Printf("  Processed day %d/%d (date: %s)", config.DaysToGenerate-dayOffset+1, config.DaysToGenerate+1, date.Format("2006-01-02"))
		}
	}

	log.Printf("Metrics seed completed: %d obitos, %d occurrences created", totalObitos, totalOccurrences)
	return nil
}

// Hospital represents a hospital for seeding
type Hospital struct {
	ID   uuid.UUID
	Nome string
}

// getActiveHospitals returns all active hospitals from the database
func getActiveHospitals(ctx context.Context, db *sql.DB) ([]Hospital, error) {
	query := `SELECT id, nome FROM hospitals WHERE ativo = true AND deleted_at IS NULL`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []Hospital
	for rows.Next() {
		var h Hospital
		if err := rows.Scan(&h.ID, &h.Nome); err != nil {
			return nil, err
		}
		hospitals = append(hospitals, h)
	}

	return hospitals, rows.Err()
}

// generateOutcome generates a realistic outcome distribution
func generateOutcome(rng *rand.Rand) string {
	// Distribution: 45% sucesso, 25% familia_recusou, 15% contraindicacao, 10% tempo_excedido, 5% outro
	roll := rng.Intn(100)
	switch {
	case roll < 45:
		return "sucesso_captacao"
	case roll < 70:
		return "familia_recusou"
	case roll < 85:
		return "contraindicacao_medica"
	case roll < 95:
		return "tempo_excedido"
	default:
		return "outro"
	}
}

// createObitoSimulado creates a simulated death record
func createObitoSimulado(ctx context.Context, db *sql.DB, id, hospitalID uuid.UUID, dataObito time.Time, rng *rand.Rand) error {
	// Generate realistic patient data
	nomes := []string{
		"Jose Carlos Silva", "Maria Helena Santos", "Antonio Pereira Lima",
		"Francisca Souza Costa", "Joao Batista Oliveira", "Ana Paula Ferreira",
		"Pedro Henrique Alves", "Lucia Maria Rodrigues", "Carlos Eduardo Martins",
		"Sandra Regina Gomes", "Roberto Carlos Nunes", "Patricia Lima Souza",
	}

	causas := []string{
		"Infarto Agudo do Miocardio", "Acidente Vascular Cerebral", "Insuficiencia Cardiaca",
		"Traumatismo Craniano", "Parada Cardiorrespiratoria", "Insuficiencia Respiratoria",
		"Pneumonia Grave", "Arritmia Cardiaca", "Embolia Pulmonar",
	}

	setores := []string{"UTI", "Emergencia", "Enfermaria", "Centro Cirurgico", "Semi-Intensiva"}

	nome := nomes[rng.Intn(len(nomes))]
	causa := causas[rng.Intn(len(causas))]
	setor := setores[rng.Intn(len(setores))]
	idade := 25 + rng.Intn(55) // 25-79 years old
	dataNascimento := dataObito.AddDate(-idade, -rng.Intn(12), -rng.Intn(28))
	leito := fmt.Sprintf("%d", rng.Intn(20)+1)
	prontuario := fmt.Sprintf("PRO-%s-%06d", dataObito.Format("20060102"), rng.Intn(1000000))

	query := `
		INSERT INTO obitos_simulados (
			id, hospital_id, nome_paciente, data_nascimento, data_obito,
			causa_mortis, prontuario, setor, leito, identificacao_desconhecida,
			processado, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, true, $10)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := db.ExecContext(ctx, query,
		id, hospitalID, nome, dataNascimento, dataObito,
		causa, prontuario, setor, leito, dataObito,
	)
	return err
}

// createOccurrenceWithHistory creates an occurrence with full history
func createOccurrenceWithHistory(
	ctx context.Context, db *sql.DB,
	occurrenceID, obitoID, hospitalID uuid.UUID,
	hospitalNome string,
	dataObito, notificadoEm, aceiteEm, conclusaoEm time.Time,
	status, outcome string,
	rng *rand.Rand,
) error {
	// Generate masked name
	nomes := []string{
		"J.C.S.", "M.H.S.", "A.P.L.", "F.S.C.", "J.B.O.", "A.P.F.",
		"P.H.A.", "L.M.R.", "C.E.M.", "S.R.G.", "R.C.N.", "P.L.S.",
	}
	nomeMascarado := nomes[rng.Intn(len(nomes))]

	// Generate complete data
	setores := []string{"UTI", "Emergencia", "Enfermaria", "Centro Cirurgico", "Semi-Intensiva"}
	setor := setores[rng.Intn(len(setores))]

	dadosCompletos := map[string]interface{}{
		"obito_id":                  obitoID.String(),
		"hospital_id":               hospitalID.String(),
		"nome_paciente":             "Paciente " + nomeMascarado,
		"data_nascimento":           dataObito.AddDate(-50, 0, 0).Format(time.RFC3339),
		"data_obito":                dataObito.Format(time.RFC3339),
		"causa_mortis":              "Causa do obito",
		"idade":                     50,
		"prontuario":                fmt.Sprintf("PRO-%06d", rng.Intn(1000000)),
		"setor":                     setor,
		"leito":                     fmt.Sprintf("%d", rng.Intn(20)+1),
		"identificacao_desconhecida": false,
	}

	dadosJSON, err := json.Marshal(dadosCompletos)
	if err != nil {
		return fmt.Errorf("failed to marshal dados_completos: %w", err)
	}

	// Create occurrence
	janelaExpiraEm := dataObito.Add(6 * time.Hour)
	scorePriorizacao := 50 + rng.Intn(50) // 50-99

	occQuery := `
		INSERT INTO occurrences (
			id, obito_id, hospital_id, status, score_priorizacao,
			nome_paciente_mascarado, dados_completos, created_at, updated_at,
			notificado_em, data_obito, janela_expira_em
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO NOTHING
	`

	_, err = db.ExecContext(ctx, occQuery,
		occurrenceID, obitoID, hospitalID, status, scorePriorizacao,
		nomeMascarado, string(dadosJSON), dataObito, conclusaoEm,
		notificadoEm, dataObito, janelaExpiraEm,
	)
	if err != nil {
		return fmt.Errorf("failed to insert occurrence: %w", err)
	}

	// Create history entries
	histories := []struct {
		Acao      string
		Status    string
		Desfecho  *string
		CreatedAt time.Time
	}{
		{"Ocorrencia criada automaticamente", "PENDENTE", nil, dataObito},
		{"Notificacao enviada", "PENDENTE", nil, notificadoEm},
		{"Ocorrencia assumida", "EM_ANDAMENTO", nil, notificadoEm.Add(30 * time.Second)},
		{"Ocorrencia aceita", "ACEITA", nil, aceiteEm},
		{"Desfecho registrado", "ACEITA", &outcome, conclusaoEm.Add(-5 * time.Minute)},
		{"Ocorrencia concluida", "CONCLUIDA", nil, conclusaoEm},
	}

	for i, h := range histories {
		historyID := uuid.New()
		var statusAnterior *string
		var statusNovo *string

		if i > 0 {
			statusAnterior = &histories[i-1].Status
		}
		statusNovo = &h.Status

		histQuery := `
			INSERT INTO occurrence_history (
				id, occurrence_id, acao, status_anterior, status_novo, desfecho, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`

		_, err = db.ExecContext(ctx, histQuery,
			historyID, occurrenceID, h.Acao, statusAnterior, statusNovo, h.Desfecho, h.CreatedAt,
		)
		if err != nil {
			log.Printf("Warning: failed to insert history entry: %v", err)
		}
	}

	return nil
}

// ClearMetricsData clears existing metrics seed data
func ClearMetricsData(ctx context.Context, db *sql.DB) error {
	log.Println("Clearing existing metrics seed data...")

	// Delete only seeded occurrences (those with status CONCLUIDA and older than today)
	// This preserves any recent real data
	_, err := db.ExecContext(ctx, `
		DELETE FROM occurrence_history WHERE occurrence_id IN (
			SELECT id FROM occurrences WHERE status = 'CONCLUIDA' AND DATE(created_at) < CURRENT_DATE
		)
	`)
	if err != nil {
		log.Printf("Warning: failed to clear occurrence history: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		DELETE FROM occurrences WHERE status = 'CONCLUIDA' AND DATE(created_at) < CURRENT_DATE
	`)
	if err != nil {
		log.Printf("Warning: failed to clear occurrences: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		DELETE FROM obitos_simulados WHERE processado = true AND DATE(created_at) < CURRENT_DATE
	`)
	if err != nil {
		log.Printf("Warning: failed to clear obitos: %v", err)
	}

	log.Println("Metrics seed data cleared")
	return nil
}
