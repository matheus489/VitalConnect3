'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { DashboardMetrics } from '@/types';

export function useMetrics() {
  return useQuery({
    queryKey: ['metrics'],
    queryFn: async () => {
      const response = await api.get<DashboardMetrics>('/metrics/dashboard');
      return response.data;
    },
    refetchInterval: 30 * 1000, // Refresh every 30 seconds
  });
}
