package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vitalconnect/backend/internal/services/listener"
	"github.com/vitalconnect/backend/internal/services/notification"
	"github.com/vitalconnect/backend/internal/services/triagem"
)

const (
	// DefaultCheckInterval is the interval between health checks
	DefaultCheckInterval = 10 * time.Second

	// DefaultAlertCooldown is the cooldown between alerts of the same type
	DefaultAlertCooldown = 5 * time.Minute

	// DefaultTimeout for service checks
	DefaultTimeout = 2 * time.Second

	// Redis keys for storing health state
	LastStatesKey    = "vitalconnect:health:last_states"
	AlertCooldownKey = "vitalconnect:health:alert_cooldowns"

	// Latency thresholds in milliseconds
	LatencyThresholdOK       = 500   // < 500ms = up
	LatencyThresholdDegraded = 2000  // 500-2000ms = degraded
)

// ServiceStatus represents the status of a single service
type ServiceStatus string

const (
	StatusUp       ServiceStatus = "up"
	StatusDegraded ServiceStatus = "degraded"
	StatusDown     ServiceStatus = "down"
)

// ComponentStatus represents the status of a component
type ComponentStatus struct {
	Name      string        `json:"name"`
	Status    ServiceStatus `json:"status"`
	LatencyMs int64         `json:"latency_ms"`
	LastCheck time.Time     `json:"last_check"`
	Message   string        `json:"message,omitempty"`
}

// HealthSummary represents the overall health status
type HealthSummary struct {
	Status     ServiceStatus              `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentStatus `json:"components"`
}

// StateTransition represents a state change for a service
type StateTransition struct {
	Service       string        `json:"service"`
	PreviousState ServiceStatus `json:"previous_state"`
	NewState      ServiceStatus `json:"new_state"`
	Timestamp     time.Time     `json:"timestamp"`
}

// HealthMonitorService monitors the health of all system components
type HealthMonitorService struct {
	db           *sql.DB
	redis        *redis.Client
	emailService *notification.EmailService
	sseHub       *notification.SSEHub
	listener     *listener.ObitoListener
	triagemMotor *triagem.TriagemMotor
	adminEmail   string

	// Last known states
	lastKnownStates map[string]ServiceStatus
	statesMu        sync.RWMutex

	// Alert cooldowns
	alertCooldowns map[string]time.Time
	cooldownMu     sync.RWMutex
	cooldownPeriod time.Duration

	// Check interval
	checkInterval time.Duration

	// Status tracking
	running int32

	// Control channels
	stopCh chan struct{}
	doneCh chan struct{}

	// Last summary cache
	lastSummary   *HealthSummary
	lastSummaryMu sync.RWMutex

	// Logger
	logger *log.Logger
}

// NewHealthMonitorService creates a new HealthMonitorService
func NewHealthMonitorService(
	db *sql.DB,
	redisClient *redis.Client,
	emailService *notification.EmailService,
	adminEmail string,
) *HealthMonitorService {
	return &HealthMonitorService{
		db:              db,
		redis:           redisClient,
		emailService:    emailService,
		adminEmail:      adminEmail,
		lastKnownStates: make(map[string]ServiceStatus),
		alertCooldowns:  make(map[string]time.Time),
		cooldownPeriod:  DefaultAlertCooldown,
		checkInterval:   DefaultCheckInterval,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
		logger:          log.Default(),
	}
}

// SetSSEHub sets the SSE hub for publishing events
func (m *HealthMonitorService) SetSSEHub(hub *notification.SSEHub) {
	m.sseHub = hub
}

// SetListener sets the listener for health checks
func (m *HealthMonitorService) SetListener(l *listener.ObitoListener) {
	m.listener = l
}

// SetTriagemMotor sets the triagem motor for health checks
func (m *HealthMonitorService) SetTriagemMotor(t *triagem.TriagemMotor) {
	m.triagemMotor = t
}

// SetCheckInterval sets the check interval
func (m *HealthMonitorService) SetCheckInterval(interval time.Duration) {
	m.checkInterval = interval
}

// SetCooldownPeriod sets the alert cooldown period
func (m *HealthMonitorService) SetCooldownPeriod(period time.Duration) {
	m.cooldownPeriod = period
}

// SetLogger sets the logger
func (m *HealthMonitorService) SetLogger(logger *log.Logger) {
	m.logger = logger
}

// Start begins the health monitoring loop
func (m *HealthMonitorService) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&m.running, 0, 1) {
		return nil // Already running
	}

	m.logger.Println("[HealthMonitor] Starting health monitor service")

	// Load last known states from Redis
	m.loadLastStates(ctx)

	go m.monitorLoop(ctx)

	return nil
}

// Stop stops the health monitoring
func (m *HealthMonitorService) Stop() {
	if atomic.CompareAndSwapInt32(&m.running, 1, 0) {
		close(m.stopCh)
		<-m.doneCh
		m.logger.Println("[HealthMonitor] Health monitor service stopped")
	}
}

// IsRunning returns true if the monitor is running
func (m *HealthMonitorService) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

// monitorLoop is the main monitoring loop
func (m *HealthMonitorService) monitorLoop(ctx context.Context) {
	defer close(m.doneCh)

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	// Initial check
	m.performHealthCheck(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck performs all health checks and handles state transitions
func (m *HealthMonitorService) performHealthCheck(ctx context.Context) {
	summary := m.GetHealthSummary(ctx)

	// Cache the summary
	m.lastSummaryMu.Lock()
	m.lastSummary = summary
	m.lastSummaryMu.Unlock()

	// Check for state transitions and send alerts
	for name, component := range summary.Components {
		m.checkStateTransition(ctx, name, component.Status)
	}

	// Save states to Redis
	m.saveLastStates(ctx)
}

// GetHealthSummary returns the current health summary
func (m *HealthMonitorService) GetHealthSummary(ctx context.Context) *HealthSummary {
	summary := &HealthSummary{
		Timestamp:  time.Now(),
		Components: make(map[string]ComponentStatus),
	}

	// Check all components in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	checks := []struct {
		name  string
		check func(context.Context) ComponentStatus
	}{
		{"database", m.checkDatabase},
		{"redis", m.checkRedis},
		{"listener", m.checkListener},
		{"triagem_motor", m.checkTriagemMotor},
		{"sse_hub", m.checkSSEHub},
		{"api", m.checkAPI},
	}

	for _, c := range checks {
		wg.Add(1)
		go func(name string, check func(context.Context) ComponentStatus) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
			defer cancel()

			status := check(checkCtx)

			mu.Lock()
			summary.Components[name] = status
			mu.Unlock()
		}(c.name, c.check)
	}

	wg.Wait()

	// Calculate overall status
	summary.Status = m.calculateOverallStatus(summary.Components)

	return summary
}

// GetLastSummary returns the cached last health summary
func (m *HealthMonitorService) GetLastSummary() *HealthSummary {
	m.lastSummaryMu.RLock()
	defer m.lastSummaryMu.RUnlock()
	return m.lastSummary
}

// calculateOverallStatus determines the overall system status
func (m *HealthMonitorService) calculateOverallStatus(components map[string]ComponentStatus) ServiceStatus {
	hasDown := false
	hasDegraded := false

	for _, comp := range components {
		switch comp.Status {
		case StatusDown:
			hasDown = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasDown {
		return StatusDown
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusUp
}

// checkDatabase checks PostgreSQL health
func (m *HealthMonitorService) checkDatabase(ctx context.Context) ComponentStatus {
	start := time.Now()

	err := m.db.PingContext(ctx)
	latencyMs := time.Since(start).Milliseconds()

	status := ComponentStatus{
		Name:      "Database (PostgreSQL)",
		LatencyMs: latencyMs,
		LastCheck: time.Now(),
	}

	if err != nil {
		status.Status = StatusDown
		status.Message = err.Error()
		return status
	}

	status.Status = m.latencyToStatus(latencyMs)
	return status
}

// checkRedis checks Redis health
func (m *HealthMonitorService) checkRedis(ctx context.Context) ComponentStatus {
	start := time.Now()

	err := m.redis.Ping(ctx).Err()
	latencyMs := time.Since(start).Milliseconds()

	status := ComponentStatus{
		Name:      "Redis",
		LatencyMs: latencyMs,
		LastCheck: time.Now(),
	}

	if err != nil {
		status.Status = StatusDown
		status.Message = err.Error()
		return status
	}

	status.Status = m.latencyToStatus(latencyMs)
	return status
}

// checkListener checks the Obito Listener health via heartbeat
func (m *HealthMonitorService) checkListener(ctx context.Context) ComponentStatus {
	status := ComponentStatus{
		Name:      "Obito Listener",
		LastCheck: time.Now(),
	}

	listenerStatus, latencyMs := listener.CheckHeartbeat(ctx, m.redis)
	status.LatencyMs = latencyMs

	switch listenerStatus {
	case "up":
		status.Status = m.latencyToStatus(latencyMs)
	default:
		status.Status = StatusDown
		status.Message = "No heartbeat detected"
	}

	return status
}

// checkTriagemMotor checks the Triagem Motor health
func (m *HealthMonitorService) checkTriagemMotor(ctx context.Context) ComponentStatus {
	start := time.Now()

	status := ComponentStatus{
		Name:      "Triagem Motor",
		LastCheck: time.Now(),
	}

	if m.triagemMotor == nil {
		status.Status = StatusDown
		status.Message = "Not initialized"
		status.LatencyMs = time.Since(start).Milliseconds()
		return status
	}

	if m.triagemMotor.IsRunning() {
		status.Status = StatusUp
	} else {
		status.Status = StatusDown
		status.Message = "Not running"
	}

	status.LatencyMs = time.Since(start).Milliseconds()
	return status
}

// checkSSEHub checks the SSE Hub health
func (m *HealthMonitorService) checkSSEHub(ctx context.Context) ComponentStatus {
	start := time.Now()

	status := ComponentStatus{
		Name:      "SSE Hub",
		LastCheck: time.Now(),
	}

	if m.sseHub == nil {
		status.Status = StatusDown
		status.Message = "Not initialized"
		status.LatencyMs = time.Since(start).Milliseconds()
		return status
	}

	if m.sseHub.IsRunning() {
		status.Status = StatusUp
	} else {
		status.Status = StatusDown
		status.Message = "Not running"
	}

	status.LatencyMs = time.Since(start).Milliseconds()
	return status
}

// checkAPI checks the API health (self-check)
func (m *HealthMonitorService) checkAPI(ctx context.Context) ComponentStatus {
	start := time.Now()

	status := ComponentStatus{
		Name:      "API",
		Status:    StatusUp,
		LatencyMs: time.Since(start).Milliseconds(),
		LastCheck: time.Now(),
	}

	return status
}

// latencyToStatus converts latency to status
func (m *HealthMonitorService) latencyToStatus(latencyMs int64) ServiceStatus {
	if latencyMs < LatencyThresholdOK {
		return StatusUp
	}
	if latencyMs < LatencyThresholdDegraded {
		return StatusDegraded
	}
	return StatusDown
}

// checkStateTransition checks if there was a state transition and handles alerts
func (m *HealthMonitorService) checkStateTransition(ctx context.Context, service string, newState ServiceStatus) {
	m.statesMu.Lock()
	previousState, exists := m.lastKnownStates[service]
	m.lastKnownStates[service] = newState
	m.statesMu.Unlock()

	if !exists {
		// First time seeing this service, no transition
		return
	}

	if previousState == newState {
		// No state change
		return
	}

	transition := StateTransition{
		Service:       service,
		PreviousState: previousState,
		NewState:      newState,
		Timestamp:     time.Now(),
	}

	m.logger.Printf("[HealthMonitor] State transition detected: %s %s -> %s",
		service, previousState, newState)

	// Send alert if listener goes down
	if service == "listener" && previousState == StatusUp && newState == StatusDown {
		m.sendListenerDownAlert(ctx, transition)
	}

	// Publish SSE event for status change (optional feature)
	m.publishStatusChangeEvent(ctx, transition)
}

// sendListenerDownAlert sends an email alert when the listener goes down
func (m *HealthMonitorService) sendListenerDownAlert(ctx context.Context, transition StateTransition) {
	if m.emailService == nil || !m.emailService.IsConfigured() {
		m.logger.Println("[HealthMonitor] Email service not configured, skipping alert")
		return
	}

	if m.adminEmail == "" {
		m.logger.Println("[HealthMonitor] Admin email not configured, skipping alert")
		return
	}

	// Check cooldown
	if !m.canSendAlert("listener") {
		m.logger.Println("[HealthMonitor] Alert cooldown active, skipping listener down alert")
		return
	}

	// Mark alert as sent
	m.markAlertSent("listener")

	// Send the alert
	err := m.emailService.SendInfrastructureAlert(ctx, m.adminEmail, &notification.InfrastructureAlertData{
		ServiceName:    "Obito Listener",
		Status:         "DOWN",
		PreviousStatus: string(transition.PreviousState),
		Timestamp:      transition.Timestamp,
		Message:        "O servico Obito Listener parou de responder. Verifique imediatamente para evitar perda de notificacoes de doacao.",
	})

	if err != nil {
		m.logger.Printf("[HealthMonitor] Error sending listener down alert: %v", err)
	} else {
		m.logger.Println("[HealthMonitor] Listener down alert sent to admin")
	}
}

// publishStatusChangeEvent publishes a status change event via SSE
func (m *HealthMonitorService) publishStatusChangeEvent(ctx context.Context, transition StateTransition) {
	if m.sseHub == nil || !m.sseHub.IsRunning() {
		return
	}

	// This could be extended to publish system_status_change events
	// For now, we just log it
	m.logger.Printf("[HealthMonitor] Status change event: %s %s -> %s",
		transition.Service, transition.PreviousState, transition.NewState)
}

// canSendAlert checks if we can send an alert (respecting cooldown)
func (m *HealthMonitorService) canSendAlert(service string) bool {
	m.cooldownMu.RLock()
	lastAlert, exists := m.alertCooldowns[service]
	m.cooldownMu.RUnlock()

	if !exists {
		return true
	}

	return time.Since(lastAlert) >= m.cooldownPeriod
}

// markAlertSent marks that an alert was sent for a service
func (m *HealthMonitorService) markAlertSent(service string) {
	m.cooldownMu.Lock()
	m.alertCooldowns[service] = time.Now()
	m.cooldownMu.Unlock()
}

// loadLastStates loads the last known states from Redis
func (m *HealthMonitorService) loadLastStates(ctx context.Context) {
	data, err := m.redis.Get(ctx, LastStatesKey).Result()
	if err != nil {
		if err != redis.Nil {
			m.logger.Printf("[HealthMonitor] Error loading last states: %v", err)
		}
		return
	}

	var states map[string]ServiceStatus
	if err := json.Unmarshal([]byte(data), &states); err != nil {
		m.logger.Printf("[HealthMonitor] Error unmarshalling last states: %v", err)
		return
	}

	m.statesMu.Lock()
	m.lastKnownStates = states
	m.statesMu.Unlock()
}

// saveLastStates saves the last known states to Redis
func (m *HealthMonitorService) saveLastStates(ctx context.Context) {
	m.statesMu.RLock()
	states := make(map[string]ServiceStatus)
	for k, v := range m.lastKnownStates {
		states[k] = v
	}
	m.statesMu.RUnlock()

	data, err := json.Marshal(states)
	if err != nil {
		m.logger.Printf("[HealthMonitor] Error marshalling last states: %v", err)
		return
	}

	err = m.redis.Set(ctx, LastStatesKey, data, 0).Err()
	if err != nil {
		m.logger.Printf("[HealthMonitor] Error saving last states: %v", err)
	}
}
