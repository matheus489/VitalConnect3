'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  X,
  LayoutDashboard,
  Building,
  Users,
  Building2,
  FileSliders,
  Settings,
  ScrollText,
  Shield,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface AdminMobileNavProps {
  open: boolean;
  onClose: () => void;
}

interface AdminNavItem {
  href: string;
  label: string;
  icon: React.ElementType;
}

const adminNavItems: AdminNavItem[] = [
  {
    href: '/admin',
    label: 'Dashboard',
    icon: LayoutDashboard,
  },
  {
    href: '/admin/tenants',
    label: 'Tenants',
    icon: Building,
  },
  {
    href: '/admin/users',
    label: 'Users',
    icon: Users,
  },
  {
    href: '/admin/hospitals',
    label: 'Hospitals',
    icon: Building2,
  },
  {
    href: '/admin/triagem-templates',
    label: 'Triagem Templates',
    icon: FileSliders,
  },
  {
    href: '/admin/settings',
    label: 'Settings',
    icon: Settings,
  },
  {
    href: '/admin/logs',
    label: 'Audit Logs',
    icon: ScrollText,
  },
];

export function AdminMobileNav({ open, onClose }: AdminMobileNavProps) {
  const pathname = usePathname();

  return (
    <>
      {/* Overlay */}
      {open && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      {/* Drawer */}
      <div
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-72 bg-slate-900 transform transition-transform duration-300 lg:hidden',
          open ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex h-16 items-center justify-between border-b border-slate-700 px-4">
            <Link href="/admin" className="flex items-center gap-2" onClick={onClose}>
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-violet-600">
                <Shield className="h-5 w-5 text-white" />
              </div>
              <div className="flex flex-col">
                <span className="text-lg font-semibold text-white">
                  SIDOT
                </span>
                <span className="text-xs font-medium text-violet-400">
                  Admin
                </span>
              </div>
            </Link>
            <Button
              variant="ghost"
              size="icon"
              onClick={onClose}
              className="text-slate-400 hover:bg-slate-800 hover:text-slate-200"
              aria-label="Fechar menu"
            >
              <X className="h-5 w-5" />
            </Button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-1 p-2">
            {adminNavItems.map((item) => {
              const isActive =
                item.href === '/admin'
                  ? pathname === '/admin'
                  : pathname === item.href || pathname?.startsWith(`${item.href}/`);
              const Icon = item.icon;

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={onClose}
                  className={cn(
                    'flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-violet-600/20 text-violet-400'
                      : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
                  )}
                >
                  <Icon className="h-5 w-5 shrink-0" />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>

          {/* Footer */}
          <div className="border-t border-slate-700 p-4">
            <p className="text-xs text-slate-500 text-center">
              SIDOT Admin v1.0
            </p>
          </div>
        </div>
      </div>
    </>
  );
}

export default AdminMobileNav;
