package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateTokenPairWithTenant(t *testing.T) {
	service, err := NewJWTService("test-access-secret-32chars!!", "test-refresh-secret-32chars!", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("should generate tokens with tenant info", func(t *testing.T) {
		userID := uuid.New().String()
		email := "test@example.com"
		role := "admin"
		hospitalID := uuid.New().String()
		tenantID := uuid.New().String()
		isSuperAdmin := false

		accessToken, refreshToken, err := service.GenerateTokenPairWithTenant(
			userID, email, role, hospitalID, tenantID, isSuperAdmin,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		// Validate access token and check tenant claims
		claims, err := service.ValidateAccessToken(accessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, hospitalID, claims.HospitalID)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.False(t, claims.IsSuperAdmin)
	})

	t.Run("should generate tokens for super admin", func(t *testing.T) {
		userID := uuid.New().String()
		email := "super@example.com"
		role := "admin"
		hospitalID := ""
		tenantID := uuid.New().String()
		isSuperAdmin := true

		accessToken, _, err := service.GenerateTokenPairWithTenant(
			userID, email, role, hospitalID, tenantID, isSuperAdmin,
		)

		require.NoError(t, err)

		claims, err := service.ValidateAccessToken(accessToken)
		require.NoError(t, err)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.True(t, claims.IsSuperAdmin)
	})

	t.Run("should work without tenant (backward compatibility)", func(t *testing.T) {
		userID := uuid.New().String()
		email := "legacy@example.com"
		role := "operador"
		hospitalID := uuid.New().String()

		// Use legacy method
		accessToken, _, err := service.GenerateTokenPair(userID, email, role, hospitalID)
		require.NoError(t, err)

		claims, err := service.ValidateAccessToken(accessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Empty(t, claims.TenantID)
		assert.False(t, claims.IsSuperAdmin)
	})

	t.Run("should validate refresh token with tenant claims", func(t *testing.T) {
		userID := uuid.New().String()
		email := "test@example.com"
		role := "gestor"
		hospitalID := uuid.New().String()
		tenantID := uuid.New().String()
		isSuperAdmin := false

		_, refreshToken, err := service.GenerateTokenPairWithTenant(
			userID, email, role, hospitalID, tenantID, isSuperAdmin,
		)
		require.NoError(t, err)

		claims, err := service.ValidateRefreshToken(refreshToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.False(t, claims.IsSuperAdmin)
	})
}

func TestJWTService_TokenValidation(t *testing.T) {
	service, err := NewJWTService("test-access-secret-32chars!!", "test-refresh-secret-32chars!", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("should reject expired access token", func(t *testing.T) {
		// Create service with very short duration
		shortService, _ := NewJWTService("test-access-secret-32chars!!", "test-refresh-secret-32chars!", 1*time.Millisecond, 7*24*time.Hour)

		accessToken, _, err := shortService.GenerateTokenPairWithTenant(
			uuid.New().String(), "test@example.com", "admin", "", uuid.New().String(), false,
		)
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortService.ValidateAccessToken(accessToken)
		assert.Error(t, err)
		assert.Equal(t, ErrExpiredToken, err)
	})

	t.Run("should reject access token validated as refresh", func(t *testing.T) {
		accessToken, _, err := service.GenerateTokenPairWithTenant(
			uuid.New().String(), "test@example.com", "admin", "", uuid.New().String(), false,
		)
		require.NoError(t, err)

		_, err = service.ValidateRefreshToken(accessToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("should reject refresh token validated as access", func(t *testing.T) {
		_, refreshToken, err := service.GenerateTokenPairWithTenant(
			uuid.New().String(), "test@example.com", "admin", "", uuid.New().String(), false,
		)
		require.NoError(t, err)

		_, err = service.ValidateAccessToken(refreshToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("should reject invalid token string", func(t *testing.T) {
		_, err := service.ValidateAccessToken("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})
}
