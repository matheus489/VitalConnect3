import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock next/navigation
const mockPush = vi.fn();
const mockPathname = '/admin';
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

// Mock API
vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

import { AdminSidebar } from './AdminSidebar';
import { AdminHeader } from './AdminHeader';
import {
  parseCommand,
  getCommandSuggestions,
  applyCommandToTheme,
  formatCommandResult,
} from './CommandParser';
import { DEFAULT_THEME_CONFIG } from '@/types/theme';

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

describe('AdminSidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders all admin navigation items', () => {
    render(<AdminSidebar collapsed={false} onToggle={() => {}} />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Tenants')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('Hospitals')).toBeInTheDocument();
    expect(screen.getByText('Triagem Templates')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
    expect(screen.getByText('Audit Logs')).toBeInTheDocument();
  });

  it('renders collapsed state correctly', () => {
    render(<AdminSidebar collapsed={true} onToggle={() => {}} />, {
      wrapper: createWrapper(),
    });

    // In collapsed state, text labels should be hidden but accessible via title
    const dashboardLink = screen.getByTitle('Dashboard');
    expect(dashboardLink).toBeInTheDocument();
  });

  it('displays SIDOT Admin branding', () => {
    render(<AdminSidebar collapsed={false} onToggle={() => {}} />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText('SIDOT')).toBeInTheDocument();
    expect(screen.getByText('Admin')).toBeInTheDocument();
  });
});

describe('AdminHeader', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('displays admin user information', () => {
    render(<AdminHeader onMenuToggle={() => {}} showMenuButton={true} />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText('Super Admin')).toBeInTheDocument();
    expect(screen.getByText('Super Administrador')).toBeInTheDocument();
  });

  it('displays back to app link', () => {
    render(<AdminHeader onMenuToggle={() => {}} showMenuButton={false} />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText('Voltar ao App')).toBeInTheDocument();
  });
});

// Command Parser Tests
describe('CommandParser', () => {
  describe('parseCommand - Color Commands', () => {
    it('parses Set Primary Color command correctly', () => {
      const result = parseCommand('Set Primary Color #0EA5E9');

      expect(result.type).toBe('SET_PRIMARY_COLOR');
      expect(result.isValid).toBe(true);
      expect(result.payload.color).toBe('#0EA5E9');
    });

    it('parses Set Background command correctly', () => {
      const result = parseCommand('Set Background #FFFFFF');

      expect(result.type).toBe('SET_BACKGROUND_COLOR');
      expect(result.isValid).toBe(true);
      expect(result.payload.color).toBe('#FFFFFF');
    });

    it('rejects invalid hex colors', () => {
      const result = parseCommand('Set Primary Color #GGGGGG');

      expect(result.type).toBe('UNKNOWN');
      expect(result.isValid).toBe(false);
    });

    it('parses 3-digit hex colors', () => {
      const result = parseCommand('Set Primary Color #FFF');

      expect(result.type).toBe('SET_PRIMARY_COLOR');
      expect(result.isValid).toBe(true);
      expect(result.payload.color).toBe('#FFF');
    });
  });

  describe('parseCommand - Sidebar Commands', () => {
    it('parses Sidebar Add Item command correctly', () => {
      const result = parseCommand('Sidebar: Add Item "Relatorios" icon="FileText" link="/reports"');

      expect(result.type).toBe('SIDEBAR_ADD_ITEM');
      expect(result.isValid).toBe(true);
      expect(result.payload.label).toBe('Relatorios');
      expect(result.payload.icon).toBe('FileText');
      expect(result.payload.link).toBe('/reports');
    });

    it('parses Sidebar Remove command correctly', () => {
      const result = parseCommand('Sidebar: Remove "Configuracoes"');

      expect(result.type).toBe('SIDEBAR_REMOVE_ITEM');
      expect(result.isValid).toBe(true);
      expect(result.payload.label).toBe('Configuracoes');
    });

    it('parses Sidebar Move command correctly', () => {
      const result = parseCommand('Sidebar: Move "Dashboard" to Top');

      expect(result.type).toBe('SIDEBAR_MOVE_ITEM');
      expect(result.isValid).toBe(true);
      expect(result.payload.label).toBe('Dashboard');
      expect(result.payload.position).toBe('top');
    });

    it('parses Sidebar Move to Bottom command correctly', () => {
      const result = parseCommand('Sidebar: Move "Settings" to Bottom');

      expect(result.type).toBe('SIDEBAR_MOVE_ITEM');
      expect(result.isValid).toBe(true);
      expect(result.payload.position).toBe('bottom');
    });
  });

  describe('parseCommand - Dashboard Widget Commands', () => {
    it('parses Dashboard Add Widget command correctly', () => {
      const result = parseCommand('Dashboard: Add Widget "stats_card"');

      expect(result.type).toBe('DASHBOARD_ADD_WIDGET');
      expect(result.isValid).toBe(true);
      expect(result.payload.widgetType).toBe('stats_card');
    });

    it('parses Dashboard Hide command correctly', () => {
      const result = parseCommand('Dashboard: Hide "chart"');

      expect(result.type).toBe('DASHBOARD_HIDE_WIDGET');
      expect(result.isValid).toBe(true);
      expect(result.payload.widgetId).toBe('chart');
    });

    it('parses Dashboard Show command correctly', () => {
      const result = parseCommand('Dashboard: Show "map"');

      expect(result.type).toBe('DASHBOARD_SHOW_WIDGET');
      expect(result.isValid).toBe(true);
      expect(result.payload.widgetId).toBe('map');
    });

    it('rejects invalid widget types', () => {
      const result = parseCommand('Dashboard: Add Widget "invalid_widget"');

      expect(result.type).toBe('DASHBOARD_ADD_WIDGET');
      expect(result.isValid).toBe(false);
      expect(result.error).toContain('Tipo de widget invalido');
    });
  });

  describe('parseCommand - Font Commands', () => {
    it('parses Set Font command correctly', () => {
      const result = parseCommand('Set Font "Inter"');

      expect(result.type).toBe('SET_FONT');
      expect(result.isValid).toBe(true);
      expect(result.payload.fontFamily).toBe('Inter');
    });
  });

  describe('parseCommand - Asset Commands', () => {
    it('parses Upload Logo command correctly', () => {
      const result = parseCommand('Upload Logo');

      expect(result.type).toBe('UPLOAD_LOGO');
      expect(result.isValid).toBe(true);
    });

    it('parses Upload Favicon command correctly', () => {
      const result = parseCommand('Upload Favicon');

      expect(result.type).toBe('UPLOAD_FAVICON');
      expect(result.isValid).toBe(true);
    });
  });

  describe('parseCommand - Edge Cases', () => {
    it('returns UNKNOWN for empty command', () => {
      const result = parseCommand('');

      expect(result.type).toBe('UNKNOWN');
      expect(result.isValid).toBe(false);
    });

    it('returns UNKNOWN for unrecognized command', () => {
      const result = parseCommand('Random text here');

      expect(result.type).toBe('UNKNOWN');
      expect(result.isValid).toBe(false);
    });

    it('is case insensitive', () => {
      const result = parseCommand('SET PRIMARY COLOR #ABC123');

      expect(result.type).toBe('SET_PRIMARY_COLOR');
      expect(result.isValid).toBe(true);
    });
  });
});

describe('getCommandSuggestions', () => {
  it('returns all suggestions for empty input', () => {
    const suggestions = getCommandSuggestions('');
    expect(suggestions.length).toBeGreaterThan(0);
  });

  it('filters suggestions based on input', () => {
    const suggestions = getCommandSuggestions('primary');
    expect(suggestions.length).toBeGreaterThan(0);
    expect(suggestions.some(s => s.command.toLowerCase().includes('primary'))).toBe(true);
  });

  it('filters by category', () => {
    const suggestions = getCommandSuggestions('sidebar');
    expect(suggestions.every(s =>
      s.command.toLowerCase().includes('sidebar') ||
      s.category === 'sidebar'
    )).toBe(true);
  });
});

describe('applyCommandToTheme', () => {
  it('applies SET_PRIMARY_COLOR action', () => {
    const action = parseCommand('Set Primary Color #FF0000');
    const newTheme = applyCommandToTheme(action, DEFAULT_THEME_CONFIG);

    expect(newTheme.theme?.colors?.primary).toBe('#FF0000');
  });

  it('applies SET_FONT action', () => {
    const action = parseCommand('Set Font "Roboto"');
    const newTheme = applyCommandToTheme(action, DEFAULT_THEME_CONFIG);

    expect(newTheme.theme?.fonts?.body).toBe('Roboto');
    expect(newTheme.theme?.fonts?.heading).toBe('Roboto');
  });

  it('applies SIDEBAR_ADD_ITEM action', () => {
    const action = parseCommand('Sidebar: Add Item "Test" icon="Home" link="/test"');
    const newTheme = applyCommandToTheme(action, DEFAULT_THEME_CONFIG);

    const addedItem = newTheme.layout?.sidebar?.find(item => item.label === 'Test');
    expect(addedItem).toBeDefined();
    expect(addedItem?.icon).toBe('Home');
    expect(addedItem?.link).toBe('/test');
  });

  it('applies SIDEBAR_REMOVE_ITEM action', () => {
    const action = parseCommand('Sidebar: Remove "Dashboard"');
    const newTheme = applyCommandToTheme(action, DEFAULT_THEME_CONFIG);

    const removedItem = newTheme.layout?.sidebar?.find(item => item.label === 'Dashboard');
    expect(removedItem).toBeUndefined();
  });

  it('applies DASHBOARD_HIDE_WIDGET action', () => {
    const action = parseCommand('Dashboard: Hide "chart"');
    const newTheme = applyCommandToTheme(action, DEFAULT_THEME_CONFIG);

    const hiddenWidget = newTheme.layout?.dashboard_widgets?.find(w => w.type === 'chart');
    expect(hiddenWidget?.visible).toBe(false);
  });

  it('does not modify theme for invalid action', () => {
    const action = parseCommand('Invalid command');
    const newTheme = applyCommandToTheme(action, DEFAULT_THEME_CONFIG);

    // Should return same config when action is invalid
    expect(newTheme).toEqual(DEFAULT_THEME_CONFIG);
  });
});

describe('formatCommandResult', () => {
  it('formats successful color change', () => {
    const action = parseCommand('Set Primary Color #FF0000');
    const result = formatCommandResult(action);

    expect(result).toContain('#FF0000');
    expect(result).toContain('primaria');
  });

  it('formats invalid command error', () => {
    const action = parseCommand('Invalid command');
    const result = formatCommandResult(action);

    expect(result).toContain('Comando nao reconhecido');
  });

  it('formats sidebar add item result', () => {
    const action = parseCommand('Sidebar: Add Item "Test" icon="Home" link="/test"');
    const result = formatCommandResult(action);

    expect(result).toContain('Test');
    expect(result).toContain('adicionado');
  });
});
