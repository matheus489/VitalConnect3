# Specification: Multi-Tenant Support

## Goal
Implement multi-tenant architecture in VitalConnect to enable multiple Transplant Centers (Centrais de Transplante) to operate on the same instance with complete data isolation and independent configurations.

## User Stories
- As a Central de Transplantes operator, I want to access only my organization's data so that patient information remains isolated and secure
- As a super admin, I want to switch between tenant contexts so that I can provide support across multiple Transplant Centers

## Specific Requirements

**Tenant Model Creation**
- Create `tenants` table as a global table (no tenant_id column) with fields: id (UUID), name, slug (unique), created_at, updated_at
- Tenant represents a Central de Transplantes (e.g., SES-GO, SES-PE, SES-SP)
- Slug must be unique and URL-safe for future extensibility
- No soft delete needed for tenants as deletion is extremely rare

**JWT Claims Extension**
- Add `tenant_id` (UUID) to the JWT Claims struct in `/backend/internal/services/auth/jwt.go`
- Add `is_super_admin` (boolean) to Claims for cross-tenant access identification
- Update `GenerateAccessToken` and `GenerateRefreshToken` to include tenant_id
- Update `UserClaims` struct in `/backend/internal/middleware/auth.go` to include TenantID and IsSuperAdmin

**Tenant Context Middleware**
- Create new middleware `TenantContext` in `/backend/internal/middleware/tenant.go`
- Extract tenant_id from JWT claims after authentication
- Store tenant_id in Gin context for downstream use
- Validate that non-super-admin users can only access their assigned tenant
- Super admins can pass optional `X-Tenant-Context` header to switch context

**User Model Updates**
- Add `tenant_id` (UUID, NOT NULL) column to users table
- Add `is_super_admin` (boolean, default false) column to users table
- Update User struct in `/backend/internal/models/user.go` with TenantID and IsSuperAdmin fields
- Super admin users have tenant_id but can access all tenants via context switch

**Database Schema Updates (tenant_id Addition)**
- Add `tenant_id` (UUID, NOT NULL) column to: hospitals, users, occurrences, obitos_simulados, triagem_rules, shifts, notifications, audit_logs
- Create foreign key constraints referencing tenants(id)
- Create indexes on tenant_id columns for query performance
- Update user_hospitals junction table to include tenant_id for integrity

**Repository Pattern Updates**
- Create `TenantScope` helper function that returns SQL WHERE clause fragment
- Update all repository queries to include tenant_id filter extracted from context
- Create `WithTenantScope(ctx context.Context, query string)` utility for consistent filtering
- Super admin bypass: when is_super_admin=true and specific tenant not in context, allow cross-tenant queries

**Triagem Rules Template System**
- Create default triagem rules template based on federal legislation
- Store template rules as seed data (not in tenants table)
- When creating new tenant via script, copy template rules with new tenant_id
- Each tenant's rules are fully independent after initial copy

**Legacy Data Migration Strategy**
- Create tenant "SES-GO" as the first tenant record with a stable UUID
- Add tenant_id column with DEFAULT set to SES-GO UUID
- Run UPDATE to set all existing records to SES-GO tenant_id
- Remove DEFAULT constraint after migration
- Add NOT NULL constraint after data backfill
- Execute in single atomic transaction with rollback capability

**Super Admin Access Control**
- Super admins identified by is_super_admin=true flag on user record
- Super admin can view aggregated data across all tenants
- Cross-tenant access must be logged in audit_logs with CRITICAL severity
- Super admin context switch requires explicit header (no implicit cross-tenant)

**Tenant-Aware Audit Logging**
- Add tenant_id to audit_logs table for tenant-scoped queries
- Log all super admin cross-tenant actions with source and target tenant context
- Create new action constants: ActionTenantContextSwitch, ActionCrossTenantAccess
- Ensure audit logs remain queryable by tenant for compliance

## Existing Code to Leverage

**JWT Service (`/backend/internal/services/auth/jwt.go`)**
- Existing Claims struct already has UserID, Email, Role, HospitalID fields
- GenerateAccessToken/GenerateRefreshToken methods accept individual parameters, extend to include tenant_id
- ValidateAccessToken returns Claims pointer, already used by middleware

**Auth Middleware (`/backend/internal/middleware/auth.go`)**
- UserClaims struct already extracted from JWT and stored in Gin context
- AuthRequired middleware pattern can be followed for TenantContext middleware
- GetUserClaims helper function pattern should be replicated for GetTenantContext

**Repository Pattern (`/backend/internal/repository/`)**
- All repositories use context.Context as first parameter, can extract tenant from context
- Consistent SQL query pattern with QueryContext and db.ExecContext
- Error handling patterns (ErrNotFound, etc.) already established

**Migration Structure (`/backend/migrations/`)**
- Numbered SQL migration files (001_, 002_, etc.) with clear naming
- Current highest migration is 018_add_telefone_to_hospitals.sql
- Migration runner in run_migrations.go handles execution

**Audit Log Model (`/backend/internal/models/audit_log.go`)**
- Already has HospitalID optional field, add TenantID following same pattern
- Severity levels (INFO, WARN, CRITICAL) already defined
- Action constants pattern established for new tenant-related actions

## Out of Scope
- Custom branding or white-label UI per tenant
- Billing or automated payment per tenant
- Cross-tenant data visibility for regular users
- Subdomain or custom DNS per tenant
- PostgreSQL Row-Level Security (RLS) - using application-level control instead
- UI for tenant management - creation via scripts/seeds only
- Tenant deletion functionality
- Tenant feature flags differentiation (all tenants have same features)
- Real-time tenant switching UI for super admins (manual header only)
- Tenant-specific configuration beyond triagem rules
