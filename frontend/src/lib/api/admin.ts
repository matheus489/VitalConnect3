import { api } from '@/lib/api';

// Re-export theme types from centralized location
export type {
  ThemeConfig,
  ThemeColors,
  ThemeFonts,
  ThemeAppearance,
  ThemeLayout,
  SidebarItem,
  TopbarConfig,
  DashboardWidget,
  DashboardWidgetType,
} from '@/types/theme';

// Import for internal use
import type { ThemeConfig } from '@/types/theme';

// =============================================================================
// Admin Types
// =============================================================================

export interface AdminTenant {
  id: string;
  name: string;
  slug: string;
  is_active: boolean;
  logo_url?: string;
  favicon_url?: string;
  theme_config: ThemeConfig;
  created_at: string;
  updated_at: string;
}

export interface AdminTenantWithMetrics extends AdminTenant {
  user_count: number;
  hospital_count: number;
  occurrence_count: number;
}

export interface AdminUser {
  id: string;
  email: string;
  nome: string;
  role: string;
  tenant_id: string;
  tenant_name?: string;
  is_super_admin: boolean;
  is_active: boolean;
  banned_at?: string;
  ban_reason?: string;
  created_at: string;
  updated_at: string;
}

export interface AdminHospital {
  id: string;
  nome: string;
  codigo: string;
  endereco?: string;
  cidade?: string;
  estado?: string;
  latitude?: number;
  longitude?: number;
  tenant_id: string;
  tenant_name?: string;
  ativo: boolean;
  created_at: string;
  updated_at: string;
}

export interface AdminTriagemTemplate {
  id: string;
  nome: string;
  tipo: string;
  condicao: Record<string, unknown>;
  descricao?: string;
  ativo: boolean;
  tenant_usage_count?: number;
  created_at: string;
  updated_at: string;
}

export interface AdminSystemSetting {
  id: string;
  key: string;
  value: Record<string, unknown>;
  description?: string;
  is_encrypted: boolean;
  created_at: string;
  updated_at: string;
}

export interface AdminAuditLog {
  id: string;
  timestamp: string;
  usuario_id?: string;
  actor_name: string;
  tenant_id?: string;
  tenant_name?: string;
  acao: string;
  entidade_tipo: string;
  entidade_id: string;
  severity: 'INFO' | 'WARN' | 'CRITICAL';
  detalhes?: Record<string, unknown>;
  ip_address?: string;
}

export interface AdminDashboardMetrics {
  total_tenants: number;
  active_tenants: number;
  inactive_tenants: number;
  total_users: number;
  total_hospitals: number;
  total_occurrences: number;
}

// Raw response from backend (flat structure)
export interface PaginatedAdminResponseRaw<T> {
  data: T[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

// Normalized response for frontend use
export interface PaginatedAdminResponse<T> {
  data: T[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

// Helper to normalize backend response to frontend format
function normalizePaginatedResponse<T>(raw: PaginatedAdminResponseRaw<T>): PaginatedAdminResponse<T> {
  return {
    data: raw.data,
    meta: {
      page: raw.page,
      per_page: raw.per_page,
      total: raw.total,
      total_pages: raw.total_pages,
    },
  };
}

// =============================================================================
// Admin API Functions - Tenants
// =============================================================================

const ADMIN_BASE = '/admin';

export async function fetchAdminTenants(params?: {
  page?: number;
  per_page?: number;
  search?: string;
  status?: 'active' | 'inactive' | 'all';
}): Promise<PaginatedAdminResponse<AdminTenant>> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', String(params.page));
  if (params?.per_page) queryParams.append('per_page', String(params.per_page));
  if (params?.search) queryParams.append('search', params.search);
  if (params?.status && params.status !== 'all') {
    queryParams.append('is_active', params.status === 'active' ? 'true' : 'false');
  }

  const { data } = await api.get<PaginatedAdminResponseRaw<AdminTenant>>(
    `${ADMIN_BASE}/tenants?${queryParams.toString()}`
  );
  return normalizePaginatedResponse(data);
}

export async function fetchAdminTenant(id: string): Promise<AdminTenantWithMetrics> {
  const { data } = await api.get<AdminTenantWithMetrics>(`${ADMIN_BASE}/tenants/${id}`);
  return data;
}

export async function createAdminTenant(input: {
  name: string;
  slug: string;
}): Promise<AdminTenant> {
  const { data } = await api.post<AdminTenant>(`${ADMIN_BASE}/tenants`, input);
  return data;
}

export async function updateAdminTenant(
  id: string,
  input: Partial<{ name: string; slug: string }>
): Promise<AdminTenant> {
  const { data } = await api.put<AdminTenant>(`${ADMIN_BASE}/tenants/${id}`, input);
  return data;
}

export async function updateAdminTenantTheme(
  id: string,
  themeConfig: ThemeConfig
): Promise<AdminTenant> {
  const { data } = await api.put<AdminTenant>(`${ADMIN_BASE}/tenants/${id}/theme`, {
    theme_config: themeConfig,
  });
  return data;
}

export async function toggleAdminTenantStatus(id: string): Promise<AdminTenant> {
  const { data } = await api.put<{ tenant: AdminTenant; is_active: boolean; message: string }>(
    `${ADMIN_BASE}/tenants/${id}/toggle`
  );
  return data.tenant;
}

export async function uploadTenantAssets(
  id: string,
  formData: FormData
): Promise<AdminTenant> {
  // Don't set Content-Type header - let axios/browser set it with boundary
  const { data } = await api.post<{ message: string; tenant: AdminTenant }>(
    `${ADMIN_BASE}/tenants/${id}/assets`,
    formData
  );
  return data.tenant;
}

// =============================================================================
// Admin API Functions - Users
// =============================================================================

export async function fetchAdminUsers(params?: {
  page?: number;
  per_page?: number;
  search?: string;
  tenant_id?: string;
  role?: string;
}): Promise<PaginatedAdminResponse<AdminUser>> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', String(params.page));
  if (params?.per_page) queryParams.append('per_page', String(params.per_page));
  if (params?.search) queryParams.append('search', params.search);
  if (params?.tenant_id) queryParams.append('tenant_id', params.tenant_id);
  if (params?.role) queryParams.append('role', params.role);

  const { data } = await api.get<PaginatedAdminResponseRaw<AdminUser>>(
    `${ADMIN_BASE}/users?${queryParams.toString()}`
  );
  return normalizePaginatedResponse(data);
}

export async function fetchAdminUser(id: string): Promise<AdminUser> {
  const { data } = await api.get<AdminUser>(`${ADMIN_BASE}/users/${id}`);
  return data;
}

export async function impersonateUser(id: string): Promise<{
  access_token: string;
  expires_at: string;
}> {
  const { data } = await api.post<{
    access_token: string;
    expires_at: string;
  }>(`${ADMIN_BASE}/users/${id}/impersonate`);
  return data;
}

export async function updateAdminUserRole(
  id: string,
  input: { role: string; is_super_admin?: boolean }
): Promise<AdminUser> {
  const { data } = await api.put<AdminUser>(`${ADMIN_BASE}/users/${id}/role`, input);
  return data;
}

export async function banAdminUser(
  id: string,
  input: { reason: string }
): Promise<AdminUser> {
  const { data } = await api.put<AdminUser>(`${ADMIN_BASE}/users/${id}/ban`, input);
  return data;
}

export async function resetAdminUserPassword(id: string): Promise<{
  message: string;
  temporary_password?: string;
}> {
  const { data } = await api.post<{
    message: string;
    temporary_password?: string;
  }>(`${ADMIN_BASE}/users/${id}/reset-password`);
  return data;
}

// =============================================================================
// Admin API Functions - Hospitals
// =============================================================================

export async function fetchAdminHospitals(params?: {
  page?: number;
  per_page?: number;
  search?: string;
  tenant_id?: string;
}): Promise<PaginatedAdminResponse<AdminHospital>> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', String(params.page));
  if (params?.per_page) queryParams.append('per_page', String(params.per_page));
  if (params?.search) queryParams.append('search', params.search);
  if (params?.tenant_id) queryParams.append('tenant_id', params.tenant_id);

  const { data } = await api.get<PaginatedAdminResponseRaw<AdminHospital>>(
    `${ADMIN_BASE}/hospitals?${queryParams.toString()}`
  );
  return normalizePaginatedResponse(data);
}

export async function fetchAdminHospital(id: string): Promise<AdminHospital> {
  const { data } = await api.get<AdminHospital>(`${ADMIN_BASE}/hospitals/${id}`);
  return data;
}

export async function updateAdminHospital(
  id: string,
  input: Partial<AdminHospital>
): Promise<AdminHospital> {
  const { data } = await api.put<AdminHospital>(
    `${ADMIN_BASE}/hospitals/${id}`,
    input
  );
  return data;
}

export async function reassignHospitalTenant(
  id: string,
  tenantId: string
): Promise<AdminHospital> {
  const { data } = await api.put<AdminHospital>(
    `${ADMIN_BASE}/hospitals/${id}/reassign`,
    { tenant_id: tenantId }
  );
  return data;
}

// =============================================================================
// Admin API Functions - Triagem Templates
// =============================================================================

export async function fetchAdminTriagemTemplates(params?: {
  page?: number;
  per_page?: number;
  tipo?: string;
  ativo?: boolean;
}): Promise<PaginatedAdminResponse<AdminTriagemTemplate>> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', String(params.page));
  if (params?.per_page) queryParams.append('per_page', String(params.per_page));
  if (params?.tipo) queryParams.append('tipo', params.tipo);
  if (params?.ativo !== undefined) {
    queryParams.append('ativo', String(params.ativo));
  }

  const { data } = await api.get<PaginatedAdminResponseRaw<AdminTriagemTemplate>>(
    `${ADMIN_BASE}/triagem-templates?${queryParams.toString()}`
  );
  return normalizePaginatedResponse(data);
}

export async function fetchAdminTriagemTemplate(
  id: string
): Promise<AdminTriagemTemplate> {
  const { data } = await api.get<AdminTriagemTemplate>(
    `${ADMIN_BASE}/triagem-templates/${id}`
  );
  return data;
}

export async function createAdminTriagemTemplate(input: {
  nome: string;
  tipo: string;
  condicao: Record<string, unknown>;
  descricao?: string;
}): Promise<AdminTriagemTemplate> {
  const { data } = await api.post<AdminTriagemTemplate>(
    `${ADMIN_BASE}/triagem-templates`,
    input
  );
  return data;
}

export async function updateAdminTriagemTemplate(
  id: string,
  input: Partial<{
    nome: string;
    tipo: string;
    condicao: Record<string, unknown>;
    descricao: string;
    ativo: boolean;
  }>
): Promise<AdminTriagemTemplate> {
  const { data } = await api.put<AdminTriagemTemplate>(
    `${ADMIN_BASE}/triagem-templates/${id}`,
    input
  );
  return data;
}

export async function cloneTriagemTemplateToTenants(
  id: string,
  tenantIds: string[]
): Promise<{ cloned_count: number }> {
  const { data } = await api.post<{ cloned_count: number }>(
    `${ADMIN_BASE}/triagem-templates/${id}/clone`,
    { tenant_ids: tenantIds }
  );
  return data;
}

export async function fetchTriagemTemplateUsage(
  id: string
): Promise<{ tenants: { id: string; name: string }[] }> {
  const { data } = await api.get<{ tenants: { id: string; name: string }[] }>(
    `${ADMIN_BASE}/triagem-templates/${id}/usage`
  );
  return data;
}

// =============================================================================
// Admin API Functions - System Settings
// =============================================================================

export async function fetchAdminSettings(): Promise<AdminSystemSetting[]> {
  const { data } = await api.get<{ data: AdminSystemSetting[] }>(
    `${ADMIN_BASE}/settings`
  );
  return data.data;
}

export async function fetchAdminSetting(key: string): Promise<AdminSystemSetting> {
  const { data } = await api.get<AdminSystemSetting>(
    `${ADMIN_BASE}/settings/${key}`
  );
  return data;
}

export async function upsertAdminSetting(
  key: string,
  input: {
    value: Record<string, unknown>;
    description?: string;
    is_encrypted?: boolean;
  }
): Promise<AdminSystemSetting> {
  const { data } = await api.put<AdminSystemSetting>(
    `${ADMIN_BASE}/settings/${key}`,
    input
  );
  return data;
}

export async function deleteAdminSetting(key: string): Promise<void> {
  await api.delete(`${ADMIN_BASE}/settings/${key}`);
}

// =============================================================================
// Admin API Functions - Audit Logs
// =============================================================================

export async function fetchAdminAuditLogs(params?: {
  page?: number;
  per_page?: number;
  tenant_id?: string;
  usuario_id?: string;
  acao?: string;
  severity?: 'INFO' | 'WARN' | 'CRITICAL';
  data_inicio?: string;
  data_fim?: string;
}): Promise<PaginatedAdminResponse<AdminAuditLog>> {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.append('page', String(params.page));
  if (params?.per_page) queryParams.append('per_page', String(params.per_page));
  if (params?.tenant_id) queryParams.append('tenant_id', params.tenant_id);
  if (params?.usuario_id) queryParams.append('usuario_id', params.usuario_id);
  if (params?.acao) queryParams.append('acao', params.acao);
  if (params?.severity) queryParams.append('severity', params.severity);
  if (params?.data_inicio) queryParams.append('data_inicio', params.data_inicio);
  if (params?.data_fim) queryParams.append('data_fim', params.data_fim);

  const { data } = await api.get<PaginatedAdminResponseRaw<AdminAuditLog>>(
    `${ADMIN_BASE}/logs?${queryParams.toString()}`
  );
  return normalizePaginatedResponse(data);
}

export async function exportAdminAuditLogsCSV(params?: {
  tenant_id?: string;
  usuario_id?: string;
  acao?: string;
  severity?: 'INFO' | 'WARN' | 'CRITICAL';
  data_inicio?: string;
  data_fim?: string;
}): Promise<Blob> {
  const queryParams = new URLSearchParams();
  if (params?.tenant_id) queryParams.append('tenant_id', params.tenant_id);
  if (params?.usuario_id) queryParams.append('usuario_id', params.usuario_id);
  if (params?.acao) queryParams.append('acao', params.acao);
  if (params?.severity) queryParams.append('severity', params.severity);
  if (params?.data_inicio) queryParams.append('data_inicio', params.data_inicio);
  if (params?.data_fim) queryParams.append('data_fim', params.data_fim);

  const { data } = await api.get<Blob>(
    `${ADMIN_BASE}/logs/export?${queryParams.toString()}`,
    { responseType: 'blob' }
  );
  return data;
}

// =============================================================================
// Admin API Functions - Dashboard Metrics
// =============================================================================

export async function fetchAdminDashboardMetrics(): Promise<AdminDashboardMetrics> {
  const { data } = await api.get<AdminDashboardMetrics>(`${ADMIN_BASE}/metrics`);
  return data;
}

export default {
  // Tenants
  fetchAdminTenants,
  fetchAdminTenant,
  createAdminTenant,
  updateAdminTenant,
  updateAdminTenantTheme,
  toggleAdminTenantStatus,
  uploadTenantAssets,
  // Users
  fetchAdminUsers,
  fetchAdminUser,
  impersonateUser,
  updateAdminUserRole,
  banAdminUser,
  resetAdminUserPassword,
  // Hospitals
  fetchAdminHospitals,
  fetchAdminHospital,
  updateAdminHospital,
  reassignHospitalTenant,
  // Triagem Templates
  fetchAdminTriagemTemplates,
  fetchAdminTriagemTemplate,
  createAdminTriagemTemplate,
  updateAdminTriagemTemplate,
  cloneTriagemTemplateToTenants,
  fetchTriagemTemplateUsage,
  // Settings
  fetchAdminSettings,
  fetchAdminSetting,
  upsertAdminSetting,
  deleteAdminSetting,
  // Audit Logs
  fetchAdminAuditLogs,
  exportAdminAuditLogsCSV,
  // Dashboard
  fetchAdminDashboardMetrics,
};
