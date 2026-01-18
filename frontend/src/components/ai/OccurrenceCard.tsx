'use client';

import { Clock, MapPin, User, Activity } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { OccurrenceCardData } from '@/types/ai';

interface OccurrenceCardProps {
  data: OccurrenceCardData;
  onClick?: (data: OccurrenceCardData) => void;
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

const urgencyColors: Record<string, { bg: string; text: string; border: string }> = {
  green: {
    bg: 'bg-green-50 dark:bg-green-950/30',
    text: 'text-green-700 dark:text-green-400',
    border: 'border-green-200 dark:border-green-800',
  },
  yellow: {
    bg: 'bg-yellow-50 dark:bg-yellow-950/30',
    text: 'text-yellow-700 dark:text-yellow-400',
    border: 'border-yellow-200 dark:border-yellow-800',
  },
  red: {
    bg: 'bg-red-50 dark:bg-red-950/30',
    text: 'text-red-700 dark:text-red-400',
    border: 'border-red-200 dark:border-red-800',
  },
};

const statusBadgeVariants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  PENDENTE: 'outline',
  EM_ANDAMENTO: 'default',
  ACEITA: 'default',
  RECUSADA: 'destructive',
  CANCELADA: 'secondary',
  CONCLUIDA: 'default',
};

export function OccurrenceCard({ data, onClick, className }: OccurrenceCardProps) {
  const urgency = data.urgencia || 'green';
  const colors = urgencyColors[urgency] || urgencyColors.green;
  const statusLabel = statusLabels[data.status] || data.status;
  const badgeVariant = statusBadgeVariants[data.status] || 'outline';

  const handleClick = () => {
    onClick?.(data);
  };

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onClick?.(data);
    }
  };

  return (
    <Card
      data-slot="occurrence-card"
      role="button"
      tabIndex={onClick ? 0 : undefined}
      onClick={onClick ? handleClick : undefined}
      onKeyDown={onClick ? handleKeyDown : undefined}
      className={cn(
        'py-3 transition-all',
        colors.bg,
        colors.border,
        onClick && 'cursor-pointer hover:shadow-md focus-visible:ring-2 focus-visible:ring-ring',
        className
      )}
    >
      <CardContent className="p-3 space-y-2">
        {/* Header: Hospital and Status */}
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0">
            <MapPin className={cn('h-4 w-4 shrink-0', colors.text)} />
            <span className="font-medium text-sm truncate">{data.hospital_nome}</span>
          </div>
          <Badge variant={badgeVariant} className="shrink-0 text-xs">
            {statusLabel}
          </Badge>
        </div>

        {/* Patient Info */}
        <div className="flex items-center gap-2 text-muted-foreground">
          <User className="h-3.5 w-3.5 shrink-0" />
          <span className="text-sm">{data.nome_paciente_mascarado}</span>
          {data.setor && (
            <>
              <span className="text-muted-foreground/50">|</span>
              <span className="text-sm">{data.setor}</span>
            </>
          )}
        </div>

        {/* Time Remaining */}
        <div className="flex items-center gap-2">
          <Clock className={cn('h-4 w-4 shrink-0', colors.text)} />
          <span className={cn('text-sm font-medium', colors.text)}>{data.tempo_restante}</span>
          {urgency === 'red' && (
            <Activity className={cn('h-4 w-4 animate-pulse', colors.text)} />
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default OccurrenceCard;
