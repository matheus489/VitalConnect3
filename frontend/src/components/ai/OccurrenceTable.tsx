'use client';

import { useState } from 'react';
import { ArrowUpDown, ArrowUp, ArrowDown, Clock, MapPin } from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import type { OccurrenceCardData, OccurrenceTableData } from '@/types/ai';

type SortField = 'hospital_nome' | 'status' | 'tempo_restante_minutos' | 'setor';
type SortDirection = 'asc' | 'desc';

interface SortConfig {
  field: SortField;
  direction: SortDirection;
}

interface OccurrenceTableProps {
  data: OccurrenceTableData;
  onRowClick?: (occurrence: OccurrenceCardData) => void;
  onSort?: (field: SortField, direction: SortDirection) => void;
  className?: string;
}

const statusLabels: Record<string, string> = {
  PENDENTE: 'Pendente',
  EM_ANDAMENTO: 'Em Andamento',
  ACEITA: 'Aceita',
  RECUSADA: 'Recusada',
  CANCELADA: 'Cancelada',
  CONCLUIDA: 'Concluida',
};

const statusBadgeVariants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  PENDENTE: 'outline',
  EM_ANDAMENTO: 'default',
  ACEITA: 'default',
  RECUSADA: 'destructive',
  CANCELADA: 'secondary',
  CONCLUIDA: 'default',
};

const urgencyColors: Record<string, string> = {
  green: 'text-green-600 dark:text-green-400',
  yellow: 'text-yellow-600 dark:text-yellow-400',
  red: 'text-red-600 dark:text-red-400',
};

export function OccurrenceTable({
  data,
  onRowClick,
  onSort,
  className,
}: OccurrenceTableProps) {
  const [sortConfig, setSortConfig] = useState<SortConfig | null>(null);
  const { occurrences, sortable = false } = data;

  const handleSort = (field: SortField) => {
    if (!sortable) return;

    const newDirection: SortDirection =
      sortConfig?.field === field && sortConfig.direction === 'asc' ? 'desc' : 'asc';

    setSortConfig({ field, direction: newDirection });
    onSort?.(field, newDirection);
  };

  const getSortIcon = (field: SortField) => {
    if (!sortable) return null;

    if (sortConfig?.field !== field) {
      return <ArrowUpDown className="h-4 w-4 ml-1 opacity-50" />;
    }

    return sortConfig.direction === 'asc' ? (
      <ArrowUp className="h-4 w-4 ml-1" />
    ) : (
      <ArrowDown className="h-4 w-4 ml-1" />
    );
  };

  const handleRowClick = (occurrence: OccurrenceCardData) => {
    onRowClick?.(occurrence);
  };

  const handleKeyDown = (event: React.KeyboardEvent, occurrence: OccurrenceCardData) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onRowClick?.(occurrence);
    }
  };

  // Sort occurrences locally if sortConfig is set
  const sortedOccurrences = sortConfig
    ? [...occurrences].sort((a, b) => {
        const aValue = a[sortConfig.field];
        const bValue = b[sortConfig.field];

        if (aValue === undefined || bValue === undefined) return 0;

        let comparison = 0;
        if (typeof aValue === 'string' && typeof bValue === 'string') {
          comparison = aValue.localeCompare(bValue, 'pt-BR');
        } else if (typeof aValue === 'number' && typeof bValue === 'number') {
          comparison = aValue - bValue;
        }

        return sortConfig.direction === 'asc' ? comparison : -comparison;
      })
    : occurrences;

  if (occurrences.length === 0) {
    return (
      <div className={cn('text-center py-6 text-muted-foreground', className)}>
        Nenhuma ocorrencia encontrada.
      </div>
    );
  }

  return (
    <div className={cn('rounded-lg border', className)}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead
              className={cn(sortable && 'cursor-pointer select-none hover:bg-muted/50')}
              onClick={() => handleSort('hospital_nome')}
            >
              <div className="flex items-center">
                Hospital
                {getSortIcon('hospital_nome')}
              </div>
            </TableHead>
            <TableHead
              className={cn(sortable && 'cursor-pointer select-none hover:bg-muted/50')}
              onClick={() => handleSort('status')}
            >
              <div className="flex items-center">
                Status
                {getSortIcon('status')}
              </div>
            </TableHead>
            <TableHead>Paciente</TableHead>
            <TableHead
              className={cn(sortable && 'cursor-pointer select-none hover:bg-muted/50')}
              onClick={() => handleSort('setor')}
            >
              <div className="flex items-center">
                Setor
                {getSortIcon('setor')}
              </div>
            </TableHead>
            <TableHead
              className={cn(sortable && 'cursor-pointer select-none hover:bg-muted/50')}
              onClick={() => handleSort('tempo_restante_minutos')}
            >
              <div className="flex items-center">
                Tempo Restante
                {getSortIcon('tempo_restante_minutos')}
              </div>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sortedOccurrences.map((occurrence) => {
            const urgency = occurrence.urgencia || 'green';
            const timeColor = urgencyColors[urgency] || urgencyColors.green;
            const statusLabel = statusLabels[occurrence.status] || occurrence.status;
            const badgeVariant = statusBadgeVariants[occurrence.status] || 'outline';

            return (
              <TableRow
                key={occurrence.id}
                className={cn(onRowClick && 'cursor-pointer')}
                tabIndex={onRowClick ? 0 : undefined}
                onClick={() => handleRowClick(occurrence)}
                onKeyDown={(e) => handleKeyDown(e, occurrence)}
              >
                <TableCell>
                  <div className="flex items-center gap-2">
                    <MapPin className="h-4 w-4 text-muted-foreground shrink-0" />
                    <span className="font-medium">{occurrence.hospital_nome}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant={badgeVariant}>{statusLabel}</Badge>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {occurrence.nome_paciente_mascarado}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {occurrence.setor || '-'}
                </TableCell>
                <TableCell>
                  <div className={cn('flex items-center gap-1 font-medium', timeColor)}>
                    <Clock className="h-4 w-4" />
                    {occurrence.tempo_restante}
                  </div>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
      {data.total > occurrences.length && (
        <div className="px-4 py-2 text-sm text-muted-foreground border-t">
          Mostrando {occurrences.length} de {data.total} ocorrencias
        </div>
      )}
    </div>
  );
}

export default OccurrenceTable;
