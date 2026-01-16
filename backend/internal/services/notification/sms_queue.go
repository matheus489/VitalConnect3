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
	// SMSQueueKey is the Redis key for the SMS queue
	SMSQueueKey = "vitalconnect:sms_queue"

	// SMSProcessingKey is the Redis key for SMS being processed
	SMSProcessingKey = "vitalconnect:sms_processing"

	// SMSMaxRetries is the maximum number of retry attempts
	SMSMaxRetries = 5

	// SMSBaseBackoffDelay is the base delay for exponential backoff
	SMSBaseBackoffDelay = 1 * time.Second
)

var (
	ErrSMSQueueFull      = errors.New("SMS queue is full")
	ErrInvalidSMSPayload = errors.New("invalid SMS payload")
)

// SMSQueueItem represents an item in the SMS queue
type SMSQueueItem struct {
	ID            string     `json:"id"`
	OccurrenceID  string     `json:"occurrence_id"`
	UserID        *string    `json:"user_id,omitempty"`
	PhoneNumber   string     `json:"phone_number"`
	Message       string     `json:"message"`
	Retries       int        `json:"retries"`
	CreatedAt     time.Time  `json:"created_at"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	Error         string     `json:"error,omitempty"`
}

// SMSQueueWorker processes SMS messages from the queue
type SMSQueueWorker struct {
	redis            *redis.Client
	smsService       *SMSService
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

// NewSMSQueueWorker creates a new SMSQueueWorker
func NewSMSQueueWorker(redisClient *redis.Client, smsService *SMSService, db *sql.DB) *SMSQueueWorker {
	return &SMSQueueWorker{
		redis:            redisClient,
		smsService:       smsService,
		notificationRepo: repository.NewNotificationRepository(db),
		stopCh:           make(chan struct{}),
		doneCh:           make(chan struct{}),
		logger:           log.Default(),
		pollInterval:     5 * time.Second,
		batchSize:        10,
	}
}

// Start begins the worker loop
func (w *SMSQueueWorker) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return nil // Already running
	}

	w.logger.Println("[SMSQueue] Starting SMS queue worker")

	go w.processLoop(ctx)

	return nil
}

// Stop stops the worker loop
func (w *SMSQueueWorker) Stop() {
	if atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		close(w.stopCh)
		<-w.doneCh
		w.logger.Println("[SMSQueue] SMS queue worker stopped")
	}
}

// IsRunning returns true if the worker is running
func (w *SMSQueueWorker) IsRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

// EnqueueSMS adds an SMS to the queue
func (w *SMSQueueWorker) EnqueueSMS(ctx context.Context, occurrenceID uuid.UUID, phoneNumber string, userID *uuid.UUID, message string) error {
	item := &SMSQueueItem{
		ID:           uuid.New().String(),
		OccurrenceID: occurrenceID.String(),
		PhoneNumber:  phoneNumber,
		Message:      message,
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

	w.logger.Printf("[SMSQueue] Enqueuing SMS for occurrence %s to %s", occurrenceID, MaskPhoneForLog(phoneNumber))

	return w.redis.LPush(ctx, SMSQueueKey, payload).Err()
}

// processLoop is the main processing loop
func (w *SMSQueueWorker) processLoop(ctx context.Context) {
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
func (w *SMSQueueWorker) processQueue(ctx context.Context) {
	for i := 0; i < w.batchSize; i++ {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
		}

		// Get an item from the queue
		result, err := w.redis.RPopLPush(ctx, SMSQueueKey, SMSProcessingKey).Result()
		if err != nil {
			if err == redis.Nil {
				return // Queue is empty
			}
			w.logger.Printf("[SMSQueue] Error getting item from queue: %v", err)
			atomic.AddInt64(&w.errors, 1)
			return
		}

		// Process the item
		var item SMSQueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			w.logger.Printf("[SMSQueue] Error unmarshalling queue item: %v", err)
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

		// Process the SMS
		w.processSMS(ctx, &item, result)
	}
}

// processSMS sends an SMS and records the notification
func (w *SMSQueueWorker) processSMS(ctx context.Context, item *SMSQueueItem, rawPayload string) {
	now := time.Now()
	item.LastAttemptAt = &now

	atomic.AddInt64(&w.totalProcessed, 1)

	// Send the SMS
	err := w.smsService.SendSMS(ctx, item.PhoneNumber, item.Message)

	occurrenceID, _ := uuid.Parse(item.OccurrenceID)
	var userID *uuid.UUID
	if item.UserID != nil {
		uid, _ := uuid.Parse(*item.UserID)
		userID = &uid
	}

	metadata := &models.NotificationMetadata{
		SMSTo:      item.PhoneNumber,
		SMSMessage: item.Message,
	}

	if err != nil {
		item.Error = err.Error()
		item.Retries++

		w.logger.Printf("[SMSQueue] Failed to send SMS to %s (attempt %d/%d): %v",
			MaskPhoneForLog(item.PhoneNumber), item.Retries, SMSMaxRetries, err)

		if item.Retries < SMSMaxRetries {
			// Calculate next retry time with exponential backoff (1s, 2s, 4s, 8s, 16s)
			backoff := time.Duration(math.Pow(2, float64(item.Retries-1))) * SMSBaseBackoffDelay
			nextRetry := time.Now().Add(backoff)
			item.NextRetryAt = &nextRetry

			w.logger.Printf("[SMSQueue] Scheduling retry %d for SMS to %s at %s (backoff: %v)",
				item.Retries, MaskPhoneForLog(item.PhoneNumber), nextRetry.Format(time.RFC3339), backoff)

			// Requeue for retry
			w.requeue(ctx, item, rawPayload)
		} else {
			// Max retries reached, record failure (DLQ)
			atomic.AddInt64(&w.totalFailed, 1)
			errMsg := err.Error()
			w.logger.Printf("[SMSQueue] SMS to %s moved to DLQ after %d attempts: %v",
				MaskPhoneForLog(item.PhoneNumber), SMSMaxRetries, err)
			_, _ = w.notificationRepo.CreateNotificationFromSMS(ctx, occurrenceID, userID, metadata, models.NotificationStatusFalha, &errMsg)
			w.removeFromProcessing(ctx, rawPayload)
		}
		return
	}

	// Success
	atomic.AddInt64(&w.totalSuccessful, 1)
	w.logger.Printf("[SMSQueue] Successfully sent SMS to %s for occurrence %s",
		MaskPhoneForLog(item.PhoneNumber), item.OccurrenceID)

	// Record successful notification
	_, _ = w.notificationRepo.CreateNotificationFromSMS(ctx, occurrenceID, userID, metadata, models.NotificationStatusEnviado, nil)
	w.removeFromProcessing(ctx, rawPayload)
}

// requeue puts an item back in the queue for retry
func (w *SMSQueueWorker) requeue(ctx context.Context, item *SMSQueueItem, rawPayload string) {
	payload, err := json.Marshal(item)
	if err != nil {
		w.logger.Printf("[SMSQueue] Error marshalling item for requeue: %v", err)
		w.removeFromProcessing(ctx, rawPayload)
		return
	}

	// Remove from processing and add back to queue
	w.redis.LRem(ctx, SMSProcessingKey, 1, rawPayload)
	w.redis.LPush(ctx, SMSQueueKey, payload)
}

// removeFromProcessing removes an item from the processing list
func (w *SMSQueueWorker) removeFromProcessing(ctx context.Context, rawPayload string) {
	w.redis.LRem(ctx, SMSProcessingKey, 1, rawPayload)
}

// GetStats returns statistics about the queue worker
func (w *SMSQueueWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":          w.IsRunning(),
		"total_processed":  atomic.LoadInt64(&w.totalProcessed),
		"total_successful": atomic.LoadInt64(&w.totalSuccessful),
		"total_failed":     atomic.LoadInt64(&w.totalFailed),
		"errors":           atomic.LoadInt64(&w.errors),
	}
}

// GetQueueLength returns the current queue length
func (w *SMSQueueWorker) GetQueueLength(ctx context.Context) (int64, error) {
	return w.redis.LLen(ctx, SMSQueueKey).Result()
}

// SetLogger sets a custom logger
func (w *SMSQueueWorker) SetLogger(logger *log.Logger) {
	w.logger = logger
}

// CalculateBackoffDelay calculates the backoff delay for a given retry count
// Uses exponential backoff: 1s, 2s, 4s, 8s, 16s
func CalculateBackoffDelay(retries int) time.Duration {
	if retries <= 0 {
		return SMSBaseBackoffDelay
	}
	return time.Duration(math.Pow(2, float64(retries-1))) * SMSBaseBackoffDelay
}
