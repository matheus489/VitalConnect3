'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  MapPin,
  ClipboardList,
  Settings,
  Users,
  Building2,
  ChevronLeft,
  ChevronRight,
  Calendar,
  FileText,
  History,
  Activity,
  Sliders,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

interface NavItem {
  href: string;
  label: string;
  icon: React.ElementType;
  roles?: string[];
}

const navItems: NavItem[] = [
  {
    href: '/dashboard',
    label: 'Dashboard',
    icon: LayoutDashboard,
  },
  {
    href: '/dashboard/map',
    label: 'Mapa',
    icon: MapPin,
  },
  {
    href: '/dashboard/occurrences',
    label: 'Ocorrencias',
    icon: ClipboardList,
  },
  {
    href: '/dashboard/shifts',
    label: 'Escalas',
    icon: Calendar,
  },
  {
    href: '/dashboard/rules',
    label: 'Regras',
    icon: Sliders,
    roles: ['admin', 'gestor'],
  },
  {
    href: '/dashboard/hospitals',
    label: 'Hospitais',
    icon: Building2,
    roles: ['admin'],
  },
  {
    href: '/dashboard/users',
    label: 'Usuarios',
    icon: Users,
    roles: ['admin'],
  },
  {
    href: '/dashboard/reports',
    label: 'Relatorios',
    icon: FileText,
    roles: ['admin', 'gestor'],
  },
  {
    href: '/dashboard/audit-logs',
    label: 'Auditoria',
    icon: History,
    roles: ['admin', 'gestor'],
  },
  {
    href: '/dashboard/status',
    label: 'Status',
    icon: Activity,
    roles: ['admin'],
  },
  {
    href: '/dashboard/settings',
    label: 'Configuracoes',
    icon: Settings,
    roles: ['admin', 'gestor'],
  },
];

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const pathname = usePathname();
  const { user } = useAuth();

  const filteredNavItems = navItems.filter((item) => {
    if (!item.roles) return true;
    return user && item.roles.includes(user.role);
  });

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
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                <span className="text-lg font-bold text-primary-foreground">V</span>
              </div>
              <span className="text-lg font-semibold text-sidebar-foreground">
                VitalConnect
              </span>
            </Link>
          )}
          {collapsed && (
            <Link href="/dashboard" className="mx-auto">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                <span className="text-lg font-bold text-primary-foreground">V</span>
              </div>
            </Link>
          )}
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 p-2">
          {filteredNavItems.map((item) => {
            const isActive =
              pathname === item.href || pathname?.startsWith(`${item.href}/`);
            const Icon = item.icon;

            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                    : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                  collapsed && 'justify-center px-2'
                )}
                title={collapsed ? item.label : undefined}
              >
                <Icon className="h-5 w-5 shrink-0" />
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
