# Task Breakdown: Backoffice Administrativo Completo

## Overview

**Total Task Groups:** 8
**Estimated Total Tasks:** 48+ sub-tasks
**Primary Stack:** Go (Gin) Backend + Next.js 14 Frontend + PostgreSQL

This feature implements a comprehensive super-admin backoffice with:
- Command Palette (CMD) for real-time tenant UI customization
- Cross-tenant management capabilities
- Dynamic frontend rendering based on JSONB theme configurations

---

## Task List

### Database Layer

#### Task Group 1: Database Migrations and Models
**Dependencies:** None
**Complexity:** Medium
**Files to Create/Modify:**
- `/backend/migrations/YYYYMMDD_add_tenant_theme_config.sql`
- `/backend/migrations/YYYYMMDD_create_system_settings.sql`
- `/backend/migrations/YYYYMMDD_create_triagem_rule_templates.sql`
- `/backend/internal/models/tenant.go` (modify)
- `/backend/internal/models/system_setting.go` (create)
- `/backend/internal/models/triagem_rule_template.go` (create)

- [x] 1.0 Complete database layer for backoffice
  - [x] 1.1 Write 4-6 focused tests for new models and migrations
    - Test Tenant model with theme_config JSONB field
    - Test SystemSetting model CRUD operations
    - Test TriagemRuleTemplate model CRUD operations
    - Test default theme_config structure validation
  - [x] 1.2 Create migration to extend tenants table
    ```sql
    ALTER TABLE tenants
    ADD COLUMN theme_config JSONB DEFAULT '{}',
    ADD COLUMN is_active BOOLEAN DEFAULT true,
    ADD COLUMN logo_url TEXT,
    ADD COLUMN favicon_url TEXT;
    ```
    - Add GIN index on theme_config for JSONB queries
    - Set default theme_config with empty structure
  - [x] 1.3 Create migration for system_settings table
    ```sql
    CREATE TABLE system_settings (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      key VARCHAR(100) UNIQUE NOT NULL,
      value JSONB NOT NULL,
      description TEXT,
      is_encrypted BOOLEAN DEFAULT false,
      created_at TIMESTAMPTZ DEFAULT NOW(),
      updated_at TIMESTAMPTZ DEFAULT NOW()
    );
    ```
    - Add unique constraint on key column
    - Include is_encrypted flag for sensitive values
  - [x] 1.4 Create migration for triagem_rule_templates table
    ```sql
    CREATE TABLE triagem_rule_templates (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      nome VARCHAR(255) NOT NULL,
      tipo VARCHAR(50) NOT NULL,
      condicao JSONB NOT NULL,
      descricao TEXT,
      ativo BOOLEAN DEFAULT true,
      created_at TIMESTAMPTZ DEFAULT NOW(),
      updated_at TIMESTAMPTZ DEFAULT NOW()
    );
    ```
  - [x] 1.5 Update Tenant model in `/backend/internal/models/tenant.go`
    - Add ThemeConfig field (JSONB)
    - Add IsActive field (bool)
    - Add LogoURL and FaviconURL fields
    - Create ThemeConfig struct with nested Theme and Layout
    - Follow existing ToResponse pattern
  - [x] 1.6 Create SystemSetting model in `/backend/internal/models/system_setting.go`
    - SystemSetting struct with Key, Value (JSONB), IsEncrypted
    - CreateSystemSettingInput and UpdateSystemSettingInput
    - SystemSettingResponse for API responses
    - Validation methods
  - [x] 1.7 Create TriagemRuleTemplate model in `/backend/internal/models/triagem_rule_template.go`
    - TriagemRuleTemplate struct with Nome, Tipo, Condicao (JSONB)
    - Input and Response types following existing patterns
    - Validation methods
  - [x] 1.8 Ensure database layer tests pass
    - Run ONLY the 4-6 tests written in 1.1
    - Verify all migrations run successfully
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 4-6 tests written in 1.1 pass
- All migrations apply cleanly to database
- Models include proper validation
- JSONB fields can store and retrieve theme configuration
- Existing tenant functionality remains intact

---

### Backend - Middleware and Routing

#### Task Group 2: SuperAdmin Middleware and Admin Route Group
**Dependencies:** Task Group 1
**Complexity:** Medium
**Files to Create/Modify:**
- `/backend/internal/middleware/super_admin.go` (create)
- `/backend/internal/routes/admin.go` (create)
- `/backend/cmd/api/main.go` (modify to register admin routes)

- [x] 2.0 Complete super admin middleware and routing
  - [x] 2.1 Write 4-6 focused tests for super admin middleware
    - Test RequireSuperAdmin allows is_super_admin=true
    - Test RequireSuperAdmin blocks is_super_admin=false
    - Test RequireSuperAdmin blocks missing claims
    - Test admin routes are protected
  - [x] 2.2 Create RequireSuperAdmin middleware in `/backend/internal/middleware/super_admin.go`
    - Follow existing RequireRole pattern from auth.go
    - Check IsSuperAdmin field from UserClaims
    - Return 403 Forbidden if not super admin
    - Include clear error messages
  - [x] 2.3 Create admin route group in `/backend/cmd/api/main.go`
    - Create `/api/v1/admin` router group
    - Apply AuthRequired() middleware
    - Apply RequireSuperAdmin() middleware
    - Register placeholder handlers (to be implemented in later tasks)
  - [x] 2.4 Update main.go to register admin routes
    - Register admin route group after existing routes
    - Ensure JWT service is available to admin routes
  - [x] 2.5 Ensure middleware tests pass
    - Run ONLY the 4-6 tests written in 2.1
    - Verify middleware blocks unauthorized access
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 4-6 tests written in 2.1 pass
- Super admin middleware correctly validates is_super_admin claim
- Admin routes are registered and protected
- Non-super-admin users receive 403 Forbidden

---

### Backend - API Handlers

#### Task Group 3: Admin Tenant Handlers
**Dependencies:** Task Groups 1, 2
**Complexity:** High
**Files to Create/Modify:**
- `/backend/internal/handlers/admin_tenants.go` (create)
- `/backend/internal/repository/admin_tenant_repo.go` (create)

- [x] 3.0 Complete admin tenant handlers
  - [x] 3.1 Write 6-8 focused tests for tenant admin endpoints
    - Test GET /admin/tenants lists all tenants (ignores tenant_id filter)
    - Test GET /admin/tenants/:id returns tenant with metrics
    - Test POST /admin/tenants creates new tenant
    - Test PUT /admin/tenants/:id updates tenant
    - Test PUT /admin/tenants/:id/theme updates theme_config
    - Test PUT /admin/tenants/:id/toggle toggles is_active
  - [x] 3.2 Create AdminTenantRepository in `/backend/internal/repository/admin_tenant_repo.go`
    - ListAllTenants (no tenant_id filter, with pagination)
    - GetTenantByID with user/hospital/occurrence counts
    - CreateTenant with default theme_config
    - UpdateTenant
    - UpdateThemeConfig (partial JSONB update)
    - ToggleTenantActive
  - [x] 3.3 Create admin tenant handlers in `/backend/internal/handlers/admin_tenants.go`
    - ListTenants: GET /admin/tenants (pagination, search, status filter)
    - GetTenant: GET /admin/tenants/:id
    - CreateTenant: POST /admin/tenants
    - UpdateTenant: PUT /admin/tenants/:id
    - UpdateThemeConfig: PUT /admin/tenants/:id/theme
    - ToggleActive: PUT /admin/tenants/:id/toggle
  - [x] 3.4 Implement asset upload endpoint
    - POST /admin/tenants/:id/assets
    - Accept multipart form (logo, favicon)
    - Store in configured storage (local/S3)
    - Update tenant logo_url/favicon_url
  - [x] 3.5 Register tenant admin routes in main.go (replace placeholders)
    - Wire up all handlers to route group
    - Follow existing routing patterns
  - [x] 3.6 Ensure tenant admin tests pass
    - Run ONLY the 6-8 tests written in 3.1
    - Verify CRUD operations work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 6-8 tests written in 3.1 pass
- All tenant CRUD operations work
- Theme config updates are persisted correctly
- Cross-tenant access works (no tenant_id filtering)
- Asset upload stores and returns URLs

---

#### Task Group 4: Admin User and Hospital Handlers
**Dependencies:** Task Groups 1, 2
**Complexity:** High
**Files to Create/Modify:**
- `/backend/internal/handlers/admin_users.go` (create)
- `/backend/internal/handlers/admin_hospitals.go` (create)
- `/backend/internal/repository/admin_user_repo.go` (create)
- `/backend/internal/repository/admin_hospital_repo.go` (create)
- `/backend/internal/services/auth/impersonate.go` (create)

- [x] 4.0 Complete admin user and hospital handlers
  - [x] 4.1 Write 6-8 focused tests for user/hospital admin endpoints
    - Test GET /admin/users lists all users with tenant filter
    - Test POST /admin/users/:id/impersonate generates temp JWT
    - Test PUT /admin/users/:id/role updates user role
    - Test PUT /admin/users/:id/ban deactivates user
    - Test GET /admin/hospitals lists all hospitals with tenant filter
    - Test PUT /admin/hospitals/:id/reassign changes tenant
  - [x] 4.2 Create AdminUserRepository
    - ListAllUsers (with optional tenant_id filter, pagination)
    - GetUserByID (cross-tenant)
    - UpdateUserRole (including is_super_admin)
    - UpdateUserStatus (active, banned, ban_reason)
    - ResetUserPassword (generate temporary or send email)
  - [x] 4.3 Implement impersonate service in `/backend/internal/services/auth/impersonate.go`
    - GenerateImpersonationToken method
    - Short-lived JWT (e.g., 1 hour)
    - Include original_admin_id in claims for audit
    - Log impersonation action to audit_logs
  - [x] 4.4 Create admin user handlers in `/backend/internal/handlers/admin_users.go`
    - ListUsers: GET /admin/users (tenant filter, search, pagination)
    - GetUser: GET /admin/users/:id
    - ImpersonateUser: POST /admin/users/:id/impersonate
    - UpdateRole: PUT /admin/users/:id/role
    - BanUser: PUT /admin/users/:id/ban
    - ResetPassword: POST /admin/users/:id/reset-password
  - [x] 4.5 Create AdminHospitalRepository
    - ListAllHospitals (with optional tenant_id filter)
    - GetHospitalByID (cross-tenant)
    - UpdateHospital
    - ReassignHospitalTenant (with validation)
  - [x] 4.6 Create admin hospital handlers in `/backend/internal/handlers/admin_hospitals.go`
    - ListHospitals: GET /admin/hospitals (tenant filter, search)
    - GetHospital: GET /admin/hospitals/:id
    - UpdateHospital: PUT /admin/hospitals/:id
    - ReassignTenant: PUT /admin/hospitals/:id/reassign
  - [x] 4.7 Register user and hospital admin routes
    - Wire up all handlers to route group
  - [x] 4.8 Ensure user/hospital admin tests pass
    - Run ONLY the 6-8 tests written in 4.1
    - Verify impersonation generates valid token
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 6-8 tests written in 4.1 pass
- User listing works with tenant filter
- Impersonation generates valid JWT with limited duration
- Hospital reassignment works with proper validation
- All actions are logged to audit_logs

---

#### Task Group 5: Admin Triagem Templates, Settings, and Logs Handlers
**Dependencies:** Task Groups 1, 2
**Complexity:** Medium
**Files to Create/Modify:**
- `/backend/internal/handlers/admin_triagem_templates.go` (create)
- `/backend/internal/handlers/admin_settings.go` (create)
- `/backend/internal/handlers/admin_audit_logs.go` (create)
- `/backend/internal/repository/admin_triagem_repo.go` (create)
- `/backend/internal/repository/admin_settings_repo.go` (create)
- `/backend/internal/services/encryption.go` (create)

- [x] 5.0 Complete admin triagem, settings, and logs handlers
  - [x] 5.1 Write 6-8 focused tests for these admin endpoints
    - Test CRUD for triagem rule templates
    - Test clone template to tenant
    - Test system settings CRUD with encryption
    - Test global audit logs retrieval
    - Test audit logs CSV export
  - [x] 5.2 Create AdminTriagemTemplateRepository
    - ListTemplates (with filters)
    - GetTemplateByID
    - CreateTemplate
    - UpdateTemplate
    - CloneToTenant (copy to tenant's triagem_rules table)
    - GetTemplateUsage (which tenants use it)
  - [x] 5.3 Create admin triagem template handlers
    - ListTemplates: GET /admin/triagem-templates
    - GetTemplate: GET /admin/triagem-templates/:id
    - CreateTemplate: POST /admin/triagem-templates
    - UpdateTemplate: PUT /admin/triagem-templates/:id
    - CloneToTenant: POST /admin/triagem-templates/:id/clone
    - GetUsage: GET /admin/triagem-templates/:id/usage
  - [x] 5.4 Create encryption service in `/backend/internal/services/encryption.go`
    - EncryptValue method (AES-256-GCM)
    - DecryptValue method
    - Use environment variable for encryption key
  - [x] 5.5 Create AdminSettingsRepository
    - GetAllSettings
    - GetSettingByKey
    - UpsertSetting (with optional encryption)
    - DeleteSetting
  - [x] 5.6 Create admin settings handlers
    - ListSettings: GET /admin/settings
    - GetSetting: GET /admin/settings/:key
    - UpsertSetting: PUT /admin/settings/:key
    - DeleteSetting: DELETE /admin/settings/:key
    - Mask encrypted values in responses
  - [x] 5.7 Extend existing audit_logs handler for global access
    - Create GET /admin/logs endpoint
    - Remove tenant_id filter for super admins
    - Add tenant name to response
    - Implement CSV export: GET /admin/logs/export
  - [x] 5.8 Register all remaining admin routes
    - Wire up triagem, settings, and logs handlers
  - [x] 5.9 Ensure admin tests pass
    - Run ONLY the 6-8 tests written in 5.1
    - Verify encryption/decryption works
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 6-8 tests written in 5.1 pass
- Triagem templates can be created and cloned
- Settings are stored with encryption for sensitive values
- Audit logs display tenant names
- CSV export generates valid file

---

### Frontend - Backoffice Application

#### Task Group 6: Admin Layout and Core Pages
**Dependencies:** Task Groups 2, 3
**Complexity:** High
**Files to Create/Modify:**
- `/frontend/src/app/admin/layout.tsx` (create)
- `/frontend/src/app/admin/page.tsx` (create)
- `/frontend/src/components/admin/AdminSidebar.tsx` (create)
- `/frontend/src/components/admin/AdminHeader.tsx` (create)
- `/frontend/src/app/admin/tenants/page.tsx` (create)
- `/frontend/src/app/admin/tenants/[id]/page.tsx` (create)
- `/frontend/src/hooks/useSuperAdmin.ts` (create)
- `/frontend/src/lib/api/admin.ts` (create)

- [x] 6.0 Complete admin layout and tenant management pages
  - [x] 6.1 Write 4-6 focused tests for admin UI components
    - Test AdminSidebar renders all admin navigation items
    - Test useSuperAdmin hook redirects non-super-admins
    - Test tenant list page displays tenants
    - Test tenant edit page loads theme config
  - [x] 6.2 Create useSuperAdmin hook
    - Check if current user has is_super_admin: true
    - Redirect to /dashboard if not super admin
    - Return loading state while checking
  - [x] 6.3 Create admin API client in `/frontend/src/lib/api/admin.ts`
    - Tenant CRUD functions
    - User management functions
    - Hospital functions
    - Settings functions
    - Logs functions
    - Base URL: /api/v1/admin
  - [x] 6.4 Create AdminSidebar component
    - Separate from regular Sidebar
    - Admin-specific navigation items:
      - Dashboard (/admin)
      - Tenants (/admin/tenants)
      - Users (/admin/users)
      - Hospitals (/admin/hospitals)
      - Triagem Templates (/admin/triagem-templates)
      - Settings (/admin/settings)
      - Audit Logs (/admin/logs)
    - Use different accent color to distinguish from tenant app
  - [x] 6.5 Create AdminHeader component
    - Display "SIDOT Admin" branding
    - Show current super admin user
    - Include logout and "Back to App" link
  - [x] 6.6 Create admin layout in `/frontend/src/app/admin/layout.tsx`
    - Use useSuperAdmin for access control
    - Render AdminSidebar and AdminHeader
    - Different styling (e.g., darker theme) to distinguish
  - [x] 6.7 Create admin dashboard page `/frontend/src/app/admin/page.tsx`
    - Global metrics cards:
      - Total tenants (active/inactive)
      - Total users across all tenants
      - Total hospitals
      - Total occurrences
    - System health indicators
    - Recent activity feed
  - [x] 6.8 Create tenant list page `/frontend/src/app/admin/tenants/page.tsx`
    - DataTable with columns: Name, Slug, Status, Users, Hospitals, Actions
    - Search by name/slug
    - Filter by status (active/inactive)
    - Pagination
    - Actions: View/Edit, Toggle Status
  - [x] 6.9 Create tenant editor page `/frontend/src/app/admin/tenants/[id]/page.tsx`
    - Tabs: Details, Theme Editor, Metrics
    - Details tab: Basic info form
    - Theme Editor tab: Placeholder for CMD interface (Task Group 7)
    - Metrics tab: User count, hospital count, occurrence count
  - [x] 6.10 Ensure admin layout tests pass
    - Run ONLY the 4-6 tests written in 6.1
    - Verify navigation works
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 4-6 tests written in 6.1 pass
- Admin layout renders with separate sidebar
- Non-super-admins are redirected
- Tenant list displays data from API
- Tenant editor loads tenant details

---

#### Task Group 7: Command Palette and Theme Editor
**Dependencies:** Task Group 6
**Complexity:** Very High
**Files to Create/Modify:**
- `/frontend/src/components/admin/CommandPalette.tsx` (create)
- `/frontend/src/components/admin/CommandParser.ts` (create)
- `/frontend/src/components/admin/ThemePreview.tsx` (create)
- `/frontend/src/components/admin/TenantThemeEditor.tsx` (create)
- `/frontend/src/types/theme.ts` (create)

- [x] 7.0 Complete command palette and theme editor
  - [x] 7.1 Write 4-6 focused tests for CMD interface
    - Test CommandParser parses color commands correctly
    - Test CommandParser parses sidebar commands correctly
    - Test CommandParser parses dashboard widget commands
    - Test CommandPalette opens on Ctrl+K
    - Test ThemePreview updates on command execution
  - [x] 7.2 Define theme TypeScript types in `/frontend/src/types/theme.ts`
    ```typescript
    interface ThemeConfig {
      theme: {
        colors: { primary: string; background: string; ... };
        fonts: { body: string; heading?: string };
      };
      layout: {
        sidebar: SidebarItem[];
        topbar: TopbarConfig;
        dashboard_widgets: DashboardWidget[];
      };
    }
    ```
  - [x] 7.3 Create CommandParser utility in `/frontend/src/components/admin/CommandParser.ts`
    - Parse command strings into actions
    - Support commands:
      - `Set Primary Color #HEXCODE`
      - `Set Background #HEXCODE`
      - `Set Font "FontName"`
      - `Sidebar: Add Item "Label" icon="Icon" link="/path"`
      - `Sidebar: Remove "Label"`
      - `Sidebar: Move "Label" to Top/Bottom`
      - `Dashboard: Add Widget "type"`
      - `Dashboard: Hide "widget_id"`
      - `Dashboard: Show "widget_id"`
      - `Upload Logo`
      - `Upload Favicon`
    - Return structured action object
    - Provide command suggestions based on partial input
  - [x] 7.4 Create CommandPalette component using cmdk library
    - Install cmdk: `npm install cmdk`
    - Ctrl+K keyboard shortcut to open
    - Command input with auto-complete
    - Command categories (Colors, Sidebar, Dashboard, Assets)
    - Execute command on Enter
    - Display command history
    - Real-time validation of commands
  - [x] 7.5 Create ThemePreview component
    - Live preview panel showing tenant UI mockup
    - Render sidebar items from config
    - Apply color variables dynamically
    - Show dashboard widget layout
    - Display logo/favicon if configured
    - Update preview instantly when theme_config changes
  - [x] 7.6 Create TenantThemeEditor component
    - Split-pane layout: CMD input on left, Preview on right
    - Manage theme_config state
    - Handle command execution and state updates
    - Save button to persist changes via API
    - Reset button to revert to last saved
    - Undo/Redo support (optional)
  - [x] 7.7 Integrate asset upload in CommandPalette
    - `Upload Logo` triggers file picker
    - `Upload Favicon` triggers file picker
    - Call API to upload and update tenant
    - Show upload progress
  - [x] 7.8 Update tenant editor page to use TenantThemeEditor
    - Replace placeholder in Theme Editor tab
    - Pass tenant theme_config as initial state
    - Handle save and refresh
  - [x] 7.9 Ensure CMD tests pass
    - Run ONLY the 4-6 tests written in 7.1
    - Verify command parsing works correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 4-6 tests written in 7.1 pass
- Ctrl+K opens command palette
- Commands are parsed and applied to preview
- Preview updates in real-time
- Changes can be saved to backend
- Asset uploads work correctly

---

#### Task Group 8: Remaining Admin Pages
**Dependencies:** Task Groups 4, 5, 6
**Complexity:** Medium
**Files to Create/Modify:**
- `/frontend/src/app/admin/users/page.tsx` (create)
- `/frontend/src/app/admin/hospitals/page.tsx` (create)
- `/frontend/src/app/admin/triagem-templates/page.tsx` (create)
- `/frontend/src/app/admin/settings/page.tsx` (create)
- `/frontend/src/app/admin/logs/page.tsx` (create)
- `/frontend/src/components/admin/ImpersonateDialog.tsx` (create)
- `/frontend/src/components/admin/CloneTemplateDialog.tsx` (create)

- [x] 8.0 Complete remaining admin pages
  - [x] 8.1 Write 4-6 focused tests for remaining admin pages
    - Test user list page with tenant filter
    - Test impersonate dialog generates token
    - Test settings page masks encrypted values
    - Test audit logs export
  - [x] 8.2 Create user management page `/frontend/src/app/admin/users/page.tsx`
    - DataTable: Email, Name, Role, Tenant, Status, Actions
    - Tenant filter dropdown
    - Search by email/name
    - Actions: View, Impersonate, Edit Role, Ban, Reset Password
  - [x] 8.3 Create ImpersonateDialog component
    - Confirmation dialog with user details
    - Warning about action being logged
    - Generate token and open new tab as impersonated user
    - "Stop Impersonating" option visible in app
  - [x] 8.4 Create hospital management page `/frontend/src/app/admin/hospitals/page.tsx`
    - DataTable: Name, City, State, Tenant, Actions
    - Tenant filter dropdown
    - Search by name/city
    - View/Edit modal
    - Reassign tenant with confirmation
  - [x] 8.5 Create triagem templates page `/frontend/src/app/admin/triagem-templates/page.tsx`
    - DataTable: Nome, Tipo, Status, Tenants Using, Actions
    - Create/Edit form (Nome, Tipo, Condicao JSON)
    - Activate/Deactivate toggle
    - Clone to tenant dialog
  - [x] 8.6 Create CloneTemplateDialog component
    - Multi-select tenants dropdown
    - Preview which tenants will receive the rule
    - Confirm and clone
  - [x] 8.7 Create global settings page `/frontend/src/app/admin/settings/page.tsx`
    - Sections: Email (SMTP), SMS (Twilio), Push (FCM)
    - Form fields for each configuration
    - Masked display for sensitive fields
    - Save per section or all at once
    - Test connection buttons (optional)
  - [x] 8.8 Create audit logs page `/frontend/src/app/admin/logs/page.tsx`
    - DataTable: Timestamp, Tenant, User, Action, Entity, Severity
    - Filters: Tenant, Date Range, User, Action, Severity
    - Tenant name column (not just ID)
    - Export CSV button
    - Pagination for large datasets
  - [x] 8.9 Ensure admin pages tests pass
    - Run ONLY the 4-6 tests written in 8.1
    - Verify pages render and fetch data
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 4-6 tests written in 8.1 pass
- User management works with impersonate
- Hospital reassignment works
- Triagem templates can be cloned
- Settings form saves correctly
- Audit logs display with export option

---

### Frontend - Dynamic Tenant UI

#### Task Group 9: Dynamic Tenant Theme System
**Dependencies:** Task Groups 3, 6
**Complexity:** High
**Files to Create/Modify:**
- `/frontend/src/contexts/TenantThemeContext.tsx` (create)
- `/frontend/src/components/layout/DynamicSidebar.tsx` (create)
- `/frontend/src/components/dashboard/DynamicDashboard.tsx` (create)
- `/frontend/src/lib/dynamicIcon.tsx` (create)
- `/frontend/src/app/dashboard/layout.tsx` (modify)
- `/frontend/src/hooks/useTenantTheme.ts` (create)

- [x] 9.0 Complete dynamic tenant theme system
  - [x] 9.1 Write 4-6 focused tests for dynamic theme components
    - Test TenantThemeProvider injects CSS variables
    - Test DynamicSidebar renders items from config
    - Test DynamicDashboard renders widgets
    - Test dynamicIcon resolves Lucide icons by name
  - [x] 9.2 Create useTenantTheme hook
    - Fetch tenant theme_config from API
    - Cache in React Query or SWR
    - Return theme_config with loading state
    - Provide fallback default config
  - [x] 9.3 Create TenantThemeContext and Provider
    - Store theme_config in context
    - Inject CSS custom properties into document root
    - Handle light/dark mode with custom colors
    - Provide theme values to child components
    ```typescript
    // Example CSS variable injection
    document.documentElement.style.setProperty('--primary', theme.colors.primary);
    ```
  - [x] 9.4 Create dynamicIcon utility in `/frontend/src/lib/dynamicIcon.tsx`
    - Map string icon names to Lucide React components
    - Support common icons: Home, Settings, Users, Building2, etc.
    - Return fallback icon for unknown names
    ```typescript
    const iconMap = {
      Home: HomeIcon,
      Settings: SettingsIcon,
      // ...
    };
    export function DynamicIcon({ name }: { name: string }) {
      const Icon = iconMap[name] || CircleIcon;
      return <Icon />;
    }
    ```
  - [x] 9.5 Create DynamicSidebar component
    - Read sidebar items from theme_config.layout.sidebar
    - Render each item with DynamicIcon
    - Support role-based visibility (item.roles)
    - Maintain active state detection
    - Fallback to static sidebar if no config
  - [x] 9.6 Create DynamicDashboard component
    - Read widgets from theme_config.layout.dashboard_widgets
    - Render widget grid based on order and visibility
    - Support widget types: stats_card, map_preview, recent_occurrences, chart
    - Conditional rendering based on visible flag
    - Fallback to default dashboard if no config
  - [x] 9.7 Update dashboard layout to use dynamic components
    - Wrap with TenantThemeProvider
    - Replace static Sidebar with DynamicSidebar
    - Update dashboard page to use DynamicDashboard
  - [x] 9.8 Ensure dynamic theme tests pass
    - Run ONLY the 4-6 tests written in 9.1
    - Verify CSS variables are injected
    - Verify sidebar renders dynamic items
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 4-6 tests written in 9.1 pass
- CSS variables update based on theme_config
- Sidebar renders items from JSONB config
- Dashboard shows/hides widgets based on config
- Fallback to defaults when no config exists

---

### Testing

#### Task Group 10: Test Review and Gap Analysis
**Dependencies:** Task Groups 1-9
**Complexity:** Medium

- [x] 10.0 Review existing tests and fill critical gaps only
  - [x] 10.1 Review tests from all Task Groups
    - Review the 4-6 tests written by database engineer (Task 1.1)
    - Review the 4-6 tests written for middleware (Task 2.1)
    - Review the 6-8 tests written for tenant handlers (Task 3.1)
    - Review the 6-8 tests written for user/hospital handlers (Task 4.1)
    - Review the 6-8 tests written for templates/settings/logs (Task 5.1)
    - Review the 4-6 tests written for admin layout (Task 6.1)
    - Review the 4-6 tests written for CMD interface (Task 7.1)
    - Review the 4-6 tests written for remaining pages (Task 8.1)
    - Review the 4-6 tests written for dynamic theme (Task 9.1)
    - Total existing tests: approximately 44-60 tests
  - [x] 10.2 Analyze test coverage gaps for THIS feature only
    - Identify critical end-to-end workflows lacking coverage
    - Focus ONLY on gaps related to backoffice feature requirements
    - Do NOT assess entire application test coverage
    - Prioritize:
      1. Super admin authentication flow
      2. Theme config save and apply flow
      3. Impersonation security
      4. Cross-tenant data access
  - [x] 10.3 Write up to 10 additional strategic tests maximum
    - End-to-end: Super admin login -> Edit tenant theme -> Verify tenant sees changes
    - End-to-end: Clone triagem template to multiple tenants
    - Integration: Impersonate user and verify limited access
    - Integration: Encrypt/decrypt system settings
    - Security: Verify non-super-admin cannot access /admin routes
    - Do NOT write comprehensive coverage for all scenarios
  - [x] 10.4 Run feature-specific tests only
    - Run ONLY tests related to backoffice feature
    - Expected total: approximately 54-70 tests maximum
    - Do NOT run the entire application test suite
    - Verify critical workflows pass

**Acceptance Criteria:**
- All feature-specific tests pass (approximately 54-70 tests total)
- Critical super admin workflows are covered
- No more than 10 additional tests added when filling gaps
- Testing focused exclusively on backoffice feature requirements

---

## Execution Order

Recommended implementation sequence:

```
Phase 1: Foundation (Backend)
  1. Task Group 1: Database Migrations and Models
  2. Task Group 2: SuperAdmin Middleware and Routing

Phase 2: Backend API
  3. Task Group 3: Admin Tenant Handlers
  4. Task Group 4: Admin User and Hospital Handlers
  5. Task Group 5: Admin Triagem, Settings, and Logs Handlers

Phase 3: Frontend Admin
  6. Task Group 6: Admin Layout and Core Pages
  7. Task Group 7: Command Palette and Theme Editor
  8. Task Group 8: Remaining Admin Pages

Phase 4: Dynamic Tenant UI
  9. Task Group 9: Dynamic Tenant Theme System

Phase 5: Validation
  10. Task Group 10: Test Review and Gap Analysis
```

---

## Dependencies Diagram

```
[Task Group 1: Database]
        |
        v
[Task Group 2: Middleware] -----> [Task Group 3: Tenant Handlers]
        |                                    |
        |                                    v
        +---> [Task Group 4: User/Hospital] [Task Group 6: Admin Layout]
        |                                    |
        +---> [Task Group 5: Templates/Settings/Logs]
                                             |
                                             v
                                    [Task Group 7: CMD Interface]
                                             |
                                             v
                                    [Task Group 8: Remaining Pages]
                                             |
                                             v
                                    [Task Group 9: Dynamic Theme]
                                             |
                                             v
                                    [Task Group 10: Test Review]
```

---

## Notes

### Technical Considerations

1. **JSONB Performance**: Add GIN indexes on theme_config for efficient queries
2. **Encryption**: Use AES-256-GCM for system settings; store encryption key in environment
3. **Impersonation Security**:
   - Short-lived tokens (1 hour max)
   - Log all impersonation actions
   - Include original_admin_id in claims for audit
4. **Theme Application**: Use CSS custom properties for instant updates without reload
5. **Command Parsing**: Build robust parser with clear error messages for invalid commands

### Existing Patterns to Follow

- **Backend**: Follow existing handler patterns in `/backend/internal/handlers/`
- **Models**: Follow ToResponse pattern from `tenant.go` and `audit_log.go`
- **Middleware**: Follow RequireRole pattern from `auth.go`
- **Frontend Layout**: Follow existing Sidebar.tsx patterns
- **CSS Variables**: Follow existing globals.css structure

### Out of Scope (Do Not Implement)

- Direct editing of patient clinical data
- Changing occurrence statuses from backoffice
- White-label domain configuration
- Billing/subscription management
- API rate limiting per tenant
- Custom CSS injection beyond predefined variables
- Multi-language/i18n configuration
- Tenant data export/import
- Automated tenant provisioning
- Real-time collaboration
