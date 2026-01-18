'use client';

import { useState, useMemo } from 'react';
import { toast } from 'sonner';
import { RefreshCw, MapPin, ClipboardList, BarChart2, Activity } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useTenantTheme } from '@/contexts/TenantThemeContext';
import { MetricsCards } from './MetricsCards';
import { OccurrencesTable } from './OccurrencesTable';
import { OccurrenceFilters } from './OccurrenceFilters';
import { OccurrenceDetailModal } from './OccurrenceDetailModal';
import { OutcomeModal } from './OutcomeModal';
import { Pagination } from './Pagination';
import {
  useOccurrences,
  useUpdateOccurrenceStatus,
  useRegisterOutcome,
} from '@/hooks/useOccurrences';
import {
  DEFAULT_THEME_CONFIG,
  type DashboardWidget,
  type DashboardWidgetType,
} from '@/types/theme';
import type {
  OccurrenceFilters as FiltersType,
  OccurrenceStatus,
  SortField,
  SortOrder,
  OutcomeType,
} from '@/types';

/**
 * Props for individual widget components
 */
interface WidgetProps {
  widget: DashboardWidget;
}

/**
 * Stats Card Widget - Renders the MetricsCards component
 */
function StatsCardWidget({ widget }: WidgetProps) {
  return (
    <div className="w-full">
      <MetricsCards />
    </div>
  );
}

/**
 * Map Preview Widget - Placeholder for map preview
 */
function MapPreviewWidget({ widget }: WidgetProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {widget.title || 'Mapa de Ocorrencias'}
        </CardTitle>
        <MapPin className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex h-48 items-center justify-center rounded-lg border border-dashed border-muted-foreground/25">
          <div className="text-center">
            <MapPin className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <p className="mt-2 text-sm text-muted-foreground">
              Visualizacao do mapa
            </p>
            <Button
              variant="link"
              size="sm"
              className="mt-1"
              onClick={() => window.location.href = '/dashboard/map'}
            >
              Abrir mapa completo
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Recent Occurrences Widget - Renders recent occurrences table
 */
function RecentOccurrencesWidget({ widget }: WidgetProps) {
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(5);
  const [filters, setFilters] = useState<FiltersType>({});
  const [sortBy, setSortBy] = useState<SortField>('created_at');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [selectedOccurrenceId, setSelectedOccurrenceId] = useState<string | null>(null);
  const [outcomeOccurrenceId, setOutcomeOccurrenceId] = useState<string | null>(null);

  const { data: occurrencesData, isLoading } = useOccurrences({
    page,
    perPage,
    filters,
    sortBy,
    sortOrder,
  });

  const updateStatus = useUpdateOccurrenceStatus();
  const registerOutcome = useRegisterOutcome();

  const handleFiltersChange = (newFilters: FiltersType) => {
    setFilters(newFilters);
    setPage(1);
  };

  const handleSortChange = (field: SortField, order: SortOrder) => {
    setSortBy(field);
    setSortOrder(order);
    setPage(1);
  };

  const handleViewDetails = (id: string) => {
    setSelectedOccurrenceId(id);
  };

  const handleStatusChange = async (id: string, status: OccurrenceStatus) => {
    try {
      await updateStatus.mutateAsync({ id, status });
      toast.success(`Status atualizado para ${status}`);
    } catch {
      toast.error('Erro ao atualizar status');
    }
  };

  const handleComplete = (id: string) => {
    setOutcomeOccurrenceId(id);
  };

  const handleOutcomeConfirm = async (id: string, outcome: OutcomeType, observacoes: string) => {
    try {
      await registerOutcome.mutateAsync({ id, desfecho: outcome, observacoes });
      toast.success('Desfecho registrado com sucesso');
    } catch {
      toast.error('Erro ao registrar desfecho');
    }
  };

  const handlePerPageChange = (newPerPage: number) => {
    setPerPage(newPerPage);
    setPage(1);
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-lg font-semibold">
          {widget.title || 'Ocorrencias Recentes'}
        </h2>
      </div>

      <OccurrenceFilters
        filters={filters}
        onFiltersChange={handleFiltersChange}
        sortBy={sortBy}
        sortOrder={sortOrder}
        onSortChange={handleSortChange}
      />

      <OccurrencesTable
        occurrences={occurrencesData?.data || []}
        isLoading={isLoading}
        onViewDetails={handleViewDetails}
        onStatusChange={handleStatusChange}
        onComplete={handleComplete}
      />

      {occurrencesData?.meta && (
        <Pagination
          currentPage={page}
          totalPages={occurrencesData.meta.total_pages || 1}
          perPage={perPage}
          totalItems={occurrencesData.meta.total || 0}
          onPageChange={setPage}
          onPerPageChange={handlePerPageChange}
        />
      )}

      <OccurrenceDetailModal
        occurrenceId={selectedOccurrenceId}
        open={!!selectedOccurrenceId}
        onClose={() => setSelectedOccurrenceId(null)}
        onStatusChange={handleStatusChange}
        onComplete={handleComplete}
      />

      <OutcomeModal
        occurrenceId={outcomeOccurrenceId}
        open={!!outcomeOccurrenceId}
        onClose={() => setOutcomeOccurrenceId(null)}
        onConfirm={handleOutcomeConfirm}
        isLoading={registerOutcome.isPending}
      />
    </div>
  );
}

/**
 * Chart Widget - Placeholder for chart visualization
 */
function ChartWidget({ widget }: WidgetProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {widget.title || 'Graficos e Estatisticas'}
        </CardTitle>
        <BarChart2 className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex h-48 items-center justify-center rounded-lg border border-dashed border-muted-foreground/25">
          <div className="text-center">
            <BarChart2 className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <p className="mt-2 text-sm text-muted-foreground">
              Graficos de desempenho
            </p>
            <Button
              variant="link"
              size="sm"
              className="mt-1"
              onClick={() => window.location.href = '/dashboard/reports'}
            >
              Ver relatorios
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Activity Feed Widget - Placeholder for activity feed
 */
function ActivityFeedWidget({ widget }: WidgetProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {widget.title || 'Atividade Recente'}
        </CardTitle>
        <Activity className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex h-48 items-center justify-center rounded-lg border border-dashed border-muted-foreground/25">
          <div className="text-center">
            <Activity className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <p className="mt-2 text-sm text-muted-foreground">
              Feed de atividades
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Quick Actions Widget - Placeholder for quick actions
 */
function QuickActionsWidget({ widget }: WidgetProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {widget.title || 'Acoes Rapidas'}
        </CardTitle>
        <RefreshCw className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" size="sm">
            Nova Ocorrencia
          </Button>
          <Button variant="outline" size="sm">
            Ver Mapa
          </Button>
          <Button variant="outline" size="sm">
            Relatorios
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Map widget type to component
 */
const widgetComponents: Record<DashboardWidgetType, React.ComponentType<WidgetProps>> = {
  stats_card: StatsCardWidget,
  map_preview: MapPreviewWidget,
  recent_occurrences: RecentOccurrencesWidget,
  chart: ChartWidget,
  activity_feed: ActivityFeedWidget,
  quick_actions: QuickActionsWidget,
};

/**
 * DynamicDashboard component that renders widgets based on theme_config.
 *
 * Features:
 * - Reads widgets from theme_config.layout.dashboard_widgets
 * - Renders widget grid based on order and visibility
 * - Supports widget types: stats_card, map_preview, recent_occurrences, chart
 * - Conditional rendering based on visible flag
 * - Falls back to default dashboard if no config
 *
 * @example
 * ```tsx
 * // In dashboard page
 * export default function DashboardPage() {
 *   return (
 *     <div className="space-y-6">
 *       <h1>Dashboard</h1>
 *       <DynamicDashboard />
 *     </div>
 *   );
 * }
 * ```
 */
export function DynamicDashboard() {
  const { themeConfig, isLoading } = useTenantTheme();

  // Get dashboard widgets from config or use defaults
  const widgets: DashboardWidget[] = useMemo(() => {
    const configWidgets = themeConfig?.layout?.dashboard_widgets;
    if (configWidgets && configWidgets.length > 0) {
      return configWidgets;
    }
    return DEFAULT_THEME_CONFIG.layout?.dashboard_widgets || [];
  }, [themeConfig]);

  // Filter visible widgets and sort by order
  const visibleWidgets = useMemo(() => {
    return [...widgets]
      .filter((widget) => widget.visible !== false)
      .sort((a, b) => (a.order ?? 999) - (b.order ?? 999));
  }, [widgets]);

  if (isLoading) {
    return (
      <div className="space-y-6">
        {/* Skeleton for metrics */}
        <div className="grid gap-4 md:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader className="pb-2">
                <div className="h-4 w-32 bg-muted rounded" />
              </CardHeader>
              <CardContent>
                <div className="h-8 w-20 bg-muted rounded" />
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Skeleton for table */}
        <Card className="animate-pulse">
          <CardHeader>
            <div className="h-6 w-48 bg-muted rounded" />
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-12 bg-muted rounded" />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Title */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Visao geral das ocorrencias e metricas do sistema
        </p>
      </div>

      {/* Render widgets based on configuration */}
      {visibleWidgets.map((widget) => {
        const WidgetComponent = widgetComponents[widget.type];

        if (!WidgetComponent) {
          console.warn(`Unknown widget type: ${widget.type}`);
          return null;
        }

        return (
          <div key={widget.id} data-widget-id={widget.id}>
            <WidgetComponent widget={widget} />
          </div>
        );
      })}
    </div>
  );
}

export default DynamicDashboard;
