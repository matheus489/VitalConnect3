'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { SystemHealth } from '@/types';

export function useSystemHealth() {
  return useQuery({
    queryKey: ['health', 'summary'],
    queryFn: async () => {
      const response = await api.get<SystemHealth>('/health/summary');
      return response.data;
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

export function useListenerHealth() {
  return useQuery({
    queryKey: ['health', 'listener'],
    queryFn: async () => {
      const response = await api.get('/health/listener');
      return response.data;
    },
    refetchInterval: 30000,
  });
}

export function useSSEHealth() {
  return useQuery({
    queryKey: ['health', 'sse'],
    queryFn: async () => {
      const response = await api.get('/health/sse');
      return response.data;
    },
    refetchInterval: 30000,
  });
}
