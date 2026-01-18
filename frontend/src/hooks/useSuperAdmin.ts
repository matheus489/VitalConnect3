'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';

interface UseSuperAdminResult {
  isSuperAdmin: boolean;
  isLoading: boolean;
  user: ReturnType<typeof useAuth>['user'];
}

/**
 * Hook to verify if the current user is a super admin.
 * Redirects to /dashboard if user is not a super admin.
 *
 * @returns Object containing isSuperAdmin status, loading state, and user data
 */
export function useSuperAdmin(): UseSuperAdminResult {
  const { user, isLoading, isAuthenticated } = useAuth();
  const router = useRouter();

  // Check if user has is_super_admin flag
  const isSuperAdmin = Boolean(user?.is_super_admin);

  useEffect(() => {
    // Wait for auth to finish loading
    if (isLoading) return;

    // If not authenticated, the AuthProvider will handle redirect to /login
    if (!isAuthenticated) return;

    // If authenticated but not super admin, redirect to dashboard
    if (!isSuperAdmin) {
      router.push('/dashboard');
    }
  }, [isLoading, isAuthenticated, isSuperAdmin, router]);

  return {
    isSuperAdmin,
    isLoading,
    user,
  };
}

export default useSuperAdmin;
