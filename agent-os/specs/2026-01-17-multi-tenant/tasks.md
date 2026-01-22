# Task Breakdown: Multi-Tenant Support

## Overview
Total Tasks: 42 sub-tasks across 6 task groups
Estimated Duration: 3-4 development cycles
Complexity: High (architectural change affecting entire codebase)

## Task List

### Database Layer

#### Task Group 1: Tenant Table and Core Schema
**Dependencies:** None
**Complexity:** Medium
**Specialist:** Database Engineer / Backend Developer

- [x] 1.0 Complete tenant table and base schema
  - [x] 1.1 Write 4-6 focused tests for Tenant model functionality
    - Test tenant creation with valid fields (name, slug)
    - Test slug uniqueness constraint
    - Test UUID generation for tenant ID
    - Test timestamps (created_at, updated_at) auto-population
  - [x] 1.2 Create migration `019_create_tenants.sql`
    - Table: `tenants` (global table, no tenant_id)
    - Fields: `id` (UUID PRIMARY KEY DEFAULT gen_random_uuid()), `name` (VARCHAR(255) NOT NULL), `slug` (VARCHAR(100) UNIQUE NOT NULL), `created_at` (TIMESTAMPTZ DEFAULT NOW()), `updated_at` (TIMESTAMPTZ DEFAULT NOW())
    - Index on `slug` for lookups
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/019_create_tenants.sql`
  - [x] 1.3 Create Tenant model struct
    - Fields: ID (uuid.UUID), Name (string), Slug (string), CreatedAt, UpdatedAt
    - Add validation methods (ValidateSlug - alphanumeric with hyphens only)
    - File: `/home/matheus_rubem/SIDOT/backend/internal/models/tenant.go`
  - [x] 1.4 Create TenantRepository with basic CRUD
    - Methods: Create, GetByID, GetBySlug, List
    - Follow existing repository patterns (use context.Context, db.QueryContext)
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/tenant_repository.go`
  - [x] 1.5 Ensure tenant table tests pass
    - Run ONLY the 4-6 tests written in 1.1
    - Verify migration runs successfully

**Acceptance Criteria:**
- Tenants table created with correct schema
- Tenant model with validation works correctly
- Repository CRUD operations functional
- Slug uniqueness enforced at database level

**Files to Modify/Create:**
- `/home/matheus_rubem/SIDOT/backend/migrations/019_create_tenants.sql` (new)
- `/home/matheus_rubem/SIDOT/backend/internal/models/tenant.go` (new)
- `/home/matheus_rubem/SIDOT/backend/internal/repository/tenant_repository.go` (new)

---

#### Task Group 2: Add tenant_id to Existing Tables
**Dependencies:** Task Group 1
**Complexity:** High
**Specialist:** Database Engineer

- [x] 2.0 Complete tenant_id column additions to all 8 tables
  - [x] 2.1 Write 4-6 focused tests for tenant_id column integrity
    - Test foreign key constraint to tenants table
    - Test NOT NULL constraint enforcement
    - Test index existence on tenant_id columns
    - Test cascade behavior (or lack thereof) on tenant deletion
  - [x] 2.2 Create migration `020_add_tenant_id_to_users.sql`
    - Add `tenant_id` (UUID) column to `users` table
    - Add `is_super_admin` (BOOLEAN DEFAULT FALSE) column to `users` table
    - Add foreign key constraint to tenants(id)
    - Create index on `tenant_id`
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/020_add_tenant_id_to_users.sql`
  - [x] 2.3 Create migration `021_add_tenant_id_to_hospitals.sql`
    - Add `tenant_id` (UUID) column to `hospitals` table
    - Add foreign key constraint and index
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/021_add_tenant_id_to_hospitals.sql`
  - [x] 2.4 Create migration `022_add_tenant_id_to_data_tables.sql`
    - Add `tenant_id` to: `occurrences`, `obitos_simulados`, `triagem_rules`, `shifts`, `notifications`, `audit_logs`
    - Add foreign key constraints and indexes for each
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/022_add_tenant_id_to_data_tables.sql`
  - [x] 2.5 Create migration `023_add_tenant_id_to_user_hospitals.sql`
    - Add `tenant_id` to `user_hospitals` junction table for integrity
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/023_add_tenant_id_to_user_hospitals.sql`
  - [x] 2.6 Ensure tenant_id migration tests pass
    - Run ONLY the 4-6 tests written in 2.1
    - Verify all migrations run successfully in sequence

**Acceptance Criteria:**
- All 8 tables have tenant_id column
- Users table has is_super_admin column
- All foreign key constraints created
- Indexes created for query performance
- Migrations reversible (down methods)

**Files to Create:**
- `/home/matheus_rubem/SIDOT/backend/migrations/020_add_tenant_id_to_users.sql`
- `/home/matheus_rubem/SIDOT/backend/migrations/021_add_tenant_id_to_hospitals.sql`
- `/home/matheus_rubem/SIDOT/backend/migrations/022_add_tenant_id_to_data_tables.sql`
- `/home/matheus_rubem/SIDOT/backend/migrations/023_add_tenant_id_to_user_hospitals.sql`

---

### Backend Core Layer

#### Task Group 3: JWT Claims and Auth Updates
**Dependencies:** Task Group 1
**Complexity:** Medium
**Specialist:** Backend Developer (Auth)

- [x] 3.0 Complete JWT claims extension for multi-tenant
  - [x] 3.1 Write 4-6 focused tests for JWT tenant claims
    - Test GenerateAccessToken includes tenant_id claim
    - Test GenerateRefreshToken includes tenant_id claim
    - Test ValidateAccessToken extracts tenant_id correctly
    - Test is_super_admin claim presence and extraction
  - [x] 3.2 Update Claims struct in JWT service
    - Add `TenantID` (string) field to Claims struct
    - Add `IsSuperAdmin` (bool) field to Claims struct
    - Update JWT claim constants
    - File: `/home/matheus_rubem/SIDOT/backend/internal/services/auth/jwt.go`
  - [x] 3.3 Update GenerateAccessToken method
    - Accept tenant_id and is_super_admin as parameters
    - Include tenant_id and is_super_admin in JWT payload
    - File: `/home/matheus_rubem/SIDOT/backend/internal/services/auth/jwt.go`
  - [x] 3.4 Update GenerateRefreshToken method
    - Include tenant_id in refresh token claims
    - Ensure is_super_admin persists through refresh
    - File: `/home/matheus_rubem/SIDOT/backend/internal/services/auth/jwt.go`
  - [x] 3.5 Update UserClaims struct in auth middleware
    - Add `TenantID` (string) field
    - Add `IsSuperAdmin` (bool) field
    - Update AuthRequired middleware to extract new claims
    - File: `/home/matheus_rubem/SIDOT/backend/internal/middleware/auth.go`
  - [x] 3.6 Update login handler to include tenant_id in token generation
    - Fetch user's tenant_id from database during login
    - Pass tenant_id and is_super_admin to token generation
    - File: `/home/matheus_rubem/SIDOT/backend/internal/services/auth/service.go`
  - [x] 3.7 Ensure JWT tests pass
    - Run ONLY the 4-6 tests written in 3.1
    - Verify existing auth tests still pass

**Acceptance Criteria:**
- JWT tokens contain tenant_id and is_super_admin claims
- Token validation extracts all claims correctly
- Login flow generates tokens with tenant context
- Refresh token maintains tenant context
- Backwards compatible with existing tokens during migration

**Files to Modify:**
- `/home/matheus_rubem/SIDOT/backend/internal/services/auth/jwt.go`
- `/home/matheus_rubem/SIDOT/backend/internal/middleware/auth.go`
- `/home/matheus_rubem/SIDOT/backend/internal/services/auth/service.go`

---

#### Task Group 4: Tenant Context Middleware
**Dependencies:** Task Group 3
**Complexity:** Medium
**Specialist:** Backend Developer (Middleware)

- [x] 4.0 Complete tenant context middleware
  - [x] 4.1 Write 4-6 focused tests for tenant context middleware
    - Test tenant_id extraction from JWT claims
    - Test non-super-admin blocked from X-Tenant-Context header
    - Test super-admin can switch context via X-Tenant-Context header
    - Test GetTenantContext helper function
    - Test missing tenant context returns error for protected routes
  - [x] 4.2 Create TenantContext struct
    - Fields: TenantID (string), IsSuperAdmin (bool), EffectiveTenantID (string)
    - EffectiveTenantID = switched context for super-admin or original tenant_id
    - File: `/home/matheus_rubem/SIDOT/backend/internal/middleware/tenant.go`
  - [x] 4.3 Create TenantContext middleware
    - Extract tenant_id from user claims (set by AuthRequired)
    - Check for X-Tenant-Context header (super-admin only)
    - Validate super-admin permission for context switch
    - Store TenantContext in Gin context
    - File: `/home/matheus_rubem/SIDOT/backend/internal/middleware/tenant.go`
  - [x] 4.4 Create GetTenantContext helper function
    - Retrieve TenantContext from Gin context
    - Return error if not found
    - Pattern matches existing GetUserClaims helper
    - File: `/home/matheus_rubem/SIDOT/backend/internal/middleware/tenant.go`
  - [x] 4.5 Create RequireTenant middleware
    - Ensures tenant context is present
    - Returns 403 if missing or invalid
    - File: `/home/matheus_rubem/SIDOT/backend/internal/middleware/tenant.go`
  - [x] 4.6 Ensure tenant middleware tests pass
    - Run ONLY the 4-6 tests written in 4.1
    - Verify middleware chain works correctly

**Acceptance Criteria:**
- TenantContext extracted from JWT for all authenticated requests
- Super-admin can switch context via X-Tenant-Context header
- Regular users cannot use X-Tenant-Context header
- GetTenantContext helper works reliably
- Middleware integrates seamlessly with existing auth flow

**Files to Create:**
- `/home/matheus_rubem/SIDOT/backend/internal/middleware/tenant.go` (new)

**Files to Modify:**
- `/home/matheus_rubem/SIDOT/backend/internal/middleware/auth.go` (update UserClaims usage)

---

### Backend Repository Layer

#### Task Group 5: Repository Tenant Scoping
**Dependencies:** Task Groups 2, 4
**Complexity:** High
**Specialist:** Backend Developer (Repository)

- [x] 5.0 Complete tenant scoping for all repositories
  - [x] 5.1 Write 6-8 focused tests for repository tenant scoping
    - Test TenantScope helper generates correct WHERE clause
    - Test WithTenantScope utility adds tenant filter
    - Test HospitalRepository filters by tenant_id
    - Test OccurrenceRepository filters by tenant_id
    - Test UserRepository filters by tenant_id (excluding super-admin lookup)
    - Test super-admin bypass for cross-tenant queries
  - [x] 5.2 Create tenant scope utilities
    - Create `TenantScope(tenantID string) string` - returns WHERE fragment
    - Create `WithTenantScope(ctx context.Context, baseQuery string) string` - appends tenant filter
    - Create `GetTenantIDFromContext(ctx context.Context) (string, error)` - extracts tenant from ctx
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/tenant_scope.go`
  - [x] 5.3 Update User model with tenant fields
    - Add TenantID (uuid.UUID) field
    - Add IsSuperAdmin (bool) field
    - Update struct tags for JSON and DB mapping
    - File: `/home/matheus_rubem/SIDOT/backend/internal/models/user.go`
  - [x] 5.4 Update UserRepository with tenant scoping
    - Add tenant_id filter to all SELECT queries
    - Add tenant_id to INSERT statements
    - Exception: GetByID for super-admin can bypass (explicit flag)
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/user_repository.go`
  - [x] 5.5 Update HospitalRepository with tenant scoping (pattern provided)
    - Add tenant_id filter to List, GetByID queries
    - Add tenant_id to Create statement
    - Ensure map queries filter by tenant
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/hospital_repository.go`
  - [x] 5.6 Update OccurrenceRepository with tenant scoping (pattern provided)
    - Add tenant_id filter to all queries
    - Include tenant_id in occurrence creation
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/occurrence_repository.go`
  - [x] 5.7 Update ObitoRepository with tenant scoping (pattern provided)
    - Add tenant_id filter to all queries
    - Include tenant_id in obito_simulado creation
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/obito_repository.go`
  - [x] 5.8 Update TriagemRuleRepository with tenant scoping (pattern provided)
    - Add tenant_id filter to rule queries
    - Critical: Rules must be tenant-specific
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/triagem_rule_repository.go`
  - [x] 5.9 Update ShiftRepository with tenant scoping (pattern provided)
    - Add tenant_id filter to shift queries
    - Include tenant_id in shift creation
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/shift_repository.go`
  - [x] 5.10 Update NotificationRepository with tenant scoping (pattern provided)
    - Add tenant_id filter to notification queries
    - Include tenant_id in notification creation
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/notification_repository.go`
  - [x] 5.11 Update AuditLogRepository with tenant scoping
    - Add tenant_id filter to queries
    - Add TenantID field to AuditLog model
    - Add new action constants: ActionTenantContextSwitch, ActionCrossTenantAccess
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/audit_log_repository.go`
    - File: `/home/matheus_rubem/SIDOT/backend/internal/models/audit_log.go`
  - [x] 5.12 Update remaining repositories (pattern provided)
    - IndicatorsRepository: Add tenant filtering
    - OccurrenceHistoryRepository: Add tenant filtering
    - PushSubscriptionRepository: Add tenant filtering
    - UserNotificationPreferencesRepository: Add tenant filtering
  - [x] 5.13 Ensure repository tenant scoping tests pass
    - Run ONLY the 6-8 tests written in 5.1
    - Verify queries include tenant filter

**Acceptance Criteria:**
- All repositories filter by tenant_id from context
- TenantScope utility works consistently
- Super-admin bypass mechanism functional
- No data leakage between tenants
- Existing functionality preserved

**Files to Create:**
- `/home/matheus_rubem/SIDOT/backend/internal/repository/tenant_scope.go` (new)

**Files to Modify:**
- `/home/matheus_rubem/SIDOT/backend/internal/models/user.go`
- `/home/matheus_rubem/SIDOT/backend/internal/models/audit_log.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/user_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/hospital_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/occurrence_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/obito_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/triagem_rule_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/shift_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/notification_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/audit_log_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/indicators_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/occurrence_history_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/push_subscription_repository.go`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/user_notification_preferences_repository.go`

---

### Data Migration Layer

#### Task Group 6: Legacy Data Migration and Seed
**Dependencies:** Task Groups 1, 2
**Complexity:** High (requires atomic transaction)
**Specialist:** Database Engineer

- [x] 6.0 Complete legacy data migration to SES-GO tenant
  - [x] 6.1 Write 4-6 focused tests for migration integrity
    - Test SES-GO tenant created with stable UUID
    - Test all existing records have tenant_id after migration
    - Test NOT NULL constraint enforced after migration
    - Test rollback works correctly (transaction atomicity)
  - [x] 6.2 Create migration `024_seed_ses_go_tenant.sql`
    - INSERT SES-GO tenant with predefined stable UUID
    - UUID should be documented and consistent across environments
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/024_seed_sesgo_tenant.sql`
  - [x] 6.3 Create migration `025_backfill_tenant_id_legacy_data.sql`
    - Set DEFAULT on tenant_id columns to SES-GO UUID
    - UPDATE all existing records to set tenant_id = SES-GO UUID
    - Tables: users, hospitals, occurrences, obitos_simulados, triagem_rules, shifts, notifications, audit_logs, user_hospitals
    - Execute in single atomic transaction
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/025_backfill_tenant_id.sql`
  - [x] 6.4 Create migration `026_enforce_tenant_id_not_null.sql`
    - Remove DEFAULT constraint from all tenant_id columns
    - Add NOT NULL constraint to all tenant_id columns
    - Must run after data backfill
    - File: `/home/matheus_rubem/SIDOT/backend/migrations/026_enforce_tenant_id_not_null.sql`
  - [x] 6.5 Create default triagem rules template seed
    - Define template rules based on federal legislation
    - Store as SQL seed data for copying to new tenants
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/triagem_rules_template.go`
  - [x] 6.6 Create new tenant provisioning helper (CopyTriagemRulesToTenant method)
    - Method to create new tenant with template rules copy
    - Accept: name, slug as parameters
    - Copy triagem_rules from template with new tenant_id
    - Document usage in function comments
    - File: `/home/matheus_rubem/SIDOT/backend/internal/repository/triagem_rules_template.go`
  - [x] 6.7 Ensure migration tests pass
    - Run ONLY the 4-6 tests written in 6.1
    - Verify complete migration flow
    - Test rollback capability

**Acceptance Criteria:**
- SES-GO tenant created as first tenant
- All existing data associated with SES-GO tenant
- NOT NULL constraints enforced after backfill
- Migration fully reversible
- Triagem rules template ready for new tenants
- Tenant creation script documented and functional

**Files to Create:**
- `/home/matheus_rubem/SIDOT/backend/migrations/024_seed_sesgo_tenant.sql`
- `/home/matheus_rubem/SIDOT/backend/migrations/025_backfill_tenant_id.sql`
- `/home/matheus_rubem/SIDOT/backend/migrations/026_enforce_tenant_id_not_null.sql`
- `/home/matheus_rubem/SIDOT/backend/internal/repository/triagem_rules_template.go`

---

### Testing Layer

#### Task Group 7: Test Review and Integration Testing
**Dependencies:** Task Groups 1-6
**Complexity:** Medium
**Specialist:** QA Engineer / Backend Developer

- [x] 7.0 Review existing tests and fill critical gaps
  - [x] 7.1 Review tests from Task Groups 1-6
    - Review database tests (Groups 1, 2): ~8-12 tests
    - Review JWT/middleware tests (Groups 3, 4): ~8-12 tests
    - Review repository tests (Group 5): ~6-8 tests
    - Review migration tests (Group 6): ~4-6 tests
    - Total existing tests: approximately 26-38 tests
  - [x] 7.2 Analyze test coverage gaps for multi-tenant feature
    - Identify critical user workflows lacking coverage
    - Focus on tenant isolation scenarios
    - Prioritize end-to-end tenant data isolation
    - Do NOT assess entire application test coverage
  - [x] 7.3 Write up to 10 additional strategic tests
    - Integration test: User login generates token with correct tenant_id
    - Integration test: API request with tenant context returns only tenant data
    - Integration test: Super-admin context switch works correctly
    - Integration test: Super-admin cross-tenant access logged to audit
    - Integration test: Regular user cannot access other tenant data
    - Integration test: New user creation assigns correct tenant_id
    - Integration test: Hospital CRUD respects tenant boundaries
    - Integration test: Occurrence queries isolated by tenant
    - Add only tests for critical gaps identified in 7.2
    - Maximum 10 new tests
  - [x] 7.4 Run feature-specific tests only
    - Run ONLY multi-tenant related tests (from groups 1-6 plus 7.3)
    - Expected total: approximately 36-48 tests
    - Do NOT run entire application test suite
    - Verify tenant isolation across all layers

**Acceptance Criteria:**
- All feature-specific tests pass
- Tenant data isolation verified end-to-end
- Super-admin context switch tested
- Cross-tenant audit logging verified
- No more than 10 additional tests written

**Test Files to Create/Modify:**
- `/home/matheus_rubem/SIDOT/backend/internal/middleware/tenant_test.go` (new)
- `/home/matheus_rubem/SIDOT/backend/internal/repository/tenant_repository_test.go` (new)
- `/home/matheus_rubem/SIDOT/backend/internal/services/auth/jwt_test.go` (modify)
- `/home/matheus_rubem/SIDOT/backend/internal/repository/tenant_scope_test.go` (new)
- `/home/matheus_rubem/SIDOT/backend/internal/models/tenant_test.go` (new)

---

## Execution Order

Recommended implementation sequence with dependencies:

```
Phase 1: Database Foundation
  [1] Task Group 1: Tenant Table and Core Schema
      |
      v
  [2] Task Group 2: Add tenant_id to Existing Tables
      |
      v
  [3] Task Group 6: Legacy Data Migration and Seed

Phase 2: Backend Core
  [4] Task Group 3: JWT Claims and Auth Updates
      |
      v
  [5] Task Group 4: Tenant Context Middleware

Phase 3: Repository Integration
  [6] Task Group 5: Repository Tenant Scoping

Phase 4: Validation
  [7] Task Group 7: Test Review and Integration Testing
```

**Parallel Execution Opportunities:**
- Task Group 3 can start in parallel with Task Group 2 (different layers)
- Task Group 6 can run after Group 1 and 2 complete
- Task Group 4 must wait for Group 3
- Task Group 5 must wait for Groups 2 and 4

---

## Risk Mitigation

**Database Migration Risk:**
- Test migrations on staging database first
- Ensure DOWN migrations work for rollback
- Use transactions for data integrity
- Keep backup before production migration

**JWT Token Compatibility:**
- Existing tokens without tenant_id should be handled gracefully during transition
- Consider token refresh flow to get new claims
- Document migration path for active sessions

**Repository Query Risk:**
- Audit all repositories for missing tenant filters
- Use static analysis to find raw SQL queries
- Consider adding lint rule for tenant_id requirement

---

## Documentation Requirements

After implementation, update:
- [x] API documentation with new JWT claims
- [x] Deployment guide with migration steps
- [x] Tenant creation documentation (using script)
- [x] Super-admin usage guide for context switching
