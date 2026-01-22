package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sidot/backend/config"
	"golang.org/x/crypto/bcrypt"
)

// Seed data constants
const (
	HGGCode  = "HGG"
	HUGOCode = "HUGO"
)

func main() {
	// Parse flags
	var (
		clearData     = flag.Bool("clear", false, "Clear existing seed data before inserting")
		hospitalsOnly = flag.Bool("hospitals", false, "Seed only hospitals")
		usersOnly     = flag.Bool("users", false, "Seed only users")
		rulesOnly     = flag.Bool("rules", false, "Seed only triagem rules")
		obitosOnly    = flag.Bool("obitos", false, "Seed only obitos")
		liveDemo      = flag.Bool("live-demo", false, "Insert live demo obito (T+10 seconds)")
		metricsDemo   = flag.Bool("metrics", false, "Seed 30 days of metrics data for dashboard demo")
		clearMetrics  = flag.Bool("clear-metrics", false, "Clear existing metrics seed data")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()

	// Clear data if requested
	if *clearData {
		log.Println("Clearing existing seed data...")
		if err := clearSeedData(ctx, db); err != nil {
			log.Printf("Warning: Failed to clear some data: %v", err)
		}
	}

	// Clear metrics data if requested
	if *clearMetrics {
		if err := ClearMetricsData(ctx, db); err != nil {
			log.Printf("Warning: Failed to clear metrics data: %v", err)
		}
	}

	// Seed based on flags or all if no specific flag
	seedAll := !*hospitalsOnly && !*usersOnly && !*rulesOnly && !*obitosOnly && !*liveDemo && !*metricsDemo && !*clearMetrics

	if seedAll || *hospitalsOnly {
		log.Println("Seeding hospitals...")
		if err := seedHospitals(ctx, db); err != nil {
			log.Fatalf("Failed to seed hospitals: %v", err)
		}
		log.Println("Hospitals seeded successfully!")
	}

	if seedAll || *usersOnly {
		log.Println("Seeding users...")
		if err := seedUsers(ctx, db); err != nil {
			log.Fatalf("Failed to seed users: %v", err)
		}
		log.Println("Users seeded successfully!")
	}

	if seedAll || *rulesOnly {
		log.Println("Seeding triagem rules...")
		if err := seedTriagemRules(ctx, db); err != nil {
			log.Fatalf("Failed to seed triagem rules: %v", err)
		}
		log.Println("Triagem rules seeded successfully!")
	}

	if seedAll || *obitosOnly {
		log.Println("Seeding demo obitos...")
		if err := seedDemoObitos(ctx, db); err != nil {
			log.Fatalf("Failed to seed obitos: %v", err)
		}
		log.Println("Demo obitos seeded successfully!")
	}

	if *liveDemo {
		log.Println("Scheduling live demo obito...")
		if err := seedLiveDemoObito(ctx, db); err != nil {
			log.Fatalf("Failed to seed live demo obito: %v", err)
		}
	}

	// Seed metrics data for dashboard demo
	if *metricsDemo {
		log.Println("\n========================================")
		log.Println("SEEDING METRICS DATA FOR DASHBOARD DEMO")
		log.Println("========================================")
		config := DefaultMetricsSeedConfig()
		if err := SeedMetricsData(ctx, db, config); err != nil {
			log.Fatalf("Failed to seed metrics data: %v", err)
		}
		log.Println("\n========================================")
		log.Println("Metrics data seeded successfully!")
		log.Println("Dashboard indicators are now populated with 30 days of data.")
		log.Println("========================================")
	}

	if seedAll {
		log.Println("\n========================================")
		log.Println("SIDOT seed data created successfully!")
		log.Println("========================================")
		log.Println("\nTest credentials:")
		log.Println("  Admin:    admin@sidot.gov.br / demo123")
		log.Println("  Gestor:   gestor@sidot.gov.br / demo123")
		log.Println("  Operador: operador@sidot.gov.br / demo123")
		log.Println("\nHospitals:")
		log.Println("  HGG:  Hospital Geral de Goiania")
		log.Println("  HUGO: Hospital de Urgencias de Goias")
		log.Println("\nTo seed 30 days of metrics data for the dashboard demo, run:")
		log.Println("  go run ./cmd/seeder -metrics")
	}
}

// clearSeedData removes existing seed data
func clearSeedData(ctx context.Context, db *sql.DB) error {
	// Order matters due to foreign keys
	tables := []string{
		"notifications",
		"occurrence_history",
		"occurrences",
		"obitos_simulados",
		"triagem_rules",
		"user_hospitals",
		"users",
		"hospitals",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Printf("Warning: Failed to clear %s: %v", table, err)
		}
	}

	return nil
}

// seedHospitals creates the demo hospitals
func seedHospitals(ctx context.Context, db *sql.DB) error {
	hospitals := []struct {
		ID       uuid.UUID
		Nome     string
		Codigo   string
		Endereco string
		Config   map[string]interface{}
	}{
		{
			ID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Nome:     "Hospital Geral de Goiania",
			Codigo:   HGGCode,
			Endereco: "Av. Anhanguera, 6479 - St. Oeste, Goiania - GO, 74110-010",
			Config: map[string]interface{}{
				"tipo":          "simulado",
				"host":          "localhost",
				"port":          5432,
				"database":      "hgg_pep",
				"poll_interval": 3,
			},
		},
		{
			ID:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Nome:     "Hospital de Urgencias de Goias",
			Codigo:   HUGOCode,
			Endereco: "Av. 31 de Marco, s/n - St. Pedro Ludovico, Goiania - GO, 74820-300",
			Config: map[string]interface{}{
				"tipo":          "simulado",
				"host":          "localhost",
				"port":          5432,
				"database":      "hugo_pep",
				"poll_interval": 3,
			},
		},
	}

	for _, h := range hospitals {
		configJSON, _ := json.Marshal(h.Config)

		// Check if hospital already exists
		var exists bool
		err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM hospitals WHERE codigo = $1 AND deleted_at IS NULL)`, h.Codigo).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check hospital %s: %w", h.Codigo, err)
		}

		if exists {
			_, err = db.ExecContext(ctx, `
				UPDATE hospitals SET nome = $1, endereco = $2, config_conexao = $3, updated_at = NOW()
				WHERE codigo = $4 AND deleted_at IS NULL
			`, h.Nome, h.Endereco, configJSON, h.Codigo)
		} else {
			_, err = db.ExecContext(ctx, `
				INSERT INTO hospitals (id, nome, codigo, endereco, config_conexao, ativo, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			`, h.ID, h.Nome, h.Codigo, h.Endereco, configJSON)
		}

		if err != nil {
			return fmt.Errorf("failed to insert hospital %s: %w", h.Codigo, err)
		}
		log.Printf("  Created hospital: %s (%s)", h.Nome, h.Codigo)
	}

	return nil
}

// seedUsers creates the demo users
func seedUsers(ctx context.Context, db *sql.DB) error {
	// Hash default password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	hggID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hugoID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	users := []struct {
		ID          uuid.UUID
		Email       string
		Nome        string
		Role        string
		HospitalIDs []uuid.UUID // Multiple hospitals via user_hospitals table
	}{
		{
			ID:          uuid.MustParse("aaaa0000-0000-0000-0000-000000000001"),
			Email:       "admin@sidot.gov.br",
			Nome:        "Administrador Sistema",
			Role:        "admin",
			HospitalIDs: []uuid.UUID{hggID, hugoID}, // Admin has access to all hospitals
		},
		{
			ID:          uuid.MustParse("aaaa0000-0000-0000-0000-000000000002"),
			Email:       "gestor@sidot.gov.br",
			Nome:        "Gestor Central Transplantes",
			Role:        "gestor",
			HospitalIDs: []uuid.UUID{hggID, hugoID}, // Gestor has access to all hospitals
		},
		{
			ID:          uuid.MustParse("aaaa0000-0000-0000-0000-000000000003"),
			Email:       "operador@sidot.gov.br",
			Nome:        "Operador Plantao",
			Role:        "operador",
			HospitalIDs: []uuid.UUID{hggID}, // Operador only has access to HGG
		},
	}

	for _, u := range users {
		// Insert user without hospital_id (now uses user_hospitals table)
		_, err := db.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, nome, role, ativo, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			ON CONFLICT (email) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				nome = EXCLUDED.nome,
				role = EXCLUDED.role,
				updated_at = NOW()
		`, u.ID, u.Email, string(passwordHash), u.Nome, u.Role)

		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", u.Email, err)
		}
		log.Printf("  Created user: %s (%s)", u.Email, u.Role)

		// Create user_hospitals associations
		for _, hospitalID := range u.HospitalIDs {
			_, err := db.ExecContext(ctx, `
				INSERT INTO user_hospitals (user_id, hospital_id, created_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT (user_id, hospital_id) DO NOTHING
			`, u.ID, hospitalID)

			if err != nil {
				return fmt.Errorf("failed to associate user %s with hospital %s: %w", u.Email, hospitalID, err)
			}
		}
	}

	return nil
}

// seedTriagemRules creates the demo triagem rules
func seedTriagemRules(ctx context.Context, db *sql.DB) error {
	rules := []struct {
		ID         uuid.UUID
		Nome       string
		Descricao  string
		Regras     map[string]interface{}
		Prioridade int
	}{
		{
			ID:        uuid.MustParse("bbbb0000-0000-0000-0000-000000000001"),
			Nome:      "Idade Maxima",
			Descricao: "Rejeita pacientes com idade acima de 80 anos",
			Regras: map[string]interface{}{
				"tipo":  "idade_maxima",
				"valor": 80,
				"acao":  "rejeitar",
			},
			Prioridade: 100,
		},
		{
			ID:        uuid.MustParse("bbbb0000-0000-0000-0000-000000000002"),
			Nome:      "Causas Excludentes",
			Descricao: "Rejeita pacientes com causas de obito que impedem doacao",
			Regras: map[string]interface{}{
				"tipo": "causas_excludentes",
				"valor": []string{
					"sepse",
					"meningite",
					"tuberculose",
					"hepatite b",
					"hepatite c",
					"hiv",
					"aids",
					"raiva",
					"doenca de creutzfeldt-jakob",
					"leucemia",
					"linfoma",
				},
				"acao": "rejeitar",
			},
			Prioridade: 90,
		},
		{
			ID:        uuid.MustParse("bbbb0000-0000-0000-0000-000000000003"),
			Nome:      "Janela de Captacao",
			Descricao: "Rejeita pacientes fora da janela de 6 horas",
			Regras: map[string]interface{}{
				"tipo":  "janela_horas",
				"valor": 6,
				"acao":  "rejeitar",
			},
			Prioridade: 80,
		},
		{
			ID:        uuid.MustParse("bbbb0000-0000-0000-0000-000000000004"),
			Nome:      "Identificacao Desconhecida",
			Descricao: "Rejeita pacientes com identificacao desconhecida (indigentes)",
			Regras: map[string]interface{}{
				"tipo":  "identificacao_desconhecida",
				"valor": true,
				"acao":  "rejeitar",
			},
			Prioridade: 70,
		},
	}

	for _, r := range rules {
		regrasJSON, _ := json.Marshal(r.Regras)

		_, err := db.ExecContext(ctx, `
			INSERT INTO triagem_rules (id, nome, descricao, regras, ativo, prioridade, created_at, updated_at)
			VALUES ($1, $2, $3, $4, true, $5, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET
				nome = EXCLUDED.nome,
				descricao = EXCLUDED.descricao,
				regras = EXCLUDED.regras,
				prioridade = EXCLUDED.prioridade,
				updated_at = NOW()
		`, r.ID, r.Nome, r.Descricao, regrasJSON, r.Prioridade)

		if err != nil {
			return fmt.Errorf("failed to insert rule %s: %w", r.Nome, err)
		}
		log.Printf("  Created rule: %s (prioridade: %d)", r.Nome, r.Prioridade)
	}

	return nil
}

// seedDemoObitos creates demo death records for demonstration
func seedDemoObitos(ctx context.Context, db *sql.DB) error {
	hggID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hugoID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	obitos := []struct {
		ID                        uuid.UUID
		HospitalID                uuid.UUID
		NomePaciente              string
		DataNascimento            time.Time
		DataObito                 time.Time
		CausaMortis               string
		Prontuario                string
		Setor                     string
		Leito                     string
		IdentificacaoDesconhecida bool
		Description               string // For logging
	}{
		{
			ID:                        uuid.MustParse("cccc0000-0000-0000-0000-000000000001"),
			HospitalID:                hggID,
			NomePaciente:              "Jose Carlos Oliveira",
			DataNascimento:            time.Date(1955, 3, 15, 0, 0, 0, 0, time.UTC),
			DataObito:                 time.Now().Add(-1 * time.Hour),
			CausaMortis:               "Infarto Agudo do Miocardio",
			Prontuario:                "HGG-2026-001234",
			Setor:                     "UTI",
			Leito:                     "3",
			IdentificacaoDesconhecida: false,
			Description:               "Elegivel - UTI, 1h atras",
		},
		{
			ID:                        uuid.MustParse("cccc0000-0000-0000-0000-000000000002"),
			HospitalID:                hugoID,
			NomePaciente:              "Maria Helena Santos",
			DataNascimento:            time.Date(1948, 7, 22, 0, 0, 0, 0, time.UTC),
			DataObito:                 time.Now().Add(-3 * time.Hour),
			CausaMortis:               "Acidente Vascular Cerebral",
			Prontuario:                "HUGO-2026-005678",
			Setor:                     "Emergencia",
			Leito:                     "12",
			IdentificacaoDesconhecida: false,
			Description:               "Elegivel - Emergencia, 3h atras",
		},
		{
			ID:                        uuid.MustParse("cccc0000-0000-0000-0000-000000000003"),
			HospitalID:                hggID,
			NomePaciente:              "Antonio Pereira Lima",
			DataNascimento:            time.Date(1940, 11, 5, 0, 0, 0, 0, time.UTC),
			DataObito:                 time.Now().Add(-5 * time.Hour),
			CausaMortis:               "Insuficiencia Cardiaca Congestiva",
			Prontuario:                "HGG-2026-002345",
			Setor:                     "Enfermaria",
			Leito:                     "8B",
			IdentificacaoDesconhecida: false,
			Description:               "INELEGIVEL - Idade > 80 anos",
		},
		{
			ID:                        uuid.MustParse("cccc0000-0000-0000-0000-000000000004"),
			HospitalID:                hugoID,
			NomePaciente:              "Francisca Souza Costa",
			DataNascimento:            time.Date(1960, 2, 28, 0, 0, 0, 0, time.UTC),
			DataObito:                 time.Now().Add(-8 * time.Hour),
			CausaMortis:               "Traumatismo Craniano Grave",
			Prontuario:                "HUGO-2026-007890",
			Setor:                     "UTI",
			Leito:                     "1",
			IdentificacaoDesconhecida: false,
			Description:               "INELEGIVEL - Fora da janela 6h",
		},
		{
			ID:                        uuid.MustParse("cccc0000-0000-0000-0000-000000000005"),
			HospitalID:                hggID,
			NomePaciente:              "Desconhecido - Indigente",
			DataNascimento:            time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), // Estimated
			DataObito:                 time.Now().Add(-2 * time.Hour),
			CausaMortis:               "Parada Cardiorrespiratoria",
			Prontuario:                "HGG-2026-003456",
			Setor:                     "Emergencia",
			Leito:                     "5",
			IdentificacaoDesconhecida: true,
			Description:               "INELEGIVEL - Identificacao desconhecida",
		},
	}

	for _, o := range obitos {
		_, err := db.ExecContext(ctx, `
			INSERT INTO obitos_simulados (
				id, hospital_id, nome_paciente, data_nascimento, data_obito,
				causa_mortis, prontuario, setor, leito, identificacao_desconhecida,
				processado, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, false, NOW())
			ON CONFLICT (id) DO UPDATE SET
				nome_paciente = EXCLUDED.nome_paciente,
				data_obito = EXCLUDED.data_obito,
				causa_mortis = EXCLUDED.causa_mortis,
				processado = false
		`, o.ID, o.HospitalID, o.NomePaciente, o.DataNascimento, o.DataObito,
			o.CausaMortis, o.Prontuario, o.Setor, o.Leito, o.IdentificacaoDesconhecida)

		if err != nil {
			return fmt.Errorf("failed to insert obito: %w", err)
		}
		log.Printf("  Created obito: %s - %s", o.NomePaciente, o.Description)
	}

	return nil
}

// seedLiveDemoObito creates an obito that will be detected in real-time during demo
func seedLiveDemoObito(ctx context.Context, db *sql.DB) error {
	hggID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	obitoID := uuid.New()

	// Schedule for 10 seconds in the future
	scheduledTime := time.Now().Add(10 * time.Second)

	log.Printf("\n========================================")
	log.Printf("LIVE DEMO SCHEDULED!")
	log.Printf("========================================")
	log.Printf("Obito ID: %s", obitoID)
	log.Printf("Hospital: HGG - Hospital Geral de Goiania")
	log.Printf("Setor: UTI")
	log.Printf("Paciente: Carlos Eduardo Martins")
	log.Printf("Scheduled for: %s", scheduledTime.Format("15:04:05"))
	log.Printf("========================================")
	log.Printf("The system will detect this obito in ~10 seconds.")
	log.Printf("Watch the dashboard for real-time notification!")
	log.Printf("========================================\n")

	// Insert with data_obito = now (so it's within the 6h window)
	// The listener checks created_at for new records
	go func() {
		time.Sleep(10 * time.Second)

		_, err := db.ExecContext(context.Background(), `
			INSERT INTO obitos_simulados (
				id, hospital_id, nome_paciente, data_nascimento, data_obito,
				causa_mortis, prontuario, setor, leito, identificacao_desconhecida,
				processado, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, false, NOW())
		`,
			obitoID,
			hggID,
			"Carlos Eduardo Martins",
			time.Date(1965, 8, 12, 0, 0, 0, 0, time.UTC),
			time.Now(), // Death just happened
			"Infarto Fulminante",
			"HGG-DEMO-999999",
			"UTI",
			"5",
			false,
		)

		if err != nil {
			log.Printf("ERROR: Failed to insert live demo obito: %v", err)
		} else {
			log.Printf("\n>>> LIVE DEMO: Obito inserted at %s", time.Now().Format("15:04:05"))
			log.Printf(">>> Check the dashboard for the notification!")
		}
	}()

	// Keep main goroutine alive to wait for insertion
	log.Println("Waiting for scheduled insertion...")
	time.Sleep(15 * time.Second)

	return nil
}
