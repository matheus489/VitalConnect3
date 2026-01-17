package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/vitalconnect/backend/internal/middleware"
)

// Errors for tenant scoping
var (
	ErrTenantContextMissing = errors.New("tenant context missing from request")
	ErrTenantIDRequired     = errors.New("tenant ID is required for this operation")
)

// TenantScope returns a SQL WHERE clause fragment for tenant filtering
// Example: "tenant_id = 'uuid-value'"
func TenantScope(tenantID string) string {
	if tenantID == "" {
		return "1=1" // No filtering if tenant ID is empty (super-admin bypass)
	}
	return fmt.Sprintf("tenant_id = '%s'", tenantID)
}

// TenantScopeWithAlias returns a SQL WHERE clause fragment for tenant filtering with table alias
// Example: "t.tenant_id = 'uuid-value'"
func TenantScopeWithAlias(tenantID, alias string) string {
	if tenantID == "" {
		return "1=1" // No filtering if tenant ID is empty (super-admin bypass)
	}
	return fmt.Sprintf("%s.tenant_id = '%s'", alias, tenantID)
}

// WithTenantScope appends tenant filtering to a base query
// Example: WithTenantScope(ctx, "SELECT * FROM hospitals WHERE deleted_at IS NULL")
//          returns "SELECT * FROM hospitals WHERE deleted_at IS NULL AND tenant_id = 'uuid'"
func WithTenantScope(ctx context.Context, baseQuery string) (string, error) {
	tenantID, isSuperAdmin, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		// If no tenant context, return base query (for backward compatibility during migration)
		return baseQuery, nil
	}

	// Super-admin without specific tenant context can see all data
	if isSuperAdmin && tenantID == "" {
		return baseQuery, nil
	}

	// Append tenant filter to query
	return fmt.Sprintf("%s AND tenant_id = '%s'", baseQuery, tenantID), nil
}

// GetTenantIDFromContext extracts the tenant ID from context
// Returns error if tenant context is missing and required
func GetTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, _, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		return "", ErrTenantContextMissing
	}
	return tenantID, nil
}

// GetTenantIDOrNil extracts the tenant ID from context, returning empty string if not found
// This is useful for queries that should work with or without tenant context
func GetTenantIDOrNil(ctx context.Context) string {
	tenantID, _, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		return ""
	}
	return tenantID
}

// IsSuperAdminContext checks if the current context has super admin privileges
func IsSuperAdminContext(ctx context.Context) bool {
	_, isSuperAdmin, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		return false
	}
	return isSuperAdmin
}

// RequireTenantID ensures tenant ID is present in context, returns error if missing
func RequireTenantID(ctx context.Context) (string, error) {
	tenantID := GetTenantIDOrNil(ctx)
	if tenantID == "" {
		return "", ErrTenantIDRequired
	}
	return tenantID, nil
}

// TenantFilter is a helper struct for building tenant-aware queries
type TenantFilter struct {
	TenantID     string
	IsSuperAdmin bool
	HasContext   bool
}

// NewTenantFilter creates a TenantFilter from context
func NewTenantFilter(ctx context.Context) *TenantFilter {
	tenantID, isSuperAdmin, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		return &TenantFilter{
			HasContext: false,
		}
	}
	return &TenantFilter{
		TenantID:     tenantID,
		IsSuperAdmin: isSuperAdmin,
		HasContext:   true,
	}
}

// ShouldFilter returns true if tenant filtering should be applied
func (f *TenantFilter) ShouldFilter() bool {
	if !f.HasContext {
		return false // No context, don't filter (backward compatibility)
	}
	if f.IsSuperAdmin && f.TenantID == "" {
		return false // Super-admin with no specific tenant, don't filter
	}
	return f.TenantID != ""
}

// WhereClause returns the WHERE clause fragment for tenant filtering
func (f *TenantFilter) WhereClause() string {
	if !f.ShouldFilter() {
		return "1=1"
	}
	return fmt.Sprintf("tenant_id = '%s'", f.TenantID)
}

// WhereClauseWithAlias returns the WHERE clause fragment with table alias
func (f *TenantFilter) WhereClauseWithAlias(alias string) string {
	if !f.ShouldFilter() {
		return "1=1"
	}
	return fmt.Sprintf("%s.tenant_id = '%s'", alias, f.TenantID)
}

// AndClause returns " AND tenant_id = 'uuid'" if filtering is needed, empty string otherwise
func (f *TenantFilter) AndClause() string {
	if !f.ShouldFilter() {
		return ""
	}
	return fmt.Sprintf(" AND tenant_id = '%s'", f.TenantID)
}

// AndClauseWithAlias returns " AND alias.tenant_id = 'uuid'" if filtering is needed
func (f *TenantFilter) AndClauseWithAlias(alias string) string {
	if !f.ShouldFilter() {
		return ""
	}
	return fmt.Sprintf(" AND %s.tenant_id = '%s'", alias, f.TenantID)
}
