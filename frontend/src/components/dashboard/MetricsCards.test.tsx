import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MetricsCards } from './MetricsCards';

// Mock useMetrics hook
vi.mock('@/hooks/useMetrics', () => ({
  useMetrics: vi.fn(),
}));

import { useMetrics } from '@/hooks/useMetrics';

const mockUseMetrics = useMetrics as ReturnType<typeof vi.fn>;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
});

function renderMetricsCards() {
  return render(
    <QueryClientProvider client={queryClient}>
      <MetricsCards />
    </QueryClientProvider>
  );
}

describe('MetricsCards', () => {
  it('renders loading state', () => {
    mockUseMetrics.mockReturnValue({
      data: null,
      isLoading: true,
      isError: false,
      refetch: vi.fn(),
      isFetching: false,
    });

    renderMetricsCards();

    // Loading state shows skeleton cards (via animate-pulse class)
    const cards = document.querySelectorAll('.animate-pulse');
    expect(cards.length).toBeGreaterThan(0);
  });

  it('renders error state', () => {
    mockUseMetrics.mockReturnValue({
      data: null,
      isLoading: false,
      isError: true,
      refetch: vi.fn(),
      isFetching: false,
    });

    renderMetricsCards();

    expect(screen.getByText(/erro ao carregar metricas/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /tentar novamente/i })).toBeInTheDocument();
  });

  it('renders metrics data correctly', () => {
    mockUseMetrics.mockReturnValue({
      data: {
        obitos_elegiveis_hoje: 5,
        tempo_medio_notificacao_segundos: 120,
        corneas_potenciais: 10,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
      isFetching: false,
    });

    renderMetricsCards();

    // Check that metrics values are displayed
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('2m 0s')).toBeInTheDocument(); // 120 seconds = 2 minutes
    expect(screen.getByText('10')).toBeInTheDocument();

    // Check card titles
    expect(screen.getByText(/obitos elegiveis/i)).toBeInTheDocument();
    expect(screen.getByText(/tempo medio de notificacao/i)).toBeInTheDocument();
    expect(screen.getByText(/corneas potenciais/i)).toBeInTheDocument();
  });

  it('formats time correctly for different values', () => {
    mockUseMetrics.mockReturnValue({
      data: {
        obitos_elegiveis_hoje: 3,
        tempo_medio_notificacao_segundos: 45,
        corneas_potenciais: 6,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
      isFetching: false,
    });

    renderMetricsCards();

    // 45 seconds should be displayed as "45s"
    expect(screen.getByText('45s')).toBeInTheDocument();
  });

  it('formats time correctly for hours', () => {
    mockUseMetrics.mockReturnValue({
      data: {
        obitos_elegiveis_hoje: 1,
        tempo_medio_notificacao_segundos: 3720, // 1 hour 2 minutes
        corneas_potenciais: 2,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
      isFetching: false,
    });

    renderMetricsCards();

    expect(screen.getByText('1h 2m')).toBeInTheDocument();
  });
});
