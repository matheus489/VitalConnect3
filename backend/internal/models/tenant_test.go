package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSlug(t *testing.T) {
	testCases := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{"valid slug", "ses-go", false},
		{"valid slug with numbers", "ses-go-2", false},
		{"valid single word", "sesgo", false},
		{"valid long slug", "secretaria-de-saude-goias", false},

		{"too short", "a", true},
		{"empty", "", true},
		{"contains uppercase", "SES-GO", true},
		{"contains underscore", "ses_go", true},
		{"contains space", "ses go", true},
		{"starts with hyphen", "-ses-go", true},
		{"ends with hyphen", "ses-go-", true},
		{"double hyphen", "ses--go", true},
		{"contains special char", "ses@go", true},
		{"too long", "this-is-a-very-long-slug-that-exceeds-the-maximum-allowed-length-for-tenant-slugs-which-is-100-characters-total", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSlug(tc.slug)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidTenantSlug, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenant_Validate(t *testing.T) {
	t.Run("valid tenant", func(t *testing.T) {
		tenant := &Tenant{
			Name: "Secretaria de Saude de Goias",
			Slug: "ses-go",
		}
		err := tenant.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid name - too short", func(t *testing.T) {
		tenant := &Tenant{
			Name: "A",
			Slug: "ses-go",
		}
		err := tenant.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid name - empty", func(t *testing.T) {
		tenant := &Tenant{
			Name: "",
			Slug: "ses-go",
		}
		err := tenant.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid slug", func(t *testing.T) {
		tenant := &Tenant{
			Name: "Valid Name",
			Slug: "Invalid Slug",
		}
		err := tenant.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTenantSlug, err)
	})
}

func TestCreateTenantInput_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		input := &CreateTenantInput{
			Name: "Secretaria de Saude de Pernambuco",
			Slug: "ses-pe",
		}
		err := input.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid name", func(t *testing.T) {
		input := &CreateTenantInput{
			Name: "X",
			Slug: "ses-pe",
		}
		err := input.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid slug", func(t *testing.T) {
		input := &CreateTenantInput{
			Name: "Valid Name",
			Slug: "SES_PE",
		}
		err := input.Validate()
		assert.Error(t, err)
	})
}

func TestTenant_ToResponse(t *testing.T) {
	tenant := &Tenant{
		Name: "Test Tenant",
		Slug: "test-tenant",
	}

	resp := tenant.ToResponse()

	assert.Equal(t, tenant.ID, resp.ID)
	assert.Equal(t, tenant.Name, resp.Name)
	assert.Equal(t, tenant.Slug, resp.Slug)
	assert.Equal(t, tenant.CreatedAt, resp.CreatedAt)
	assert.Equal(t, tenant.UpdatedAt, resp.UpdatedAt)
}
