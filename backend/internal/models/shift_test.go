package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestShiftIsNightShift tests the night shift detection logic
func TestShiftIsNightShift(t *testing.T) {
	tests := []struct {
		name      string
		startTime ShiftTime
		endTime   ShiftTime
		expected  bool
	}{
		{
			name:      "Day shift 07:00-19:00",
			startTime: "07:00",
			endTime:   "19:00",
			expected:  false,
		},
		{
			name:      "Night shift 19:00-07:00",
			startTime: "19:00",
			endTime:   "07:00",
			expected:  true,
		},
		{
			name:      "Morning shift 06:00-14:00",
			startTime: "06:00",
			endTime:   "14:00",
			expected:  false,
		},
		{
			name:      "Overnight shift 22:00-06:00",
			startTime: "22:00",
			endTime:   "06:00",
			expected:  true,
		},
		{
			name:      "Midnight crossing 23:00-01:00",
			startTime: "23:00",
			endTime:   "01:00",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shift := &Shift{
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}
			if got := shift.IsNightShift(); got != tt.expected {
				t.Errorf("IsNightShift() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

// TestShiftContainsTime tests the time containment logic for both day and night shifts
func TestShiftContainsTime(t *testing.T) {
	tests := []struct {
		name      string
		startTime ShiftTime
		endTime   ShiftTime
		checkTime time.Time
		expected  bool
	}{
		{
			name:      "Day shift contains 12:00",
			startTime: "07:00",
			endTime:   "19:00",
			checkTime: time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name:      "Day shift does not contain 20:00",
			startTime: "07:00",
			endTime:   "19:00",
			checkTime: time.Date(2026, 1, 16, 20, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name:      "Day shift contains start time 07:00",
			startTime: "07:00",
			endTime:   "19:00",
			checkTime: time.Date(2026, 1, 16, 7, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name:      "Day shift does not contain end time 19:00 (exclusive)",
			startTime: "07:00",
			endTime:   "19:00",
			checkTime: time.Date(2026, 1, 16, 19, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name:      "Night shift contains 02:00",
			startTime: "19:00",
			endTime:   "07:00",
			checkTime: time.Date(2026, 1, 16, 2, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name:      "Night shift contains 22:00",
			startTime: "19:00",
			endTime:   "07:00",
			checkTime: time.Date(2026, 1, 16, 22, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name:      "Night shift does not contain 12:00",
			startTime: "19:00",
			endTime:   "07:00",
			checkTime: time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shift := &Shift{
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}
			if got := shift.ContainsTime(tt.checkTime); got != tt.expected {
				t.Errorf("ContainsTime(%v) = %v, expected %v", tt.checkTime, got, tt.expected)
			}
		})
	}
}

// TestDayOfWeekValidation tests the day of week validation
func TestDayOfWeekValidation(t *testing.T) {
	tests := []struct {
		day      DayOfWeek
		expected bool
	}{
		{Sunday, true},
		{Monday, true},
		{Tuesday, true},
		{Wednesday, true},
		{Thursday, true},
		{Friday, true},
		{Saturday, true},
		{DayOfWeek(-1), false},
		{DayOfWeek(7), false},
		{DayOfWeek(10), false},
	}

	for _, tt := range tests {
		t.Run(tt.day.String(), func(t *testing.T) {
			if got := tt.day.IsValid(); got != tt.expected {
				t.Errorf("DayOfWeek(%d).IsValid() = %v, expected %v", tt.day, got, tt.expected)
			}
		})
	}
}

// TestShiftTimeValidation tests the time format validation
func TestShiftTimeValidation(t *testing.T) {
	tests := []struct {
		time     ShiftTime
		expected bool
	}{
		{"07:00", true},
		{"19:00", true},
		{"00:00", true},
		{"23:59", true},
		{"12:30", true},
		{"7:00", false},    // Missing leading zero
		{"25:00", false},   // Invalid hour
		{"12:60", false},   // Invalid minute
		{"invalid", false}, // Not a time
		{"", false},        // Empty
	}

	for _, tt := range tests {
		t.Run(string(tt.time), func(t *testing.T) {
			if got := tt.time.IsValid(); got != tt.expected {
				t.Errorf("ShiftTime(%s).IsValid() = %v, expected %v", tt.time, got, tt.expected)
			}
		})
	}
}

// TestCreateShiftInputValidation tests the shift input validation
func TestCreateShiftInputValidation(t *testing.T) {
	validInput := CreateShiftInput{
		HospitalID: uuid.New(),
		UserID:     uuid.New(),
		DayOfWeek:  Monday,
		StartTime:  "07:00",
		EndTime:    "19:00",
	}

	if err := validInput.Validate(); err != nil {
		t.Errorf("Valid input should not return error, got: %v", err)
	}

	// Test invalid day of week
	invalidDay := validInput
	invalidDay.DayOfWeek = DayOfWeek(7)
	if err := invalidDay.Validate(); err != ErrInvalidDayOfWeek {
		t.Errorf("Expected ErrInvalidDayOfWeek, got: %v", err)
	}

	// Test invalid start time
	invalidStart := validInput
	invalidStart.StartTime = "invalid"
	if err := invalidStart.Validate(); err != ErrInvalidStartTime {
		t.Errorf("Expected ErrInvalidStartTime, got: %v", err)
	}

	// Test invalid end time
	invalidEnd := validInput
	invalidEnd.EndTime = "25:00"
	if err := invalidEnd.Validate(); err != ErrInvalidEndTime {
		t.Errorf("Expected ErrInvalidEndTime, got: %v", err)
	}
}

// TestShiftToResponse tests the response conversion
func TestShiftToResponse(t *testing.T) {
	hospitalID := uuid.New()
	userID := uuid.New()
	shiftID := uuid.New()
	now := time.Now()

	shift := &Shift{
		ID:         shiftID,
		HospitalID: hospitalID,
		UserID:     userID,
		DayOfWeek:  Monday,
		StartTime:  "19:00",
		EndTime:    "07:00",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	resp := shift.ToResponse()

	if resp.ID != shiftID {
		t.Errorf("ID mismatch: got %v, expected %v", resp.ID, shiftID)
	}
	if resp.DayName != "Segunda-feira" {
		t.Errorf("DayName mismatch: got %v, expected Segunda-feira", resp.DayName)
	}
	if !resp.IsNight {
		t.Error("Expected IsNight to be true for 19:00-07:00 shift")
	}
}

// TestUserCanManageShifts tests the user permission methods for shifts
func TestUserCanManageShifts(t *testing.T) {
	hospitalID := uuid.New()
	otherHospitalID := uuid.New()

	// Create hospital for gestor
	hospital := Hospital{ID: hospitalID}

	tests := []struct {
		name       string
		user       *User
		hospitalID uuid.UUID
		canManage  bool
		canView    bool
	}{
		{
			name: "Admin can manage any hospital",
			user: &User{
				Role: RoleAdmin,
			},
			hospitalID: hospitalID,
			canManage:  true,
			canView:    true,
		},
		{
			name: "Gestor can manage own hospital",
			user: &User{
				Role:      RoleGestor,
				Hospitals: []Hospital{hospital},
			},
			hospitalID: hospitalID,
			canManage:  true,
			canView:    true,
		},
		{
			name: "Gestor cannot manage other hospital",
			user: &User{
				Role:      RoleGestor,
				Hospitals: []Hospital{hospital},
			},
			hospitalID: otherHospitalID,
			canManage:  false,
			canView:    true,
		},
		{
			name: "Operador cannot manage but can view",
			user: &User{
				Role:      RoleOperador,
				Hospitals: []Hospital{hospital},
			},
			hospitalID: hospitalID,
			canManage:  false,
			canView:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.CanManageShiftsForHospital(tt.hospitalID); got != tt.canManage {
				t.Errorf("CanManageShiftsForHospital() = %v, expected %v", got, tt.canManage)
			}
			if got := tt.user.CanViewShifts(); got != tt.canView {
				t.Errorf("CanViewShifts() = %v, expected %v", got, tt.canView)
			}
		})
	}
}
