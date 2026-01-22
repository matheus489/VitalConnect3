// Package database provides multi-driver database connectivity for the PEP Agent
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	// Oracle driver requires CGO and godror, commented out for portability
	// _ "github.com/godror/godror"

	"github.com/sidot/pep-agent/internal/config"
	"github.com/sidot/pep-agent/internal/models"
)

// Connector provides database connectivity for the PEP Agent
type Connector struct {
	config *config.AgentConfig
	db     *sql.DB
}

// NewConnector creates a new database connector based on configuration
func NewConnector(cfg *config.AgentConfig) *Connector {
	return &Connector{
		config: cfg,
	}
}

// Connect establishes a connection to the PEP database
func (c *Connector) Connect() error {
	var driverName string
	switch c.config.Database.Driver {
	case "postgres":
		driverName = "postgres"
	case "mysql":
		driverName = "mysql"
	case "oracle":
		driverName = "godror"
	default:
		return fmt.Errorf("unsupported database driver: %s", c.config.Database.Driver)
	}

	dsn := c.config.GetDSN()

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool (conservative for read-only agent)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = db
	return nil
}

// Close closes the database connection
func (c *Connector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// IsConnected checks if the database connection is alive
func (c *Connector) IsConnected() bool {
	if c.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.db.PingContext(ctx) == nil
}

// FetchNewRecords fetches new death records since the watermark
func (c *Connector) FetchNewRecords(ctx context.Context, watermark string) ([]*models.PEPRecord, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := c.config.BuildSelectQuery(watermark)

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var records []*models.PEPRecord

	for rows.Next() {
		record, err := c.scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return records, nil
}

// scanRecord scans a database row into a PEPRecord based on field mappings
func (c *Connector) scanRecord(rows *sql.Rows) (*models.PEPRecord, error) {
	mapping := c.config.Mapping.Fields
	record := &models.PEPRecord{}

	// Build scan destinations based on configured fields
	var scanDest []interface{}
	var id, nomePaciente, causaMortis string
	var dataObito time.Time
	var dataNascimento sql.NullTime
	var idade sql.NullInt64
	var cns, cpf, setor, leito, prontuario, idDesconhecida sql.NullString

	// Required fields (always present in query)
	scanDest = append(scanDest, &id)
	scanDest = append(scanDest, &nomePaciente)
	scanDest = append(scanDest, &dataObito)
	scanDest = append(scanDest, &causaMortis)

	// Optional fields based on config
	if mapping.DataNascimento != "" {
		scanDest = append(scanDest, &dataNascimento)
	}
	if mapping.Idade != "" {
		scanDest = append(scanDest, &idade)
	}
	if mapping.CNS != "" {
		scanDest = append(scanDest, &cns)
	}
	if mapping.CPF != "" {
		scanDest = append(scanDest, &cpf)
	}
	if mapping.Setor != "" {
		scanDest = append(scanDest, &setor)
	}
	if mapping.Leito != "" {
		scanDest = append(scanDest, &leito)
	}
	if mapping.Prontuario != "" {
		scanDest = append(scanDest, &prontuario)
	}
	if mapping.IdentificacaoDesconhecida != "" {
		scanDest = append(scanDest, &idDesconhecida)
	}

	if err := rows.Scan(scanDest...); err != nil {
		return nil, err
	}

	// Populate record
	record.ID = id
	record.NomePaciente = nomePaciente
	record.DataObito = dataObito
	record.CausaMortis = causaMortis

	if mapping.DataNascimento != "" && dataNascimento.Valid {
		record.DataNascimento = &dataNascimento.Time
	}
	if mapping.Idade != "" && idade.Valid {
		idadeInt := int(idade.Int64)
		record.Idade = &idadeInt
	}
	if mapping.CNS != "" && cns.Valid {
		record.CNS = &cns.String
	}
	if mapping.CPF != "" && cpf.Valid {
		record.CPF = &cpf.String
	}
	if mapping.Setor != "" && setor.Valid {
		record.Setor = &setor.String
	}
	if mapping.Leito != "" && leito.Valid {
		record.Leito = &leito.String
	}
	if mapping.Prontuario != "" && prontuario.Valid {
		record.Prontuario = &prontuario.String
	}
	if mapping.IdentificacaoDesconhecida != "" && idDesconhecida.Valid {
		record.IdentificacaoDesconhecida = idDesconhecida.String
	}

	return record, nil
}

// GetDriver returns the configured database driver name
func (c *Connector) GetDriver() string {
	return c.config.Database.Driver
}

// GetDB returns the underlying sql.DB (for testing)
func (c *Connector) GetDB() *sql.DB {
	return c.db
}
