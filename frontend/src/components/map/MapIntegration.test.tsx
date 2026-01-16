import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import type { MapHospital, UrgencyLevel } from '@/types';
import { HospitalDrawer } from './HospitalDrawer';

// Mock next/navigation
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

// Mock data with different urgency levels
const createMockHospital = (urgencyLevel: UrgencyLevel): MapHospital => ({
  id: '1',
  nome: 'Hospital Teste',
  codigo: 'HT001',
  latitude: -16.6799,
  longitude: -49.2556,
  ativo: true,
  urgencia_maxima: urgencyLevel,
  ocorrencias_count: urgencyLevel === 'none' ? 0 : 1,
  ocorrencias: urgencyLevel === 'none' ? [] : [
    {
      id: '101',
      nome_mascarado: 'J.S.',
      setor: 'UTI',
      tempo_restante: urgencyLevel === 'red' ? '1h 30m' : urgencyLevel === 'yellow' ? '3h 0m' : '5h 0m',
      tempo_restante_minutos: urgencyLevel === 'red' ? 90 : urgencyLevel === 'yellow' ? 180 : 300,
      status: 'PENDENTE',
      urgencia: urgencyLevel === 'none' ? 'green' : urgencyLevel,
    },
  ],
  operador_plantao: {
    id: '201',
    nome: 'Dr. Carlos Silva',
  },
});

describe('Fluxo de Integracao do Mapa', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPush.mockClear();
  });

  it('navega para detalhes da ocorrencia ao clicar em "Ver Detalhes"', async () => {
    const hospital = createMockHospital('red');
    const onClose = vi.fn();

    render(
      <HospitalDrawer
        hospital={hospital}
        open={true}
        onClose={onClose}
      />
    );

    // Wait for drawer content to render
    await waitFor(() => {
      expect(screen.getByText('Hospital Teste')).toBeInTheDocument();
    });

    // Find and click the "Ver Detalhes" button
    const detailsButton = screen.getByRole('button', { name: /Ver Detalhes/i });
    fireEvent.click(detailsButton);

    // Verify navigation was called with correct URL
    expect(mockPush).toHaveBeenCalledWith('/dashboard/occurrences?id=101');
    expect(onClose).toHaveBeenCalled();
  });

  it('exibe badge de urgencia correta para nivel vermelho (Critico)', async () => {
    const hospital = createMockHospital('red');

    render(
      <HospitalDrawer
        hospital={hospital}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      // Use getAllByText since both hospital and occurrence urgency badges show 'Critico'
      const criticos = screen.getAllByText('Critico');
      expect(criticos.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('exibe badge de urgencia correta para nivel amarelo (Atencao)', async () => {
    const hospital = createMockHospital('yellow');

    render(
      <HospitalDrawer
        hospital={hospital}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      // Use getAllByText since both hospital and occurrence urgency badges show 'Atencao'
      const atencoes = screen.getAllByText('Atencao');
      expect(atencoes.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('exibe badge de urgencia correta para nivel verde (Normal)', async () => {
    const hospital = createMockHospital('green');

    render(
      <HospitalDrawer
        hospital={hospital}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      // Use getAllByText since both hospital and occurrence urgency badges show 'Normal'
      const normals = screen.getAllByText('Normal');
      expect(normals.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('atualiza dados do drawer quando hospital muda de urgencia', async () => {
    const hospitalRed = createMockHospital('red');
    const hospitalYellow: MapHospital = {
      ...createMockHospital('yellow'),
      nome: 'Hospital Atualizado',
      ocorrencias: [
        {
          id: '102',
          nome_mascarado: 'M.A.',
          setor: 'Emergencia',
          tempo_restante: '3h 0m',
          tempo_restante_minutos: 180,
          status: 'EM_ANDAMENTO',
          urgencia: 'yellow',
        },
      ],
    };

    const { rerender } = render(
      <HospitalDrawer
        hospital={hospitalRed}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Hospital Teste')).toBeInTheDocument();
    });

    // Verify red urgency is displayed
    expect(screen.getAllByText('Critico').length).toBeGreaterThanOrEqual(1);

    // Rerender with updated hospital data
    rerender(
      <HospitalDrawer
        hospital={hospitalYellow}
        open={true}
        onClose={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Hospital Atualizado')).toBeInTheDocument();
    });

    // Verify yellow urgency is now displayed and red is gone
    expect(screen.getAllByText('Atencao').length).toBeGreaterThanOrEqual(1);
    expect(screen.queryByText('Critico')).not.toBeInTheDocument();
  });
});
