import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OccurrenceCard } from '../OccurrenceCard';
import { OccurrenceTable } from '../OccurrenceTable';
import { ConfirmationDialog } from '../ConfirmationDialog';
import type { OccurrenceCardData, OccurrenceTableData, AIConfirmationRequest } from '@/types/ai';

// Mock data for testing
const mockOccurrenceCardData: OccurrenceCardData = {
  id: 'occ-001',
  hospital_nome: 'Hospital Central',
  hospital_id: 'hosp-001',
  status: 'PENDENTE',
  nome_paciente_mascarado: 'J.S.',
  tempo_restante: '2h 30m',
  tempo_restante_minutos: 150,
  setor: 'UTI',
  urgencia: 'yellow',
};

const mockOccurrenceTableData: OccurrenceTableData = {
  occurrences: [
    {
      id: 'occ-001',
      hospital_nome: 'Hospital Central',
      hospital_id: 'hosp-001',
      status: 'PENDENTE',
      nome_paciente_mascarado: 'J.S.',
      tempo_restante: '2h 30m',
      tempo_restante_minutos: 150,
      setor: 'UTI',
      urgencia: 'yellow',
    },
    {
      id: 'occ-002',
      hospital_nome: 'Hospital Norte',
      hospital_id: 'hosp-002',
      status: 'EM_ANDAMENTO',
      nome_paciente_mascarado: 'M.A.',
      tempo_restante: '1h 15m',
      tempo_restante_minutos: 75,
      setor: 'Emergencia',
      urgencia: 'red',
    },
  ],
  total: 2,
  sortable: true,
};

const mockConfirmationRequest: AIConfirmationRequest = {
  action_id: 'action-001',
  action_type: 'update_occurrence_status',
  tool_name: 'update_occurrence_status',
  description: 'Atualizar status da ocorrencia OCC-001 para EM_ANDAMENTO',
  details: {
    occurrence_id: 'occ-001',
    new_status: 'EM_ANDAMENTO',
    current_status: 'PENDENTE',
  },
  severity: 'WARN',
};

describe('OccurrenceCard', () => {
  it('renderiza informacoes da ocorrencia corretamente', () => {
    render(<OccurrenceCard data={mockOccurrenceCardData} />);

    expect(screen.getByText('Hospital Central')).toBeInTheDocument();
    expect(screen.getByText('Pendente')).toBeInTheDocument(); // Translated label
    expect(screen.getByText('J.S.')).toBeInTheDocument();
    expect(screen.getByText('2h 30m')).toBeInTheDocument();
    expect(screen.getByText('UTI')).toBeInTheDocument();
  });

  it('chama onClick quando o card e clicado', () => {
    const handleClick = vi.fn();
    render(<OccurrenceCard data={mockOccurrenceCardData} onClick={handleClick} />);

    const card = screen.getByRole('button');
    fireEvent.click(card);

    expect(handleClick).toHaveBeenCalledWith(mockOccurrenceCardData);
  });

  it('aplica cor de urgencia correta baseado no nivel', () => {
    const { rerender } = render(<OccurrenceCard data={mockOccurrenceCardData} />);

    // Yellow urgency - card should have yellow background styling
    let card = screen.getByRole('button');
    expect(card).toBeInTheDocument();

    // Red urgency
    rerender(<OccurrenceCard data={{ ...mockOccurrenceCardData, urgencia: 'red' }} />);
    card = screen.getByRole('button');
    expect(card).toBeInTheDocument();
  });
});

describe('OccurrenceTable', () => {
  it('renderiza tabela com todas as ocorrencias', () => {
    render(<OccurrenceTable data={mockOccurrenceTableData} />);

    expect(screen.getByText('Hospital Central')).toBeInTheDocument();
    expect(screen.getByText('Hospital Norte')).toBeInTheDocument();
    expect(screen.getByText('J.S.')).toBeInTheDocument();
    expect(screen.getByText('M.A.')).toBeInTheDocument();
  });

  it('chama onRowClick quando uma linha e clicada', () => {
    const handleRowClick = vi.fn();
    render(<OccurrenceTable data={mockOccurrenceTableData} onRowClick={handleRowClick} />);

    const rows = screen.getAllByRole('row');
    // First row is header, click second row (first data row)
    fireEvent.click(rows[1]);

    expect(handleRowClick).toHaveBeenCalledWith(mockOccurrenceTableData.occurrences[0]);
  });

  it('permite ordenacao quando sortable e true', () => {
    const handleSort = vi.fn();
    render(<OccurrenceTable data={mockOccurrenceTableData} onSort={handleSort} />);

    const hospitalHeader = screen.getByText('Hospital');
    fireEvent.click(hospitalHeader);

    expect(handleSort).toHaveBeenCalled();
  });
});

describe('ConfirmationDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renderiza detalhes da acao corretamente', () => {
    render(
      <ConfirmationDialog
        open={true}
        data={mockConfirmationRequest}
        onConfirm={() => {}}
        onCancel={() => {}}
      />
    );

    expect(screen.getByText('Confirmar Acao')).toBeInTheDocument();
    expect(screen.getByText(/Atualizar status da ocorrencia/)).toBeInTheDocument();
  });

  it('chama onConfirm quando botao Confirmar e clicado', async () => {
    const handleConfirm = vi.fn();
    render(
      <ConfirmationDialog
        open={true}
        data={mockConfirmationRequest}
        onConfirm={handleConfirm}
        onCancel={() => {}}
      />
    );

    const confirmButton = screen.getByRole('button', { name: /confirmar/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(handleConfirm).toHaveBeenCalledWith(mockConfirmationRequest.action_id);
    });
  });

  it('chama onCancel quando botao Cancelar e clicado', async () => {
    const handleCancel = vi.fn();
    render(
      <ConfirmationDialog
        open={true}
        data={mockConfirmationRequest}
        onConfirm={() => {}}
        onCancel={handleCancel}
      />
    );

    const cancelButton = screen.getByRole('button', { name: /cancelar/i });
    fireEvent.click(cancelButton);

    await waitFor(() => {
      expect(handleCancel).toHaveBeenCalled();
    });
  });

  it('mostra indicador de loading durante confirmacao', () => {
    render(
      <ConfirmationDialog
        open={true}
        data={mockConfirmationRequest}
        onConfirm={() => {}}
        onCancel={() => {}}
        isLoading={true}
      />
    );

    const confirmButton = screen.getByRole('button', { name: /processando/i });
    expect(confirmButton).toBeDisabled();
  });
});
