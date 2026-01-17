package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// DefaultAccessTokenDuration is the default expiration for access tokens (15 minutes)
	DefaultAccessTokenDuration = 15 * time.Minute

	// DefaultRefreshTokenDuration is the default expiration for refresh tokens (7 days)
	DefaultRefreshTokenDuration = 7 * 24 * time.Hour
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.New("token has expired")

	// ErrInvalidClaims is returned when token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")

	// ErrMissingSecret is returned when JWT secret is not configured
	ErrMissingSecret = errors.New("JWT secret is not configured")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims for authentication
type Claims struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	HospitalID   string `json:"hospital_id,omitempty"`
	TenantID     string `json:"tenant_id,omitempty"`
	IsSuperAdmin bool   `json:"is_super_admin,omitempty"`
	TokenType    string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	accessSecret    []byte
	refreshSecret   []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
	issuer          string
}

// NewJWTService creates a new JWT service instance
func NewJWTService(accessSecret, refreshSecret string, accessDuration, refreshDuration time.Duration) (*JWTService, error) {
	if accessSecret == "" {
		return nil, ErrMissingSecret
	}
	if refreshSecret == "" {
		return nil, ErrMissingSecret
	}

	if accessDuration == 0 {
		accessDuration = DefaultAccessTokenDuration
	}
	if refreshDuration == 0 {
		refreshDuration = DefaultRefreshTokenDuration
	}

	return &JWTService{
		accessSecret:    []byte(accessSecret),
		refreshSecret:   []byte(refreshSecret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		issuer:          "vitalconnect",
	}, nil
}

// GenerateTokenPair generates both access and refresh tokens
// Deprecated: Use GenerateTokenPairWithTenant for multi-tenant support
func (s *JWTService) GenerateTokenPair(userID, email, role, hospitalID string) (accessToken, refreshToken string, err error) {
	// For backward compatibility, call the new method with empty tenant
	return s.GenerateTokenPairWithTenant(userID, email, role, hospitalID, "", false)
}

// GenerateTokenPairWithTenant generates both access and refresh tokens with tenant context
func (s *JWTService) GenerateTokenPairWithTenant(userID, email, role, hospitalID, tenantID string, isSuperAdmin bool) (accessToken, refreshToken string, err error) {
	accessToken, err = s.GenerateAccessTokenWithTenant(userID, email, role, hospitalID, tenantID, isSuperAdmin)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.GenerateRefreshTokenWithTenant(userID, email, role, hospitalID, tenantID, isSuperAdmin)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken generates a new access token (15 minutes expiration)
// Deprecated: Use GenerateAccessTokenWithTenant for multi-tenant support
func (s *JWTService) GenerateAccessToken(userID, email, role, hospitalID string) (string, error) {
	return s.GenerateAccessTokenWithTenant(userID, email, role, hospitalID, "", false)
}

// GenerateAccessTokenWithTenant generates a new access token with tenant context (15 minutes expiration)
func (s *JWTService) GenerateAccessTokenWithTenant(userID, email, role, hospitalID, tenantID string, isSuperAdmin bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		Email:        email,
		Role:         role,
		HospitalID:   hospitalID,
		TenantID:     tenantID,
		IsSuperAdmin: isSuperAdmin,
		TokenType:    string(AccessToken),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

// GenerateRefreshToken generates a new refresh token (7 days expiration)
// Deprecated: Use GenerateRefreshTokenWithTenant for multi-tenant support
func (s *JWTService) GenerateRefreshToken(userID, email, role, hospitalID string) (string, error) {
	return s.GenerateRefreshTokenWithTenant(userID, email, role, hospitalID, "", false)
}

// GenerateRefreshTokenWithTenant generates a new refresh token with tenant context (7 days expiration)
func (s *JWTService) GenerateRefreshTokenWithTenant(userID, email, role, hospitalID, tenantID string, isSuperAdmin bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		Email:        email,
		Role:         role,
		HospitalID:   hospitalID,
		TenantID:     tenantID,
		IsSuperAdmin: isSuperAdmin,
		TokenType:    string(RefreshToken),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecret)
}

// ValidateAccessToken validates an access token and returns the claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.accessSecret, AccessToken)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.refreshSecret, RefreshToken)
}

// validateToken validates a token with the given secret and expected type
func (s *JWTService) validateToken(tokenString string, secret []byte, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Verify token type
	if claims.TokenType != string(expectedType) {
		return nil, ErrInvalidToken
	}

	// Verify issuer
	if claims.Issuer != s.issuer {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetAccessTokenDuration returns the access token duration
func (s *JWTService) GetAccessTokenDuration() time.Duration {
	return s.accessDuration
}

// GetRefreshTokenDuration returns the refresh token duration
func (s *JWTService) GetRefreshTokenDuration() time.Duration {
	return s.refreshDuration
}
