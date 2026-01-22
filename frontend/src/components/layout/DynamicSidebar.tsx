'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { useTenantTheme } from '@/contexts/TenantThemeContext';
import { DynamicIcon } from '@/lib/dynamicIcon';
import { DEFAULT_THEME_CONFIG, type SidebarItem } from '@/types/theme';

interface DynamicSidebarProps {
  /** Whether the sidebar is collapsed */
  collapsed: boolean;
  /** Callback when collapse toggle is clicked */
  onToggle: () => void;
}

/**
 * Dynamic sidebar component that renders navigation items from theme_config.
 *
 * Features:
 * - Reads sidebar items from tenant's theme_config.layout.sidebar
 * - Renders icons dynamically using DynamicIcon component
 * - Supports role-based visibility (item.roles)
 * - Maintains active state detection based on current pathname
 * - Falls back to default sidebar items if no config exists
 * - Supports collapsed/expanded states
 *
 * @example
 * ```tsx
 * function Layout() {
 *   const [collapsed, setCollapsed] = useState(false);
 *
 *   return (
 *     <DynamicSidebar
 *       collapsed={collapsed}
 *       onToggle={() => setCollapsed(!collapsed)}
 *     />
 *   );
 * }
 * ```
 */
export function DynamicSidebar({ collapsed, onToggle }: DynamicSidebarProps) {
  const pathname = usePathname();
  const { user } = useAuth();
  const { themeConfig, isLoading, logoUrl } = useTenantTheme();

  // Get sidebar items from config or use defaults
  const sidebarItems: SidebarItem[] =
    themeConfig?.layout?.sidebar ||
    DEFAULT_THEME_CONFIG.layout?.sidebar ||
    [];

  // Filter items based on user role
  const filteredNavItems = sidebarItems.filter((item) => {
    // If no roles specified, show to everyone
    if (!item.roles || item.roles.length === 0) return true;
    // Check if user's role is in the allowed roles
    return user && item.roles.includes(user.role);
  });

  // Sort items by order if specified
  const sortedNavItems = [...filteredNavItems].sort((a, b) => {
    const orderA = a.order ?? 999;
    const orderB = b.order ?? 999;
    return orderA - orderB;
  });

  /**
   * Check if a nav item is currently active
   */
  const isItemActive = (item: SidebarItem): boolean => {
    if (!pathname) return false;
    // Exact match or starts with the link path
    return pathname === item.link || pathname.startsWith(`${item.link}/`);
  };

  if (isLoading) {
    return (
      <aside
        className={cn(
          'fixed left-0 top-0 z-40 h-screen bg-sidebar border-r border-sidebar-border transition-all duration-300',
          collapsed ? 'w-16' : 'w-64'
        )}
      >
        <div className="flex h-full flex-col">
          {/* Logo Skeleton */}
          <div className="flex h-16 items-center justify-between border-b border-sidebar-border px-4">
            <div className="h-8 w-8 animate-pulse rounded-lg bg-muted" />
            {!collapsed && (
              <div className="ml-2 h-6 w-24 animate-pulse rounded bg-muted" />
            )}
          </div>

          {/* Navigation Skeleton */}
          <nav className="flex-1 space-y-1 p-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <div
                key={i}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2',
                  collapsed && 'justify-center px-2'
                )}
              >
                <div className="h-5 w-5 animate-pulse rounded bg-muted" />
                {!collapsed && (
                  <div className="h-4 w-20 animate-pulse rounded bg-muted" />
                )}
              </div>
            ))}
          </nav>
        </div>
      </aside>
    );
  }

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 z-40 h-screen bg-sidebar border-r border-sidebar-border transition-all duration-300',
        collapsed ? 'w-16' : 'w-64'
      )}
    >
      <div className="flex h-full flex-col">
        {/* Logo */}
        <div className="flex h-16 items-center justify-between border-b border-sidebar-border px-4">
          {!collapsed && (
            <Link href="/dashboard" className="flex items-center gap-2">
              {logoUrl ? (
                <img
                  src={logoUrl}
                  alt="Logo"
                  className="h-8 w-8 rounded-lg object-contain"
                />
              ) : (
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                  <span className="text-lg font-bold text-primary-foreground">S</span>
                </div>
              )}
              <span className="text-lg font-semibold text-sidebar-foreground">
                SIDOT
              </span>
            </Link>
          )}
          {collapsed && (
            <Link href="/dashboard" className="mx-auto">
              {logoUrl ? (
                <img
                  src={logoUrl}
                  alt="Logo"
                  className="h-8 w-8 rounded-lg object-contain"
                />
              ) : (
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                  <span className="text-lg font-bold text-primary-foreground">S</span>
                </div>
              )}
            </Link>
          )}
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 overflow-y-auto p-2 scrollbar-thin">
          {sortedNavItems.map((item) => {
            const isActive = isItemActive(item);

            return (
              <Link
                key={item.id || item.link}
                href={item.link}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                    : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                  collapsed && 'justify-center px-2'
                )}
                title={collapsed ? item.label : undefined}
              >
                <DynamicIcon name={item.icon} className="h-5 w-5 shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        {/* Collapse Toggle */}
        <div className="border-t border-sidebar-border p-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggle}
            className="w-full justify-center"
            aria-label={collapsed ? 'Expandir menu' : 'Recolher menu'}
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>
    </aside>
  );
}

export default DynamicSidebar;
