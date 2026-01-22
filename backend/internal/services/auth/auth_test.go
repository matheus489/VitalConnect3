package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test 1: Testar login com credenciais validas
func TestHashPasswordAndCheck(t *testing.T) {
	password := "securePassword123"

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify the hash is not empty
	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}

	// Verify correct password matches
	err = CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash failed for correct password: %v", err)
	}
}

// Test 2: Testar rejeicao de credenciais invalidas
func TestCheckPasswordHashRejectsInvalid(t *testing.T) {
	password := "correctPassword123"
	wrongPassword := "wrongPassword456"

	// Hash the correct password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify wrong password is rejected
	err = CheckPasswordHash(wrongPassword, hash)
	if err == nil {
		t.Error("CheckPasswordHash should reject wrong password")
	}
	if err != ErrInvalidPassword {
		t.Errorf("Expected ErrInvalidPassword, got: %v", err)
	}
}

// Test 3: Testar geracao e validacao de JWT
func TestJWTGenerationAndValidation(t *testing.T) {
	// Create JWT service
	jwtService, err := NewJWTService(
		"test-access-secret-key-32-chars!",
		"test-refresh-secret-key-32chars!",
		15*time.Minute,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Test data
	userID := uuid.New().String()
	email := "test@sidot.gov.br"
	role := "operador"
	hospitalID := uuid.New().String()

	// Generate access token
	accessToken, err := jwtService.GenerateAccessToken(userID, email, role, hospitalID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if accessToken == "" {
		t.Error("GenerateAccessToken returned empty token")
	}

	// Validate access token
	claims, err := jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	// Verify claims
	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %s, expected %s", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email mismatch: got %s, expected %s", claims.Email, email)
	}
	if claims.Role != role {
		t.Errorf("Role mismatch: got %s, expected %s", claims.Role, role)
	}
	if claims.HospitalID != hospitalID {
		t.Errorf("HospitalID mismatch: got %s, expected %s", claims.HospitalID, hospitalID)
	}
	if claims.TokenType != string(AccessToken) {
		t.Errorf("TokenType mismatch: got %s, expected %s", claims.TokenType, AccessToken)
	}
}

// Test 4: Testar refresh token flow
func TestRefreshTokenFlow(t *testing.T) {
	// Create JWT service
	jwtService, err := NewJWTService(
		"test-access-secret-key-32-chars!",
		"test-refresh-secret-key-32chars!",
		15*time.Minute,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Test data
	userID := uuid.New().String()
	email := "gestor@sidot.gov.br"
	role := "gestor"
	hospitalID := ""

	// Generate token pair
	accessToken, refreshToken, err := jwtService.GenerateTokenPair(userID, email, role, hospitalID)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Error("GenerateTokenPair returned empty tokens")
	}

	// Access token should be different from refresh token
	if accessToken == refreshToken {
		t.Error("Access token should be different from refresh token")
	}

	// Validate refresh token
	refreshClaims, err := jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}

	if refreshClaims.TokenType != string(RefreshToken) {
		t.Errorf("TokenType mismatch: got %s, expected %s", refreshClaims.TokenType, RefreshToken)
	}

	// Refresh token should not be valid as access token
	_, err = jwtService.ValidateAccessToken(refreshToken)
	if err == nil {
		t.Error("Refresh token should not be valid as access token")
	}

	// Access token should not be valid as refresh token
	_, err = jwtService.ValidateRefreshToken(accessToken)
	if err == nil {
		t.Error("Access token should not be valid as refresh token")
	}
}

// Test 5: Testar expiracao de token
func TestTokenExpiration(t *testing.T) {
	// Create JWT service with very short expiration
	jwtService, err := NewJWTService(
		"test-access-secret-key-32-chars!",
		"test-refresh-secret-key-32chars!",
		1*time.Millisecond, // Very short expiration for testing
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Generate access token
	accessToken, err := jwtService.GenerateAccessToken(
		uuid.New().String(),
		"test@test.com",
		"operador",
		"",
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate should fail with expired error
	_, err = jwtService.ValidateAccessToken(accessToken)
	if err == nil {
		t.Error("Validation should fail for expired token")
	}
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got: %v", err)
	}
}

// Test 6: Testar validacao de senha com requisitos
func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError error
	}{
		{"valid password", "validPass123", nil},
		{"too short", "short", ErrPasswordTooShort},
		{"exactly 8 chars", "12345678", nil},
		{"long password", "thisIsAVeryLongPasswordThatIsStillValid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("Expected error %v, got %v", tt.expectError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*User),
	}
}

func (r *MockUserRepository) AddUser(user *User) {
	r.users[user.Email] = user
	r.users[user.ID.String()] = user
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, ok := r.users[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, ok := r.users[id.String()]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Test para login com mock repository
func TestAuthServiceLogin(t *testing.T) {
	// Create JWT service
	jwtService, err := NewJWTService(
		"test-access-secret-key-32-chars!",
		"test-refresh-secret-key-32chars!",
		15*time.Minute,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Create mock repository with test user
	userRepo := NewMockUserRepository()
	passwordHash, _ := HashPassword("demo123")
	testUser := &User{
		ID:           uuid.New(),
		Email:        "operador@sidot.gov.br",
		PasswordHash: passwordHash,
		Nome:         "Operador Teste",
		Role:         "operador",
		Ativo:        true,
	}
	userRepo.AddUser(testUser)

	// Create auth service without Redis (nil)
	authService := NewAuthService(jwtService, userRepo, nil)

	// Test successful login
	ctx := context.Background()
	result, err := authService.Login(ctx, "operador@sidot.gov.br", "demo123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if result.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if result.TokenType != "Bearer" {
		t.Errorf("TokenType should be Bearer, got %s", result.TokenType)
	}
	if result.User.Email != testUser.Email {
		t.Errorf("User email mismatch: got %s, expected %s", result.User.Email, testUser.Email)
	}

	// Test login with wrong password
	_, err = authService.Login(ctx, "operador@sidot.gov.br", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}

	// Test login with non-existent user
	_, err = authService.Login(ctx, "notexist@test.com", "demo123")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials for non-existent user, got: %v", err)
	}

	// Test login with inactive user
	inactiveUser := &User{
		ID:           uuid.New(),
		Email:        "inactive@sidot.gov.br",
		PasswordHash: passwordHash,
		Nome:         "Inactive User",
		Role:         "operador",
		Ativo:        false,
	}
	userRepo.AddUser(inactiveUser)

	_, err = authService.Login(ctx, "inactive@sidot.gov.br", "demo123")
	if err != ErrUserInactive {
		t.Errorf("Expected ErrUserInactive, got: %v", err)
	}
}

// Benchmark para hash de senha
func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

// Benchmark para validacao de JWT
func BenchmarkValidateAccessToken(b *testing.B) {
	jwtService, _ := NewJWTService(
		"test-access-secret-key-32-chars!",
		"test-refresh-secret-key-32chars!",
		15*time.Minute,
		7*24*time.Hour,
	)

	token, _ := jwtService.GenerateAccessToken(
		uuid.New().String(),
		"benchmark@test.com",
		"operador",
		"",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jwtService.ValidateAccessToken(token)
	}
}
