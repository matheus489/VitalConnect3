package notification

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
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
)

const (
	// SSEChannelName is the Redis Pub/Sub channel for SSE events
	SSEChannelName = "sidot:sse_events"

	// HeartbeatInterval is the interval for sending heartbeat events
	HeartbeatInterval = 30 * time.Second

	// ClientTimeout is the timeout for client connection check
	ClientTimeout = 60 * time.Second
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID        string
	UserID    string
	Role      string
	Channel   chan *models.SSEEvent
	Done      chan struct{}
	CreatedAt time.Time
}

// NewSSEClient creates a new SSE client
func NewSSEClient(userID, role string) *SSEClient {
	return &SSEClient{
		ID:        uuid.New().String(),
		UserID:    userID,
		Role:      role,
		Channel:   make(chan *models.SSEEvent, 100),
		Done:      make(chan struct{}),
		CreatedAt: time.Now(),
	}
}

// Close closes the client connection
func (c *SSEClient) Close() {
	select {
	case <-c.Done:
		// Already closed
	default:
		close(c.Done)
	}
}

// SSEHub manages SSE connections and event distribution
type SSEHub struct {
	redis            *redis.Client
	notificationRepo *repository.NotificationRepository

	// Connected clients
	clients   map[string]*SSEClient
	clientsMu sync.RWMutex

	// Status tracking
	running           int32
	totalConnections  int64
	totalBroadcasts   int64
	totalEventsPublished int64

	// Control
	stopCh chan struct{}
	doneCh chan struct{}

	// Logger
	logger *log.Logger
}

// NewSSEHub creates a new SSEHub
func NewSSEHub(redisClient *redis.Client, db *sql.DB) *SSEHub {
	return &SSEHub{
		redis:            redisClient,
		notificationRepo: repository.NewNotificationRepository(db),
		clients:          make(map[string]*SSEClient),
		stopCh:           make(chan struct{}),
		doneCh:           make(chan struct{}),
		logger:           log.Default(),
	}
}

// Start begins the SSE hub (subscribes to Redis Pub/Sub)
func (h *SSEHub) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&h.running, 0, 1) {
		return nil // Already running
	}

	h.logger.Println("[SSE] Starting SSE hub")

	go h.subscribeLoop(ctx)
	go h.heartbeatLoop(ctx)

	return nil
}

// Stop stops the SSE hub
func (h *SSEHub) Stop() {
	if atomic.CompareAndSwapInt32(&h.running, 1, 0) {
		close(h.stopCh)
		<-h.doneCh

		// Close all client connections
		h.clientsMu.Lock()
		for _, client := range h.clients {
			client.Close()
		}
		h.clients = make(map[string]*SSEClient)
		h.clientsMu.Unlock()

		h.logger.Println("[SSE] SSE hub stopped")
	}
}

// IsRunning returns true if the hub is running
func (h *SSEHub) IsRunning() bool {
	return atomic.LoadInt32(&h.running) == 1
}

// RegisterClient registers a new SSE client
func (h *SSEHub) RegisterClient(client *SSEClient) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	h.clients[client.ID] = client
	atomic.AddInt64(&h.totalConnections, 1)

	h.logger.Printf("[SSE] Client registered: %s (user: %s, role: %s)", client.ID, client.UserID, client.Role)
}

// UnregisterClient removes an SSE client
func (h *SSEHub) UnregisterClient(clientID string) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if client, exists := h.clients[clientID]; exists {
		client.Close()
		delete(h.clients, clientID)
		h.logger.Printf("[SSE] Client unregistered: %s", clientID)
	}
}

// GetClientCount returns the number of connected clients
func (h *SSEHub) GetClientCount() int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	return len(h.clients)
}

// PublishEvent publishes an event to all connected clients via Redis Pub/Sub
func (h *SSEHub) PublishEvent(ctx context.Context, event *models.SSEEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = h.redis.Publish(ctx, SSEChannelName, data).Err()
	if err != nil {
		return err
	}

	atomic.AddInt64(&h.totalEventsPublished, 1)
	h.logger.Printf("[SSE] Event published: type=%s, occurrence_id=%s", event.Type, event.OccurrenceID)

	return nil
}

// PublishNewOccurrence publishes a new occurrence event and records the notification
func (h *SSEHub) PublishNewOccurrence(ctx context.Context, occurrence *models.Occurrence, hospitalNome string) error {
	event := models.NewOccurrenceSSEEvent(occurrence, hospitalNome)

	// Publish to Redis Pub/Sub
	if err := h.PublishEvent(ctx, &event); err != nil {
		h.logger.Printf("[SSE] Error publishing event: %v", err)
		return err
	}

	// Record notification
	metadata := &models.NotificationMetadata{
		HospitalNome:  hospitalNome,
		Setor:         event.Setor,
		TempoRestante: event.TempoRestante,
	}

	_, err := h.notificationRepo.CreateNotificationFromSSE(ctx, occurrence.ID, metadata)
	if err != nil {
		h.logger.Printf("[SSE] Warning: Failed to record notification: %v", err)
		// Don't fail the whole operation for notification recording error
	}

	return nil
}

// subscribeLoop subscribes to Redis Pub/Sub and broadcasts events to clients
func (h *SSEHub) subscribeLoop(ctx context.Context) {
	defer close(h.doneCh)

	pubsub := h.redis.Subscribe(ctx, SSEChannelName)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var event models.SSEEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				h.logger.Printf("[SSE] Error unmarshalling event: %v", err)
				continue
			}

			h.broadcastToClients(&event)
		}
	}
}

// broadcastToClients sends an event to all connected clients
func (h *SSEHub) broadcastToClients(event *models.SSEEvent) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	atomic.AddInt64(&h.totalBroadcasts, 1)

	for _, client := range h.clients {
		select {
		case client.Channel <- event:
			// Event sent successfully
		default:
			// Client channel is full, skip
			h.logger.Printf("[SSE] Client %s channel full, skipping event", client.ID)
		}
	}
}

// heartbeatLoop sends periodic heartbeat events to keep connections alive
func (h *SSEHub) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat event to all clients
func (h *SSEHub) sendHeartbeat() {
	heartbeat := &models.SSEEvent{
		Type:      "heartbeat",
		CreatedAt: time.Now(),
	}

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Channel <- heartbeat:
		default:
			// Channel full, skip
		}
	}
}

// GetStats returns statistics about the SSE hub
func (h *SSEHub) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":               h.IsRunning(),
		"connected_clients":     h.GetClientCount(),
		"total_connections":     atomic.LoadInt64(&h.totalConnections),
		"total_broadcasts":      atomic.LoadInt64(&h.totalBroadcasts),
		"total_events_published": atomic.LoadInt64(&h.totalEventsPublished),
	}
}

// SetLogger sets a custom logger
func (h *SSEHub) SetLogger(logger *log.Logger) {
	h.logger = logger
}
