import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import type { MapHospital, UrgencyLevel } from '@/types';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

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
    latitude: -16.65,
    longitude: -49.21,
    ativo: true,
    urgencia_maxima: 'none' as UrgencyLevel,
    ocorrencias_count: 0,
    ocorrencias: [],
    operador_plantao: null,
  },
];

// Import the HospitalDrawer component (this doesn't depend on Leaflet)
import { HospitalDrawer } from './HospitalDrawer';

describe('HospitalDrawer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('exibe informacoes do hospital quando aberto', async () => {
    render(
      <HospitalDrawer
        hospital={mockHospitals[0]}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Hospital Central')).toBeInTheDocument();
    });
    // Check for occurrences count text (the number appears separately)
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('Dr. Carlos Silva')).toBeInTheDocument();
  });

  it('exibe lista de ocorrencias com detalhes', async () => {
    render(
      <HospitalDrawer
        hospital={mockHospitals[0]}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('J.S.')).toBeInTheDocument();
    });
    expect(screen.getByText(/UTI/)).toBeInTheDocument();
    expect(screen.getByText(/1h 30m/)).toBeInTheDocument();
  });

  it('exibe mensagem quando hospital nao tem operador de plantao', async () => {
    render(
      <HospitalDrawer
        hospital={mockHospitals[1]}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Hospital Norte')).toBeInTheDocument();
    });
    expect(screen.getByText(/Nao definido/i)).toBeInTheDocument();
  });

  it('exibe mensagem quando hospital nao tem ocorrencias', async () => {
    render(
      <HospitalDrawer
        hospital={mockHospitals[1]}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Hospital Norte')).toBeInTheDocument();
    });
    expect(screen.getByText(/Nenhuma ocorrencia ativa/i)).toBeInTheDocument();
  });

  it('nao renderiza quando hospital e null', () => {
    const { container } = render(
      <HospitalDrawer
        hospital={null}
        open={true}
        onClose={vi.fn()}
      />
    );

    // The drawer should not render any content when hospital is null
    expect(container.querySelector('[data-slot="sheet-content"]')).not.toBeInTheDocument();
  });
});

// Test map utility functions
import { calculateUrgencyLevel, getUrgencyColor, getUrgencyLabel, formatTimeRemaining } from '@/lib/map-utils';

describe('Funcoes utilitarias para marcadores do mapa', () => {
  it('calculateUrgencyLevel retorna cores corretas baseadas no tempo', () => {
    // >= 240 minutos (>= 4 horas) = green
    expect(calculateUrgencyLevel(300)).toBe('green');
    expect(calculateUrgencyLevel(241)).toBe('green');
    expect(calculateUrgencyLevel(240)).toBe('green'); // Exactly 4 hours is green

    // 120-239 minutos (2-4 horas) = yellow
    expect(calculateUrgencyLevel(239)).toBe('yellow');
    expect(calculateUrgencyLevel(180)).toBe('yellow');
    expect(calculateUrgencyLevel(120)).toBe('yellow');

    // < 120 minutos (< 2 horas) = red
    expect(calculateUrgencyLevel(119)).toBe('red');
    expect(calculateUrgencyLevel(60)).toBe('red');
    expect(calculateUrgencyLevel(0)).toBe('red');
    expect(calculateUrgencyLevel(-10)).toBe('red');
  });

  it('getUrgencyColor retorna cores hex corretas', () => {
    expect(getUrgencyColor('green')).toBe('#22c55e');
    expect(getUrgencyColor('yellow')).toBe('#eab308');
    expect(getUrgencyColor('red')).toBe('#ef4444');
    expect(getUrgencyColor('none')).toBe('#6b7280');
  });

  it('getUrgencyLabel retorna labels em portugues', () => {
    expect(getUrgencyLabel('green')).toBe('Normal');
    expect(getUrgencyLabel('yellow')).toBe('Atencao');
    expect(getUrgencyLabel('red')).toBe('Critico');
    expect(getUrgencyLabel('none')).toBe('Sem ocorrencias');
  });

  it('formatTimeRemaining formata tempo corretamente', () => {
    expect(formatTimeRemaining(150)).toBe('2h 30m');
    expect(formatTimeRemaining(60)).toBe('1h 0m');
    expect(formatTimeRemaining(45)).toBe('45m');
    expect(formatTimeRemaining(0)).toBe('0m');
    expect(formatTimeRemaining(-10)).toBe('Expirado');
  });
});
