package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Environment string
	ServerPort  string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret          string
	JWTRefreshSecret   string
	JWTAccessDuration  time.Duration
	JWTRefreshDuration time.Duration

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// Twilio (SMS)
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioPhoneNumber string

	// CORS
	CORSOrigins []string

	// Rate Limiting
	LoginRateLimit int // attempts per minute

	// Listener
	ListenerPollInterval time.Duration

	// Health Check
	AdminAlertEmail     string
	HealthCheckInterval time.Duration
	AlertCooldownMinutes int

	// Dashboard URL (for notification links)
	DashboardURL string

	// Firebase Cloud Messaging (Push Notifications)
	FCMServerKey string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Server defaults
		Environment: getEnv("ENVIRONMENT", "development"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/vitalconnect?sslmode=disable"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", ""),
		JWTAccessDuration:  getDurationEnv("JWT_ACCESS_DURATION", 15*time.Minute),
		JWTRefreshDuration: getDurationEnv("JWT_REFRESH_DURATION", 7*24*time.Hour),

		// SMTP
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getIntEnv("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@vitalconnect.gov.br"),

		// Twilio (SMS)
		TwilioAccountSID:  getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:   getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber: getEnv("TWILIO_PHONE_NUMBER", ""),

		// CORS
		CORSOrigins: getSliceEnv("CORS_ORIGINS", []string{"http://localhost:3000"}),

		// Rate Limiting
		LoginRateLimit: getIntEnv("LOGIN_RATE_LIMIT", 5),

		// Listener
		ListenerPollInterval: getDurationEnv("LISTENER_POLL_INTERVAL", 3*time.Second),

		// Health Check
		AdminAlertEmail:      getEnv("ADMIN_ALERT_EMAIL", ""),
		HealthCheckInterval:  getDurationEnv("HEALTH_CHECK_INTERVAL", 10*time.Second),
		AlertCooldownMinutes: getIntEnv("ALERT_COOLDOWN_MINUTES", 5),

		// Dashboard URL
		DashboardURL: getEnv("DASHBOARD_URL", "http://localhost:3000"),

		// FCM (Push Notifications)
		FCMServerKey: getEnv("FCM_SERVER_KEY", ""),
	}

	// Validate required fields in production
	if cfg.Environment == "production" {
		if cfg.JWTSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET is required in production")
		}
		if cfg.JWTRefreshSecret == "" {
			return nil, fmt.Errorf("JWT_REFRESH_SECRET is required in production")
		}
	}

	// Set defaults for development
	if cfg.Environment == "development" {
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = "dev-jwt-secret-change-in-production"
		}
		if cfg.JWTRefreshSecret == "" {
			cfg.JWTRefreshSecret = "dev-jwt-refresh-secret-change-in-production"
		}
	}

	return cfg, nil
}

// IsTwilioConfigured returns true if Twilio SMS is configured
func (c *Config) IsTwilioConfigured() bool {
	return c.TwilioAccountSID != "" && c.TwilioAuthToken != "" && c.TwilioPhoneNumber != ""
}

// IsFCMConfigured returns true if Firebase Cloud Messaging is configured
func (c *Config) IsFCMConfigured() bool {
	return c.FCMServerKey != ""
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv retrieves an integer environment variable or returns a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getDurationEnv retrieves a duration environment variable or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getSliceEnv retrieves a comma-separated environment variable as a slice
func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
