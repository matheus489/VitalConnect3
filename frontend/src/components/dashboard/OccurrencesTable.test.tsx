import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { OccurrencesTable } from './OccurrencesTable';
import type { Occurrence } from '@/types';

const mockOccurrences: Occurrence[] = [
  {
    id: '1',
    obito_id: 'obito-1',
    hospital_id: 'hospital-1',
    hospital: {
      id: 'hospital-1',
      nome: 'Hospital Geral de Goiania',
      codigo: 'HGG',
      endereco: 'Rua Test, 123',
      config_conexao: {},
      ativo: true,
      created_at: '2026-01-15T10:00:00Z',
      updated_at: '2026-01-15T10:00:00Z',
    },
    status: 'PENDENTE',
    score_priorizacao: 100,
    nome_paciente_mascarado: 'Jo** Si***',
    dados_completos: {
      nome_paciente: 'Joao Silva',
      data_nascimento: '1970-01-01',
      data_obito: '2026-01-15T10:00:00Z',
      causa_mortis: 'Infarto',
      prontuario: '12345',
      setor: 'UTI',
      leito: '3',
      identificacao_desconhecida: false,
      idade: 56,
    },
    created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
    updated_at: new Date().toISOString(),
  },
  {
    id: '2',
    obito_id: 'obito-2',
    hospital_id: 'hospital-2',
    hospital: {
      id: 'hospital-2',
      nome: 'Hospital de Urgencias de Goias',
      codigo: 'HUGO',
      endereco: 'Av Test, 456',
      config_conexao: {},
      ativo: true,
      created_at: '2026-01-15T10:00:00Z',
      updated_at: '2026-01-15T10:00:00Z',
    },
    status: 'EM_ANDAMENTO',
    score_priorizacao: 80,
    nome_paciente_mascarado: 'Ma*** Pe*****',
    dados_completos: {
      nome_paciente: 'Maria Pereira',
      data_nascimento: '1960-05-15',
      data_obito: '2026-01-15T09:00:00Z',
      causa_mortis: 'AVC',
      prontuario: '67890',
      setor: 'Emergencia',
      leito: '5',
      identificacao_desconhecida: false,
      idade: 65,
    },
    created_at: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(), // 3 hours ago
    updated_at: new Date().toISOString(),
  },
];

describe('OccurrencesTable', () => {
  it('renders table with occurrence data', () => {
    const onViewDetails = vi.fn();
    const onStatusChange = vi.fn();
    const onComplete = vi.fn();

    render(
      <OccurrencesTable
        occurrences={mockOccurrences}
        onViewDetails={onViewDetails}
        onStatusChange={onStatusChange}
        onComplete={onComplete}
      />
    );

    // Check that hospital codes are rendered
    expect(screen.getByText('HGG')).toBeInTheDocument();
    expect(screen.getByText('HUGO')).toBeInTheDocument();

    // Check that masked patient names are rendered
    expect(screen.getByText('Jo** Si***')).toBeInTheDocument();
    expect(screen.getByText('Ma*** Pe*****')).toBeInTheDocument();

    // Check that sectors are rendered
    expect(screen.getByText('UTI')).toBeInTheDocument();
    expect(screen.getByText('Emergencia')).toBeInTheDocument();

    // Check that status badges are rendered
    expect(screen.getByText('Pendente')).toBeInTheDocument();
    expect(screen.getByText('Em Andamento')).toBeInTheDocument();
  });

  it('calls onViewDetails when view button is clicked', () => {
    const onViewDetails = vi.fn();
    const onStatusChange = vi.fn();
    const onComplete = vi.fn();

    render(
      <OccurrencesTable
        occurrences={mockOccurrences}
        onViewDetails={onViewDetails}
        onStatusChange={onStatusChange}
        onComplete={onComplete}
      />
    );

    const viewButtons = screen.getAllByLabelText('Ver detalhes');
    fireEvent.click(viewButtons[0]);

    expect(onViewDetails).toHaveBeenCalledWith('1');
  });

  it('shows empty state when no occurrences', () => {
    const onViewDetails = vi.fn();
    const onStatusChange = vi.fn();
    const onComplete = vi.fn();

    render(
      <OccurrencesTable
        occurrences={[]}
        onViewDetails={onViewDetails}
        onStatusChange={onStatusChange}
        onComplete={onComplete}
      />
    );

    expect(screen.getByText('Nenhuma ocorrencia encontrada')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    const onViewDetails = vi.fn();
    const onStatusChange = vi.fn();
    const onComplete = vi.fn();

    render(
      <OccurrencesTable
        occurrences={[]}
        isLoading={true}
        onViewDetails={onViewDetails}
        onStatusChange={onStatusChange}
        onComplete={onComplete}
      />
    );

    // Loading state shows skeleton rows
    const tableRows = screen.getAllByRole('row');
    expect(tableRows.length).toBeGreaterThan(1); // Header + skeleton rows
  });
});
