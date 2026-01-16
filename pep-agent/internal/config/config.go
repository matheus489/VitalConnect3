// Package config provides YAML configuration parsing for the PEP Agent.
// It supports environment variable substitution for sensitive values.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// AgentConfig is the root configuration structure for the PEP Agent
type AgentConfig struct {
	Database DatabaseConfig `yaml:"database"`
	Mapping  MappingConfig  `yaml:"mapping"`
	Central  CentralConfig  `yaml:"central"`
	Agent    AgentSettings  `yaml:"agent"`
}

// DatabaseConfig defines the connection to the hospital PEP database
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`   // postgres, mysql, oracle
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"` // supports ${ENV_VAR} syntax
	SSLMode  string `yaml:"ssl_mode"` // disable, require, verify-full
}

// MappingConfig defines the field mapping from PEP to VitalConnect
type MappingConfig struct {
	SourceTable  string       `yaml:"source_table"`  // e.g., "TASY.TB_PACIENTE_OBITO"
	Fields       FieldMapping `yaml:"fields"`
	FilterColumn string       `yaml:"filter_column"` // column for watermark filtering
	CustomQuery  string       `yaml:"custom_query"`  // optional: override auto-generated query
}

// FieldMapping maps PEP database columns to VitalConnect standard fields
type FieldMapping struct {
	// Required fields
	ID             string `yaml:"id"`              // unique identifier in source
	NomePaciente   string `yaml:"nome_paciente"`
	DataObito      string `yaml:"data_obito"`
	CausaMortis    string `yaml:"causa_mortis"`
	DataNascimento string `yaml:"data_nascimento"` // or use idade
	Idade          string `yaml:"idade"`           // alternative to data_nascimento

	// Patient identification (at least one required)
	CNS string `yaml:"cns"` // Cartao Nacional de Saude
	CPF string `yaml:"cpf"` // CPF (will be masked)

	// Optional fields
	Setor                     string `yaml:"setor"`
	Leito                     string `yaml:"leito"`
	Prontuario                string `yaml:"prontuario"`
	IdentificacaoDesconhecida string `yaml:"identificacao_desconhecida"` // 'S' or 'N'
}

// CentralConfig defines the connection to the VitalConnect central server
type CentralConfig struct {
	URL      string `yaml:"url"`      // e.g., "https://vitalconnect.example.com/api/v1/pep/eventos"
	APIKey   string `yaml:"api_key"`  // supports ${ENV_VAR} syntax
	Insecure bool   `yaml:"insecure"` // skip TLS verification (dev only)
	Timeout  string `yaml:"timeout"`  // request timeout (default: 30s)
}

// AgentSettings defines operational parameters for the agent
type AgentSettings struct {
	HospitalID     string `yaml:"hospital_id"`     // UUID of the hospital in VitalConnect
	PollInterval   string `yaml:"poll_interval"`   // polling interval (default: 3s)
	StateFile      string `yaml:"state_file"`      // path to watermark state file
	LogLevel       string `yaml:"log_level"`       // debug, info, warn, error
	AlertThreshold string `yaml:"alert_threshold"` // offline duration to trigger alert (default: 10m)
}

// Load reads and parses a YAML configuration file
func Load(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Substitute environment variables
	content := substituteEnvVars(string(data))

	var config AgentConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	config.SetDefaults()

	return &config, nil
}

// substituteEnvVars replaces ${VAR_NAME} with environment variable values
func substituteEnvVars(content string) string {
	re := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		varName := match[2 : len(match)-1] // Extract VAR_NAME from ${VAR_NAME}
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Keep original if env var not set
	})
}

// Validate checks that all required configuration fields are present
func (c *AgentConfig) Validate() error {
	// Database validation
	if c.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}
	validDrivers := map[string]bool{"postgres": true, "mysql": true, "oracle": true}
	if !validDrivers[c.Database.Driver] {
		return fmt.Errorf("database.driver must be one of: postgres, mysql, oracle")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}

	// Mapping validation
	if c.Mapping.SourceTable == "" && c.Mapping.CustomQuery == "" {
		return fmt.Errorf("mapping.source_table or mapping.custom_query is required")
	}

	// Field mapping validation (required fields)
	if c.Mapping.Fields.ID == "" {
		return fmt.Errorf("mapping.fields.id is required")
	}
	if c.Mapping.Fields.NomePaciente == "" {
		return fmt.Errorf("mapping.fields.nome_paciente is required")
	}
	if c.Mapping.Fields.DataObito == "" {
		return fmt.Errorf("mapping.fields.data_obito is required")
	}
	if c.Mapping.Fields.CausaMortis == "" {
		return fmt.Errorf("mapping.fields.causa_mortis is required")
	}
	if c.Mapping.Fields.DataNascimento == "" && c.Mapping.Fields.Idade == "" {
		return fmt.Errorf("mapping.fields.data_nascimento or mapping.fields.idade is required")
	}

	// Central validation
	if c.Central.URL == "" {
		return fmt.Errorf("central.url is required")
	}
	if c.Central.APIKey == "" {
		return fmt.Errorf("central.api_key is required")
	}

	// Agent validation
	if c.Agent.HospitalID == "" {
		return fmt.Errorf("agent.hospital_id is required")
	}

	return nil
}

// SetDefaults sets default values for optional configuration fields
func (c *AgentConfig) SetDefaults() {
	if c.Database.Port == 0 {
		switch c.Database.Driver {
		case "postgres":
			c.Database.Port = 5432
		case "mysql":
			c.Database.Port = 3306
		case "oracle":
			c.Database.Port = 1521
		}
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Agent.PollInterval == "" {
		c.Agent.PollInterval = "3s"
	}
	if c.Agent.StateFile == "" {
		c.Agent.StateFile = "/var/lib/pep-agent/state.json"
	}
	if c.Agent.LogLevel == "" {
		c.Agent.LogLevel = "info"
	}
	if c.Agent.AlertThreshold == "" {
		c.Agent.AlertThreshold = "10m"
	}
	if c.Central.Timeout == "" {
		c.Central.Timeout = "30s"
	}
	if c.Mapping.FilterColumn == "" {
		c.Mapping.FilterColumn = c.Mapping.Fields.DataObito
	}
}

// GetPollInterval returns the poll interval as a time.Duration
func (c *AgentConfig) GetPollInterval() time.Duration {
	d, err := time.ParseDuration(c.Agent.PollInterval)
	if err != nil {
		return 3 * time.Second
	}
	return d
}

// GetAlertThreshold returns the alert threshold as a time.Duration
func (c *AgentConfig) GetAlertThreshold() time.Duration {
	d, err := time.ParseDuration(c.Agent.AlertThreshold)
	if err != nil {
		return 10 * time.Minute
	}
	return d
}

// GetTimeout returns the HTTP timeout as a time.Duration
func (c *AgentConfig) GetTimeout() time.Duration {
	d, err := time.ParseDuration(c.Central.Timeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// GetDSN returns the database connection string
func (c *AgentConfig) GetDSN() string {
	switch c.Database.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.User,
			c.Database.Password,
			c.Database.Database,
			c.Database.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s",
			c.Database.User,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
		)
	case "oracle":
		return fmt.Sprintf(
			"%s/%s@%s:%d/%s",
			c.Database.User,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
		)
	default:
		return ""
	}
}

// BuildSelectQuery generates the SQL SELECT query based on field mappings
func (c *AgentConfig) BuildSelectQuery(watermark string) string {
	if c.Mapping.CustomQuery != "" {
		// Use custom query with watermark placeholder
		return strings.ReplaceAll(c.Mapping.CustomQuery, "{{WATERMARK}}", watermark)
	}

	// Build field list
	fields := []string{
		c.Mapping.Fields.ID,
		c.Mapping.Fields.NomePaciente,
		c.Mapping.Fields.DataObito,
		c.Mapping.Fields.CausaMortis,
	}

	if c.Mapping.Fields.DataNascimento != "" {
		fields = append(fields, c.Mapping.Fields.DataNascimento)
	}
	if c.Mapping.Fields.Idade != "" {
		fields = append(fields, c.Mapping.Fields.Idade)
	}
	if c.Mapping.Fields.CNS != "" {
		fields = append(fields, c.Mapping.Fields.CNS)
	}
	if c.Mapping.Fields.CPF != "" {
		fields = append(fields, c.Mapping.Fields.CPF)
	}
	if c.Mapping.Fields.Setor != "" {
		fields = append(fields, c.Mapping.Fields.Setor)
	}
	if c.Mapping.Fields.Leito != "" {
		fields = append(fields, c.Mapping.Fields.Leito)
	}
	if c.Mapping.Fields.Prontuario != "" {
		fields = append(fields, c.Mapping.Fields.Prontuario)
	}
	if c.Mapping.Fields.IdentificacaoDesconhecida != "" {
		fields = append(fields, c.Mapping.Fields.IdentificacaoDesconhecida)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s > '%s' ORDER BY %s ASC",
		strings.Join(fields, ", "),
		c.Mapping.SourceTable,
		c.Mapping.FilterColumn,
		watermark,
		c.Mapping.FilterColumn,
	)

	return query
}
