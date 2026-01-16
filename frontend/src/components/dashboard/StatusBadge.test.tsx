import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusBadge, getStatusConfig } from './StatusBadge';

describe('StatusBadge', () => {
  it('renders PENDENTE status with correct label', () => {
    render(<StatusBadge status="PENDENTE" />);
    expect(screen.getByText('Pendente')).toBeInTheDocument();
  });

  it('renders EM_ANDAMENTO status with correct label', () => {
    render(<StatusBadge status="EM_ANDAMENTO" />);
    expect(screen.getByText('Em Andamento')).toBeInTheDocument();
  });

  it('renders ACEITA status with correct label', () => {
    render(<StatusBadge status="ACEITA" />);
    expect(screen.getByText('Aceita')).toBeInTheDocument();
  });

  it('renders RECUSADA status with correct label', () => {
    render(<StatusBadge status="RECUSADA" />);
    expect(screen.getByText('Recusada')).toBeInTheDocument();
  });

  it('renders CANCELADA status with correct label', () => {
    render(<StatusBadge status="CANCELADA" />);
    expect(screen.getByText('Cancelada')).toBeInTheDocument();
  });

  it('renders CONCLUIDA status with correct label', () => {
    render(<StatusBadge status="CONCLUIDA" />);
    expect(screen.getByText('Concluida')).toBeInTheDocument();
  });
});

describe('getStatusConfig', () => {
  it('returns correct config for PENDENTE', () => {
    const config = getStatusConfig('PENDENTE');
    expect(config.label).toBe('Pendente');
    expect(config.variant).toBe('destructive');
    expect(config.className).toContain('animate-pulse-alert');
  });

  it('returns correct config for EM_ANDAMENTO', () => {
    const config = getStatusConfig('EM_ANDAMENTO');
    expect(config.label).toBe('Em Andamento');
    expect(config.variant).toBe('default');
  });

  it('returns correct config for ACEITA', () => {
    const config = getStatusConfig('ACEITA');
    expect(config.label).toBe('Aceita');
    expect(config.variant).toBe('secondary');
    expect(config.className).toContain('emerald');
  });

  it('returns correct config for CONCLUIDA', () => {
    const config = getStatusConfig('CONCLUIDA');
    expect(config.label).toBe('Concluida');
    expect(config.variant).toBe('secondary');
    expect(config.className).toContain('sky');
  });
});
