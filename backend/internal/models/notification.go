package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationChannel represents the notification channel enum
type NotificationChannel string

const (
	ChannelDashboard NotificationChannel = "dashboard"
	ChannelEmail     NotificationChannel = "email"
	ChannelSMS       NotificationChannel = "sms"
)

// ValidChannels contains all valid notification channels
var ValidChannels = []NotificationChannel{ChannelDashboard, ChannelEmail, ChannelSMS}

// IsValid checks if the channel is a valid notification channel
func (c NotificationChannel) IsValid() bool {
	for _, valid := range ValidChannels {
		if c == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the channel
func (c NotificationChannel) String() string {
	return string(c)
}

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusEnviado  NotificationStatus = "enviado"
	NotificationStatusFalha    NotificationStatus = "falha"
	NotificationStatusPendente NotificationStatus = "pendente"
)

// Notification represents a sent notification
type Notification struct {
	ID            uuid.UUID           `json:"id" db:"id"`
	OccurrenceID  uuid.UUID           `json:"occurrence_id" db:"occurrence_id" validate:"required"`
	UserID        *uuid.UUID          `json:"user_id,omitempty" db:"user_id"`
	Canal         NotificationChannel `json:"canal" db:"canal" validate:"required,oneof=dashboard email sms"`
	EnviadoEm     time.Time           `json:"enviado_em" db:"enviado_em"`
	StatusEnvio   NotificationStatus  `json:"status_envio" db:"status_envio"`
	ErroMensagem  *string             `json:"erro_mensagem,omitempty" db:"erro_mensagem"`
	Metadata      json.RawMessage     `json:"metadata,omitempty" db:"metadata"`

	// Related data (populated by queries)
	Occurrence *Occurrence `json:"occurrence,omitempty" db:"-"`
	User       *User       `json:"user,omitempty" db:"-"`
}

// NotificationMetadata represents additional notification data
type NotificationMetadata struct {
	// Email fields
	EmailTo       string `json:"email_to,omitempty"`
	EmailSubject  string `json:"email_subject,omitempty"`

	// SMS fields
	SMSTo      string `json:"sms_to,omitempty"`
	SMSMessage string `json:"sms_message,omitempty"`

	// Common fields
	HospitalNome  string `json:"hospital_nome,omitempty"`
	Setor         string `json:"setor,omitempty"`
	TempoRestante string `json:"tempo_restante,omitempty"`
}

// CreateNotificationInput represents input for creating a notification
type CreateNotificationInput struct {
	OccurrenceID uuid.UUID           `json:"occurrence_id" validate:"required"`
	UserID       *uuid.UUID          `json:"user_id,omitempty"`
	Canal        NotificationChannel `json:"canal" validate:"required,oneof=dashboard email sms"`
	StatusEnvio  NotificationStatus  `json:"status_envio,omitempty"`
	ErroMensagem *string             `json:"erro_mensagem,omitempty"`
	Metadata     json.RawMessage     `json:"metadata,omitempty"`
}

// NotificationResponse represents the API response for a notification
type NotificationResponse struct {
	ID           uuid.UUID           `json:"id"`
	OccurrenceID uuid.UUID           `json:"occurrence_id"`
	UserID       *uuid.UUID          `json:"user_id,omitempty"`
	Canal        NotificationChannel `json:"canal"`
	EnviadoEm    time.Time           `json:"enviado_em"`
	StatusEnvio  NotificationStatus  `json:"status_envio"`
	ErroMensagem *string             `json:"erro_mensagem,omitempty"`
}

// ToResponse converts Notification to NotificationResponse
func (n *Notification) ToResponse() NotificationResponse {
	return NotificationResponse{
		ID:           n.ID,
		OccurrenceID: n.OccurrenceID,
		UserID:       n.UserID,
		Canal:        n.Canal,
		EnviadoEm:    n.EnviadoEm,
		StatusEnvio:  n.StatusEnvio,
		ErroMensagem: n.ErroMensagem,
	}
}

// PushSubscription represents a user's push notification subscription
type PushSubscription struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`          // FCM registration token
	Platform  string    `json:"platform" db:"platform"`    // web, android, ios
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SSEEvent represents a Server-Sent Event for dashboard notifications
type SSEEvent struct {
	Type         string    `json:"type"`
	OccurrenceID uuid.UUID `json:"occurrence_id"`
	HospitalNome string    `json:"hospital_nome"`
	Setor        string    `json:"setor"`
	DataObito    time.Time `json:"data_obito"`
	TempoRestante string   `json:"tempo_restante"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewOccurrenceSSEEvent creates a new SSE event for a new occurrence
func NewOccurrenceSSEEvent(occurrence *Occurrence, hospitalNome string) SSEEvent {
	setor := ""
	var data OccurrenceCompleteData
	if err := json.Unmarshal(occurrence.DadosCompletos, &data); err == nil {
		setor = data.Setor
	}

	return SSEEvent{
		Type:          "new_occurrence",
		OccurrenceID:  occurrence.ID,
		HospitalNome:  hospitalNome,
		Setor:         setor,
		DataObito:     occurrence.DataObito,
		TempoRestante: occurrence.FormatTimeRemaining(),
		CreatedAt:     time.Now(),
	}
}

// SSE Event Types for AI Assistant
const (
	// SSEEventTypeAIResponseChunk represents a chunk of AI response text
	SSEEventTypeAIResponseChunk = "ai_response_chunk"
	// SSEEventTypeAIThinking represents the AI is processing/thinking
	SSEEventTypeAIThinking = "ai_thinking"
	// SSEEventTypeAIToolCall represents a tool/function call by the AI
	SSEEventTypeAIToolCall = "ai_tool_call"
	// SSEEventTypeAIDone represents the AI has finished responding
	SSEEventTypeAIDone = "ai_done"
	// SSEEventTypeAIError represents an error occurred during AI processing
	SSEEventTypeAIError = "ai_error"
	// SSEEventTypeAIConfirmationRequired represents an action needs user confirmation
	SSEEventTypeAIConfirmationRequired = "ai_confirmation_required"
)

// AISSEEvent represents a Server-Sent Event for AI assistant responses
type AISSEEvent struct {
	Type      string                 `json:"type"`
	SessionID string                 `json:"session_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
	ToolCall  *AIToolCallEvent       `json:"tool_call,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// AIToolCallEvent represents details of an AI tool call
type AIToolCallEvent struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Result     interface{}            `json:"result,omitempty"`
	Status     string                 `json:"status"` // executing, completed, failed
}

// AIConfirmationEvent represents a pending action requiring user confirmation
type AIConfirmationEvent struct {
	ActionID    string                 `json:"action_id"`
	ActionType  string                 `json:"action_type"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// NewAIResponseChunkEvent creates an AI response chunk event
func NewAIResponseChunkEvent(sessionID, content string) AISSEEvent {
	return AISSEEvent{
		Type:      SSEEventTypeAIResponseChunk,
		SessionID: sessionID,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewAIThinkingEvent creates an AI thinking event
func NewAIThinkingEvent(sessionID, step string) AISSEEvent {
	return AISSEEvent{
		Type:      SSEEventTypeAIThinking,
		SessionID: sessionID,
		Content:   step,
		Timestamp: time.Now(),
	}
}

// NewAIToolCallEvent creates an AI tool call event
func NewAIToolCallEvent(sessionID string, toolCall *AIToolCallEvent) AISSEEvent {
	return AISSEEvent{
		Type:      SSEEventTypeAIToolCall,
		SessionID: sessionID,
		ToolCall:  toolCall,
		Timestamp: time.Now(),
	}
}

// NewAIDoneEvent creates an AI done event
func NewAIDoneEvent(sessionID string, metadata map[string]interface{}) AISSEEvent {
	return AISSEEvent{
		Type:      SSEEventTypeAIDone,
		SessionID: sessionID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}
}

// NewAIErrorEvent creates an AI error event
func NewAIErrorEvent(sessionID, errorMsg string) AISSEEvent {
	return AISSEEvent{
		Type:      SSEEventTypeAIError,
		SessionID: sessionID,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
}

// NewAIConfirmationRequiredEvent creates an AI confirmation required event
func NewAIConfirmationRequiredEvent(sessionID string, confirmation *AIConfirmationEvent) AISSEEvent {
	return AISSEEvent{
		Type:      SSEEventTypeAIConfirmationRequired,
		SessionID: sessionID,
		Metadata: map[string]interface{}{
			"confirmation": confirmation,
		},
		Timestamp: time.Now(),
	}
}
