import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createElement } from 'react';
import type { MapHospital, UrgencyLevel } from '@/types';
import { calculateUrgencyLevel, getUrgencyColor, getUrgencyLabel, formatTimeRemaining } from '@/lib/map-utils';

// Mock the API module
vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
  },
}));

import { api } from '@/lib/api';
import { useMapHospitals } from './useMap';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const mockApiGet = api.get as Mock<any>;

// Helper to create a wrapper with QueryClientProvider
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
};

// Mock data for testing
const mockHospitals: MapHospital[] = [
  {
    id: '1',
    nome: 'Hospital Central',
    codigo: 'HC001',
    latitude: -16.6799,
    longitude: -49.2556,
    ativo: true,
    urgencia_maxima: 'red' as UrgencyLevel,
    ocorrencias_count: 2,
    ocorrencias: [
      {
        id: '101',
        nome_mascarado: 'J.S.',
        setor: 'UTI',
        tempo_restante: '1h 30m',
        tempo_restante_minutos: 90,
        status: 'PENDENTE',
        urgencia: 'red' as UrgencyLevel,
      },
      {
        id: '102',
        nome_mascarado: 'M.A.',
        setor: 'Emergencia',
        tempo_restante: '3h 00m',
        tempo_restante_minutos: 180,
        status: 'EM_ANDAMENTO',
        urgencia: 'yellow' as UrgencyLevel,
      },
    ],
    operador_plantao: {
      id: '201',
      nome: 'Dr. Carlos Silva',
    },
  },
  {
    id: '2',
    nome: 'Hospital Norte',
    codigo: 'HN002',
    latitude: -16.6500,
    longitude: -49.2100,
    ativo: true,
    urgencia_maxima: 'none' as UrgencyLevel,
    ocorrencias_count: 0,
    ocorrencias: [],
    operador_plantao: null,
  },
];

describe('useMapHospitals', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('retorna dados corretamente quando a API responde com sucesso', async () => {
    mockApiGet.mockResolvedValueOnce({ data: { hospitals: mockHospitals, total: 2 } });

    const { result } = renderHook(() => useMapHospitals(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual({ hospitals: mockHospitals, total: 2 });
    expect(result.current.error).toBeNull();
    expect(mockApiGet).toHaveBeenCalledWith('/map/hospitals');
  });

  it('trata erro corretamente quando a API falha', async () => {
    const apiError = new Error('Network error');
    mockApiGet.mockRejectedValueOnce(apiError);

    const { result } = renderHook(() => useMapHospitals(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeTruthy();
    expect(result.current.data).toBeUndefined();
  });

  it('usa refetchInterval configurado para atualizacao periodica', async () => {
    mockApiGet.mockResolvedValue({ data: { hospitals: mockHospitals, total: 2 } });

    const { result } = renderHook(() => useMapHospitals({ refetchInterval: 15000 }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // The hook should be configured with refetchInterval
    // This verifies the hook accepts the option without error
    expect(result.current.data).toBeDefined();
  });
});

describe('Funcoes utilitarias do mapa - calculateUrgencyLevel', () => {
  it('retorna "green" para tempo restante > 4 horas (> 240 minutos)', () => {
    expect(calculateUrgencyLevel(241)).toBe('green');
    expect(calculateUrgencyLevel(300)).toBe('green');
    expect(calculateUrgencyLevel(500)).toBe('green');
  });

  it('retorna "yellow" para tempo restante entre 2 e 4 horas (120-240 minutos)', () => {
    expect(calculateUrgencyLevel(120)).toBe('yellow');
    expect(calculateUrgencyLevel(180)).toBe('yellow');
    expect(calculateUrgencyLevel(239)).toBe('yellow');
  });

  it('retorna "red" para tempo restante < 2 horas (< 120 minutos)', () => {
    expect(calculateUrgencyLevel(119)).toBe('red');
    expect(calculateUrgencyLevel(60)).toBe('red');
    expect(calculateUrgencyLevel(0)).toBe('red');
    expect(calculateUrgencyLevel(-10)).toBe('red');
  });

  it('retorna "red" para valores negativos ou zero (tempo expirado)', () => {
    expect(calculateUrgencyLevel(0)).toBe('red');
    expect(calculateUrgencyLevel(-1)).toBe('red');
    expect(calculateUrgencyLevel(-100)).toBe('red');
  });
});

describe('Funcoes utilitarias do mapa - getUrgencyColor', () => {
  it('retorna cor verde correta para nivel green', () => {
    expect(getUrgencyColor('green')).toBe('#22c55e');
  });

  it('retorna cor amarela correta para nivel yellow', () => {
    expect(getUrgencyColor('yellow')).toBe('#eab308');
  });

  it('retorna cor vermelha correta para nivel red', () => {
    expect(getUrgencyColor('red')).toBe('#ef4444');
  });

  it('retorna cor cinza correta para nivel none', () => {
    expect(getUrgencyColor('none')).toBe('#6b7280');
  });
});

describe('Funcoes utilitarias do mapa - getUrgencyLabel', () => {
  it('retorna label em portugues para cada nivel de urgencia', () => {
    expect(getUrgencyLabel('green')).toBe('Normal');
    expect(getUrgencyLabel('yellow')).toBe('Atencao');
    expect(getUrgencyLabel('red')).toBe('Critico');
    expect(getUrgencyLabel('none')).toBe('Sem ocorrencias');
  });
});

describe('Funcoes utilitarias do mapa - formatTimeRemaining', () => {
  it('formata minutos corretamente para horas e minutos', () => {
    expect(formatTimeRemaining(90)).toBe('1h 30m');
    expect(formatTimeRemaining(120)).toBe('2h 0m');
    expect(formatTimeRemaining(180)).toBe('3h 0m');
  });

  it('formata apenas minutos quando menor que 1 hora', () => {
    expect(formatTimeRemaining(45)).toBe('45m');
    expect(formatTimeRemaining(5)).toBe('5m');
    expect(formatTimeRemaining(59)).toBe('59m');
  });

  it('trata valores zero ou negativos apropriadamente', () => {
    expect(formatTimeRemaining(0)).toBe('0m');
    expect(formatTimeRemaining(-10)).toBe('Expirado');
  });
});
