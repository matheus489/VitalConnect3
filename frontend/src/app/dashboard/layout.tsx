'use client';

import { AuthProvider } from '@/hooks/useAuth';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ChatWidget } from '@/components/ai';

export default function DashboardRootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthProvider>
      <DashboardLayout>{children}</DashboardLayout>
      <ChatWidget />
    </AuthProvider>
  );
}
