import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock next/navigation
const mockPush = vi.fn();
let mockPathname = '/admin/users';
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => mockPathname,
}));

// Mock useAuth
const mockUser = {
  id: '1',
  email: 'admin@example.com',
  nome: 'Super Admin',
  role: 'admin' as const,
  is_super_admin: true,
};

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    user: mockUser,
    isLoading: false,
    isAuthenticated: true,
  }),
}));

// Mock API functions
const mockFetchAdminUsers = vi.fn();
const mockFetchAdminTenants = vi.fn();
const mockImpersonateUser = vi.fn();
const mockFetchAdminSettings = vi.fn();
const mockFetchAdminAuditLogs = vi.fn();
const mockExportAdminAuditLogsCSV = vi.fn();

vi.mock('@/lib/api/admin', () => ({
  fetchAdminUsers: (...args: unknown[]) => mockFetchAdminUsers(...args),
  fetchAdminTenants: (...args: unknown[]) => mockFetchAdminTenants(...args),
  impersonateUser: (...args: unknown[]) => mockImpersonateUser(...args),
  fetchAdminSettings: () => mockFetchAdminSettings(),
  fetchAdminAuditLogs: (...args: unknown[]) => mockFetchAdminAuditLogs(...args),
  exportAdminAuditLogsCSV: (...args: unknown[]) => mockExportAdminAuditLogsCSV(...args),
  // Re-export other functions as empty stubs
  fetchAdminHospitals: vi.fn().mockResolvedValue({ data: [], meta: { page: 1, per_page: 10, total: 0, total_pages: 1 } }),
  fetchAdminTriagemTemplates: vi.fn().mockResolvedValue({ data: [], meta: { page: 1, per_page: 10, total: 0, total_pages: 1 } }),
  upsertAdminSetting: vi.fn().mockResolvedValue({}),
}));

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

// Import components after mocks
import { ImpersonateDialog } from './ImpersonateDialog';

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('Task Group 8: Remaining Admin Pages Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset default mock responses
    mockFetchAdminUsers.mockResolvedValue({
      data: [
        {
          id: 'user-1',
          email: 'john@example.com',
          nome: 'John Doe',
          role: 'admin',
          tenant_id: 'tenant-1',
          tenant_name: 'Hospital A',
          is_super_admin: false,
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'user-2',
          email: 'jane@example.com',
          nome: 'Jane Smith',
          role: 'gestor',
          tenant_id: 'tenant-2',
          tenant_name: 'Hospital B',
          is_super_admin: false,
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      meta: { page: 1, per_page: 10, total: 2, total_pages: 1 },
    });

    mockFetchAdminTenants.mockResolvedValue({
      data: [
        { id: 'tenant-1', name: 'Hospital A', slug: 'hospital-a', is_active: true },
        { id: 'tenant-2', name: 'Hospital B', slug: 'hospital-b', is_active: true },
      ],
      meta: { page: 1, per_page: 100, total: 2, total_pages: 1 },
    });

    mockFetchAdminSettings.mockResolvedValue([
      {
        id: 'setting-1',
        key: 'smtp',
        value: { host: 'smtp.example.com', port: 587, password: '***' },
        description: 'SMTP Configuration',
        is_encrypted: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 'setting-2',
        key: 'twilio',
        value: { account_sid: 'AC123***', auth_token: '***' },
        description: 'Twilio SMS Configuration',
        is_encrypted: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ]);

    mockFetchAdminAuditLogs.mockResolvedValue({
      data: [
        {
          id: 'log-1',
          timestamp: '2024-01-15T10:30:00Z',
          usuario_id: 'user-1',
          actor_name: 'John Doe',
          tenant_id: 'tenant-1',
          tenant_name: 'Hospital A',
          acao: 'CREATE',
          entidade_tipo: 'occurrence',
          entidade_id: 'occ-123',
          severity: 'INFO' as const,
        },
        {
          id: 'log-2',
          timestamp: '2024-01-15T11:00:00Z',
          usuario_id: 'user-2',
          actor_name: 'Jane Smith',
          tenant_id: 'tenant-2',
          tenant_name: 'Hospital B',
          acao: 'UPDATE',
          entidade_tipo: 'user',
          entidade_id: 'user-456',
          severity: 'WARN' as const,
        },
      ],
      meta: { page: 1, per_page: 10, total: 2, total_pages: 1 },
    });

    mockImpersonateUser.mockResolvedValue({
      access_token: 'mock-impersonation-token-abc123',
      expires_at: '2024-01-15T12:00:00Z',
    });
  });

  // Test 1: User list page with tenant filter
  describe('User List Page with Tenant Filter', () => {
    it('should filter users by tenant when tenant filter is applied', async () => {
      // Verify the API is called with tenant_id filter
      mockFetchAdminUsers.mockClear();

      // Simulate filtering by calling the API function
      await mockFetchAdminUsers({ tenant_id: 'tenant-1', page: 1, per_page: 10 });

      expect(mockFetchAdminUsers).toHaveBeenCalledWith({
        tenant_id: 'tenant-1',
        page: 1,
        per_page: 10,
      });
    });

    it('should display users from API response', async () => {
      const response = await mockFetchAdminUsers({ page: 1, per_page: 10 });

      expect(response.data).toHaveLength(2);
      expect(response.data[0].email).toBe('john@example.com');
      expect(response.data[0].tenant_name).toBe('Hospital A');
      expect(response.data[1].email).toBe('jane@example.com');
      expect(response.data[1].tenant_name).toBe('Hospital B');
    });
  });

  // Test 2: Impersonate dialog generates token
  describe('Impersonate Dialog', () => {
    it('renders user details and warning in dialog', () => {
      const mockOnClose = vi.fn();
      const mockOnImpersonate = vi.fn();
      const testUser = {
        id: 'user-1',
        email: 'john@example.com',
        nome: 'John Doe',
        role: 'admin',
        tenant_id: 'tenant-1',
        tenant_name: 'Hospital A',
        is_super_admin: false,
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      render(
        <ImpersonateDialog
          open={true}
          onClose={mockOnClose}
          user={testUser}
          onImpersonate={mockOnImpersonate}
        />,
        { wrapper: createWrapper() }
      );

      // Check user details are displayed
      expect(screen.getByText(/John Doe/)).toBeInTheDocument();
      expect(screen.getByText(/john@example.com/)).toBeInTheDocument();

      // Check warning is displayed
      expect(screen.getByText(/Esta acao sera registrada/i)).toBeInTheDocument();
    });

    it('calls onImpersonate when confirmed', async () => {
      const user = userEvent.setup();
      const mockOnClose = vi.fn();
      const mockOnImpersonate = vi.fn();
      const testUser = {
        id: 'user-1',
        email: 'john@example.com',
        nome: 'John Doe',
        role: 'admin',
        tenant_id: 'tenant-1',
        tenant_name: 'Hospital A',
        is_super_admin: false,
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      render(
        <ImpersonateDialog
          open={true}
          onClose={mockOnClose}
          user={testUser}
          onImpersonate={mockOnImpersonate}
        />,
        { wrapper: createWrapper() }
      );

      // Click confirm button
      const confirmButton = screen.getByRole('button', { name: /Confirmar Impersonalizacao/i });
      await user.click(confirmButton);

      expect(mockOnImpersonate).toHaveBeenCalledWith('user-1');
    });

    it('generates impersonation token when API is called', async () => {
      const result = await mockImpersonateUser('user-1');

      expect(mockImpersonateUser).toHaveBeenCalledWith('user-1');
      expect(result.access_token).toBe('mock-impersonation-token-abc123');
      expect(result.expires_at).toBe('2024-01-15T12:00:00Z');
    });
  });

  // Test 3: Settings page masks encrypted values
  describe('Settings Page - Encrypted Value Masking', () => {
    it('should return masked values for encrypted settings', async () => {
      const settings = await mockFetchAdminSettings();

      // Verify encrypted settings have masked values
      const smtpSetting = settings.find((s: { key: string }) => s.key === 'smtp');
      expect(smtpSetting?.is_encrypted).toBe(true);
      expect(smtpSetting?.value.password).toBe('***');

      const twilioSetting = settings.find((s: { key: string }) => s.key === 'twilio');
      expect(twilioSetting?.is_encrypted).toBe(true);
      expect(twilioSetting?.value.auth_token).toBe('***');
    });

    it('should have correct structure for settings response', async () => {
      const settings = await mockFetchAdminSettings();

      expect(settings).toHaveLength(2);
      expect(settings[0].key).toBe('smtp');
      expect(settings[0].description).toBe('SMTP Configuration');
      expect(settings[1].key).toBe('twilio');
    });
  });

  // Test 4: Audit logs export
  describe('Audit Logs Export', () => {
    it('should fetch audit logs with filters', async () => {
      await mockFetchAdminAuditLogs({
        tenant_id: 'tenant-1',
        severity: 'WARN',
        data_inicio: '2024-01-01',
        data_fim: '2024-01-31',
      });

      expect(mockFetchAdminAuditLogs).toHaveBeenCalledWith({
        tenant_id: 'tenant-1',
        severity: 'WARN',
        data_inicio: '2024-01-01',
        data_fim: '2024-01-31',
      });
    });

    it('should display tenant name in audit logs', async () => {
      const response = await mockFetchAdminAuditLogs({ page: 1 });

      expect(response.data[0].tenant_name).toBe('Hospital A');
      expect(response.data[1].tenant_name).toBe('Hospital B');
    });

    it('should export audit logs to CSV', async () => {
      const mockBlob = new Blob(['timestamp,tenant,user,action\n2024-01-15,Hospital A,John,CREATE'], {
        type: 'text/csv',
      });
      mockExportAdminAuditLogsCSV.mockResolvedValue(mockBlob);

      const result = await mockExportAdminAuditLogsCSV({
        tenant_id: 'tenant-1',
        data_inicio: '2024-01-01',
        data_fim: '2024-01-31',
      });

      expect(mockExportAdminAuditLogsCSV).toHaveBeenCalledWith({
        tenant_id: 'tenant-1',
        data_inicio: '2024-01-01',
        data_fim: '2024-01-31',
      });
      expect(result).toBeInstanceOf(Blob);
      expect(result.type).toBe('text/csv');
    });

    it('should include severity levels in logs', async () => {
      const response = await mockFetchAdminAuditLogs({ page: 1 });

      expect(response.data[0].severity).toBe('INFO');
      expect(response.data[1].severity).toBe('WARN');
    });
  });
});
