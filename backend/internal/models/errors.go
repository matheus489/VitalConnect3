package models

import "errors"

// Shift-related errors
var (
	ErrInvalidDayOfWeek = errors.New("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	ErrInvalidStartTime = errors.New("start_time must be in HH:MM format")
	ErrInvalidEndTime   = errors.New("end_time must be in HH:MM format")
	ErrShiftNotFound    = errors.New("shift not found")
	ErrShiftExists      = errors.New("shift already exists for this user on this day at this time")
)
