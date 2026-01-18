# Specification: Backoffice Administrativo Completo

## Goal
Create a comprehensive super-admin backoffice interface with a CMD (Command Palette) for real-time tenant UI customization, cross-tenant management capabilities, and dynamic frontend rendering based on JSONB theme configurations.

## User Stories
- As a super admin, I want to customize each tenant's UI through a command palette so that I can quickly configure branding and layout without editing code
- As a super admin, I want to manage users, hospitals, and settings across all tenants so that I can maintain the entire platform from a single interface

## Specific Requirements

**CMD Interface (Command Palette)**
- Implement Ctrl+K keyboard shortcut to open command palette using cmdk library
- Support color commands: `> Set Primary Color #HEXCODE`, `> Set Background #HEXCODE`
- Support sidebar commands: `> Sidebar: Add Item "Label" icon="IconName" link="/path"`, `> Sidebar: Remove "Label"`, `> Sidebar: Move "Label" to Top/Bottom`
- Support dashboard commands: `> Dashboard: Add Widget "widget_type"`, `> Dashboard: Hide "widget_id"`, `> Dashboard: Show "widget_id"`
- Support asset commands: `> Upload Logo`, `> Upload Favicon` (trigger file picker)
- Support typography: `> Set Font "FontFamily"`
- Implement real-time preview panel showing changes before saving
- Auto-complete suggestions based on command context

**Dynamic Theme Configuration (JSONB)**
- Store complete UI configuration in `theme_config` JSONB column on tenants table
- Structure: `{ theme: { colors, fonts }, layout: { sidebar[], topbar{}, dashboard_widgets[] } }`
- Sidebar items: `{ label: string, icon: string, link: string, roles?: string[] }`
- Dashboard widgets: `{ type: string, visible: boolean, order: number, config?: object }`
- Topbar config: `{ show_user_info: boolean, show_tenant_logo: boolean }`
- Default theme_config seeded for new tenants

**Dynamic Frontend Components (Tenant App)**
- DynamicSidebar component that reads `layout.sidebar` array and renders navigation items dynamically
- DynamicDashboard component with configurable widget grid based on `dashboard_widgets` array
- ThemeProvider context that injects CSS custom properties from `theme.colors` into `:root`
- Dynamic icon resolution using Lucide icons by string name
- Fallback to default theme if tenant has no theme_config

**Super Admin Middleware and Routing**
- Create RequireSuperAdmin middleware checking `is_super_admin = true` on user claims
- All admin routes under `/api/v1/admin/*` protected by this middleware
- Admin queries ignore tenant_id (cross-tenant access)
- Frontend routes under `/admin/*` with client-side super_admin check

**Tenant Management (/admin/tenants)**
- List all tenants with pagination, search, and status filter (active/inactive)
- Full CRUD: Create, Read, Update, Delete (soft-delete via is_active flag)
- Visual Editor page with CMD interface and live preview
- Toggle tenant active/inactive status
- Display tenant metrics: user count, hospital count, occurrence count

**User Management (/admin/users)**
- List all users globally with tenant filter dropdown
- Impersonate feature: Generate temporary JWT to "login as" any user for testing
- Promote/demote user roles including super_admin flag
- Reset password (generate temporary password or send reset email)
- Ban/Deactivate users with reason logging

**Hospital Management (/admin/hospitals)**
- List all hospitals globally with tenant filter
- View/Edit hospital details including coordinates
- Reassign hospital to different tenant (with confirmation)

**Triagem Rule Templates (/admin/triagem-templates)**
- Create master triagem rules (not tied to any tenant)
- Clone template to one or multiple tenants
- Edit/deactivate templates
- View which tenants are using each template

**Global Settings (/admin/settings)**
- SMTP/SendGrid configuration: host, port, user, password, from address
- SMS/Twilio configuration: account_sid, auth_token, from_number
- FCM Push configuration: server_key
- Settings stored in system_settings table as key-value JSONB pairs
- Encrypt sensitive values before storage

**Audit Logs Global View (/admin/logs)**
- Read-only view of all audit_logs across tenants
- Filter by: tenant, date range, user, action type, severity
- Export to CSV functionality
- Display tenant name alongside each log entry
- Clinical data (occurrence details) visible but NOT editable

## Visual Design
No visual mockups provided for this specification.

## Existing Code to Leverage

**`/backend/internal/middleware/auth.go`**
- Existing AuthRequired middleware pattern for JWT validation
- RequireRole middleware pattern to replicate for RequireSuperAdmin
- UserClaims struct already contains IsSuperAdmin field
- GetUserClaims helper function for extracting claims from context

**`/backend/internal/models/tenant.go`**
- Existing Tenant model with ID, Name, Slug, timestamps
- CreateTenantInput and TenantResponse patterns to extend
- ValidateSlug function for slug validation
- ToResponse pattern for API responses

**`/backend/internal/models/audit_log.go`**
- AuditLog model with TenantID (nullable for cross-tenant)
- AuditLogFilter struct for query parameters
- Severity levels (INFO, WARN, CRITICAL)
- Action constants pattern to extend for admin actions

**`/frontend/src/components/layout/Sidebar.tsx`**
- Existing static sidebar with NavItem interface
- Role-based filtering pattern (item.roles)
- Active state detection using usePathname
- Lucide icon components already imported

**`/frontend/src/app/globals.css`**
- CSS custom properties structure for theming
- Both light and dark mode variables
- Theme token naming convention (--primary, --sidebar-*, etc.)
- Pattern for injecting dynamic values into :root

## Out of Scope
- Direct editing of patient clinical data (occurrences, medical info)
- Changing occurrence statuses from backoffice
- White-label domain configuration (custom domains per tenant)
- Billing/subscription management
- API rate limiting configuration per tenant
- Custom CSS injection beyond predefined variables
- Multi-language/i18n configuration
- Tenant data export/import functionality
- Automated tenant provisioning workflows
- Real-time collaboration (multiple admins editing same tenant)
