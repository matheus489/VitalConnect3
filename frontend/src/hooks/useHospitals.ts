'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Hospital, CreateHospitalInput, UpdateHospitalInput } from '@/types';

interface HospitalsResponse {
  data: Hospital[];
  total: number;
}

/**
 * Hook for fetching and managing hospitals data
 * Provides query for listing hospitals and mutations for create/update operations
 */
export function useHospitals() {
  const queryClient = useQueryClient();

  const hospitalsQuery = useQuery({
    queryKey: ['hospitals'],
    queryFn: async () => {
      const response = await api.get<HospitalsResponse>('/hospitals');
      return response.data.data;
    },
  });

  const createHospitalMutation = useMutation({
    mutationFn: async (input: CreateHospitalInput) => {
      const response = await api.post<Hospital>('/hospitals', input);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hospitals'] });
    },
  });

  const updateHospitalMutation = useMutation({
    mutationFn: async ({ id, input }: { id: string; input: UpdateHospitalInput }) => {
      const response = await api.patch<Hospital>(`/hospitals/${id}`, input);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hospitals'] });
    },
  });

  return {
    ...hospitalsQuery,
    createHospital: createHospitalMutation.mutateAsync,
    updateHospital: updateHospitalMutation.mutateAsync,
    isCreating: createHospitalMutation.isPending,
    isUpdating: updateHospitalMutation.isPending,
    createError: createHospitalMutation.error,
    updateError: updateHospitalMutation.error,
  };
}
