package notification

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/sidot/backend/internal/models"
)

var (
	ErrTwilioNotConfigured  = errors.New("Twilio not configured")
	ErrInvalidPhoneNumber   = errors.New("invalid phone number")
	ErrSMSSendFailed        = errors.New("failed to send SMS")
	ErrSMSRateLimited       = errors.New("SMS rate limited")
	ErrTwilioCredentials    = errors.New("invalid Twilio credentials")
)

// SMSConfig holds the configuration for SMS sending via Twilio
type SMSConfig struct {
	AccountSID      string
	AuthToken       string
	FromPhoneNumber string
}

// SMSService handles sending SMS messages via Twilio
type SMSService struct {
	config *SMSConfig
	client *twilio.RestClient
}

// NewSMSService creates a new SMSService
func NewSMSService(config *SMSConfig) *SMSService {
	if config == nil {
		config = &SMSConfig{}
	}

	svc := &SMSService{
		config: config,
	}

	// Initialize Twilio client if configured
	if svc.IsConfigured() {
		svc.client = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: config.AccountSID,
			Password: config.AuthToken,
		})
	}

	return svc
}

// NewSMSServiceFromEnv creates a new SMSService from environment variables
func NewSMSServiceFromEnv() *SMSService {
	config := &SMSConfig{
		AccountSID:      os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:       os.Getenv("TWILIO_AUTH_TOKEN"),
		FromPhoneNumber: os.Getenv("TWILIO_PHONE_NUMBER"),
	}
	return NewSMSService(config)
}

// IsConfigured returns true if Twilio is properly configured
func (s *SMSService) IsConfigured() bool {
	return s.config != nil &&
		s.config.AccountSID != "" &&
		s.config.AuthToken != "" &&
		s.config.FromPhoneNumber != ""
}

// SendSMS sends an SMS message to the specified phone number
func (s *SMSService) SendSMS(ctx context.Context, to, message string) error {
	if !s.IsConfigured() {
		return ErrTwilioNotConfigured
	}

	if to == "" || !models.ValidateMobilePhone(to) {
		return ErrInvalidPhoneNumber
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(s.config.FromPhoneNumber)
	params.SetBody(message)

	_, err := s.client.Api.CreateMessage(params)
	if err != nil {
		// Check for specific Twilio errors
		errStr := err.Error()
		if strings.Contains(errStr, "21610") || strings.Contains(errStr, "21614") {
			return fmt.Errorf("%w: %v", ErrInvalidPhoneNumber, err)
		}
		if strings.Contains(errStr, "20003") || strings.Contains(errStr, "20001") {
			return fmt.Errorf("%w: %v", ErrTwilioCredentials, err)
		}
		if strings.Contains(errStr, "14107") || strings.Contains(errStr, "rate") {
			return fmt.Errorf("%w: %v", ErrSMSRateLimited, err)
		}
		return fmt.Errorf("%w: %v", ErrSMSSendFailed, err)
	}

	return nil
}

// SMSNotificationData contains data for building SMS messages
type SMSNotificationData struct {
	HospitalNome  string
	Idade         int
	HorasRestante int
	OccurrenceID  uuid.UUID
	BaseURL       string
}

// BuildSMSMessage builds the SMS message from occurrence data
// Template: [SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: {hospital_name} Idade: {age} Janela: {hours_left}h restantes. Acao: {short_link}
// Max 160 characters to avoid fragmentation
func BuildSMSMessage(data *SMSNotificationData) string {
	// Build the short link
	shortLink := fmt.Sprintf("/ocorrencias/%s", data.OccurrenceID.String())
	if data.BaseURL != "" {
		shortLink = data.BaseURL + shortLink
	}

	// Build the message
	hospitalName := data.HospitalNome

	// Calculate available space for hospital name
	// Fixed parts: "[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: " (59 chars)
	// "Idade: XX Janela: Xh restantes. Acao: " (38 chars max)
	// URL part varies but typically around 50 chars
	// Total fixed: ~147 chars, leaving ~13 chars for hospital name with some buffer

	baseMessage := fmt.Sprintf("[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: %s Idade: %d Janela: %dh restantes. Acao: %s",
		hospitalName, data.Idade, data.HorasRestante, shortLink)

	// If message is too long, truncate hospital name
	if len(baseMessage) > 160 {
		// Calculate how much we need to trim
		excess := len(baseMessage) - 160 + 3 // +3 for "..."
		if excess < len(hospitalName) {
			hospitalName = hospitalName[:len(hospitalName)-excess] + "..."
		} else {
			hospitalName = hospitalName[:10] + "..." // Minimum truncation
		}
		baseMessage = fmt.Sprintf("[SIDOT] ALERTA CRITICO: Obito PCR detectado. Hosp: %s Idade: %d Janela: %dh restantes. Acao: %s",
			hospitalName, data.Idade, data.HorasRestante, shortLink)
	}

	// Final truncation if still too long
	if len(baseMessage) > 160 {
		baseMessage = baseMessage[:157] + "..."
	}

	return baseMessage
}

// MaskPhoneForLog masks a phone number for logging (+55119****9999)
func MaskPhoneForLog(phone string) string {
	return models.MaskMobilePhone(phone)
}
