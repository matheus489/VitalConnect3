package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

// NotificationRepository handles database operations for notifications
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification record
func (r *NotificationRepository) Create(ctx context.Context, input *models.CreateNotificationInput) (*models.Notification, error) {
	notification := &models.Notification{
		ID:           uuid.New(),
		OccurrenceID: input.OccurrenceID,
		UserID:       input.UserID,
		Canal:        input.Canal,
		EnviadoEm:    time.Now(),
		StatusEnvio:  input.StatusEnvio,
		ErroMensagem: input.ErroMensagem,
		Metadata:     input.Metadata,
	}

	if notification.StatusEnvio == "" {
		notification.StatusEnvio = models.NotificationStatusEnviado
	}

	query := `
		INSERT INTO notifications (id, occurrence_id, user_id, canal, enviado_em, status_envio, erro_mensagem, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, occurrence_id, user_id, canal, enviado_em, status_envio, erro_mensagem, metadata
	`

	var metadataBytes []byte
	if notification.Metadata != nil {
		metadataBytes = notification.Metadata
	}

	err := r.db.QueryRowContext(ctx, query,
		notification.ID,
		notification.OccurrenceID,
		notification.UserID,
		notification.Canal,
		notification.EnviadoEm,
		notification.StatusEnvio,
		notification.ErroMensagem,
		metadataBytes,
	).Scan(
		&notification.ID,
		&notification.OccurrenceID,
		&notification.UserID,
		&notification.Canal,
		&notification.EnviadoEm,
		&notification.StatusEnvio,
		&notification.ErroMensagem,
		&notification.Metadata,
	)

	if err != nil {
		return nil, err
	}

	return notification, nil
}

// GetByID retrieves a notification by its ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, occurrence_id, user_id, canal, enviado_em, status_envio, erro_mensagem, metadata
		FROM notifications
		WHERE id = $1
	`

	notification := &models.Notification{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.OccurrenceID,
		&notification.UserID,
		&notification.Canal,
		&notification.EnviadoEm,
		&notification.StatusEnvio,
		&notification.ErroMensagem,
		&notification.Metadata,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}

	return notification, nil
}

// GetByOccurrenceID retrieves all notifications for an occurrence
func (r *NotificationRepository) GetByOccurrenceID(ctx context.Context, occurrenceID uuid.UUID) ([]models.Notification, error) {
	query := `
		SELECT id, occurrence_id, user_id, canal, enviado_em, status_envio, erro_mensagem, metadata
		FROM notifications
		WHERE occurrence_id = $1
		ORDER BY enviado_em DESC
	`

	rows, err := r.db.QueryContext(ctx, query, occurrenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID,
			&n.OccurrenceID,
			&n.UserID,
			&n.Canal,
			&n.EnviadoEm,
			&n.StatusEnvio,
			&n.ErroMensagem,
			&n.Metadata,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// UpdateStatus updates the status of a notification
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus, errorMsg *string) error {
	query := `
		UPDATE notifications
		SET status_envio = $2, erro_mensagem = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status, errorMsg)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// GetAverageNotificationTime calculates the average notification time in seconds
func (r *NotificationRepository) GetAverageNotificationTime(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(
			AVG(EXTRACT(EPOCH FROM (n.enviado_em - o.created_at))),
			0
		)
		FROM notifications n
		JOIN occurrences o ON n.occurrence_id = o.id
		WHERE n.status_envio = 'enviado'
		AND n.canal = 'dashboard'
		AND n.enviado_em >= CURRENT_DATE
	`

	var avgTime float64
	err := r.db.QueryRowContext(ctx, query).Scan(&avgTime)
	if err != nil {
		return 0, err
	}

	return avgTime, nil
}

// CountTodayByChannel counts notifications sent today by channel
func (r *NotificationRepository) CountTodayByChannel(ctx context.Context, canal models.NotificationChannel) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE canal = $1
		AND status_envio = 'enviado'
		AND enviado_em >= CURRENT_DATE
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, canal).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateNotificationFromSSE creates a notification record for SSE events
func (r *NotificationRepository) CreateNotificationFromSSE(ctx context.Context, occurrenceID uuid.UUID, metadata *models.NotificationMetadata) (*models.Notification, error) {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metadataJSON = data
	}

	input := &models.CreateNotificationInput{
		OccurrenceID: occurrenceID,
		Canal:        models.ChannelDashboard,
		StatusEnvio:  models.NotificationStatusEnviado,
		Metadata:     metadataJSON,
	}

	return r.Create(ctx, input)
}

// CreateNotificationFromEmail creates a notification record for email events
func (r *NotificationRepository) CreateNotificationFromEmail(ctx context.Context, occurrenceID uuid.UUID, userID *uuid.UUID, metadata *models.NotificationMetadata, status models.NotificationStatus, errorMsg *string) (*models.Notification, error) {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metadataJSON = data
	}

	input := &models.CreateNotificationInput{
		OccurrenceID: occurrenceID,
		UserID:       userID,
		Canal:        models.ChannelEmail,
		StatusEnvio:  status,
		ErroMensagem: errorMsg,
		Metadata:     metadataJSON,
	}

	return r.Create(ctx, input)
}

// CreateNotificationFromSMS creates a notification record for SMS events
func (r *NotificationRepository) CreateNotificationFromSMS(ctx context.Context, occurrenceID uuid.UUID, userID *uuid.UUID, metadata *models.NotificationMetadata, status models.NotificationStatus, errorMsg *string) (*models.Notification, error) {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metadataJSON = data
	}

	input := &models.CreateNotificationInput{
		OccurrenceID: occurrenceID,
		UserID:       userID,
		Canal:        models.ChannelSMS,
		StatusEnvio:  status,
		ErroMensagem: errorMsg,
		Metadata:     metadataJSON,
	}

	return r.Create(ctx, input)
}

// ExistsSMSForOccurrenceAndUser checks if an SMS notification already exists for an occurrence and user
// Used to prevent duplicate SMS notifications (limit 1 per occurrence per user)
func (r *NotificationRepository) ExistsSMSForOccurrenceAndUser(ctx context.Context, occurrenceID uuid.UUID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM notifications
			WHERE occurrence_id = $1
			AND user_id = $2
			AND canal = 'sms'
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, occurrenceID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
