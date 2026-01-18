import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TenantThemeProvider, useTenantTheme } from './TenantThemeContext';
import { DynamicIcon, getIconByName } from '@/lib/dynamicIcon';
import type { ThemeConfig } from '@/types/theme';

// Mock useAuth hook
vi.mock('@/hooks/useAuth', () => ({
  useAuth: vi.fn(() => ({
    user: { id: '1', email: 'test@test.com', nome: 'Test', role: 'admin' },
    isAuthenticated: true,
    isLoading: false,
  })),
}));

// Mock fetch for API calls
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock document.documentElement.style.setProperty
const setPropertyMock = vi.fn();
const originalSetProperty = document.documentElement.style.setProperty;

beforeEach(() => {
  document.documentElement.style.setProperty = setPropertyMock;
  mockFetch.mockReset();
  setPropertyMock.mockClear();
});

afterEach(() => {
  document.documentElement.style.setProperty = originalSetProperty;
});

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
});

function TestConsumer() {
  const { themeConfig, isLoading } = useTenantTheme();

  if (isLoading) return <div>Loading...</div>;

  return (
    <div>
      <div data-testid="primary-color">{themeConfig?.theme?.colors?.primary || 'no-color'}</div>
      <div data-testid="sidebar-items-count">
        {themeConfig?.layout?.sidebar?.length || 0}
      </div>
    </div>
  );
}

describe('TenantThemeContext', () => {
  it('provides default theme config when no tenant config exists', async () => {
    render(
      <QueryClientProvider client={queryClient}>
        <TenantThemeProvider>
          <TestConsumer />
        </TenantThemeProvider>
      </QueryClientProvider>
    );

    // Should render with default values after loading
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    // Default config has primary color
    const primaryColor = screen.getByTestId('primary-color');
    expect(primaryColor.textContent).not.toBe('no-color');
  });

  it('injects CSS variables into document root', async () => {
    const mockConfig: ThemeConfig = {
      theme: {
        colors: {
          primary: '#FF5733',
          background: '#FFFFFF',
        },
      },
      layout: {
        sidebar: [
          { label: 'Test', icon: 'Home', link: '/test' },
        ],
      },
    };

    render(
      <QueryClientProvider client={queryClient}>
        <TenantThemeProvider initialConfig={mockConfig}>
          <TestConsumer />
        </TenantThemeProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      // Check that CSS variables were set
      expect(setPropertyMock).toHaveBeenCalled();
    });
  });
});

describe('DynamicIcon', () => {
  it('resolves known Lucide icons by name', () => {
    const { container } = render(<DynamicIcon name="Home" />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('returns fallback icon for unknown names', () => {
    const { container } = render(<DynamicIcon name="UnknownIconXYZ" />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('getIconByName returns component for valid name', () => {
    const Icon = getIconByName('Settings');
    expect(Icon).toBeDefined();
  });

  it('getIconByName returns fallback for invalid name', () => {
    const Icon = getIconByName('InvalidIconName');
    expect(Icon).toBeDefined();
  });
});
