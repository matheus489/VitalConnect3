package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment or use default
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/vitalconnect?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Get migrations directory
	migrationsDir := "."
	if len(os.Args) > 1 {
		migrationsDir = os.Args[1]
	}

	// Find all SQL migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to find migration files: %v", err)
	}

	// Sort files to ensure correct order
	sort.Strings(files)

	// Run init.sql first if exists
	for i, f := range files {
		if strings.HasSuffix(f, "init.sql") {
			files = append([]string{f}, append(files[:i], files[i+1:]...)...)
			break
		}
	}

	fmt.Printf("Found %d migration files\n", len(files))

	// Execute each migration
	for _, file := range files {
		fmt.Printf("Running migration: %s\n", filepath.Base(file))

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// Skip DOWN sections and comments
		sql := extractUpMigration(string(content))

		if sql == "" {
			fmt.Printf("  Skipping (no UP migration found)\n")
			continue
		}

		_, err = db.Exec(sql)
		if err != nil {
			// Check if it's a "already exists" error
			if strings.Contains(err.Error(), "already exists") ||
			   strings.Contains(err.Error(), "duplicate") {
				fmt.Printf("  Already applied (skipping)\n")
				continue
			}
			log.Fatalf("Failed to execute migration %s: %v", file, err)
		}

		fmt.Printf("  Applied successfully\n")
	}

	fmt.Println("\nAll migrations completed successfully!")

	// Verify tables were created
	verifyTables(db)
}

// extractUpMigration extracts the UP migration section from the SQL file
func extractUpMigration(content string) string {
	// Remove comments starting with --
	lines := strings.Split(content, "\n")
	var result []string
	inDownSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the start
		if len(result) == 0 && trimmed == "" {
			continue
		}

		// Check for DOWN section marker
		if strings.HasPrefix(strings.ToUpper(trimmed), "-- DOWN") {
			inDownSection = true
			continue
		}

		// Skip DOWN section
		if inDownSection {
			continue
		}

		// Skip pure comment lines (but keep inline comments)
		if strings.HasPrefix(trimmed, "-- ") && !strings.Contains(line, ";") {
			continue
		}

		// Skip DROP statements (rollback)
		if strings.HasPrefix(strings.ToUpper(trimmed), "DROP ") {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// verifyTables checks that all expected tables were created
func verifyTables(db *sql.DB) {
	fmt.Println("\nVerifying created tables:")

	expectedTables := []string{
		"hospitals",
		"users",
		"obitos_simulados",
		"occurrences",
		"occurrence_history",
		"triagem_rules",
		"notifications",
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	existingTables := make(map[string]bool)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		existingTables[tableName] = true
	}

	allFound := true
	for _, table := range expectedTables {
		if existingTables[table] {
			fmt.Printf("  [OK] %s\n", table)
		} else {
			fmt.Printf("  [MISSING] %s\n", table)
			allFound = false
		}
	}

	if !allFound {
		fmt.Println("\nWarning: Some tables are missing!")
	} else {
		fmt.Println("\nAll expected tables exist!")
	}

	// Verify indexes
	fmt.Println("\nVerifying indexes:")
	verifyIndexes(db)
}

// verifyIndexes checks that key indexes were created
func verifyIndexes(db *sql.DB) {
	expectedIndexes := []string{
		"idx_hospitals_codigo",
		"idx_users_email",
		"idx_obitos_simulados_data_obito",
		"idx_occurrences_status",
		"idx_occurrence_history_occurrence_id",
		"idx_triagem_rules_regras",
		"idx_notifications_occurrence_id",
	}

	query := `
		SELECT indexname
		FROM pg_indexes
		WHERE schemaname = 'public'
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to query indexes: %v", err)
		return
	}
	defer rows.Close()

	existingIndexes := make(map[string]bool)
	for rows.Next() {
		var indexName string
		if err := rows.Scan(&indexName); err != nil {
			log.Printf("Failed to scan index name: %v", err)
			continue
		}
		existingIndexes[indexName] = true
	}

	for _, idx := range expectedIndexes {
		if existingIndexes[idx] {
			fmt.Printf("  [OK] %s\n", idx)
		} else {
			fmt.Printf("  [MISSING] %s\n", idx)
		}
	}
}
