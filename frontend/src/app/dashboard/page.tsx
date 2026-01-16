'use client';

import { useState } from 'react';
import { toast } from 'sonner';
import { MetricsCards } from '@/components/dashboard/MetricsCards';
import { OccurrencesTable } from '@/components/dashboard/OccurrencesTable';
import { OccurrenceFilters } from '@/components/dashboard/OccurrenceFilters';
import { OccurrenceDetailModal } from '@/components/dashboard/OccurrenceDetailModal';
import { OutcomeModal } from '@/components/dashboard/OutcomeModal';
import { Pagination } from '@/components/dashboard/Pagination';
import {
  useOccurrences,
  useUpdateOccurrenceStatus,
  useRegisterOutcome,
} from '@/hooks/useOccurrences';
import type { OccurrenceFilters as FiltersType, OccurrenceStatus, SortField, SortOrder, OutcomeType } from '@/types';

export default function DashboardPage() {
  // State
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(10);
  const [filters, setFilters] = useState<FiltersType>({});
  const [sortBy, setSortBy] = useState<SortField>('created_at');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [selectedOccurrenceId, setSelectedOccurrenceId] = useState<string | null>(null);
  const [outcomeOccurrenceId, setOutcomeOccurrenceId] = useState<string | null>(null);

  // Queries & Mutations
  const { data: occurrencesData, isLoading } = useOccurrences({
    page,
    perPage,
    filters,
    sortBy,
    sortOrder,
  });
  const updateStatus = useUpdateOccurrenceStatus();
  const registerOutcome = useRegisterOutcome();

  // Handlers
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
    <div className="space-y-6">
      {/* Page Title */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Visao geral das ocorrencias e metricas do sistema
        </p>
      </div>

      {/* Metrics Cards */}
      <MetricsCards />

      {/* Occurrences Section */}
      <div className="space-y-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-lg font-semibold">Ocorrencias Recentes</h2>
        </div>

        {/* Filters */}
        <OccurrenceFilters
          filters={filters}
          onFiltersChange={handleFiltersChange}
          sortBy={sortBy}
          sortOrder={sortOrder}
          onSortChange={handleSortChange}
        />

        {/* Table */}
        <OccurrencesTable
          occurrences={occurrencesData?.data || []}
          isLoading={isLoading}
          onViewDetails={handleViewDetails}
          onStatusChange={handleStatusChange}
          onComplete={handleComplete}
        />

        {/* Pagination */}
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
      </div>

      {/* Modals */}
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
