'use client';

import { useState } from 'react';
import { Eye, Play, Check, X, Ban, Flag } from 'lucide-react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { StatusBadge } from './StatusBadge';
import type { Occurrence, OccurrenceStatus } from '@/types';
import { cn } from '@/lib/utils';
import { formatDistanceToNow, differenceInMinutes, parseISO } from 'date-fns';
import { ptBR } from 'date-fns/locale';

interface OccurrencesTableProps {
  occurrences: Occurrence[];
  isLoading?: boolean;
  onViewDetails: (id: string) => void;
  onStatusChange: (id: string, status: OccurrenceStatus) => void;
  onComplete: (id: string) => void;
}

function calculateTimeRemaining(createdAt: string): { minutes: number; label: string; isUrgent: boolean } {
  const created = parseISO(createdAt);
  const deadline = new Date(created.getTime() + 6 * 60 * 60 * 1000); // 6 hours from creation
  const now = new Date();
  const minutes = differenceInMinutes(deadline, now);

  if (minutes <= 0) {
    return { minutes: 0, label: 'Expirado', isUrgent: true };
  }

  if (minutes < 60) {
    return { minutes, label: `${minutes}min`, isUrgent: minutes <= 30 };
  }

  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  return {
    minutes,
    label: `${hours}h ${remainingMinutes}min`,
    isUrgent: hours < 1,
  };
}

function getAvailableActions(status: OccurrenceStatus): Array<{
  action: OccurrenceStatus | 'view' | 'complete';
  label: string;
  icon: React.ElementType;
  variant?: 'default' | 'outline' | 'destructive' | 'secondary';
}> {
  const viewAction = {
    action: 'view' as const,
    label: 'Ver detalhes',
    icon: Eye,
    variant: 'outline' as const,
  };

  switch (status) {
    case 'PENDENTE':
      return [
        viewAction,
        { action: 'EM_ANDAMENTO' as const, label: 'Assumir', icon: Play, variant: 'default' as const },
        { action: 'CANCELADA' as const, label: 'Cancelar', icon: Ban, variant: 'secondary' as const },
      ];
    case 'EM_ANDAMENTO':
      return [
        viewAction,
        { action: 'ACEITA' as const, label: 'Aceitar', icon: Check, variant: 'default' as const },
        { action: 'RECUSADA' as const, label: 'Recusar', icon: X, variant: 'secondary' as const },
        { action: 'CANCELADA' as const, label: 'Cancelar', icon: Ban, variant: 'secondary' as const },
      ];
    case 'ACEITA':
      return [
        viewAction,
        { action: 'complete' as const, label: 'Concluir', icon: Flag, variant: 'default' as const },
        { action: 'CANCELADA' as const, label: 'Cancelar', icon: Ban, variant: 'secondary' as const },
      ];
    case 'RECUSADA':
      return [
        viewAction,
        { action: 'complete' as const, label: 'Concluir', icon: Flag, variant: 'default' as const },
        { action: 'CANCELADA' as const, label: 'Cancelar', icon: Ban, variant: 'secondary' as const },
      ];
    default:
      return [viewAction];
  }
}

export function OccurrencesTable({
  occurrences,
  isLoading,
  onViewDetails,
  onStatusChange,
  onComplete,
}: OccurrencesTableProps) {
  const [expandedActions, setExpandedActions] = useState<string | null>(null);

  if (isLoading) {
    return (
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hospital</TableHead>
              <TableHead>Setor</TableHead>
              <TableHead>Paciente</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Tempo Restante</TableHead>
              <TableHead className="text-right">Acoes</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {[1, 2, 3, 4, 5].map((i) => (
              <TableRow key={i}>
                {[1, 2, 3, 4, 5, 6].map((j) => (
                  <TableCell key={j}>
                    <div className="h-4 w-20 animate-pulse rounded bg-muted" />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  if (occurrences.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-md border py-16">
        <p className="text-muted-foreground">Nenhuma ocorrencia encontrada</p>
      </div>
    );
  }

  return (
    <div className="rounded-md border overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="min-w-[120px]">Hospital</TableHead>
            <TableHead className="min-w-[100px]">Setor</TableHead>
            <TableHead className="min-w-[150px]">Paciente</TableHead>
            <TableHead className="min-w-[120px]">Status</TableHead>
            <TableHead className="min-w-[120px]">Tempo Restante</TableHead>
            <TableHead className="text-right min-w-[200px]">Acoes</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {occurrences.map((occurrence) => {
            const timeRemaining = calculateTimeRemaining(occurrence.created_at);
            const actions = getAvailableActions(occurrence.status);
            const isExpanded = expandedActions === occurrence.id;

            return (
              <TableRow key={occurrence.id}>
                <TableCell className="font-medium">
                  {occurrence.hospital?.codigo || 'N/A'}
                </TableCell>
                <TableCell>
                  {occurrence.dados_completos?.setor || 'N/A'}
                </TableCell>
                <TableCell>{occurrence.nome_paciente_mascarado}</TableCell>
                <TableCell>
                  <StatusBadge status={occurrence.status} />
                </TableCell>
                <TableCell>
                  <span
                    className={cn(
                      'font-medium',
                      timeRemaining.isUrgent && 'text-destructive'
                    )}
                  >
                    {timeRemaining.label}
                  </span>
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-1">
                    {/* Primary action - View */}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onViewDetails(occurrence.id)}
                      aria-label="Ver detalhes"
                    >
                      <Eye className="h-4 w-4" />
                    </Button>

                    {/* Status action buttons */}
                    {actions.slice(1).map((actionConfig) => (
                      <Button
                        key={actionConfig.action}
                        variant={actionConfig.variant}
                        size="sm"
                        onClick={() => {
                          if (actionConfig.action === 'complete') {
                            onComplete(occurrence.id);
                          } else {
                            onStatusChange(occurrence.id, actionConfig.action as OccurrenceStatus);
                          }
                        }}
                        aria-label={actionConfig.label}
                        className={cn(
                          !isExpanded && actions.length > 3 && 'hidden sm:inline-flex'
                        )}
                      >
                        <actionConfig.icon className="h-4 w-4" />
                        <span className="hidden lg:inline ml-1">{actionConfig.label}</span>
                      </Button>
                    ))}

                    {/* Mobile expand button */}
                    {actions.length > 3 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="sm:hidden"
                        onClick={() =>
                          setExpandedActions(isExpanded ? null : occurrence.id)
                        }
                      >
                        ...
                      </Button>
                    )}
                  </div>

                  {/* Expanded mobile actions */}
                  {isExpanded && actions.length > 3 && (
                    <div className="mt-2 flex flex-col gap-1 sm:hidden">
                      {actions.slice(1).map((actionConfig) => (
                        <Button
                          key={actionConfig.action}
                          variant={actionConfig.variant}
                          size="sm"
                          onClick={() => {
                            if (actionConfig.action === 'complete') {
                              onComplete(occurrence.id);
                            } else {
                              onStatusChange(occurrence.id, actionConfig.action as OccurrenceStatus);
                            }
                            setExpandedActions(null);
                          }}
                          className="w-full justify-start"
                        >
                          <actionConfig.icon className="mr-2 h-4 w-4" />
                          {actionConfig.label}
                        </Button>
                      ))}
                    </div>
                  )}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
