package models

import (
	"testing"
)

// Test 1: Test phone validation in E.164 format
func TestValidateMobilePhone_ValidFormats(t *testing.T) {
	validPhones := []string{
		"+5511999999999",
		"+15551234567",
		"+447911123456",
		"+33612345678",
		"",       // Empty is valid (optional field)
	}

	for _, phone := range validPhones {
		if !ValidateMobilePhone(phone) {
			t.Errorf("Expected phone %s to be valid, but it was invalid", phone)
		}
	}
}

func TestValidateMobilePhone_InvalidFormats(t *testing.T) {
	invalidPhones := []string{
		"5511999999999",       // Missing +
		"+0511999999999",      // Starts with 0
		"11999999999",         // No country code
		"+55119999999",        // Too short
		"+551199999999999999", // Too long
		"phone",              // Text
		"+",                  // Just +
	}

	for _, phone := range invalidPhones {
		if ValidateMobilePhone(phone) {
			t.Errorf("Expected phone %s to be invalid, but it was valid", phone)
		}
	}
}

// Test 2: Test phone masking for logs
func TestMaskMobilePhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+5511999999999", "+5511****9999"},
		{"+15551234567", "+1555***4567"},
		{"", ""},                              // Empty remains empty
		{"+55123", "+55123"},                  // Too short to mask, returns as is
	}

	for _, tc := range tests {
		result := MaskMobilePhone(tc.input)
		if result != tc.expected {
			t.Errorf("MaskMobilePhone(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// Test 3: Test user can receive SMS notifications
func TestUser_CanReceiveSMSNotifications(t *testing.T) {
	phone := "+5511999999999"
	emptyPhone := ""

	tests := []struct {
		name     string
		user     User
		expected bool
	}{
		{
			name:     "Active user with phone can receive SMS",
			user:     User{Ativo: true, MobilePhone: &phone},
			expected: true,
		},
		{
			name:     "Inactive user with phone cannot receive SMS",
			user:     User{Ativo: false, MobilePhone: &phone},
			expected: false,
		},
		{
			name:     "Active user without phone cannot receive SMS",
			user:     User{Ativo: true, MobilePhone: nil},
			expected: false,
		},
		{
			name:     "Active user with empty phone cannot receive SMS",
			user:     User{Ativo: true, MobilePhone: &emptyPhone},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.user.CanReceiveSMSNotifications()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// Test 4: Test CreateUserInput and UpdateUserInput with MobilePhone
func TestUserInputStructsHaveMobilePhone(t *testing.T) {
	phone := "+5511999999999"

	// Test CreateUserInput
	createInput := CreateUserInput{
		Email:       "test@example.com",
		Password:    "password123",
		Nome:        "Test User",
		Role:        RoleOperador,
		MobilePhone: &phone,
	}

	if createInput.MobilePhone == nil || *createInput.MobilePhone != phone {
		t.Error("CreateUserInput.MobilePhone not set correctly")
	}

	// Test UpdateUserInput
	updateInput := UpdateUserInput{
		MobilePhone: &phone,
	}

	if updateInput.MobilePhone == nil || *updateInput.MobilePhone != phone {
		t.Error("UpdateUserInput.MobilePhone not set correctly")
	}
}
