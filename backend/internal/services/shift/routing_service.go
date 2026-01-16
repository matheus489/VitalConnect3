package shift

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

const (
	// CacheTTL is the time-to-live for cached shift data
	CacheTTL = 5 * time.Minute

	// CacheKeyPrefix is the prefix for Redis keys
	CacheKeyPrefix = "shift:hospital:"
)

var (
	ErrNoOperatorsOnDuty = errors.New("no operators on duty")
)

// ShiftRoutingService handles shift-based notification routing
type ShiftRoutingService struct {
	db          *sql.DB
	redis       *redis.Client
	shiftRepo   *repository.ShiftRepository
	userRepo    *repository.UserRepository
}

// NewShiftRoutingService creates a new shift routing service
func NewShiftRoutingService(db *sql.DB, redis *redis.Client) *ShiftRoutingService {
	return &ShiftRoutingService{
		db:        db,
		redis:     redis,
		shiftRepo: repository.NewShiftRepository(db),
		userRepo:  repository.NewUserRepository(db),
	}
}

// cachedOperator represents a cached operator entry
type cachedOperator struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Nome  string `json:"nome"`
}

// GetOnDutyOperators returns the operators currently on duty for a hospital at a given event time
// Uses the obito timestamp (not notification time) to determine the correct shift
// Implements fallback: if no operators scheduled, returns all Gestors for the hospital
func (s *ShiftRoutingService) GetOnDutyOperators(ctx context.Context, hospitalID uuid.UUID, eventTime time.Time) ([]models.User, error) {
	// Try to get from cache first
	operators, err := s.getFromCache(ctx, hospitalID, eventTime)
	if err == nil && len(operators) > 0 {
		return operators, nil
	}

	// Cache miss - query database
	dayOfWeek := int(eventTime.Weekday())
	shifts, err := s.shiftRepo.GetActiveShifts(ctx, hospitalID, dayOfWeek, eventTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get active shifts: %w", err)
	}

	// Extract unique user IDs from shifts
	userIDs := make(map[uuid.UUID]bool)
	for _, shift := range shifts {
		userIDs[shift.UserID] = true
	}

	// If operators found, return them
	if len(userIDs) > 0 {
		operators = make([]models.User, 0, len(userIDs))
		for _, shift := range shifts {
			if shift.User != nil && shift.User.Ativo {
				operators = append(operators, *shift.User)
			}
		}

		// Cache the result
		if err := s.setCache(ctx, hospitalID, eventTime, operators); err != nil {
			log.Printf("Warning: failed to cache shift operators: %v", err)
		}

		return operators, nil
	}

	// FALLBACK: No operators scheduled - return all Gestors for this hospital
	log.Printf("No operators on duty for hospital %s at %s, falling back to gestors", hospitalID, eventTime.Format(time.RFC3339))
	return s.getFallbackGestors(ctx, hospitalID)
}

// getFallbackGestors returns all active Gestors linked to the hospital
func (s *ShiftRoutingService) getFallbackGestors(ctx context.Context, hospitalID uuid.UUID) ([]models.User, error) {
	gestors, err := s.userRepo.ListByRoleAndHospital(ctx, string(models.RoleGestor), hospitalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fallback gestors: %w", err)
	}

	// Filter only active gestors
	activeGestors := make([]models.User, 0)
	for _, g := range gestors {
		if g.Ativo {
			activeGestors = append(activeGestors, g)
		}
	}

	if len(activeGestors) == 0 {
		log.Printf("Warning: No active gestors found for hospital %s", hospitalID)
		return nil, ErrNoOperatorsOnDuty
	}

	return activeGestors, nil
}

// GetOnDutyOperatorIDs returns just the IDs of operators on duty
func (s *ShiftRoutingService) GetOnDutyOperatorIDs(ctx context.Context, hospitalID uuid.UUID, eventTime time.Time) ([]uuid.UUID, error) {
	operators, err := s.GetOnDutyOperators(ctx, hospitalID, eventTime)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(operators))
	for _, op := range operators {
		ids = append(ids, op.ID)
	}

	return ids, nil
}

// cacheKey generates the Redis cache key for a hospital's current shift
func (s *ShiftRoutingService) cacheKey(hospitalID uuid.UUID) string {
	return fmt.Sprintf("%s%s:current", CacheKeyPrefix, hospitalID.String())
}

// getFromCache attempts to get operators from Redis cache
func (s *ShiftRoutingService) getFromCache(ctx context.Context, hospitalID uuid.UUID, eventTime time.Time) ([]models.User, error) {
	if s.redis == nil {
		return nil, errors.New("redis not configured")
	}

	key := s.cacheKey(hospitalID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("cache miss")
		}
		return nil, err
	}

	var cached struct {
		Operators []cachedOperator `json:"operators"`
		DayOfWeek int              `json:"day_of_week"`
		Hour      int              `json:"hour"`
	}

	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	// Validate that cached data is still relevant for this time
	if cached.DayOfWeek != int(eventTime.Weekday()) || cached.Hour != eventTime.Hour() {
		return nil, errors.New("cache stale")
	}

	// Convert cached operators to models
	operators := make([]models.User, 0, len(cached.Operators))
	for _, co := range cached.Operators {
		id, err := uuid.Parse(co.ID)
		if err != nil {
			continue
		}
		operators = append(operators, models.User{
			ID:    id,
			Email: co.Email,
			Nome:  co.Nome,
		})
	}

	return operators, nil
}

// setCache stores operators in Redis cache
func (s *ShiftRoutingService) setCache(ctx context.Context, hospitalID uuid.UUID, eventTime time.Time, operators []models.User) error {
	if s.redis == nil {
		return errors.New("redis not configured")
	}

	cached := struct {
		Operators []cachedOperator `json:"operators"`
		DayOfWeek int              `json:"day_of_week"`
		Hour      int              `json:"hour"`
	}{
		Operators: make([]cachedOperator, 0, len(operators)),
		DayOfWeek: int(eventTime.Weekday()),
		Hour:      eventTime.Hour(),
	}

	for _, op := range operators {
		cached.Operators = append(cached.Operators, cachedOperator{
			ID:    op.ID.String(),
			Email: op.Email,
			Nome:  op.Nome,
		})
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	key := s.cacheKey(hospitalID)
	return s.redis.Set(ctx, key, data, CacheTTL).Err()
}

// InvalidateCache invalidates the cache for a hospital
// Should be called when shifts are created, updated, or deleted
func (s *ShiftRoutingService) InvalidateCache(ctx context.Context, hospitalID uuid.UUID) error {
	if s.redis == nil {
		return nil // No-op if Redis not configured
	}

	key := s.cacheKey(hospitalID)
	return s.redis.Del(ctx, key).Err()
}

// InvalidateAllCaches invalidates all shift caches
// Use with caution - typically only needed for bulk operations
func (s *ShiftRoutingService) InvalidateAllCaches(ctx context.Context) error {
	if s.redis == nil {
		return nil
	}

	pattern := CacheKeyPrefix + "*"
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// GetCoverage returns the coverage analysis for a hospital
func (s *ShiftRoutingService) GetCoverage(ctx context.Context, hospitalID uuid.UUID) (*models.CoverageAnalysis, error) {
	return s.shiftRepo.GetCoverageGaps(ctx, hospitalID)
}

// HasCoverageGaps returns true if the hospital has any coverage gaps
func (s *ShiftRoutingService) HasCoverageGaps(ctx context.Context, hospitalID uuid.UUID) (bool, error) {
	analysis, err := s.GetCoverage(ctx, hospitalID)
	if err != nil {
		return false, err
	}
	return analysis.HasGaps, nil
}
