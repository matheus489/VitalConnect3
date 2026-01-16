import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createElement } from 'react';
import type { SSENotificationEvent } from '@/types';
import { useMapSSEHandler } from './useMap';

// Helper to create a wrapper with QueryClientProvider
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children),
    queryClient,
  };
};

describe('useMapSSEHandler', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('invalida cache do mapa quando recebe evento new_occurrence', () => {
    const { wrapper, queryClient } = createWrapper();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useMapSSEHandler(), { wrapper });

    const event: SSENotificationEvent = {
      type: 'new_occurrence',
      occurrence_id: '123',
      hospital: 'Hospital Central',
      hospital_id: '456',
      setor: 'UTI',
      tempo_restante_minutos: 90,
      timestamp: new Date().toISOString(),
    };

    act(() => {
      result.current(event);
    });

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['map', 'hospitals'] });
  });

  it('invalida cache do mapa quando recebe evento status_update', () => {
    const { wrapper, queryClient } = createWrapper();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useMapSSEHandler(), { wrapper });

    const event: SSENotificationEvent = {
      type: 'status_update',
      occurrence_id: '123',
      hospital: 'Hospital Norte',
      hospital_id: '789',
      setor: 'Emergencia',
      tempo_restante_minutos: 180,
      timestamp: new Date().toISOString(),
    };

    act(() => {
      result.current(event);
    });

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['map', 'hospitals'] });
  });

  it('invalida cache do mapa quando recebe evento map_update', () => {
    const { wrapper, queryClient } = createWrapper();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useMapSSEHandler(), { wrapper });

    const event: SSENotificationEvent = {
      type: 'map_update',
      occurrence_id: '123',
      hospital: 'Hospital Sul',
      setor: 'CCU',
      tempo_restante_minutos: 240,
      timestamp: new Date().toISOString(),
    };

    act(() => {
      result.current(event);
    });

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['map', 'hospitals'] });
  });
});
