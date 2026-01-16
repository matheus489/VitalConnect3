package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserInactive is returned when user account is inactive
	ErrUserInactive = errors.New("user account is inactive")

	// ErrInvalidCredentials is returned for invalid email/password
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrTokenRevoked is returned when token has been revoked
	ErrTokenRevoked = errors.New("token has been revoked")
)

// User represents the user data needed for authentication
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Nome         string
	Role         string
	HospitalID   *uuid.UUID
	Ativo        bool
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// AuthService handles authentication operations
type AuthService struct {
	jwtService     *JWTService
	userRepo       UserRepository
	redisClient    *redis.Client
	revokedKeyPrefix string
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtService *JWTService, userRepo UserRepository, redisClient *redis.Client) *AuthService {
	return &AuthService{
		jwtService:       jwtService,
		userRepo:         userRepo,
		redisClient:      redisClient,
		revokedKeyPrefix: "revoked_token",
	}
}

// LoginResult contains the result of a successful login
type LoginResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

// UserInfo contains user information returned after login
type UserInfo struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Nome       string     `json:"nome"`
	Role       string     `json:"role"`
	HospitalID *uuid.UUID `json:"hospital_id,omitempty"`
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	// Fetch user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.Ativo {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := CheckPasswordHash(password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	hospitalID := ""
	if user.HospitalID != nil {
		hospitalID = user.HospitalID.String()
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Role,
		hospitalID,
	)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtService.GetAccessTokenDuration().Seconds()),
		User: &UserInfo{
			ID:         user.ID,
			Email:      user.Email,
			Nome:       user.Nome,
			Role:       user.Role,
			HospitalID: user.HospitalID,
		},
	}, nil
}

// RefreshResult contains the result of a successful token refresh
type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Refresh generates new tokens using a valid refresh token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if token is revoked
	isRevoked, err := s.isTokenRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if isRevoked {
		return nil, ErrTokenRevoked
	}

	// Verify user still exists and is active
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.Ativo {
		return nil, ErrUserInactive
	}

	// Revoke old refresh token
	if err := s.revokeToken(ctx, claims.ID, s.jwtService.GetRefreshTokenDuration()); err != nil {
		// Log error but continue - token rotation is best effort
		_ = err
	}

	// Generate new tokens
	hospitalID := ""
	if user.HospitalID != nil {
		hospitalID = user.HospitalID.String()
	}

	newAccessToken, newRefreshToken, err := s.jwtService.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Role,
		hospitalID,
	)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtService.GetAccessTokenDuration().Seconds()),
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// Validate refresh token to get its ID
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		// Even if token is invalid/expired, return success
		// This prevents information leakage
		return nil
	}

	// Revoke the token
	return s.revokeToken(ctx, claims.ID, s.jwtService.GetRefreshTokenDuration())
}

// revokeToken adds a token ID to the revoked tokens set
func (s *AuthService) revokeToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if s.redisClient == nil {
		return nil
	}

	key := fmt.Sprintf("%s:%s", s.revokedKeyPrefix, tokenID)
	return s.redisClient.Set(ctx, key, "revoked", ttl).Err()
}

// isTokenRevoked checks if a token has been revoked
func (s *AuthService) isTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	if s.redisClient == nil {
		return false, nil
	}

	key := fmt.Sprintf("%s:%s", s.revokedKeyPrefix, tokenID)
	result, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

// GetCurrentUser retrieves the current user from their ID
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*UserInfo, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:         user.ID,
		Email:      user.Email,
		Nome:       user.Nome,
		Role:       user.Role,
		HospitalID: user.HospitalID,
	}, nil
}
