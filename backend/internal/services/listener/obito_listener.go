package listener

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

const (
	// Redis stream name for detected obitos
	ObitosStreamName = "obitos:detectados"

	// Default poll interval
	DefaultPollInterval = 3 * time.Second
)

// ObitoEvent represents an event published to Redis Streams
type ObitoEvent struct {
	ObitoID            string `json:"obito_id"`
	HospitalID         string `json:"hospital_id"`
	TimestampDeteccao  string `json:"timestamp_deteccao"`
	NomePaciente       string `json:"nome_paciente"`
	DataObito          string `json:"data_obito"`
	CausaMortis        string `json:"causa_mortis"`
	Setor              string `json:"setor,omitempty"`
	Leito              string `json:"leito,omitempty"`
	Idade              int    `json:"idade"`
	IdentificacaoDesconhecida bool `json:"identificacao_desconhecida"`
}

// ListenerStatus represents the current status of the listener
type ListenerStatus struct {
	Status              string    `json:"status"`
	Running             bool      `json:"running"`
	UltimoProcessamento *time.Time `json:"ultimo_processamento,omitempty"`
	ObitosDetectadosHoje int      `json:"obitos_detectados_hoje"`
	TotalProcessados    int64     `json:"total_processados"`
	Errors              int64     `json:"errors"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
}

// ObitoListener polls the obitos_simulados table and detects new deaths
type ObitoListener struct {
	db           *sql.DB
	redis        *redis.Client
	obitoRepo    *repository.ObitoRepository
	pollInterval time.Duration

	// Status tracking
	running             int32
	ultimoProcessamento atomic.Value // *time.Time
	totalProcessados    int64
	errors              int64
	startedAt           time.Time

	// Control channels
	stopCh chan struct{}
	doneCh chan struct{}

	// Mutex for status updates
	mu sync.RWMutex

	// Logger
	logger *log.Logger
}

// NewObitoListener creates a new ObitoListener
func NewObitoListener(db *sql.DB, redisClient *redis.Client, pollInterval time.Duration) *ObitoListener {
	if pollInterval == 0 {
		pollInterval = DefaultPollInterval
	}

	listener := &ObitoListener{
		db:           db,
		redis:        redisClient,
		obitoRepo:    repository.NewObitoRepository(db),
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
		logger:       log.Default(),
	}

	return listener
}

// Start begins the polling loop
func (l *ObitoListener) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&l.running, 0, 1) {
		return nil // Already running
	}

	l.startedAt = time.Now()
	l.logger.Printf("[Listener] Starting obito listener with poll interval: %v", l.pollInterval)

	go l.pollLoop(ctx)

	return nil
}

// Stop stops the polling loop
func (l *ObitoListener) Stop() {
	if atomic.CompareAndSwapInt32(&l.running, 1, 0) {
		close(l.stopCh)
		<-l.doneCh
		l.logger.Println("[Listener] Obito listener stopped")
	}
}

// IsRunning returns true if the listener is running
func (l *ObitoListener) IsRunning() bool {
	return atomic.LoadInt32(&l.running) == 1
}

// GetStatus returns the current status of the listener
func (l *ObitoListener) GetStatus(ctx context.Context) *ListenerStatus {
	status := &ListenerStatus{
		Running:          l.IsRunning(),
		TotalProcessados: atomic.LoadInt64(&l.totalProcessados),
		Errors:           atomic.LoadInt64(&l.errors),
	}

	if status.Running {
		status.Status = "running"
		status.StartedAt = &l.startedAt
	} else {
		status.Status = "stopped"
	}

	// Get ultimo processamento
	if val := l.ultimoProcessamento.Load(); val != nil {
		t := val.(*time.Time)
		status.UltimoProcessamento = t
	}

	// Get count of obitos detected today
	count, err := l.obitoRepo.CountTodayDetected(ctx)
	if err == nil {
		status.ObitosDetectadosHoje = count
	}

	return status
}

// pollLoop is the main polling loop
func (l *ObitoListener) pollLoop(ctx context.Context) {
	defer close(l.doneCh)

	ticker := time.NewTicker(l.pollInterval)
	defer ticker.Stop()

	// Initial poll
	l.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.poll(ctx)
		}
	}
}

// poll checks for new obitos and processes them
func (l *ObitoListener) poll(ctx context.Context) {
	// Get unprocessed obitos from the last 24 hours
	since := time.Now().Add(-24 * time.Hour)

	obitos, err := l.obitoRepo.GetUnprocessed(ctx, since)
	if err != nil {
		l.logger.Printf("[Listener] Error fetching unprocessed obitos: %v", err)
		atomic.AddInt64(&l.errors, 1)
		return
	}

	if len(obitos) == 0 {
		return
	}

	l.logger.Printf("[Listener] Detected %d new obito(s)", len(obitos))

	// Process each obito in a goroutine
	var wg sync.WaitGroup
	for _, obito := range obitos {
		wg.Add(1)
		go func(o models.ObitoSimulado) {
			defer wg.Done()
			l.processObito(ctx, &o)
		}(obito)
	}

	wg.Wait()

	// Update last processed time
	now := time.Now()
	l.ultimoProcessamento.Store(&now)
}

// processObito processes a single obito
func (l *ObitoListener) processObito(ctx context.Context, obito *models.ObitoSimulado) {
	// Check if already processed (idempotency)
	processed, err := l.obitoRepo.IsProcessed(ctx, obito.ID)
	if err != nil {
		l.logger.Printf("[Listener] Error checking if obito %s is processed: %v", obito.ID, err)
		atomic.AddInt64(&l.errors, 1)
		return
	}

	if processed {
		l.logger.Printf("[Listener] Obito %s already processed, skipping", obito.ID)
		return
	}

	hospitalNome := "Unknown"
	if obito.Hospital != nil {
		hospitalNome = obito.Hospital.Nome
	}

	l.logger.Printf("[Listener] Processing obito: ID=%s, Hospital=%s, Paciente=%s",
		obito.ID, hospitalNome, models.MaskName(obito.NomePaciente))

	// Publish to Redis Streams
	err = l.publishToStream(ctx, obito)
	if err != nil {
		l.logger.Printf("[Listener] Error publishing obito %s to stream: %v", obito.ID, err)
		atomic.AddInt64(&l.errors, 1)
		return
	}

	// Mark as processed
	err = l.obitoRepo.MarkAsProcessed(ctx, obito.ID)
	if err != nil {
		l.logger.Printf("[Listener] Error marking obito %s as processed: %v", obito.ID, err)
		atomic.AddInt64(&l.errors, 1)
		return
	}

	atomic.AddInt64(&l.totalProcessados, 1)
	l.logger.Printf("[Listener] Successfully processed obito: ID=%s, Hospital=%s, Timestamp=%s",
		obito.ID, hospitalNome, time.Now().Format(time.RFC3339))
}

// publishToStream publishes an obito event to Redis Streams
func (l *ObitoListener) publishToStream(ctx context.Context, obito *models.ObitoSimulado) error {
	setor := ""
	if obito.Setor != nil {
		setor = *obito.Setor
	}

	leito := ""
	if obito.Leito != nil {
		leito = *obito.Leito
	}

	event := ObitoEvent{
		ObitoID:                   obito.ID.String(),
		HospitalID:                obito.HospitalID.String(),
		TimestampDeteccao:         time.Now().Format(time.RFC3339),
		NomePaciente:              obito.NomePaciente,
		DataObito:                 obito.DataObito.Format(time.RFC3339),
		CausaMortis:               obito.CausaMortis,
		Setor:                     setor,
		Leito:                     leito,
		Idade:                     obito.CalculateAge(),
		IdentificacaoDesconhecida: obito.IdentificacaoDesconhecida,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Use XADD to publish to Redis Streams
	args := &redis.XAddArgs{
		Stream: ObitosStreamName,
		Values: map[string]interface{}{
			"data": string(eventJSON),
		},
	}

	_, err = l.redis.XAdd(ctx, args).Result()
	if err != nil {
		return err
	}

	l.logger.Printf("[Listener] Published obito %s to stream %s", obito.ID, ObitosStreamName)

	return nil
}

// SetLogger sets a custom logger for the listener
func (l *ObitoListener) SetLogger(logger *log.Logger) {
	l.logger = logger
}

// ParseObitoEvent parses an obito event from Redis stream data
func ParseObitoEvent(data string) (*ObitoEvent, error) {
	var event ObitoEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetObitoID returns the obito ID as UUID
func (e *ObitoEvent) GetObitoID() (uuid.UUID, error) {
	return uuid.Parse(e.ObitoID)
}

// GetHospitalID returns the hospital ID as UUID
func (e *ObitoEvent) GetHospitalID() (uuid.UUID, error) {
	return uuid.Parse(e.HospitalID)
}

// GetDataObito returns the data obito as time.Time
func (e *ObitoEvent) GetDataObito() (time.Time, error) {
	return time.Parse(time.RFC3339, e.DataObito)
}

// GetTimestampDeteccao returns the detection timestamp as time.Time
func (e *ObitoEvent) GetTimestampDeteccao() (time.Time, error) {
	return time.Parse(time.RFC3339, e.TimestampDeteccao)
}
