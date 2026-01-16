'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type {
  Occurrence,
  OccurrenceDetail,
  OccurrenceFilters,
  OccurrenceStatus,
  OutcomeType,
  PaginatedResponse,
  OccurrenceHistoryItem,
} from '@/types';

interface UseOccurrencesOptions {
  page?: number;
  perPage?: number;
  filters?: OccurrenceFilters;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

interface StatusUpdateRequest {
  id: string;
  status: OccurrenceStatus;
  observacoes?: string;
}

interface OutcomeRequest {
  id: string;
  desfecho: OutcomeType;
  observacoes?: string;
}

export function useOccurrences(options: UseOccurrencesOptions = {}) {
  const { page = 1, perPage = 10, filters, sortBy, sortOrder } = options;

  return useQuery({
    queryKey: ['occurrences', page, perPage, filters, sortBy, sortOrder],
    queryFn: async () => {
      const params = new URLSearchParams();
      params.append('page', String(page));
      params.append('per_page', String(perPage));

      if (filters?.status) {
        params.append('status', filters.status);
      }
      if (filters?.hospital_id) {
        params.append('hospital_id', filters.hospital_id);
      }
      if (filters?.date_from) {
        params.append('date_from', filters.date_from);
      }
      if (filters?.date_to) {
        params.append('date_to', filters.date_to);
      }
      if (sortBy) {
        params.append('sort_by', sortBy);
      }
      if (sortOrder) {
        params.append('sort_order', sortOrder);
      }

      const response = await api.get(`/occurrences?${params.toString()}`);
      const apiData = response.data as {
        data: Occurrence[];
        page: number;
        page_size: number;
        total_items: number;
        total_pages: number;
        has_next: boolean;
        has_prev: boolean;
      };
      // Transform to expected format
      return {
        data: apiData.data,
        meta: {
          page: apiData.page,
          per_page: apiData.page_size,
          total: apiData.total_items,
          total_pages: apiData.total_pages,
        },
      } as PaginatedResponse<Occurrence>;
    },
  });
}

export function useOccurrenceDetail(id: string | null) {
  return useQuery({
    queryKey: ['occurrence', id],
    queryFn: async () => {
      if (!id) return null;
      const response = await api.get<OccurrenceDetail>(`/occurrences/${id}`);
      return response.data;
    },
    enabled: !!id,
  });
}

export function useOccurrenceHistory(id: string | null) {
  return useQuery({
    queryKey: ['occurrence-history', id],
    queryFn: async () => {
      if (!id) return [];
      const response = await api.get<OccurrenceHistoryItem[]>(`/occurrences/${id}/history`);
      return response.data;
    },
    enabled: !!id,
  });
}

export function useUpdateOccurrenceStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, status, observacoes }: StatusUpdateRequest) => {
      const response = await api.patch(`/occurrences/${id}/status`, {
        status,
        observacoes,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['occurrences'] });
      queryClient.invalidateQueries({ queryKey: ['occurrence'] });
      queryClient.invalidateQueries({ queryKey: ['occurrence-history'] });
      queryClient.invalidateQueries({ queryKey: ['metrics'] });
    },
  });
}

export function useRegisterOutcome() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, desfecho, observacoes }: OutcomeRequest) => {
      const response = await api.post(`/occurrences/${id}/outcome`, {
        desfecho,
        observacoes,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['occurrences'] });
      queryClient.invalidateQueries({ queryKey: ['occurrence'] });
      queryClient.invalidateQueries({ queryKey: ['occurrence-history'] });
      queryClient.invalidateQueries({ queryKey: ['metrics'] });
    },
  });
}
