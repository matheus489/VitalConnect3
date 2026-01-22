package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

// ShiftRepository handles shift data access
type ShiftRepository struct {
	db *sql.DB
}

// NewShiftRepository creates a new shift repository
func NewShiftRepository(db *sql.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

// Create creates a new shift
func (r *ShiftRepository) Create(ctx context.Context, input *models.CreateShiftInput) (*models.Shift, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	shift := &models.Shift{
		ID:         uuid.New(),
		HospitalID: input.HospitalID,
		UserID:     input.UserID,
		DayOfWeek:  input.DayOfWeek,
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
		INSERT INTO shifts (id, hospital_id, user_id, day_of_week, start_time, end_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		shift.ID,
		shift.HospitalID,
		shift.UserID,
		shift.DayOfWeek,
		shift.StartTime,
		shift.EndTime,
		shift.CreatedAt,
		shift.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return nil, models.ErrShiftExists
		}
		return nil, err
	}

	return shift, nil
}

// GetByID retrieves a shift by ID
func (r *ShiftRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Shift, error) {
	query := `
		SELECT
			s.id, s.hospital_id, s.user_id, s.day_of_week,
			s.start_time::text, s.end_time::text, s.created_at, s.updated_at,
			u.id, u.email, u.nome, u.role, u.ativo
		FROM shifts s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`

	var shift models.Shift
	var user models.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shift.ID, &shift.HospitalID, &shift.UserID, &shift.DayOfWeek,
		&shift.StartTime, &shift.EndTime, &shift.CreatedAt, &shift.UpdatedAt,
		&user.ID, &user.Email, &user.Nome, &user.Role, &user.Ativo,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrShiftNotFound
		}
		return nil, err
	}

	shift.User = &user

	return &shift, nil
}

// Update updates an existing shift
func (r *ShiftRepository) Update(ctx context.Context, id uuid.UUID, input *models.UpdateShiftInput) (*models.Shift, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// First get the existing shift
	shift, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.UserID != nil {
		shift.UserID = *input.UserID
	}
	if input.DayOfWeek != nil {
		shift.DayOfWeek = *input.DayOfWeek
	}
	if input.StartTime != nil {
		shift.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		shift.EndTime = *input.EndTime
	}
	shift.UpdatedAt = time.Now()

	query := `
		UPDATE shifts
		SET user_id = $1, day_of_week = $2, start_time = $3, end_time = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		shift.UserID,
		shift.DayOfWeek,
		shift.StartTime,
		shift.EndTime,
		shift.UpdatedAt,
		id,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, models.ErrShiftExists
		}
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, models.ErrShiftNotFound
	}

	return shift, nil
}

// Delete deletes a shift by ID
func (r *ShiftRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM shifts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrShiftNotFound
	}

	return nil
}

// ListByHospitalID retrieves all shifts for a hospital
func (r *ShiftRepository) ListByHospitalID(ctx context.Context, hospitalID uuid.UUID) ([]models.Shift, error) {
	query := `
		SELECT
			s.id, s.hospital_id, s.user_id, s.day_of_week,
			s.start_time::text, s.end_time::text, s.created_at, s.updated_at,
			u.id, u.email, u.nome, u.role, u.ativo
		FROM shifts s
		JOIN users u ON s.user_id = u.id
		WHERE s.hospital_id = $1
		ORDER BY s.day_of_week, s.start_time
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.Shift
	for rows.Next() {
		var shift models.Shift
		var user models.User

		err := rows.Scan(
			&shift.ID, &shift.HospitalID, &shift.UserID, &shift.DayOfWeek,
			&shift.StartTime, &shift.EndTime, &shift.CreatedAt, &shift.UpdatedAt,
			&user.ID, &user.Email, &user.Nome, &user.Role, &user.Ativo,
		)
		if err != nil {
			return nil, err
		}

		shift.User = &user
		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// GetShiftsByUserID retrieves all shifts for a user
func (r *ShiftRepository) GetShiftsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Shift, error) {
	query := `
		SELECT
			s.id, s.hospital_id, s.user_id, s.day_of_week,
			s.start_time::text, s.end_time::text, s.created_at, s.updated_at
		FROM shifts s
		WHERE s.user_id = $1
		ORDER BY s.day_of_week, s.start_time
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.Shift
	for rows.Next() {
		var shift models.Shift
		err := rows.Scan(
			&shift.ID, &shift.HospitalID, &shift.UserID, &shift.DayOfWeek,
			&shift.StartTime, &shift.EndTime, &shift.CreatedAt, &shift.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// GetActiveShifts retrieves operators currently on duty based on hospital, day of week, and current time
// This handles night shifts that cross midnight correctly
func (r *ShiftRepository) GetActiveShifts(ctx context.Context, hospitalID uuid.UUID, dayOfWeek int, currentTime time.Time) ([]models.Shift, error) {
	// For night shifts, we need to check both the current day and the previous day
	// because a shift starting on Monday at 19:00 covers early Tuesday morning
	previousDay := dayOfWeek - 1
	if previousDay < 0 {
		previousDay = 6
	}

	query := `
		SELECT
			s.id, s.hospital_id, s.user_id, s.day_of_week,
			s.start_time::text, s.end_time::text, s.created_at, s.updated_at,
			u.id, u.email, u.nome, u.role, u.ativo
		FROM shifts s
		JOIN users u ON s.user_id = u.id
		WHERE s.hospital_id = $1
		  AND u.ativo = true
		  AND (
			-- Day shift on current day: start_time < end_time and time within range
			(s.day_of_week = $2 AND s.start_time < s.end_time
			 AND $3::time >= s.start_time AND $3::time < s.end_time)
			OR
			-- Night shift starting current day: start_time > end_time and time >= start_time
			(s.day_of_week = $2 AND s.start_time > s.end_time
			 AND $3::time >= s.start_time)
			OR
			-- Night shift started previous day: start_time > end_time and time < end_time
			(s.day_of_week = $4 AND s.start_time > s.end_time
			 AND $3::time < s.end_time)
		  )
		ORDER BY s.start_time
	`

	timeStr := currentTime.Format("15:04:05")

	rows, err := r.db.QueryContext(ctx, query, hospitalID, dayOfWeek, timeStr, previousDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.Shift
	for rows.Next() {
		var shift models.Shift
		var user models.User

		err := rows.Scan(
			&shift.ID, &shift.HospitalID, &shift.UserID, &shift.DayOfWeek,
			&shift.StartTime, &shift.EndTime, &shift.CreatedAt, &shift.UpdatedAt,
			&user.ID, &user.Email, &user.Nome, &user.Role, &user.Ativo,
		)
		if err != nil {
			return nil, err
		}

		shift.User = &user
		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// GetTodayShifts retrieves all shifts scheduled for today for a hospital
func (r *ShiftRepository) GetTodayShifts(ctx context.Context, hospitalID uuid.UUID) ([]models.TodayShift, error) {
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	previousDay := dayOfWeek - 1
	if previousDay < 0 {
		previousDay = 6
	}

	query := `
		SELECT
			s.id, s.hospital_id, s.user_id, s.day_of_week,
			s.start_time::text, s.end_time::text, s.created_at, s.updated_at,
			u.id, u.email, u.nome, u.role, u.ativo
		FROM shifts s
		JOIN users u ON s.user_id = u.id
		WHERE s.hospital_id = $1
		  AND u.ativo = true
		  AND (
			-- Shifts that are scheduled for today
			s.day_of_week = $2
			OR
			-- Night shifts from yesterday that extend into today
			(s.day_of_week = $3 AND s.start_time > s.end_time)
		  )
		ORDER BY s.day_of_week, s.start_time
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID, dayOfWeek, previousDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todayShifts []models.TodayShift
	for rows.Next() {
		var shift models.Shift
		var user models.User

		err := rows.Scan(
			&shift.ID, &shift.HospitalID, &shift.UserID, &shift.DayOfWeek,
			&shift.StartTime, &shift.EndTime, &shift.CreatedAt, &shift.UpdatedAt,
			&user.ID, &user.Email, &user.Nome, &user.Role, &user.Ativo,
		)
		if err != nil {
			return nil, err
		}

		shift.User = &user

		// Determine if this shift is currently active
		isActive := shift.ContainsTime(now)

		todayShifts = append(todayShifts, models.TodayShift{
			Shift:    shift,
			IsActive: isActive,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todayShifts, nil
}

// GetCoverageGaps analyzes shift coverage and returns gaps (hours without any operator scheduled)
func (r *ShiftRepository) GetCoverageGaps(ctx context.Context, hospitalID uuid.UUID) (*models.CoverageAnalysis, error) {
	shifts, err := r.ListByHospitalID(ctx, hospitalID)
	if err != nil {
		return nil, err
	}

	analysis := &models.CoverageAnalysis{
		HospitalID:  hospitalID,
		TotalShifts: len(shifts),
		Gaps:        []models.CoverageGap{},
		HasGaps:     false,
	}

	// For each day of the week, check if there's full 24-hour coverage
	for day := models.Sunday; day <= models.Saturday; day++ {
		// Get shifts for this day
		dayShifts := filterShiftsByDay(shifts, day)

		// If no shifts for this day, the entire day is a gap
		if len(dayShifts) == 0 {
			analysis.Gaps = append(analysis.Gaps, models.CoverageGap{
				DayOfWeek: day,
				DayName:   day.String(),
				StartTime: "00:00",
				EndTime:   "23:59",
			})
			analysis.HasGaps = true
			continue
		}

		// Check for gaps in coverage
		// This is a simplified implementation that checks for obvious gaps
		// A more sophisticated implementation would merge overlapping shifts
		gaps := findGapsForDay(dayShifts)
		for _, gap := range gaps {
			gap.DayOfWeek = day
			gap.DayName = day.String()
			analysis.Gaps = append(analysis.Gaps, gap)
			analysis.HasGaps = true
		}
	}

	return analysis, nil
}

// filterShiftsByDay returns shifts that cover the specified day
func filterShiftsByDay(shifts []models.Shift, day models.DayOfWeek) []models.Shift {
	var result []models.Shift
	for _, s := range shifts {
		if s.DayOfWeek == day {
			result = append(result, s)
		}
		// Also include night shifts from the previous day that extend into this day
		previousDay := day - 1
		if previousDay < 0 {
			previousDay = 6
		}
		if s.DayOfWeek == previousDay && s.IsNightShift() {
			result = append(result, s)
		}
	}
	return result
}

// findGapsForDay finds gaps in coverage for a single day
func findGapsForDay(shifts []models.Shift) []models.CoverageGap {
	if len(shifts) == 0 {
		return []models.CoverageGap{{
			StartTime: "00:00",
			EndTime:   "23:59",
		}}
	}

	// Create a 24-hour coverage bitmap (resolution: 1 hour)
	covered := make([]bool, 24)

	for _, s := range shifts {
		markCoverage(covered, s)
	}

	// Find gaps
	var gaps []models.CoverageGap
	inGap := false
	gapStart := 0

	for hour := 0; hour < 24; hour++ {
		if !covered[hour] {
			if !inGap {
				inGap = true
				gapStart = hour
			}
		} else {
			if inGap {
				gaps = append(gaps, models.CoverageGap{
					StartTime: formatHour(gapStart),
					EndTime:   formatHour(hour),
				})
				inGap = false
			}
		}
	}

	// Handle gap that extends to end of day
	if inGap {
		gaps = append(gaps, models.CoverageGap{
			StartTime: formatHour(gapStart),
			EndTime:   "23:59",
		})
	}

	return gaps
}

// markCoverage marks hours covered by a shift
func markCoverage(covered []bool, s models.Shift) {
	startHour := s.StartTime.Hour()
	endHour := s.EndTime.Hour()

	if s.IsNightShift() {
		// Night shift: cover from start to midnight, then midnight to end
		for h := startHour; h < 24; h++ {
			covered[h] = true
		}
		for h := 0; h < endHour; h++ {
			covered[h] = true
		}
	} else {
		// Day shift: cover from start to end
		for h := startHour; h < endHour; h++ {
			covered[h] = true
		}
	}
}

// formatHour formats an hour as HH:00
func formatHour(hour int) string {
	if hour < 10 {
		return "0" + string(rune('0'+hour)) + ":00"
	}
	return string(rune('0'+hour/10)) + string(rune('0'+hour%10)) + ":00"
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code is 23505
	return containsErrorCode(err, "23505")
}

// containsErrorCode checks if the error message contains a specific PostgreSQL error code
func containsErrorCode(err error, code string) bool {
	return err != nil && (contains(err.Error(), code) || contains(err.Error(), "duplicate key"))
}

// contains is a simple string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
