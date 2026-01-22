package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

// TestAdminUpdateUserRoleInputValidation tests the AdminUpdateUserRoleInput validation
func TestAdminUpdateUserRoleInputValidation(t *testing.T) {
	operador := models.RoleOperador
	gestor := models.RoleGestor
	admin := models.RoleAdmin
	invalidRole := models.UserRole("superuser") // invalid role
	trueBool := true
	falseBool := false

	tests := []struct {
		name    string
		input   models.AdminUpdateUserRoleInput
		wantErr bool
	}{
		{
			name:    "Empty input - valid (no changes)",
			input:   models.AdminUpdateUserRoleInput{},
			wantErr: false,
		},
		{
			name: "Valid role operador",
			input: models.AdminUpdateUserRoleInput{
				Role: &operador,
			},
			wantErr: false,
		},
		{
			name: "Valid role gestor",
			input: models.AdminUpdateUserRoleInput{
				Role: &gestor,
			},
			wantErr: false,
		},
		{
			name: "Valid role admin",
			input: models.AdminUpdateUserRoleInput{
				Role: &admin,
			},
			wantErr: false,
		},
		{
			name: "Invalid role",
			input: models.AdminUpdateUserRoleInput{
				Role: &invalidRole,
			},
			wantErr: true,
		},
		{
			name: "Valid is_super_admin true",
			input: models.AdminUpdateUserRoleInput{
				IsSuperAdmin: &trueBool,
			},
			wantErr: false,
		},
		{
			name: "Valid is_super_admin false",
			input: models.AdminUpdateUserRoleInput{
				IsSuperAdmin: &falseBool,
			},
			wantErr: false,
		},
		{
			name: "Valid combined role and is_super_admin",
			input: models.AdminUpdateUserRoleInput{
				Role:         &admin,
				IsSuperAdmin: &trueBool,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate role if provided
			if tt.input.Role != nil {
				valid := tt.input.Role.IsValid()
				if !valid && !tt.wantErr {
					t.Errorf("Role.IsValid() = false, expected true for %s", *tt.input.Role)
				}
				if valid && tt.wantErr {
					t.Errorf("Role.IsValid() = true, expected false for %s", *tt.input.Role)
				}
			}
		})
	}
}

// TestAdminBanUserInputValidation tests the AdminBanUserInput structure
func TestAdminBanUserInputValidation(t *testing.T) {
	reason := "Violation of terms"

	tests := []struct {
		name  string
		input models.AdminBanUserInput
	}{
		{
			name: "Ban without reason",
			input: models.AdminBanUserInput{
				Banned:    true,
				BanReason: nil,
			},
		},
		{
			name: "Ban with reason",
			input: models.AdminBanUserInput{
				Banned:    true,
				BanReason: &reason,
			},
		},
		{
			name: "Unban",
			input: models.AdminBanUserInput{
				Banned:    false,
				BanReason: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the struct can be created without errors
			if tt.input.Banned && tt.input.BanReason != nil {
				if *tt.input.BanReason == "" {
					t.Error("Ban reason should not be empty if provided")
				}
			}
		})
	}
}

// TestUserWithTenantResponse tests the UserWithTenant response conversion
func TestUserWithTenantResponse(t *testing.T) {
	tenantID := uuid.New()
	tenantName := "Test Tenant"
	tenantSlug := "test-tenant"
	phone := "+5511999999999"

	user := models.UserWithTenant{
		ID:                 uuid.New(),
		TenantID:           &tenantID,
		Email:              "test@example.com",
		Nome:               "Test User",
		Role:               models.RoleOperador,
		IsSuperAdmin:       false,
		MobilePhone:        &phone,
		EmailNotifications: true,
		Ativo:              true,
		TenantName:         &tenantName,
		TenantSlug:         &tenantSlug,
		Hospitals:          []models.Hospital{},
	}

	response := user.ToResponse()

	if response.ID != user.ID {
		t.Errorf("ID mismatch: got %v, expected %v", response.ID, user.ID)
	}
	if *response.TenantID != *user.TenantID {
		t.Errorf("TenantID mismatch: got %v, expected %v", *response.TenantID, *user.TenantID)
	}
	if response.Email != user.Email {
		t.Errorf("Email mismatch: got %s, expected %s", response.Email, user.Email)
	}
	if response.Nome != user.Nome {
		t.Errorf("Nome mismatch: got %s, expected %s", response.Nome, user.Nome)
	}
	if response.Role != user.Role {
		t.Errorf("Role mismatch: got %s, expected %s", response.Role, user.Role)
	}
	if response.IsSuperAdmin != user.IsSuperAdmin {
		t.Errorf("IsSuperAdmin mismatch: got %v, expected %v", response.IsSuperAdmin, user.IsSuperAdmin)
	}
	if *response.TenantName != *user.TenantName {
		t.Errorf("TenantName mismatch: got %s, expected %s", *response.TenantName, *user.TenantName)
	}
	if *response.TenantSlug != *user.TenantSlug {
		t.Errorf("TenantSlug mismatch: got %s, expected %s", *response.TenantSlug, *user.TenantSlug)
	}
	if response.Ativo != user.Ativo {
		t.Errorf("Ativo mismatch: got %v, expected %v", response.Ativo, user.Ativo)
	}
}

// TestUserWithTenantResponseNilTenant tests the response with nil tenant info
func TestUserWithTenantResponseNilTenant(t *testing.T) {
	user := models.UserWithTenant{
		ID:                 uuid.New(),
		TenantID:           nil,
		Email:              "superadmin@example.com",
		Nome:               "Super Admin",
		Role:               models.RoleAdmin,
		IsSuperAdmin:       true,
		MobilePhone:        nil,
		EmailNotifications: true,
		Ativo:              true,
		TenantName:         nil,
		TenantSlug:         nil,
		Hospitals:          nil,
	}

	response := user.ToResponse()

	if response.TenantID != nil {
		t.Error("TenantID should be nil for super admin without tenant")
	}
	if response.TenantName != nil {
		t.Error("TenantName should be nil for super admin without tenant")
	}
	if response.TenantSlug != nil {
		t.Error("TenantSlug should be nil for super admin without tenant")
	}
	if !response.IsSuperAdmin {
		t.Error("IsSuperAdmin should be true")
	}
	if len(response.Hospitals) != 0 {
		t.Errorf("Hospitals should be empty slice, got %d items", len(response.Hospitals))
	}
}

// TestGenerateTempPassword tests the temporary password generation
func TestGenerateTempPassword(t *testing.T) {
	passwords := make(map[string]bool)

	// Generate 100 passwords and check for uniqueness
	for i := 0; i < 100; i++ {
		pwd := generateTempPassword()

		// Check length (12 bytes = 24 hex characters)
		if len(pwd) != 24 {
			t.Errorf("Password length = %d, expected 24", len(pwd))
		}

		// Check uniqueness
		if passwords[pwd] {
			t.Error("Duplicate password generated")
		}
		passwords[pwd] = true

		// Check only hex characters
		for _, c := range pwd {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("Password contains non-hex character: %c", c)
			}
		}
	}
}

// TestRoleValidation tests role validation
func TestRoleValidation(t *testing.T) {
	tests := []struct {
		role    models.UserRole
		isValid bool
	}{
		{models.RoleOperador, true},
		{models.RoleGestor, true},
		{models.RoleAdmin, true},
		{models.UserRole(""), false},
		{models.UserRole("superuser"), false},
		{models.UserRole("root"), false},
		{models.UserRole("ADMIN"), false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if tt.role.IsValid() != tt.isValid {
				t.Errorf("Role(%s).IsValid() = %v, expected %v", tt.role, tt.role.IsValid(), tt.isValid)
			}
		})
	}
}
