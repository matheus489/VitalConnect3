'use client';

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { api } from '@/lib/api';
import type { MapDataResponse, SSENotificationEvent } from '@/types';

/**
 * Opcoes para configurar o hook useMapHospitals
 */
interface UseMapHospitalsOptions {
  /**
   * Intervalo de atualizacao automatica em milissegundos
   * @default 30000 (30 segundos)
   */
  refetchInterval?: number;
  /**
   * Se o hook deve estar habilitado
   * @default true
   */
  enabled?: boolean;
}

/**
 * Hook para buscar dados de hospitais para renderizacao no mapa geografico
 *
 * Usa React Query para gerenciamento de cache e atualizacao automatica.
 * Os dados incluem hospitais ativos com coordenadas, suas ocorrencias ativas
 * e o operador de plantao atual.
 *
 * @param options - Opcoes de configuracao do hook
 * @returns Query result com isLoading, error, data e funcoes de refetch
 *
 * @example
 * ```tsx
 * const { data, isLoading, error } = useMapHospitals();
 *
 * if (isLoading) return <Skeleton />;
 * if (error) return <ErrorMessage />;
 *
 * return <Map hospitals={data?.hospitals} />;
 * ```
 */
export function useMapHospitals(options: UseMapHospitalsOptions = {}) {
  const { refetchInterval = 30000, enabled = true } = options;

  return useQuery({
    queryKey: ['map', 'hospitals'],
    queryFn: async () => {
      const response = await api.get<MapDataResponse>('/map/hospitals');
      return response.data;
    },
    refetchInterval,
    enabled,
  });
}

/**
 * Hook para invalidar o cache do mapa quando receber eventos SSE
 *
 * Retorna uma funcao callback que pode ser passada para o hook useSSE
 * como onNotification. Quando uma ocorrencia muda de status ou uma nova
 * ocorrencia e criada, o cache do mapa e invalidado para buscar dados atualizados.
 *
 * @returns Callback para tratar eventos SSE relacionados ao mapa
 *
 * @example
 * ```tsx
 * const handleMapUpdate = useMapSSEHandler();
 *
 * useSSE({
 *   onNotification: handleMapUpdate,
 * });
 * ```
 */
export function useMapSSEHandler() {
  const queryClient = useQueryClient();

  const handleSSEEvent = useCallback(
    (event: SSENotificationEvent) => {
      // Invalidar cache do mapa quando houver mudanca de status ou nova ocorrencia
      if (event.type === 'new_occurrence' || event.type === 'status_update' || event.type === 'map_update') {
        queryClient.invalidateQueries({ queryKey: ['map', 'hospitals'] });
      }
    },
    [queryClient]
  );

  return handleSSEEvent;
}
