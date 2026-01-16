package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TestShiftCreateInputValidation tests the create input validation
func TestShiftCreateInputValidation(t *testing.T) {
	validInput := &models.CreateShiftInput{
		HospitalID: uuid.New(),
		UserID:     uuid.New(),
		DayOfWeek:  models.Monday,
		StartTime:  "07:00",
		EndTime:    "19:00",
	}

	if err := validInput.Validate(); err != nil {
		t.Errorf("Valid input should not return error, got: %v", err)
	}

	// Test invalid day of week
	invalidDay := &models.CreateShiftInput{
		HospitalID: uuid.New(),
		UserID:     uuid.New(),
		DayOfWeek:  models.DayOfWeek(7),
		StartTime:  "07:00",
		EndTime:    "19:00",
	}
	if err := invalidDay.Validate(); err != models.ErrInvalidDayOfWeek {
		t.Errorf("Expected ErrInvalidDayOfWeek, got: %v", err)
	}

	// Test invalid start time
	invalidStart := &models.CreateShiftInput{
		HospitalID: uuid.New(),
		UserID:     uuid.New(),
		DayOfWeek:  models.Monday,
		StartTime:  "invalid",
		EndTime:    "19:00",
	}
	if err := invalidStart.Validate(); err != models.ErrInvalidStartTime {
		t.Errorf("Expected ErrInvalidStartTime, got: %v", err)
	}
}

// TestShiftUpdateInputValidation tests the update input validation
func TestShiftUpdateInputValidation(t *testing.T) {
	// Valid empty update
	emptyUpdate := &models.UpdateShiftInput{}
	if err := emptyUpdate.Validate(); err != nil {
		t.Errorf("Empty update should not return error, got: %v", err)
	}

	// Valid partial update
	dayOfWeek := models.Tuesday
	startTime := models.ShiftTime("08:00")
	validUpdate := &models.UpdateShiftInput{
		DayOfWeek: &dayOfWeek,
		StartTime: &startTime,
	}
	if err := validUpdate.Validate(); err != nil {
		t.Errorf("Valid update should not return error, got: %v", err)
	}

	// Invalid day of week
	invalidDay := models.DayOfWeek(10)
	invalidDayUpdate := &models.UpdateShiftInput{
		DayOfWeek: &invalidDay,
	}
	if err := invalidDayUpdate.Validate(); err != models.ErrInvalidDayOfWeek {
		t.Errorf("Expected ErrInvalidDayOfWeek, got: %v", err)
	}
}

// TestGetActiveShiftsQueryLogic tests the logic for determining active shifts
// This is a unit test for the ContainsTime method which is used by the repository
func TestGetActiveShiftsQueryLogic(t *testing.T) {
	tests := []struct {
		name      string
		shift     *models.Shift
		checkTime time.Time
		expected  bool
	}{
		{
			name: "Day shift 07:00-19:00 at 10:00 - should be active",
			shift: &models.Shift{
				DayOfWeek: models.Monday,
				StartTime: "07:00",
				EndTime:   "19:00",
			},
			checkTime: time.Date(2026, 1, 19, 10, 0, 0, 0, time.UTC), // Monday 10:00
			expected:  true,
		},
		{
			name: "Day shift 07:00-19:00 at 20:00 - should not be active",
			shift: &models.Shift{
				DayOfWeek: models.Monday,
				StartTime: "07:00",
				EndTime:   "19:00",
			},
			checkTime: time.Date(2026, 1, 19, 20, 0, 0, 0, time.UTC), // Monday 20:00
			expected:  false,
		},
		{
			name: "Night shift 19:00-07:00 at 02:00 next day - should be active",
			shift: &models.Shift{
				DayOfWeek: models.Monday,
				StartTime: "19:00",
				EndTime:   "07:00",
			},
			checkTime: time.Date(2026, 1, 20, 2, 0, 0, 0, time.UTC), // Tuesday 02:00
			expected:  true,
		},
		{
			name: "Night shift 19:00-07:00 at 22:00 same day - should be active",
			shift: &models.Shift{
				DayOfWeek: models.Monday,
				StartTime: "19:00",
				EndTime:   "07:00",
			},
			checkTime: time.Date(2026, 1, 19, 22, 0, 0, 0, time.UTC), // Monday 22:00
			expected:  true,
		},
		{
			name: "Night shift 19:00-07:00 at 12:00 - should not be active",
			shift: &models.Shift{
				DayOfWeek: models.Monday,
				StartTime: "19:00",
				EndTime:   "07:00",
			},
			checkTime: time.Date(2026, 1, 19, 12, 0, 0, 0, time.UTC), // Monday 12:00
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.shift.ContainsTime(tt.checkTime); got != tt.expected {
				t.Errorf("ContainsTime() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

// TestCoverageGapDetection tests the coverage gap detection logic
func TestCoverageGapDetection(t *testing.T) {
	// Test case: Day with no shifts - entire day should be a gap
	t.Run("Day with no shifts", func(t *testing.T) {
		gaps := findGapsForDay([]models.Shift{})
		if len(gaps) != 1 {
			t.Errorf("Expected 1 gap for empty day, got %d", len(gaps))
		}
		if gaps[0].StartTime != "00:00" || gaps[0].EndTime != "23:59" {
			t.Errorf("Expected full day gap, got %s-%s", gaps[0].StartTime, gaps[0].EndTime)
		}
	})

	// Test case: Day with full coverage (day + night shifts)
	t.Run("Day with full coverage", func(t *testing.T) {
		shifts := []models.Shift{
			{StartTime: "07:00", EndTime: "19:00"},
			{StartTime: "19:00", EndTime: "07:00"},
		}
		gaps := findGapsForDay(shifts)
		if len(gaps) != 0 {
			t.Errorf("Expected no gaps for full coverage, got %d gaps", len(gaps))
		}
	})

	// Test case: Day with partial coverage
	t.Run("Day with gap in the middle", func(t *testing.T) {
		shifts := []models.Shift{
			{StartTime: "06:00", EndTime: "10:00"},
			{StartTime: "14:00", EndTime: "20:00"},
		}
		gaps := findGapsForDay(shifts)
		// Should have gaps: 00:00-06:00, 10:00-14:00, 20:00-23:59
		if len(gaps) != 3 {
			t.Errorf("Expected 3 gaps, got %d", len(gaps))
		}
	})
}

// TestFilterShiftsByDay tests the filtering logic for shifts by day
func TestFilterShiftsByDay(t *testing.T) {
	shifts := []models.Shift{
		{
			DayOfWeek: models.Monday,
			StartTime: "07:00",
			EndTime:   "19:00",
		},
		{
			DayOfWeek: models.Monday,
			StartTime: "19:00",
			EndTime:   "07:00", // Night shift
		},
		{
			DayOfWeek: models.Tuesday,
			StartTime: "07:00",
			EndTime:   "19:00",
		},
	}

	// Monday shifts should include both Monday day shift and Monday night shift
	mondayShifts := filterShiftsByDay(shifts, models.Monday)
	if len(mondayShifts) != 2 {
		t.Errorf("Expected 2 Monday shifts, got %d", len(mondayShifts))
	}

	// Tuesday shifts should include Tuesday day shift AND Monday night shift (extends to Tuesday morning)
	tuesdayShifts := filterShiftsByDay(shifts, models.Tuesday)
	if len(tuesdayShifts) != 2 {
		t.Errorf("Expected 2 Tuesday shifts (including Monday night), got %d", len(tuesdayShifts))
	}
}

// TestFormatHour tests the hour formatting utility
func TestFormatHour(t *testing.T) {
	tests := []struct {
		hour     int
		expected string
	}{
		{0, "00:00"},
		{7, "07:00"},
		{12, "12:00"},
		{19, "19:00"},
		{23, "23:00"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := formatHour(tt.hour); got != tt.expected {
				t.Errorf("formatHour(%d) = %s, expected %s", tt.hour, got, tt.expected)
			}
		})
	}
}
