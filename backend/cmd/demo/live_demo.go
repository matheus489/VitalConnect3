package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sidot/backend/config"
)

func main() {
	log.Println("========================================")
	log.Println("SIDOT - Live Demo Script")
	log.Println("========================================")

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

	log.Println("Database connected successfully!")

	// Generate new obito ID
	obitoID := uuid.New()
	hggID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// Patient data for demo
	patientName := "Carlos Eduardo Martins"
	setor := "UTI"
	leito := "5"
	prontuario := fmt.Sprintf("HGG-DEMO-%d", time.Now().Unix())

	// Calculate scheduled time
	scheduledTime := time.Now().Add(10 * time.Second)

	log.Println("")
	log.Println("========================================")
	log.Println("LIVE DEMO SCHEDULED!")
	log.Println("========================================")
	log.Printf("Obito ID: %s", obitoID)
	log.Println("Hospital: HGG - Hospital Geral de Goiania")
	log.Printf("Setor: %s", setor)
	log.Printf("Leito: %s", leito)
	log.Printf("Paciente: %s", patientName)
	log.Printf("Prontuario: %s", prontuario)
	log.Println("")
	log.Printf("Current time: %s", time.Now().Format("15:04:05"))
	log.Printf("Scheduled for: %s", scheduledTime.Format("15:04:05"))
	log.Println("========================================")
	log.Println("")
	log.Println("The system will detect this obito in ~10 seconds.")
	log.Println("Watch the dashboard for real-time notification!")
	log.Println("")
	log.Println("========================================")
	log.Println("Countdown:")

	// Countdown
	for i := 10; i > 0; i-- {
		log.Printf("  %d...", i)
		time.Sleep(1 * time.Second)
	}

	log.Println("")
	log.Println(">>> INSERTING OBITO NOW!")
	log.Println("")

	// Insert obito
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO obitos_simulados (
			id, hospital_id, nome_paciente, data_nascimento, data_obito,
			causa_mortis, prontuario, setor, leito, identificacao_desconhecida,
			processado, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, false, NOW())
	`,
		obitoID,
		hggID,
		patientName,
		time.Date(1965, 8, 12, 0, 0, 0, 0, time.UTC),
		time.Now(), // Death just happened
		"Infarto Fulminante",
		prontuario,
		setor,
		leito,
		false,
	)

	if err != nil {
		log.Fatalf("ERROR: Failed to insert obito: %v", err)
	}

	log.Println("========================================")
	log.Println("SUCCESS! Obito inserted at:", time.Now().Format("15:04:05"))
	log.Println("========================================")
	log.Println("")
	log.Println("The listener service should detect this obito within 3-5 seconds.")
	log.Println("Watch the dashboard for:")
	log.Println("  - Badge piscando (red notification badge)")
	log.Println("  - Toast notification with obito details")
	log.Println("  - Sound alert (if enabled)")
	log.Println("  - New occurrence in the table")
	log.Println("")
	log.Println("========================================")
}
