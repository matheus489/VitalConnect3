package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/sidot/backend/internal/middleware"
)

func TestTenantFilter(t *testing.T) {
	t.Run("should filter when tenant ID present", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		filter := NewTenantFilter(ctx)

		assert.True(t, filter.HasContext)
		assert.True(t, filter.ShouldFilter())
		assert.Equal(t, tenantID, filter.TenantID)
		assert.False(t, filter.IsSuperAdmin)
	})

	t.Run("should not filter when no context", func(t *testing.T) {
		ctx := context.Background()

		filter := NewTenantFilter(ctx)

		assert.False(t, filter.HasContext)
		assert.False(t, filter.ShouldFilter())
	})

	t.Run("should not filter for super admin without tenant", func(t *testing.T) {
		ctx := middleware.WithTenantContext(context.Background(), "", true)

		filter := NewTenantFilter(ctx)

		assert.True(t, filter.HasContext)
		assert.False(t, filter.ShouldFilter())
		assert.True(t, filter.IsSuperAdmin)
	})

	t.Run("should filter for super admin with specific tenant", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, true)

		filter := NewTenantFilter(ctx)

		assert.True(t, filter.HasContext)
		assert.True(t, filter.ShouldFilter())
		assert.True(t, filter.IsSuperAdmin)
		assert.Equal(t, tenantID, filter.TenantID)
	})
}

func TestTenantFilter_WhereClause(t *testing.T) {
	t.Run("should return proper WHERE clause", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		filter := NewTenantFilter(ctx)
		clause := filter.WhereClause()

		expected := "tenant_id = '" + tenantID + "'"
		assert.Equal(t, expected, clause)
	})

	t.Run("should return 1=1 when not filtering", func(t *testing.T) {
		ctx := context.Background()

		filter := NewTenantFilter(ctx)
		clause := filter.WhereClause()

		assert.Equal(t, "1=1", clause)
	})

	t.Run("should return WHERE clause with alias", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		filter := NewTenantFilter(ctx)
		clause := filter.WhereClauseWithAlias("u")

		expected := "u.tenant_id = '" + tenantID + "'"
		assert.Equal(t, expected, clause)
	})
}

func TestTenantFilter_AndClause(t *testing.T) {
	t.Run("should return AND clause when filtering", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		filter := NewTenantFilter(ctx)
		clause := filter.AndClause()

		expected := " AND tenant_id = '" + tenantID + "'"
		assert.Equal(t, expected, clause)
	})

	t.Run("should return empty string when not filtering", func(t *testing.T) {
		ctx := context.Background()

		filter := NewTenantFilter(ctx)
		clause := filter.AndClause()

		assert.Equal(t, "", clause)
	})

	t.Run("should return AND clause with alias", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		filter := NewTenantFilter(ctx)
		clause := filter.AndClauseWithAlias("h")

		expected := " AND h.tenant_id = '" + tenantID + "'"
		assert.Equal(t, expected, clause)
	})
}

func TestTenantScope(t *testing.T) {
	t.Run("should return scope when tenant ID provided", func(t *testing.T) {
		tenantID := uuid.New().String()
		scope := TenantScope(tenantID)

		expected := "tenant_id = '" + tenantID + "'"
		assert.Equal(t, expected, scope)
	})

	t.Run("should return 1=1 when tenant ID empty", func(t *testing.T) {
		scope := TenantScope("")

		assert.Equal(t, "1=1", scope)
	})

	t.Run("should return scope with alias", func(t *testing.T) {
		tenantID := uuid.New().String()
		scope := TenantScopeWithAlias(tenantID, "t")

		expected := "t.tenant_id = '" + tenantID + "'"
		assert.Equal(t, expected, scope)
	})
}

func TestGetTenantIDOrNil(t *testing.T) {
	t.Run("should return tenant ID when present", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		result := GetTenantIDOrNil(ctx)

		assert.Equal(t, tenantID, result)
	})

	t.Run("should return empty string when no context", func(t *testing.T) {
		ctx := context.Background()

		result := GetTenantIDOrNil(ctx)

		assert.Equal(t, "", result)
	})
}

func TestIsSuperAdminContext(t *testing.T) {
	t.Run("should return true for super admin", func(t *testing.T) {
		ctx := middleware.WithTenantContext(context.Background(), uuid.New().String(), true)

		result := IsSuperAdminContext(ctx)

		assert.True(t, result)
	})

	t.Run("should return false for regular user", func(t *testing.T) {
		ctx := middleware.WithTenantContext(context.Background(), uuid.New().String(), false)

		result := IsSuperAdminContext(ctx)

		assert.False(t, result)
	})

	t.Run("should return false when no context", func(t *testing.T) {
		ctx := context.Background()

		result := IsSuperAdminContext(ctx)

		assert.False(t, result)
	})
}

func TestRequireTenantID(t *testing.T) {
	t.Run("should return tenant ID when present", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := middleware.WithTenantContext(context.Background(), tenantID, false)

		result, err := RequireTenantID(ctx)

		assert.NoError(t, err)
		assert.Equal(t, tenantID, result)
	})

	t.Run("should return error when tenant ID empty", func(t *testing.T) {
		ctx := middleware.WithTenantContext(context.Background(), "", true)

		_, err := RequireTenantID(ctx)

		assert.Error(t, err)
		assert.Equal(t, ErrTenantIDRequired, err)
	})

	t.Run("should return error when no context", func(t *testing.T) {
		ctx := context.Background()

		_, err := RequireTenantID(ctx)

		assert.Error(t, err)
		assert.Equal(t, ErrTenantIDRequired, err)
	})
}
