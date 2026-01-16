package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

const (
	// EmailQueueKey is the Redis key for the email queue
	EmailQueueKey = "vitalconnect:email_queue"

	// EmailProcessingKey is the Redis key for emails being processed
	EmailProcessingKey = "vitalconnect:email_processing"

	// MaxRetries is the maximum number of retry attempts
	MaxRetries = 3

	// BaseBackoffDelay is the base delay for exponential backoff
	BaseBackoffDelay = 1 * time.Second
)

var (
	ErrQueueFull      = errors.New("email queue is full")
	ErrInvalidPayload = errors.New("invalid email payload")
)

// EmailQueueItem represents an item in the email queue
type EmailQueueItem struct {
	ID            string                 `json:"id"`
	OccurrenceID  string                 `json:"occurrence_id"`
	To            string                 `json:"to"`
	UserID        *string                `json:"user_id,omitempty"`
	Data          *ObitoNotificationData `json:"data"`
	Retries       int                    `json:"retries"`
	CreatedAt     time.Time              `json:"created_at"`
	LastAttemptAt *time.Time             `json:"last_attempt_at,omitempty"`
	NextRetryAt   *time.Time             `json:"next_retry_at,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

// EmailQueueWorker processes emails from the queue
type EmailQueueWorker struct {
	redis            *redis.Client
	emailService     *EmailService
	notificationRepo *repository.NotificationRepository

	// Status tracking
	running         int32
	totalProcessed  int64
	totalSuccessful int64
	totalFailed     int64
	errors          int64

	// Control
	stopCh chan struct{}
	doneCh chan struct{}
	wg     sync.WaitGroup

	// Logger
	logger *log.Logger

	// Configuration
	pollInterval time.Duration
	batchSize    int
}

// NewEmailQueueWorker creates a new EmailQueueWorker
func NewEmailQueueWorker(redisClient *redis.Client, emailService *EmailService, db *sql.DB) *EmailQueueWorker {
	return &EmailQueueWorker{
		redis:            redisClient,
		emailService:     emailService,
		notificationRepo: repository.NewNotificationRepository(db),
		stopCh:           make(chan struct{}),
		doneCh:           make(chan struct{}),
		logger:           log.Default(),
		pollInterval:     5 * time.Second,
		batchSize:        10,
	}
}

// Start begins the worker loop
func (w *EmailQueueWorker) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return nil // Already running
	}

	w.logger.Println("[EmailQueue] Starting email queue worker")

	go w.processLoop(ctx)

	return nil
}

// Stop stops the worker loop
func (w *EmailQueueWorker) Stop() {
	if atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		close(w.stopCh)
		<-w.doneCh
		w.logger.Println("[EmailQueue] Email queue worker stopped")
	}
}

// IsRunning returns true if the worker is running
func (w *EmailQueueWorker) IsRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

// EnqueueEmail adds an email to the queue
func (w *EmailQueueWorker) EnqueueEmail(ctx context.Context, occurrenceID uuid.UUID, to string, userID *uuid.UUID, data *ObitoNotificationData) error {
	item := &EmailQueueItem{
		ID:           uuid.New().String(),
		OccurrenceID: occurrenceID.String(),
		To:           to,
		Data:         data,
		Retries:      0,
		CreatedAt:    time.Now(),
	}

	if userID != nil {
		userIDStr := userID.String()
		item.UserID = &userIDStr
	}

	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return w.redis.LPush(ctx, EmailQueueKey, payload).Err()
}

// processLoop is the main processing loop
func (w *EmailQueueWorker) processLoop(ctx context.Context) {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processQueue(ctx)
		}
	}
}

// processQueue processes items from the queue
func (w *EmailQueueWorker) processQueue(ctx context.Context) {
	for i := 0; i < w.batchSize; i++ {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
		}

		// Get an item from the queue
		result, err := w.redis.RPopLPush(ctx, EmailQueueKey, EmailProcessingKey).Result()
		if err != nil {
			if err == redis.Nil {
				return // Queue is empty
			}
			w.logger.Printf("[EmailQueue] Error getting item from queue: %v", err)
			atomic.AddInt64(&w.errors, 1)
			return
		}

		// Process the item
		var item EmailQueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			w.logger.Printf("[EmailQueue] Error unmarshalling queue item: %v", err)
			w.removeFromProcessing(ctx, result)
			atomic.AddInt64(&w.errors, 1)
			continue
		}

		// Check if we should retry (exponential backoff)
		if item.NextRetryAt != nil && time.Now().Before(*item.NextRetryAt) {
			// Put back in queue for later
			w.requeue(ctx, &item, result)
			continue
		}

		// Process the email
		w.processEmail(ctx, &item, result)
	}
}

// processEmail sends an email and records the notification
func (w *EmailQueueWorker) processEmail(ctx context.Context, item *EmailQueueItem, rawPayload string) {
	now := time.Now()
	item.LastAttemptAt = &now

	atomic.AddInt64(&w.totalProcessed, 1)

	// Send the email
	err := w.emailService.SendObitoNotification(ctx, item.To, item.Data)

	occurrenceID, _ := uuid.Parse(item.OccurrenceID)
	var userID *uuid.UUID
	if item.UserID != nil {
		uid, _ := uuid.Parse(*item.UserID)
		userID = &uid
	}

	metadata := &models.NotificationMetadata{
		EmailTo:       item.To,
		EmailSubject:  "[URGENTE] Nova Ocorrencia Elegivel - " + item.Data.HospitalNome,
		HospitalNome:  item.Data.HospitalNome,
		Setor:         item.Data.Setor,
		TempoRestante: item.Data.TempoRestante,
	}

	if err != nil {
		item.Error = err.Error()
		item.Retries++

		w.logger.Printf("[EmailQueue] Failed to send email to %s (attempt %d/%d): %v",
			item.To, item.Retries, MaxRetries, err)

		if item.Retries < MaxRetries {
			// Calculate next retry time with exponential backoff
			backoff := time.Duration(math.Pow(2, float64(item.Retries))) * BaseBackoffDelay
			nextRetry := time.Now().Add(backoff)
			item.NextRetryAt = &nextRetry

			// Requeue for retry
			w.requeue(ctx, item, rawPayload)
		} else {
			// Max retries reached, record failure
			atomic.AddInt64(&w.totalFailed, 1)
			errMsg := err.Error()
			_, _ = w.notificationRepo.CreateNotificationFromEmail(ctx, occurrenceID, userID, metadata, models.NotificationStatusFalha, &errMsg)
			w.removeFromProcessing(ctx, rawPayload)
		}
		return
	}

	// Success
	atomic.AddInt64(&w.totalSuccessful, 1)
	w.logger.Printf("[EmailQueue] Successfully sent email to %s for occurrence %s", item.To, item.OccurrenceID)

	// Record successful notification
	_, _ = w.notificationRepo.CreateNotificationFromEmail(ctx, occurrenceID, userID, metadata, models.NotificationStatusEnviado, nil)
	w.removeFromProcessing(ctx, rawPayload)
}

// requeue puts an item back in the queue for retry
func (w *EmailQueueWorker) requeue(ctx context.Context, item *EmailQueueItem, rawPayload string) {
	payload, err := json.Marshal(item)
	if err != nil {
		w.logger.Printf("[EmailQueue] Error marshalling item for requeue: %v", err)
		w.removeFromProcessing(ctx, rawPayload)
		return
	}

	// Remove from processing and add back to queue
	w.redis.LRem(ctx, EmailProcessingKey, 1, rawPayload)
	w.redis.LPush(ctx, EmailQueueKey, payload)
}

// removeFromProcessing removes an item from the processing list
func (w *EmailQueueWorker) removeFromProcessing(ctx context.Context, rawPayload string) {
	w.redis.LRem(ctx, EmailProcessingKey, 1, rawPayload)
}

// GetStats returns statistics about the queue worker
func (w *EmailQueueWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":          w.IsRunning(),
		"total_processed":  atomic.LoadInt64(&w.totalProcessed),
		"total_successful": atomic.LoadInt64(&w.totalSuccessful),
		"total_failed":     atomic.LoadInt64(&w.totalFailed),
		"errors":           atomic.LoadInt64(&w.errors),
	}
}

// GetQueueLength returns the current queue length
func (w *EmailQueueWorker) GetQueueLength(ctx context.Context) (int64, error) {
	return w.redis.LLen(ctx, EmailQueueKey).Result()
}

// SetLogger sets a custom logger
func (w *EmailQueueWorker) SetLogger(logger *log.Logger) {
	w.logger = logger
}
