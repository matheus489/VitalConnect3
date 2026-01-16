'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { AuditLog, AuditLogFilters, PaginatedResponse } from '@/types';

export function useAuditLogs(filters: AuditLogFilters = {}) {
  const params = new URLSearchParams();

  if (filters.data_inicio) params.append('data_inicio', filters.data_inicio);
  if (filters.data_fim) params.append('data_fim', filters.data_fim);
  if (filters.usuario_id) params.append('usuario_id', filters.usuario_id);
  if (filters.acao) params.append('acao', filters.acao);
  if (filters.entidade_tipo) params.append('entidade_tipo', filters.entidade_tipo);
  if (filters.severity) params.append('severity', filters.severity);
  if (filters.hospital_id) params.append('hospital_id', filters.hospital_id);
  if (filters.page) params.append('page', String(filters.page));
  if (filters.page_size) params.append('page_size', String(filters.page_size));

  return useQuery({
    queryKey: ['audit-logs', filters],
    queryFn: async () => {
      const response = await api.get<PaginatedResponse<AuditLog>>(`/audit-logs?${params.toString()}`);
      return response.data;
    },
  });
}

export function useOccurrenceTimeline(occurrenceId?: string) {
  return useQuery({
    queryKey: ['occurrence-timeline', occurrenceId],
    queryFn: async () => {
      if (!occurrenceId) return [];
      const response = await api.get<{ data: AuditLog[]; total: number }>(
        `/occurrences/${occurrenceId}/timeline`
      );
      return response.data.data;
    },
    enabled: !!occurrenceId,
  });
}
