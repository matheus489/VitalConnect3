'use client';

import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useHospitals } from '@/hooks/useHospitals';
import type { OccurrenceFilters as FiltersType, OccurrenceStatus, SortField, SortOrder } from '@/types';

interface OccurrenceFiltersProps {
  filters: FiltersType;
  onFiltersChange: (filters: FiltersType) => void;
  sortBy: SortField;
  sortOrder: SortOrder;
  onSortChange: (field: SortField, order: SortOrder) => void;
}

const statusOptions: { value: OccurrenceStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'Todos os status' },
  { value: 'PENDENTE', label: 'Pendente' },
  { value: 'EM_ANDAMENTO', label: 'Em Andamento' },
  { value: 'ACEITA', label: 'Aceita' },
  { value: 'RECUSADA', label: 'Recusada' },
  { value: 'CANCELADA', label: 'Cancelada' },
  { value: 'CONCLUIDA', label: 'Concluida' },
];

const sortOptions: { value: SortField; label: string }[] = [
  { value: 'created_at', label: 'Data de Criacao' },
  { value: 'score_priorizacao', label: 'Prioridade' },
  { value: 'tempo_restante', label: 'Tempo Restante' },
];

export function OccurrenceFilters({
  filters,
  onFiltersChange,
  sortBy,
  sortOrder,
  onSortChange,
}: OccurrenceFiltersProps) {
  const { data: hospitals } = useHospitals();

  const handleStatusChange = (value: string) => {
    onFiltersChange({
      ...filters,
      status: value === 'all' ? undefined : value as OccurrenceStatus,
    });
  };

  const handleHospitalChange = (value: string) => {
    onFiltersChange({
      ...filters,
      hospital_id: value === 'all' ? undefined : value || undefined,
    });
  };

  const handleDateFromChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onFiltersChange({
      ...filters,
      date_from: e.target.value || undefined,
    });
  };

  const handleDateToChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onFiltersChange({
      ...filters,
      date_to: e.target.value || undefined,
    });
  };

  const handleSortFieldChange = (value: string) => {
    onSortChange(value as SortField, sortOrder);
  };

  const handleSortOrderChange = (value: string) => {
    onSortChange(sortBy, value as SortOrder);
  };

  const clearFilters = () => {
    onFiltersChange({});
    onSortChange('created_at', 'desc');
  };

  const hasFilters =
    filters.status ||
    filters.hospital_id ||
    filters.date_from ||
    filters.date_to;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-3">
        {/* Status Filter */}
        <div className="w-full sm:w-auto">
          <Select value={filters.status || 'all'} onValueChange={handleStatusChange}>
            <SelectTrigger className="w-full sm:w-[180px]">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              {statusOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Hospital Filter */}
        <div className="w-full sm:w-auto">
          <Select value={filters.hospital_id || 'all'} onValueChange={handleHospitalChange}>
            <SelectTrigger className="w-full sm:w-[200px]">
              <SelectValue placeholder="Hospital" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Todos os hospitais</SelectItem>
              {hospitals?.map((hospital) => (
                <SelectItem key={hospital.id} value={hospital.id}>
                  {hospital.codigo} - {hospital.nome}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Date From */}
        <div className="w-full sm:w-auto">
          <Input
            type="date"
            value={filters.date_from || ''}
            onChange={handleDateFromChange}
            className="w-full sm:w-[160px]"
            aria-label="Data inicial"
          />
        </div>

        {/* Date To */}
        <div className="w-full sm:w-auto">
          <Input
            type="date"
            value={filters.date_to || ''}
            onChange={handleDateToChange}
            className="w-full sm:w-[160px]"
            aria-label="Data final"
          />
        </div>

        {/* Sort Field */}
        <div className="w-full sm:w-auto">
          <Select value={sortBy} onValueChange={handleSortFieldChange}>
            <SelectTrigger className="w-full sm:w-[180px]">
              <SelectValue placeholder="Ordenar por" />
            </SelectTrigger>
            <SelectContent>
              {sortOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Sort Order */}
        <div className="w-full sm:w-auto">
          <Select value={sortOrder} onValueChange={handleSortOrderChange}>
            <SelectTrigger className="w-full sm:w-[140px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="desc">Decrescente</SelectItem>
              <SelectItem value="asc">Crescente</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Clear Filters */}
        {hasFilters && (
          <Button variant="outline" size="icon" onClick={clearFilters} aria-label="Limpar filtros">
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  );
}
