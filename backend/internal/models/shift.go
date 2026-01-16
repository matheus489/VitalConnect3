package models

import (
	"time"

	"github.com/google/uuid"
)

// DayOfWeek represents a day of the week (0 = Sunday, 6 = Saturday)
type DayOfWeek int

const (
	Sunday    DayOfWeek = 0
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
)

// DayNames maps DayOfWeek to Portuguese names
var DayNames = map[DayOfWeek]string{
	Sunday:    "Domingo",
	Monday:    "Segunda-feira",
	Tuesday:   "Terca-feira",
	Wednesday: "Quarta-feira",
	Thursday:  "Quinta-feira",
	Friday:    "Sexta-feira",
	Saturday:  "Sabado",
}

// String returns the Portuguese name of the day
func (d DayOfWeek) String() string {
	if name, ok := DayNames[d]; ok {
		return name
	}
	return "Invalido"
}

// IsValid checks if the day is a valid day of the week (0-6)
func (d DayOfWeek) IsValid() bool {
	return d >= Sunday && d <= Saturday
}

// ShiftTime represents a time of day for shift scheduling (HH:MM format)
type ShiftTime string

// IsValid validates the time format (HH:MM)
func (st ShiftTime) IsValid() bool {
	if len(st) != 5 {
		return false
	}
	_, err := time.Parse("15:04", string(st))
	return err == nil
}

// ToTime converts ShiftTime to a time.Time on a reference date
func (st ShiftTime) ToTime(referenceDate time.Time) (time.Time, error) {
	parsed, err := time.Parse("15:04", string(st))
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(
		referenceDate.Year(),
		referenceDate.Month(),
		referenceDate.Day(),
		parsed.Hour(),
		parsed.Minute(),
		0, 0,
		referenceDate.Location(),
	), nil
}

// Hour returns the hour component
func (st ShiftTime) Hour() int {
	t, err := time.Parse("15:04", string(st))
	if err != nil {
		return 0
	}
	return t.Hour()
}

// Minute returns the minute component
func (st ShiftTime) Minute() int {
	t, err := time.Parse("15:04", string(st))
	if err != nil {
		return 0
	}
	return t.Minute()
}

// Shift represents a work schedule entry for an operator
type Shift struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	HospitalID uuid.UUID  `json:"hospital_id" db:"hospital_id" validate:"required"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id" validate:"required"`
	DayOfWeek  DayOfWeek  `json:"day_of_week" db:"day_of_week" validate:"gte=0,lte=6"`
	StartTime  ShiftTime  `json:"start_time" db:"start_time" validate:"required"`
	EndTime    ShiftTime  `json:"end_time" db:"end_time" validate:"required"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`

	// Related data (populated by queries)
	User     *User     `json:"user,omitempty" db:"-"`
	Hospital *Hospital `json:"hospital,omitempty" db:"-"`
}

// IsNightShift returns true if the shift crosses midnight (e.g., 19:00-07:00)
func (s *Shift) IsNightShift() bool {
	if !s.StartTime.IsValid() || !s.EndTime.IsValid() {
		return false
	}
	startHour := s.StartTime.Hour()
	startMin := s.StartTime.Minute()
	endHour := s.EndTime.Hour()
	endMin := s.EndTime.Minute()

	// If start time is greater than end time, it crosses midnight
	// e.g., 19:00 > 07:00 means night shift
	return startHour > endHour || (startHour == endHour && startMin > endMin)
}

// ContainsTime checks if a given time falls within this shift
// Handles night shifts that cross midnight correctly
func (s *Shift) ContainsTime(checkTime time.Time) bool {
	checkHour := checkTime.Hour()
	checkMin := checkTime.Minute()
	checkMinutes := checkHour*60 + checkMin

	startHour := s.StartTime.Hour()
	startMin := s.StartTime.Minute()
	startMinutes := startHour*60 + startMin

	endHour := s.EndTime.Hour()
	endMin := s.EndTime.Minute()
	endMinutes := endHour*60 + endMin

	if s.IsNightShift() {
		// Night shift: 19:00-07:00
		// Time is within if it's >= 19:00 OR < 07:00
		return checkMinutes >= startMinutes || checkMinutes < endMinutes
	}

	// Day shift: 07:00-19:00
	// Time is within if it's >= 07:00 AND < 19:00
	return checkMinutes >= startMinutes && checkMinutes < endMinutes
}

// CreateShiftInput represents input for creating a shift
type CreateShiftInput struct {
	HospitalID uuid.UUID `json:"hospital_id" validate:"required"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	DayOfWeek  DayOfWeek `json:"day_of_week" validate:"gte=0,lte=6"`
	StartTime  ShiftTime `json:"start_time" validate:"required"`
	EndTime    ShiftTime `json:"end_time" validate:"required"`
}

// Validate validates the CreateShiftInput
func (i *CreateShiftInput) Validate() error {
	if !i.DayOfWeek.IsValid() {
		return ErrInvalidDayOfWeek
	}
	if !i.StartTime.IsValid() {
		return ErrInvalidStartTime
	}
	if !i.EndTime.IsValid() {
		return ErrInvalidEndTime
	}
	return nil
}

// UpdateShiftInput represents input for updating a shift
type UpdateShiftInput struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	DayOfWeek *DayOfWeek `json:"day_of_week,omitempty" validate:"omitempty,gte=0,lte=6"`
	StartTime *ShiftTime `json:"start_time,omitempty"`
	EndTime   *ShiftTime `json:"end_time,omitempty"`
}

// Validate validates the UpdateShiftInput
func (i *UpdateShiftInput) Validate() error {
	if i.DayOfWeek != nil && !i.DayOfWeek.IsValid() {
		return ErrInvalidDayOfWeek
	}
	if i.StartTime != nil && !i.StartTime.IsValid() {
		return ErrInvalidStartTime
	}
	if i.EndTime != nil && !i.EndTime.IsValid() {
		return ErrInvalidEndTime
	}
	return nil
}

// ShiftResponse represents the API response for a shift
type ShiftResponse struct {
	ID         uuid.UUID     `json:"id"`
	HospitalID uuid.UUID     `json:"hospital_id"`
	UserID     uuid.UUID     `json:"user_id"`
	DayOfWeek  DayOfWeek     `json:"day_of_week"`
	DayName    string        `json:"day_name"`
	StartTime  ShiftTime     `json:"start_time"`
	EndTime    ShiftTime     `json:"end_time"`
	IsNight    bool          `json:"is_night"`
	User       *UserResponse `json:"user,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// ToResponse converts Shift to ShiftResponse
func (s *Shift) ToResponse() ShiftResponse {
	resp := ShiftResponse{
		ID:         s.ID,
		HospitalID: s.HospitalID,
		UserID:     s.UserID,
		DayOfWeek:  s.DayOfWeek,
		DayName:    s.DayOfWeek.String(),
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		IsNight:    s.IsNightShift(),
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}

	if s.User != nil {
		userResp := s.User.ToResponse()
		resp.User = &userResp
	}

	return resp
}

// CoverageGap represents a gap in shift coverage for a day
type CoverageGap struct {
	DayOfWeek DayOfWeek `json:"day_of_week"`
	DayName   string    `json:"day_name"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
}

// CoverageAnalysis represents the coverage analysis for a hospital
type CoverageAnalysis struct {
	HospitalID  uuid.UUID     `json:"hospital_id"`
	TotalShifts int           `json:"total_shifts"`
	Gaps        []CoverageGap `json:"gaps"`
	HasGaps     bool          `json:"has_gaps"`
}

// TodayShift represents a shift scheduled for today with user details
type TodayShift struct {
	Shift
	IsActive bool `json:"is_active"`
}

// ToTodayResponse converts TodayShift to a response
func (ts *TodayShift) ToTodayResponse() map[string]interface{} {
	resp := ts.Shift.ToResponse()
	return map[string]interface{}{
		"id":          resp.ID,
		"hospital_id": resp.HospitalID,
		"user_id":     resp.UserID,
		"day_of_week": resp.DayOfWeek,
		"day_name":    resp.DayName,
		"start_time":  resp.StartTime,
		"end_time":    resp.EndTime,
		"is_night":    resp.IsNight,
		"is_active":   ts.IsActive,
		"user":        resp.User,
		"created_at":  resp.CreatedAt,
		"updated_at":  resp.UpdatedAt,
	}
}
