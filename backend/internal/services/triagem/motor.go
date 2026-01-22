package triagem

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
	"github.com/sidot/backend/internal/services/listener"
)

const (
	// Consumer group name for triagem motor
	ConsumerGroupName = "triagem-motor"

	// Consumer name (can be unique per instance)
	ConsumerName = "triagem-consumer-1"

	// Block time for XREADGROUP
	BlockTime = 5 * time.Second

	// Default rules cache TTL
	DefaultRulesCacheTTL = 5 * time.Minute
)

var (
	ErrIneligible = errors.New("obito is not eligible for donation")
)

// OccurrenceCreatedCallback is called when a new occurrence is created
type OccurrenceCreatedCallback func(ctx context.Context, occurrence *models.Occurrence, hospitalNome string)

// TriagemResult represents the result of triagem evaluation
type TriagemResult struct {
	Elegivel     bool     `json:"elegivel"`
	Score        int      `json:"score"`
	Motivos      []string `json:"motivos,omitempty"`
	RulesApplied []string `json:"rules_applied,omitempty"`
}

// TriagemMotor consumes obito events and applies triagem rules
type TriagemMotor struct {
	db           *sql.DB
	redis        *redis.Client
	obitoRepo    *repository.ObitoRepository
	occRepo      *repository.OccurrenceRepository
	historyRepo  *repository.OccurrenceHistoryRepository
	ruleRepo     *repository.TriagemRuleRepository
	hospitalRepo *repository.HospitalRepository

	// Cached rules
	cachedRules    []models.TriagemRule
	rulesCacheTime time.Time
	rulesCacheTTL  time.Duration
	rulesMu        sync.RWMutex

	// Status tracking
	running          int32
	totalProcessados int64
	totalElegiveis   int64
	totalInelegiveis int64
	errors           int64
	startedAt        time.Time

	// Control channels
	stopCh chan struct{}
	doneCh chan struct{}

	// Callback for new occurrences (used for SSE notifications)
	onOccurrenceCreated OccurrenceCreatedCallback

	// Logger
	logger *log.Logger
}

// NewTriagemMotor creates a new TriagemMotor
func NewTriagemMotor(db *sql.DB, redisClient *redis.Client) *TriagemMotor {
	return &TriagemMotor{
		db:            db,
		redis:         redisClient,
		obitoRepo:     repository.NewObitoRepository(db),
		occRepo:       repository.NewOccurrenceRepository(db),
		historyRepo:   repository.NewOccurrenceHistoryRepository(db),
		ruleRepo:      repository.NewTriagemRuleRepository(db, redisClient),
		hospitalRepo:  repository.NewHospitalRepository(db),
		rulesCacheTTL: DefaultRulesCacheTTL,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
		logger:        log.Default(),
	}
}

// SetOnOccurrenceCreated sets the callback for when a new occurrence is created
func (m *TriagemMotor) SetOnOccurrenceCreated(callback OccurrenceCreatedCallback) {
	m.onOccurrenceCreated = callback
}

// Start begins the consumer loop
func (m *TriagemMotor) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&m.running, 0, 1) {
		return nil // Already running
	}

	m.startedAt = time.Now()
	m.logger.Println("[Triagem] Starting triagem motor")

	// Create consumer group if it doesn't exist
	err := m.createConsumerGroup(ctx)
	if err != nil {
		m.logger.Printf("[Triagem] Warning: Could not create consumer group: %v", err)
		// Continue anyway - group might already exist
	}

	go m.consumeLoop(ctx)

	return nil
}

// Stop stops the consumer loop
func (m *TriagemMotor) Stop() {
	if atomic.CompareAndSwapInt32(&m.running, 1, 0) {
		close(m.stopCh)
		<-m.doneCh
		m.logger.Println("[Triagem] Triagem motor stopped")
	}
}

// IsRunning returns true if the motor is running
func (m *TriagemMotor) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

// createConsumerGroup creates the consumer group for the stream
func (m *TriagemMotor) createConsumerGroup(ctx context.Context) error {
	// Try to create the group, starting from the beginning
	err := m.redis.XGroupCreateMkStream(ctx, listener.ObitosStreamName, ConsumerGroupName, "0").Err()
	if err != nil {
		// Check if the group already exists
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil // Group already exists
		}
		return err
	}
	return nil
}

// consumeLoop is the main consumer loop
func (m *TriagemMotor) consumeLoop(ctx context.Context) {
	defer close(m.doneCh)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		default:
			m.consumeMessages(ctx)
		}
	}
}

// consumeMessages reads and processes messages from the stream
func (m *TriagemMotor) consumeMessages(ctx context.Context) {
	// Read messages using XREADGROUP
	streams, err := m.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroupName,
		Consumer: ConsumerName,
		Streams:  []string{listener.ObitosStreamName, ">"},
		Count:    10,
		Block:    BlockTime,
	}).Result()

	if err != nil {
		if err == redis.Nil || strings.Contains(err.Error(), "context") {
			return // No messages or context cancelled
		}
		m.logger.Printf("[Triagem] Error reading from stream: %v", err)
		atomic.AddInt64(&m.errors, 1)
		time.Sleep(1 * time.Second) // Back off on error
		return
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			m.processMessage(ctx, message)
		}
	}
}

// processMessage processes a single message from the stream
func (m *TriagemMotor) processMessage(ctx context.Context, message redis.XMessage) {
	data, ok := message.Values["data"].(string)
	if !ok {
		m.logger.Printf("[Triagem] Invalid message format: %v", message.ID)
		m.ackMessage(ctx, message.ID)
		return
	}

	event, err := listener.ParseObitoEvent(data)
	if err != nil {
		m.logger.Printf("[Triagem] Error parsing obito event: %v", err)
		m.ackMessage(ctx, message.ID)
		atomic.AddInt64(&m.errors, 1)
		return
	}

	m.logger.Printf("[Triagem] Processing obito: ID=%s, Hospital=%s", event.ObitoID, event.HospitalID)

	// Get the full obito data
	obitoID, err := event.GetObitoID()
	if err != nil {
		m.logger.Printf("[Triagem] Error parsing obito ID: %v", err)
		m.ackMessage(ctx, message.ID)
		atomic.AddInt64(&m.errors, 1)
		return
	}

	obito, err := m.obitoRepo.GetByID(ctx, obitoID)
	if err != nil {
		m.logger.Printf("[Triagem] Error fetching obito %s: %v", obitoID, err)
		m.ackMessage(ctx, message.ID)
		atomic.AddInt64(&m.errors, 1)
		return
	}

	// Check if occurrence already exists (idempotency)
	exists, err := m.occRepo.ExistsByObitoID(ctx, obitoID)
	if err != nil {
		m.logger.Printf("[Triagem] Error checking occurrence existence: %v", err)
		m.ackMessage(ctx, message.ID)
		atomic.AddInt64(&m.errors, 1)
		return
	}

	if exists {
		m.logger.Printf("[Triagem] Occurrence already exists for obito %s, skipping", obitoID)
		m.ackMessage(ctx, message.ID)
		return
	}

	// Apply triagem rules
	result, err := m.ApplyRules(ctx, obito)
	if err != nil {
		m.logger.Printf("[Triagem] Error applying rules to obito %s: %v", obitoID, err)
		m.ackMessage(ctx, message.ID)
		atomic.AddInt64(&m.errors, 1)
		return
	}

	atomic.AddInt64(&m.totalProcessados, 1)

	if result.Elegivel {
		// Create occurrence for eligible obito
		occurrence, err := m.createOccurrence(ctx, obito, result)
		if err != nil {
			m.logger.Printf("[Triagem] Error creating occurrence for obito %s: %v", obitoID, err)
			atomic.AddInt64(&m.errors, 1)
		} else {
			atomic.AddInt64(&m.totalElegiveis, 1)
			m.logger.Printf("[Triagem] Obito %s is ELIGIBLE - Occurrence created with score %d", obitoID, result.Score)

			// Trigger notification callback if set
			if m.onOccurrenceCreated != nil && occurrence != nil {
				hospitalNome := m.getHospitalName(ctx, obito.HospitalID)
				m.onOccurrenceCreated(ctx, occurrence, hospitalNome)
			}
		}
	} else {
		atomic.AddInt64(&m.totalInelegiveis, 1)
		m.logger.Printf("[Triagem] Obito %s is INELIGIBLE - Reasons: %s", obitoID, strings.Join(result.Motivos, ", "))
	}

	// Acknowledge the message
	m.ackMessage(ctx, message.ID)
}

// getHospitalName retrieves the hospital name for notifications
func (m *TriagemMotor) getHospitalName(ctx context.Context, hospitalID uuid.UUID) string {
	hospital, err := m.hospitalRepo.GetByID(ctx, hospitalID)
	if err != nil {
		return "Hospital Desconhecido"
	}
	return hospital.Nome
}

// ackMessage acknowledges a message in the stream
func (m *TriagemMotor) ackMessage(ctx context.Context, messageID string) {
	err := m.redis.XAck(ctx, listener.ObitosStreamName, ConsumerGroupName, messageID).Err()
	if err != nil {
		m.logger.Printf("[Triagem] Error acknowledging message %s: %v", messageID, err)
	}
}

// ApplyRules applies all active triagem rules to an obito
func (m *TriagemMotor) ApplyRules(ctx context.Context, obito *models.ObitoSimulado) (*TriagemResult, error) {
	result := &TriagemResult{
		Elegivel:     true,
		Score:        0,
		Motivos:      []string{},
		RulesApplied: []string{},
	}

	// Get cached rules
	rules, err := m.getCachedRules(ctx)
	if err != nil {
		// If we can't get rules, apply default rules
		m.logger.Printf("[Triagem] Warning: Could not get rules, using defaults: %v", err)
		rules = m.getDefaultRules()
	}

	// Apply each rule
	for _, rule := range rules {
		if !rule.Ativo {
			continue
		}

		ruleResult := m.applyRule(obito, &rule)
		result.RulesApplied = append(result.RulesApplied, rule.Nome)

		if !ruleResult.Elegivel {
			result.Elegivel = false
			result.Motivos = append(result.Motivos, ruleResult.Motivos...)
		}

		result.Score += ruleResult.Score
	}

	// Calculate final score based on sector if eligible
	if result.Elegivel {
		result.Score = m.calculatePriorityScore(obito)
	}

	return result, nil
}

// getCachedRules gets rules from cache or database
func (m *TriagemMotor) getCachedRules(ctx context.Context) ([]models.TriagemRule, error) {
	m.rulesMu.RLock()
	if time.Since(m.rulesCacheTime) < m.rulesCacheTTL && len(m.cachedRules) > 0 {
		rules := m.cachedRules
		m.rulesMu.RUnlock()
		return rules, nil
	}
	m.rulesMu.RUnlock()

	// Fetch from database
	m.rulesMu.Lock()
	defer m.rulesMu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(m.rulesCacheTime) < m.rulesCacheTTL && len(m.cachedRules) > 0 {
		return m.cachedRules, nil
	}

	rules, err := m.ruleRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	m.cachedRules = rules
	m.rulesCacheTime = time.Now()

	return rules, nil
}

// InvalidateRulesCache invalidates the rules cache
func (m *TriagemMotor) InvalidateRulesCache() {
	m.rulesMu.Lock()
	defer m.rulesMu.Unlock()
	m.rulesCacheTime = time.Time{}
	m.cachedRules = nil
}

// getDefaultRules returns default triagem rules
func (m *TriagemMotor) getDefaultRules() []models.TriagemRule {
	return []models.TriagemRule{
		{
			ID:     uuid.New(),
			Nome:   "Idade Maxima",
			Ativo:  true,
			Regras: json.RawMessage(`{"tipo": "idade_maxima", "valor": 80, "acao": "rejeitar"}`),
		},
		{
			ID:     uuid.New(),
			Nome:   "Janela 6 Horas",
			Ativo:  true,
			Regras: json.RawMessage(`{"tipo": "janela_horas", "valor": 6, "acao": "rejeitar"}`),
		},
		{
			ID:     uuid.New(),
			Nome:   "Identificacao Desconhecida",
			Ativo:  true,
			Regras: json.RawMessage(`{"tipo": "identificacao_desconhecida", "valor": true, "acao": "rejeitar"}`),
		},
	}
}

// applyRule applies a single rule to an obito
func (m *TriagemMotor) applyRule(obito *models.ObitoSimulado, rule *models.TriagemRule) *TriagemResult {
	result := &TriagemResult{
		Elegivel: true,
		Score:    0,
		Motivos:  []string{},
	}

	var config models.RuleConfig
	if err := json.Unmarshal(rule.Regras, &config); err != nil {
		m.logger.Printf("[Triagem] Error parsing rule config: %v", err)
		return result
	}

	switch config.Tipo {
	case models.RuleTypeIdadeMaxima:
		return m.applyIdadeMaximaRule(obito, config.Valor)

	case models.RuleTypeJanelaHoras:
		return m.applyJanelaHorasRule(obito, config.Valor)

	case models.RuleTypeIdentificacaoDesconhecida:
		return m.applyIdentificacaoDesconhecidaRule(obito, config.Valor)

	case models.RuleTypeCausasExcludentes:
		return m.applyCausasExcludentesRule(obito, config.Valor)

	case models.RuleTypeSetorPriorizacao:
		// This rule only affects score, not eligibility
		return result
	}

	return result
}

// applyIdadeMaximaRule applies the maximum age rule
func (m *TriagemMotor) applyIdadeMaximaRule(obito *models.ObitoSimulado, valor interface{}) *TriagemResult {
	result := &TriagemResult{Elegivel: true, Motivos: []string{}}

	maxAge, ok := valor.(float64) // JSON numbers are float64
	if !ok {
		return result
	}

	idade := obito.CalculateAge()
	if idade > int(maxAge) {
		result.Elegivel = false
		result.Motivos = append(result.Motivos, "Idade acima do limite")
	}

	return result
}

// applyJanelaHorasRule applies the time window rule
func (m *TriagemMotor) applyJanelaHorasRule(obito *models.ObitoSimulado, valor interface{}) *TriagemResult {
	result := &TriagemResult{Elegivel: true, Motivos: []string{}}

	windowHours, ok := valor.(float64)
	if !ok {
		return result
	}

	if !obito.IsWithinWindow(int(windowHours)) {
		result.Elegivel = false
		result.Motivos = append(result.Motivos, "Fora da janela de captacao")
	}

	return result
}

// applyIdentificacaoDesconhecidaRule applies the unknown identification rule
func (m *TriagemMotor) applyIdentificacaoDesconhecidaRule(obito *models.ObitoSimulado, valor interface{}) *TriagemResult {
	result := &TriagemResult{Elegivel: true, Motivos: []string{}}

	rejectUnknown, ok := valor.(bool)
	if !ok {
		return result
	}

	if rejectUnknown && obito.IdentificacaoDesconhecida {
		result.Elegivel = false
		result.Motivos = append(result.Motivos, "Identificacao desconhecida (indigente)")
	}

	return result
}

// applyCausasExcludentesRule applies the excluded causes rule
func (m *TriagemMotor) applyCausasExcludentesRule(obito *models.ObitoSimulado, valor interface{}) *TriagemResult {
	result := &TriagemResult{Elegivel: true, Motivos: []string{}}

	causasInterface, ok := valor.([]interface{})
	if !ok {
		return result
	}

	var causas []string
	for _, c := range causasInterface {
		if s, ok := c.(string); ok {
			causas = append(causas, strings.ToLower(s))
		}
	}

	causaMortisLower := strings.ToLower(obito.CausaMortis)
	for _, causa := range causas {
		if strings.Contains(causaMortisLower, causa) {
			result.Elegivel = false
			result.Motivos = append(result.Motivos, "Causa de morte excludente: "+causa)
			break
		}
	}

	return result
}

// calculatePriorityScore calculates the priority score based on sector and time remaining
func (m *TriagemMotor) calculatePriorityScore(obito *models.ObitoSimulado) int {
	baseScore := 50 // Default score

	// Get sector score
	if obito.Setor != nil {
		baseScore = models.GetSectorScore(*obito.Setor)
	}

	// Adjust by time remaining (more urgent = higher score)
	remaining := obito.TimeRemaining(6) // 6 hour window
	if remaining > 0 {
		hoursRemaining := remaining.Hours()
		if hoursRemaining <= 1 {
			baseScore += 20 // Very urgent
		} else if hoursRemaining <= 2 {
			baseScore += 10 // Urgent
		} else if hoursRemaining <= 3 {
			baseScore += 5 // Moderate
		}
	}

	// Cap at 100
	if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

// createOccurrence creates a new occurrence for an eligible obito
func (m *TriagemMotor) createOccurrence(ctx context.Context, obito *models.ObitoSimulado, result *TriagemResult) (*models.Occurrence, error) {
	// Prepare complete data
	completeData := obito.ToOccurrenceData()
	completeDataJSON, err := json.Marshal(completeData)
	if err != nil {
		return nil, err
	}

	// Create occurrence input
	input := &models.CreateOccurrenceInput{
		ObitoID:               obito.ID,
		HospitalID:            obito.HospitalID,
		ScorePriorizacao:      result.Score,
		NomePacienteMascarado: models.MaskName(obito.NomePaciente),
		DadosCompletos:        completeDataJSON,
		DataObito:             obito.DataObito,
	}

	// Create the occurrence
	occurrence, err := m.occRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	// Register in history
	historyInput := &models.CreateHistoryInput{
		OccurrenceID: occurrence.ID,
		UserID:       nil, // System created
		Acao:         models.ActionOccurrenceCreated,
		StatusNovo:   &occurrence.Status,
	}

	_, err = m.historyRepo.Create(ctx, historyInput)
	if err != nil {
		m.logger.Printf("[Triagem] Warning: Could not create history entry: %v", err)
		// Don't fail the whole operation for history error
	}

	m.logger.Printf("[Triagem] Created occurrence %s for obito %s with score %d",
		occurrence.ID, obito.ID, result.Score)

	return occurrence, nil
}

// SetLogger sets a custom logger for the motor
func (m *TriagemMotor) SetLogger(logger *log.Logger) {
	m.logger = logger
}

// GetStats returns the current statistics of the motor
func (m *TriagemMotor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":           m.IsRunning(),
		"total_processados": atomic.LoadInt64(&m.totalProcessados),
		"total_elegiveis":   atomic.LoadInt64(&m.totalElegiveis),
		"total_inelegiveis": atomic.LoadInt64(&m.totalInelegiveis),
		"errors":            atomic.LoadInt64(&m.errors),
		"started_at":        m.startedAt,
	}
}
