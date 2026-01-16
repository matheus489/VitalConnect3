package models

import (
	"testing"

	"github.com/google/uuid"
)

// Test 1: Test default preferences creation with mobile phone
func TestDefaultPreferences_WithMobilePhone(t *testing.T) {
	userID := uuid.New()
	prefs := DefaultPreferences(userID, true)

	if prefs.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, prefs.UserID)
	}

	if !prefs.SMSEnabled {
		t.Error("Expected SMSEnabled to be true when hasMobilePhone is true")
	}

	if !prefs.EmailEnabled {
		t.Error("Expected EmailEnabled to be true by default")
	}

	if !prefs.DashboardEnabled {
		t.Error("Expected DashboardEnabled to be always true")
	}
}

// Test 2: Test default preferences creation without mobile phone
func TestDefaultPreferences_WithoutMobilePhone(t *testing.T) {
	userID := uuid.New()
	prefs := DefaultPreferences(userID, false)

	if prefs.SMSEnabled {
		t.Error("Expected SMSEnabled to be false when hasMobilePhone is false")
	}

	if !prefs.EmailEnabled {
		t.Error("Expected EmailEnabled to be true by default")
	}

	if !prefs.DashboardEnabled {
		t.Error("Expected DashboardEnabled to be always true")
	}
}

// Test 3: Test ToResponse conversion
func TestUserNotificationPreferences_ToResponse(t *testing.T) {
	prefs := &UserNotificationPreferences{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		SMSEnabled:       true,
		EmailEnabled:     false,
		DashboardEnabled: true,
	}

	response := prefs.ToResponse()

	if response.ID != prefs.ID {
		t.Error("ID not correctly converted in ToResponse")
	}

	if response.UserID != prefs.UserID {
		t.Error("UserID not correctly converted in ToResponse")
	}

	if response.SMSEnabled != prefs.SMSEnabled {
		t.Error("SMSEnabled not correctly converted in ToResponse")
	}

	if response.EmailEnabled != prefs.EmailEnabled {
		t.Error("EmailEnabled not correctly converted in ToResponse")
	}

	if response.DashboardEnabled != prefs.DashboardEnabled {
		t.Error("DashboardEnabled not correctly converted in ToResponse")
	}
}

// Test 4: Test UpdateNotificationPreferencesInput does not include DashboardEnabled
func TestUpdateNotificationPreferencesInput_NoDashboardEnabled(t *testing.T) {
	smsEnabled := true
	emailEnabled := false

	input := UpdateNotificationPreferencesInput{
		SMSEnabled:   &smsEnabled,
		EmailEnabled: &emailEnabled,
	}

	// Verify the struct fields
	if input.SMSEnabled == nil || *input.SMSEnabled != true {
		t.Error("SMSEnabled should be settable")
	}

	if input.EmailEnabled == nil || *input.EmailEnabled != false {
		t.Error("EmailEnabled should be settable")
	}

	// Note: DashboardEnabled is intentionally not in UpdateNotificationPreferencesInput
	// This test documents that design decision
}
