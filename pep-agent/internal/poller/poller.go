// Package poller implements periodic polling of the PEP database
package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vitalconnect/pep-agent/internal/config"
	"github.com/vitalconnect/pep-agent/internal/database"
	"github.com/vitalconnect/pep-agent/internal/models"
	"github.com/vitalconnect/pep-agent/internal/pusher"
)

// Poller polls the PEP database and pushes events to central server
type Poller struct {
	config    *config.AgentConfig
	connector *database.Connector
	pusher    *pusher.Pusher
	logger    *log.Logger

	// State management
	state        *models.AgentState
	stateMu      sync.RWMutex
	stateChanged bool

	// Status tracking
	running        int32
	offlineSince   time.Time
	lastPollTime   time.Time
	totalProcessed int64
	totalErrors    int64

	// Control
	stopCh chan struct{}
	doneCh chan struct{}
}

// PollerStatus represents the current status of the poller
type PollerStatus struct {
	Running          bool       `json:"running"`
	OfflineSince     *time.Time `json:"offline_since,omitempty"`
	LastPollTime     *time.Time `json:"last_poll_time,omitempty"`
	LastProcessedID  string     `json:"last_processed_id,omitempty"`
	TotalProcessed   int64      `json:"total_processed"`
	TotalErrors      int64      `json:"total_errors"`
	DatabaseConnected bool      `json:"database_connected"`
}

// NewPoller creates a new Poller instance
func NewPoller(cfg *config.AgentConfig) *Poller {
	connector := database.NewConnector(cfg)
	push := pusher.NewPusher(cfg)

	return &Poller{
		config:    cfg,
		connector: connector,
		pusher:    push,
		logger:    log.New(os.Stdout, "[PEP-Agent] ", log.LstdFlags|log.Lmsgprefix),
		state:     &models.AgentState{},
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// Start begins the polling loop
func (p *Poller) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return nil // Already running
	}

	// Load persisted state
	if err := p.loadState(); err != nil {
		p.logger.Printf("Warning: Could not load state: %v (starting fresh)", err)
	}

	// Connect to database
	if err := p.connector.Connect(); err != nil {
		atomic.StoreInt32(&p.running, 0)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	p.logger.Printf("Starting poller with interval: %v", p.config.GetPollInterval())
	p.logger.Printf("Database: %s@%s:%d/%s",
		p.config.Database.User,
		p.config.Database.Host,
		p.config.Database.Port,
		p.config.Database.Database,
	)
	p.logger.Printf("Central URL: %s", p.config.Central.URL)

	go p.pollLoop(ctx)

	return nil
}

// Stop stops the polling loop
func (p *Poller) Stop() {
	if atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		close(p.stopCh)
		<-p.doneCh

		// Save state before shutdown
		if err := p.saveState(); err != nil {
			p.logger.Printf("Warning: Failed to save state: %v", err)
		}

		// Close database connection
		if err := p.connector.Close(); err != nil {
			p.logger.Printf("Warning: Error closing database: %v", err)
		}

		p.logger.Println("Poller stopped")
	}
}

// IsRunning returns true if the poller is running
func (p *Poller) IsRunning() bool {
	return atomic.LoadInt32(&p.running) == 1
}

// GetStatus returns the current poller status
func (p *Poller) GetStatus() *PollerStatus {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()

	status := &PollerStatus{
		Running:           p.IsRunning(),
		TotalProcessed:    atomic.LoadInt64(&p.totalProcessed),
		TotalErrors:       atomic.LoadInt64(&p.totalErrors),
		DatabaseConnected: p.connector.IsConnected(),
		LastProcessedID:   p.state.LastProcessedID,
	}

	if !p.lastPollTime.IsZero() {
		status.LastPollTime = &p.lastPollTime
	}
	if !p.offlineSince.IsZero() {
		status.OfflineSince = &p.offlineSince
	}

	return status
}

// pollLoop is the main polling loop
func (p *Poller) pollLoop(ctx context.Context) {
	defer close(p.doneCh)

	ticker := time.NewTicker(p.config.GetPollInterval())
	defer ticker.Stop()

	// State save ticker (every minute)
	saveTicker := time.NewTicker(1 * time.Minute)
	defer saveTicker.Stop()

	// Initial poll
	p.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.poll(ctx)
		case <-saveTicker.C:
			if p.stateChanged {
				if err := p.saveState(); err != nil {
					p.logger.Printf("Warning: Failed to save state: %v", err)
				}
				p.stateChanged = false
			}
		}
	}
}

// poll performs a single poll cycle
func (p *Poller) poll(ctx context.Context) {
	p.lastPollTime = time.Now()

	// Check database connection
	if !p.connector.IsConnected() {
		if err := p.reconnectWithBackoff(ctx); err != nil {
			p.logger.Printf("Database connection failed: %v", err)
			atomic.AddInt64(&p.totalErrors, 1)
			p.checkOfflineAlert()
			return
		}
	}

	// Clear offline status on successful connection
	p.offlineSince = time.Time{}

	// Get watermark
	watermark := p.getWatermark()

	// Fetch new records
	records, err := p.connector.FetchNewRecords(ctx, watermark)
	if err != nil {
		p.logger.Printf("Error fetching records: %v", err)
		atomic.AddInt64(&p.totalErrors, 1)
		return
	}

	if len(records) == 0 {
		return
	}

	p.logger.Printf("Detected %d new record(s)", len(records))

	// Process each record
	for _, record := range records {
		select {
		case <-ctx.Done():
			return
		default:
			p.processRecord(ctx, record)
		}
	}
}

// processRecord processes a single PEP record
func (p *Poller) processRecord(ctx context.Context, record *models.PEPRecord) {
	// Convert to event (with LGPD masking)
	event := record.ToObitoEvent(p.config.Agent.HospitalID)

	// Log without sensitive data
	p.logger.Printf("Processing: ID=%s, Patient=%s, Time=%s",
		record.ID,
		models.MaskName(record.NomePaciente),
		record.DataObito.Format(time.RFC3339),
	)

	// Push to central server with retry
	result, err := p.pusher.PushWithRetry(ctx, event, nil)
	if err != nil {
		p.logger.Printf("Error pushing event %s: %v", record.ID, err)
		atomic.AddInt64(&p.totalErrors, 1)
		p.updateStateError(err.Error())
		return
	}

	if result.Success {
		p.logger.Printf("Successfully pushed event: ID=%s", record.ID)
		atomic.AddInt64(&p.totalProcessed, 1)
		p.updateWatermark(record.ID, record.DataObito)
	} else {
		p.logger.Printf("Failed to push event %s: %s", record.ID, result.Message)
		atomic.AddInt64(&p.totalErrors, 1)
	}
}

// reconnectWithBackoff attempts to reconnect with exponential backoff
func (p *Poller) reconnectWithBackoff(ctx context.Context) error {
	intervals := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		1 * time.Minute,
		2 * time.Minute,
		5 * time.Minute, // cap
	}

	for attempt, interval := range intervals {
		p.logger.Printf("Attempting to reconnect (attempt %d/%d)...", attempt+1, len(intervals))

		if err := p.connector.Connect(); err == nil {
			p.logger.Println("Reconnected to database")
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.stopCh:
			return fmt.Errorf("poller stopped")
		case <-time.After(interval):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("failed to reconnect after all attempts")
}

// checkOfflineAlert checks if offline duration exceeds threshold
func (p *Poller) checkOfflineAlert() {
	if p.offlineSince.IsZero() {
		p.offlineSince = time.Now()
		return
	}

	offlineDuration := time.Since(p.offlineSince)
	if offlineDuration >= p.config.GetAlertThreshold() {
		p.logger.Printf("ALERT: Offline for %v (threshold: %v)", offlineDuration, p.config.GetAlertThreshold())
	}
}

// getWatermark returns the current watermark for filtering
func (p *Poller) getWatermark() string {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()

	if !p.state.LastProcessedAt.IsZero() {
		return p.state.LastProcessedAt.Format("2006-01-02 15:04:05")
	}

	// Default: 24 hours ago
	return time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
}

// updateWatermark updates the watermark after successful processing
func (p *Poller) updateWatermark(id string, timestamp time.Time) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	p.state.LastProcessedID = id
	p.state.LastProcessedAt = timestamp
	p.state.TotalProcessed++
	p.stateChanged = true
}

// updateStateError records the last error in state
func (p *Poller) updateStateError(errMsg string) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	p.state.LastError = errMsg
	p.state.LastErrorAt = time.Now()
	p.stateChanged = true
}

// loadState loads persisted state from file
func (p *Poller) loadState() error {
	stateFile := p.config.Agent.StateFile

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	return json.Unmarshal(data, p.state)
}

// saveState persists state to file
func (p *Poller) saveState() error {
	stateFile := p.config.Agent.StateFile

	// Ensure directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	p.stateMu.RLock()
	data, err := json.MarshalIndent(p.state, "", "  ")
	p.stateMu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(stateFile, data, 0644)
}

// SetLogger sets a custom logger
func (p *Poller) SetLogger(logger *log.Logger) {
	p.logger = logger
}
