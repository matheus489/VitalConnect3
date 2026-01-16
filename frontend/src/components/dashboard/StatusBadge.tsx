'use client';

import { Badge } from '@/components/ui/badge';
import type { OccurrenceStatus } from '@/types';
import { cn } from '@/lib/utils';

interface StatusConfig {
  label: string;
  variant: 'default' | 'secondary' | 'destructive' | 'outline';
  className?: string;
}

export function getStatusConfig(status: OccurrenceStatus): StatusConfig {
  switch (status) {
    case 'PENDENTE':
      return {
        label: 'Pendente',
        variant: 'destructive',
        className: 'animate-pulse-alert',
      };
    case 'EM_ANDAMENTO':
      return {
        label: 'Em Andamento',
        variant: 'default',
      };
    case 'ACEITA':
      return {
        label: 'Aceita',
        variant: 'secondary',
        className: 'bg-emerald-100 text-emerald-800 border-emerald-200',
      };
    case 'RECUSADA':
      return {
        label: 'Recusada',
        variant: 'secondary',
        className: 'bg-amber-100 text-amber-800 border-amber-200',
      };
    case 'CANCELADA':
      return {
        label: 'Cancelada',
        variant: 'secondary',
        className: 'bg-gray-100 text-gray-600 border-gray-200',
      };
    case 'CONCLUIDA':
      return {
        label: 'Concluida',
        variant: 'secondary',
        className: 'bg-sky-100 text-sky-800 border-sky-200',
      };
    default:
      return {
        label: status,
        variant: 'outline',
      };
  }
}

interface StatusBadgeProps {
  status: OccurrenceStatus;
  className?: string;
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const config = getStatusConfig(status);

  return (
    <Badge
      variant={config.variant}
      className={cn(config.className, className)}
    >
      {config.label}
    </Badge>
  );
}
