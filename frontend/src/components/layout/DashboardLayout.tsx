'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { DynamicSidebar } from './DynamicSidebar';
import { Header } from './Header';
import { MobileNav } from './MobileNav';
import { useAuth } from '@/hooks/useAuth';
import { useSSE } from '@/hooks/useSSE';
import { TenantThemeProvider } from '@/contexts/TenantThemeContext';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';
import type { SSENotificationEvent } from '@/types';

interface DashboardLayoutProps {
  children: React.ReactNode;
}

/**
 * Inner layout component that uses the DynamicSidebar
 * Must be inside TenantThemeProvider to access theme context
 */
function DashboardLayoutInner({ children }: DashboardLayoutProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const { user, isLoading, isAuthenticated } = useAuth();
  const router = useRouter();

  const handleNotification = (event: SSENotificationEvent) => {
    if (event.type === 'new_occurrence') {
      toast.warning('Nova Ocorrencia Detectada', {
        description: `${event.hospital} - ${event.setor}. Tempo restante: ${event.tempo_restante_minutos} minutos`,
        duration: 10000,
        action: {
          label: 'Ver',
          onClick: () => router.push(`/dashboard/occurrences?id=${event.occurrence_id}`),
        },
      });
    }
  };

  useSSE({
    onNotification: handleNotification,
    enabled: isAuthenticated,
  });

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isLoading, isAuthenticated, router]);

  // Close mobile nav on route change
  useEffect(() => {
    setMobileNavOpen(false);
  }, [router]);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">Carregando...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated || !user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Desktop Sidebar - Now uses DynamicSidebar */}
      <div className="hidden lg:block">
        <DynamicSidebar
          collapsed={sidebarCollapsed}
          onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
        />
      </div>

      {/* Mobile Navigation */}
      <MobileNav open={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />

      {/* Main Content */}
      <div
        className={cn(
          'flex flex-col transition-all duration-300',
          sidebarCollapsed ? 'lg:pl-16' : 'lg:pl-64'
        )}
      >
        <Header
          onMenuToggle={() => setMobileNavOpen(true)}
          showMenuButton
        />

        <main className="flex-1 p-4 lg:p-6">{children}</main>
      </div>
    </div>
  );
}

/**
 * Dashboard layout component with dynamic theme support.
 *
 * This component:
 * - Wraps children with TenantThemeProvider for dynamic theming
 * - Uses DynamicSidebar that reads from theme_config
 * - Handles authentication and redirects
 * - Manages SSE notifications
 * - Handles responsive sidebar collapse
 *
 * @example
 * ```tsx
 * // In app/dashboard/layout.tsx
 * export default function DashboardRootLayout({ children }) {
 *   return (
 *     <AuthProvider>
 *       <DashboardLayout>{children}</DashboardLayout>
 *     </AuthProvider>
 *   );
 * }
 * ```
 */
export function DashboardLayout({ children }: DashboardLayoutProps) {
  return (
    <TenantThemeProvider>
      <DashboardLayoutInner>{children}</DashboardLayoutInner>
    </TenantThemeProvider>
  );
}

export default DashboardLayout;
