package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	ErrSubscriptionNotFound = errors.New("push subscription not found")
	ErrSubscriptionExists   = errors.New("push subscription already exists")
)

// PushSubscriptionRepository handles push subscription data access
type PushSubscriptionRepository struct {
	db *sql.DB
}

// NewPushSubscriptionRepository creates a new push subscription repository
func NewPushSubscriptionRepository(db *sql.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

// Create creates a new push subscription
func (r *PushSubscriptionRepository) Create(ctx context.Context, userID uuid.UUID, token, platform, userAgent string) (*models.PushSubscription, error) {
	// Check if token already exists for this user
	existing, err := r.GetByToken(ctx, token)
	if err == nil && existing != nil {
		// Update the existing subscription
		return r.updateSubscription(ctx, existing.ID, userID, platform, userAgent)
	}

	sub := &models.PushSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		Platform:  platform,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO push_subscriptions (id, user_id, token, platform, user_agent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		sub.ID,
		sub.UserID,
		sub.Token,
		sub.Platform,
		sub.UserAgent,
		sub.CreatedAt,
		sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *PushSubscriptionRepository) updateSubscription(ctx context.Context, id, userID uuid.UUID, platform, userAgent string) (*models.PushSubscription, error) {
	query := `
		UPDATE push_subscriptions
		SET user_id = $1, platform = $2, user_agent = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, user_id, token, platform, user_agent, created_at, updated_at
	`

	var sub models.PushSubscription
	err := r.db.QueryRowContext(ctx, query, userID, platform, userAgent, time.Now(), id).Scan(
		&sub.ID, &sub.UserID, &sub.Token, &sub.Platform, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

// GetByToken retrieves a subscription by its FCM token
func (r *PushSubscriptionRepository) GetByToken(ctx context.Context, token string) (*models.PushSubscription, error) {
	query := `
		SELECT id, user_id, token, platform, user_agent, created_at, updated_at
		FROM push_subscriptions
		WHERE token = $1
	`

	var sub models.PushSubscription
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&sub.ID, &sub.UserID, &sub.Token, &sub.Platform, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}

	return &sub, nil
}

// GetByUserID retrieves all subscriptions for a user
func (r *PushSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.PushSubscription, error) {
	query := `
		SELECT id, user_id, token, platform, user_agent, created_at, updated_at
		FROM push_subscriptions
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.Token, &sub.Platform, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// GetByHospitalID retrieves all subscriptions for users linked to a hospital
func (r *PushSubscriptionRepository) GetByHospitalID(ctx context.Context, hospitalID uuid.UUID) ([]models.PushSubscription, error) {
	query := `
		SELECT ps.id, ps.user_id, ps.token, ps.platform, ps.user_agent, ps.created_at, ps.updated_at
		FROM push_subscriptions ps
		JOIN user_hospitals uh ON ps.user_id = uh.user_id
		JOIN users u ON ps.user_id = u.id
		WHERE uh.hospital_id = $1 AND u.ativo = true
		ORDER BY ps.updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.Token, &sub.Platform, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// Delete removes a subscription by token
func (r *PushSubscriptionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM push_subscriptions WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

// DeleteByUserID removes all subscriptions for a user
func (r *PushSubscriptionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM push_subscriptions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// CleanupStale removes subscriptions not updated in the given duration
func (r *PushSubscriptionRepository) CleanupStale(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM push_subscriptions WHERE updated_at < $1`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
