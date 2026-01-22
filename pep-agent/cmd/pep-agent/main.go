// PEP Agent - SIDOT Hospital Integration Agent
//
// This agent runs on-premise at hospital servers and polls the local PEP
// (Prontuario Eletronico do Paciente) database for death records.
// Detected events are pushed to the SIDOT central server via HTTPS.
//
// Features:
// - Multi-driver support: PostgreSQL, MySQL, Oracle
// - Configurable field mapping via YAML
// - LGPD-compliant data masking (CPF masked, CNS preserved)
// - Exponential backoff retry for network failures
// - Persistent watermark state for crash recovery
// - Outbound-only connectivity (no incoming connections)
//
// Usage:
//
//	pep-agent -config mapping.yaml
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sidot/pep-agent/internal/config"
	"github.com/sidot/pep-agent/internal/poller"
)

const (
	version = "1.0.0"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "mapping.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	validateOnly := flag.Bool("validate", false, "Validate configuration and exit")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("SIDOT PEP Agent v%s\n", version)
		os.Exit(0)
	}

	// Set up logging
	logger := log.New(os.Stdout, "[PEP-Agent] ", log.LstdFlags|log.Lmsgprefix)
	logger.Printf("SIDOT PEP Agent v%s starting...", version)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Printf("Configuration loaded from: %s", *configPath)
	logger.Printf("Database: %s://%s@%s:%d/%s",
		cfg.Database.Driver,
		cfg.Database.User,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
	logger.Printf("Source table: %s", cfg.Mapping.SourceTable)
	logger.Printf("Central server: %s", cfg.Central.URL)
	logger.Printf("Hospital ID: %s", cfg.Agent.HospitalID)

	// Validate only mode
	if *validateOnly {
		logger.Println("Configuration validated successfully")
		os.Exit(0)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start poller
	p := poller.NewPoller(cfg)
	p.SetLogger(logger)

	if err := p.Start(ctx); err != nil {
		logger.Fatalf("Failed to start poller: %v", err)
	}

	logger.Println("Agent started successfully")
	logger.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Println("Shutting down...")
	cancel()
	p.Stop()

	// Log final status
	status := p.GetStatus()
	logger.Printf("Final status: Processed=%d, Errors=%d", status.TotalProcessed, status.TotalErrors)
	logger.Println("Agent stopped gracefully")
}
