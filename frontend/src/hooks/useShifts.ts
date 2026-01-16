'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Shift, TodayShift, CreateShiftInput, CoverageAnalysis } from '@/types';

export function useShifts(hospitalId?: string) {
  return useQuery({
    queryKey: ['shifts', hospitalId],
    queryFn: async () => {
      if (!hospitalId) return [];
      const response = await api.get<Shift[]>(`/hospitals/${hospitalId}/shifts`);
      return response.data;
    },
    enabled: !!hospitalId,
  });
}

export function useMyShifts() {
  return useQuery({
    queryKey: ['shifts', 'me'],
    queryFn: async () => {
      const response = await api.get<Shift[]>('/shifts/me');
      return response.data;
    },
  });
}

export function useTodayShifts(hospitalId?: string) {
  return useQuery({
    queryKey: ['shifts', 'today', hospitalId],
    queryFn: async () => {
      if (!hospitalId) return [];
      const response = await api.get<TodayShift[]>(`/hospitals/${hospitalId}/shifts/today`);
      return response.data;
    },
    enabled: !!hospitalId,
    refetchInterval: 60000, // Refetch every minute
  });
}

export function useCoverageGaps(hospitalId?: string) {
  return useQuery({
    queryKey: ['shifts', 'coverage', hospitalId],
    queryFn: async () => {
      if (!hospitalId) return null;
      const response = await api.get<CoverageAnalysis>(`/hospitals/${hospitalId}/shifts/coverage`);
      return response.data;
    },
    enabled: !!hospitalId,
  });
}

export function useCreateShift() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateShiftInput) => {
      const response = await api.post<Shift>('/shifts', input);
      return response.data;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['shifts', variables.hospital_id] });
      queryClient.invalidateQueries({ queryKey: ['shifts', 'coverage', variables.hospital_id] });
    },
  });
}

export function useUpdateShift() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, ...input }: { id: string } & Partial<CreateShiftInput>) => {
      const response = await api.put<Shift>(`/shifts/${id}`, input);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
    },
  });
}

export function useDeleteShift() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/shifts/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
    },
  });
}
