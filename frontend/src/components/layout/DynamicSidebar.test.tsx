import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DynamicSidebar } from './DynamicSidebar';
import type { SidebarItem } from '@/types/theme';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  usePathname: vi.fn(() => '/dashboard'),
}));

// Mock useAuth
vi.mock('@/hooks/useAuth', () => ({
  useAuth: vi.fn(() => ({
    user: { id: '1', email: 'test@test.com', nome: 'Test', role: 'admin' },
    isAuthenticated: true,
    isLoading: false,
  })),
}));

// Mock useTenantTheme
vi.mock('@/contexts/TenantThemeContext', () => ({
  useTenantTheme: vi.fn(),
}));

import { useTenantTheme } from '@/contexts/TenantThemeContext';

const mockUseTenantTheme = useTenantTheme as ReturnType<typeof vi.fn>;

describe('DynamicSidebar', () => {
  it('renders sidebar items from theme config', () => {
    const sidebarItems: SidebarItem[] = [
      { label: 'Dashboard', icon: 'LayoutDashboard', link: '/dashboard' },
      { label: 'Mapa', icon: 'MapPin', link: '/dashboard/map' },
      { label: 'Configuracoes', icon: 'Settings', link: '/dashboard/settings' },
    ];

    mockUseTenantTheme.mockReturnValue({
      themeConfig: {
        layout: { sidebar: sidebarItems },
      },
      isLoading: false,
    });

    render(<DynamicSidebar collapsed={false} onToggle={() => {}} />);

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Mapa')).toBeInTheDocument();
    expect(screen.getByText('Configuracoes')).toBeInTheDocument();
  });

  it('filters items based on user role', () => {
    const sidebarItems: SidebarItem[] = [
      { label: 'Dashboard', icon: 'LayoutDashboard', link: '/dashboard' },
      { label: 'Admin Only', icon: 'Shield', link: '/admin', roles: ['admin'] },
      { label: 'Gestor Only', icon: 'Users', link: '/gestor', roles: ['gestor'] },
    ];

    mockUseTenantTheme.mockReturnValue({
      themeConfig: {
        layout: { sidebar: sidebarItems },
      },
      isLoading: false,
    });

    render(<DynamicSidebar collapsed={false} onToggle={() => {}} />);

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Admin Only')).toBeInTheDocument();
    expect(screen.queryByText('Gestor Only')).not.toBeInTheDocument();
  });

  it('renders fallback sidebar when no config exists', () => {
    mockUseTenantTheme.mockReturnValue({
      themeConfig: null,
      isLoading: false,
    });

    render(<DynamicSidebar collapsed={false} onToggle={() => {}} />);

    // Should render default items from DEFAULT_THEME_CONFIG
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('collapses sidebar correctly', () => {
    const sidebarItems: SidebarItem[] = [
      { label: 'Dashboard', icon: 'LayoutDashboard', link: '/dashboard' },
    ];

    mockUseTenantTheme.mockReturnValue({
      themeConfig: {
        layout: { sidebar: sidebarItems },
      },
      isLoading: false,
    });

    const { container } = render(<DynamicSidebar collapsed={true} onToggle={() => {}} />);

    // When collapsed, sidebar should have reduced width class
    const sidebar = container.querySelector('aside');
    expect(sidebar).toHaveClass('w-16');
  });
});
