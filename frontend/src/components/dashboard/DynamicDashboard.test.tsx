import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DynamicDashboard } from './DynamicDashboard';
import type { DashboardWidget } from '@/types/theme';

// Mock useTenantTheme
vi.mock('@/contexts/TenantThemeContext', () => ({
  useTenantTheme: vi.fn(),
}));

// Mock child widgets
vi.mock('@/components/dashboard/MetricsCards', () => ({
  MetricsCards: () => <div data-testid="metrics-cards">Metrics Cards</div>,
}));

// Mock all hooks from useOccurrences
vi.mock('@/hooks/useOccurrences', () => ({
  useOccurrences: () => ({
    data: { data: [], meta: { total: 0, page: 1, per_page: 10, total_pages: 1 } },
    isLoading: false,
  }),
  useOccurrenceDetail: () => ({
    data: null,
    isLoading: false,
  }),
  useOccurrenceHistory: () => ({
    data: [],
    isLoading: false,
  }),
  useUpdateOccurrenceStatus: () => ({ mutateAsync: vi.fn() }),
  useRegisterOutcome: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

// Mock OccurrencesTable and related components
vi.mock('@/components/dashboard/OccurrencesTable', () => ({
  OccurrencesTable: ({ occurrences }: { occurrences: unknown[] }) => (
    <div data-testid="occurrences-table">
      Occurrences Table ({occurrences?.length || 0} items)
    </div>
  ),
}));

vi.mock('@/components/dashboard/OccurrenceFilters', () => ({
  OccurrenceFilters: () => <div data-testid="occurrence-filters">Filters</div>,
}));

vi.mock('@/components/dashboard/OccurrenceDetailModal', () => ({
  OccurrenceDetailModal: () => null,
}));

vi.mock('@/components/dashboard/OutcomeModal', () => ({
  OutcomeModal: () => null,
}));

vi.mock('@/components/dashboard/Pagination', () => ({
  Pagination: () => <div data-testid="pagination">Pagination</div>,
}));

import { useTenantTheme } from '@/contexts/TenantThemeContext';

const mockUseTenantTheme = useTenantTheme as ReturnType<typeof vi.fn>;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
});

function renderDynamicDashboard() {
  return render(
    <QueryClientProvider client={queryClient}>
      <DynamicDashboard />
    </QueryClientProvider>
  );
}

describe('DynamicDashboard', () => {
  it('renders widgets based on config order', () => {
    const widgets: DashboardWidget[] = [
      { id: 'stats', type: 'stats_card', visible: true, order: 1, title: 'Estatisticas' },
      { id: 'recent', type: 'recent_occurrences', visible: true, order: 2, title: 'Recentes' },
    ];

    mockUseTenantTheme.mockReturnValue({
      themeConfig: {
        layout: { dashboard_widgets: widgets },
      },
      isLoading: false,
    });

    renderDynamicDashboard();

    // Should render stats card (which renders MetricsCards mock)
    expect(screen.getByTestId('metrics-cards')).toBeInTheDocument();
    // Should render occurrences table
    expect(screen.getByTestId('occurrences-table')).toBeInTheDocument();
  });

  it('hides widgets based on visible flag', () => {
    const widgets: DashboardWidget[] = [
      { id: 'stats', type: 'stats_card', visible: true, order: 1, title: 'Estatisticas' },
      { id: 'map', type: 'map_preview', visible: false, order: 2, title: 'Mapa' },
    ];

    mockUseTenantTheme.mockReturnValue({
      themeConfig: {
        layout: { dashboard_widgets: widgets },
      },
      isLoading: false,
    });

    renderDynamicDashboard();

    expect(screen.getByTestId('metrics-cards')).toBeInTheDocument();
    // Map preview should not be rendered since visible: false
    // The map preview card contains title text
    expect(screen.queryByText('Mapa de Ocorrencias')).not.toBeInTheDocument();
  });

  it('renders default dashboard when no config exists', () => {
    mockUseTenantTheme.mockReturnValue({
      themeConfig: null,
      isLoading: false,
    });

    renderDynamicDashboard();

    // Default dashboard should show metrics cards
    expect(screen.getByTestId('metrics-cards')).toBeInTheDocument();
  });

  it('sorts widgets by order property', () => {
    const widgets: DashboardWidget[] = [
      { id: 'second', type: 'recent_occurrences', visible: true, order: 2, title: 'Segundo' },
      { id: 'first', type: 'stats_card', visible: true, order: 1, title: 'Primeiro' },
    ];

    mockUseTenantTheme.mockReturnValue({
      themeConfig: {
        layout: { dashboard_widgets: widgets },
      },
      isLoading: false,
    });

    const { container } = renderDynamicDashboard();

    // Both should be present
    expect(screen.getByTestId('metrics-cards')).toBeInTheDocument();
    expect(screen.getByTestId('occurrences-table')).toBeInTheDocument();

    // Verify order by checking data-widget-id attributes
    const widgetElements = container.querySelectorAll('[data-widget-id]');
    expect(widgetElements.length).toBe(2);
    // First widget should be 'first' (stats_card with order 1)
    expect(widgetElements[0].getAttribute('data-widget-id')).toBe('first');
    // Second widget should be 'second' (recent_occurrences with order 2)
    expect(widgetElements[1].getAttribute('data-widget-id')).toBe('second');
  });
});
