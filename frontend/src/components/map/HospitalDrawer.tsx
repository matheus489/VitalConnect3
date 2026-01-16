'use client';

import { useRouter } from 'next/navigation';
import { Building2, Clock, User, AlertCircle, ChevronRight } from 'lucide-react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { MapHospital, MapOccurrence } from '@/types';
import { getUrgencyLabel, getUrgencyBadgeClasses } from '@/lib/map-utils';

interface HospitalDrawerProps {
  /**
   * Dados do hospital selecionado
   */
  hospital: MapHospital | null;
  /**
   * Se o drawer esta aberto
   */
  open: boolean;
  /**
   * Callback executado ao fechar o drawer
   */
  onClose: () => void;
}

/**
 * Componente para exibir o badge de status da ocorrencia
 */
function OccurrenceStatusBadge({ status }: { status: string }) {
  const getStatusConfig = (status: string) => {
    switch (status) {
      case 'PENDENTE':
        return { label: 'Pendente', className: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' };
      case 'EM_ANDAMENTO':
        return { label: 'Em Andamento', className: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' };
      case 'ACEITA':
        return { label: 'Aceita', className: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' };
      case 'RECUSADA':
        return { label: 'Recusada', className: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200' };
      default:
        return { label: status, className: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200' };
    }
  };

  const config = getStatusConfig(status);
  return (
    <Badge variant="secondary" className={cn('text-xs', config.className)}>
      {config.label}
    </Badge>
  );
}

/**
 * Item da lista de ocorrencias
 */
function OccurrenceItem({ occurrence, onViewDetails }: { occurrence: MapOccurrence; onViewDetails: (id: string) => void }) {
  return (
    <div className="flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50 transition-colors">
      <div className="flex-1 space-y-1">
        <div className="flex items-center gap-2">
          <span className="font-medium text-sm">{occurrence.nome_mascarado}</span>
          <OccurrenceStatusBadge status={occurrence.status} />
        </div>
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span>{occurrence.setor}</span>
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {occurrence.tempo_restante}
          </span>
        </div>
        <Badge
          variant="outline"
          className={cn('text-xs mt-1', getUrgencyBadgeClasses(occurrence.urgencia))}
        >
          {getUrgencyLabel(occurrence.urgencia)}
        </Badge>
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onViewDetails(occurrence.id)}
        className="shrink-0"
      >
        Ver Detalhes
        <ChevronRight className="ml-1 h-4 w-4" />
      </Button>
    </div>
  );
}

/**
 * Drawer lateral com detalhes do hospital e lista de ocorrencias
 *
 * Exibe:
 * - Nome e codigo do hospital
 * - Quantidade de ocorrencias ativas
 * - Operador de plantao atual
 * - Lista de ocorrencias com detalhes
 * - Botao para navegar aos detalhes de cada ocorrencia
 */
export function HospitalDrawer({ hospital, open, onClose }: HospitalDrawerProps) {
  const router = useRouter();

  const handleViewDetails = (occurrenceId: string) => {
    router.push(`/dashboard/occurrences?id=${occurrenceId}`);
    onClose();
  };

  if (!hospital) {
    return null;
  }

  return (
    <Sheet open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <SheetContent side="right" className="w-full sm:max-w-md overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-primary" />
            {hospital.nome}
          </SheetTitle>
          <SheetDescription>
            Codigo: {hospital.codigo}
          </SheetDescription>
        </SheetHeader>

        <div className="mt-6 space-y-6">
          {/* Summary */}
          <div className="grid grid-cols-2 gap-4">
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
                <AlertCircle className="h-3.5 w-3.5" />
                Ocorrencias Ativas
              </div>
              <p className="text-2xl font-bold">
                {hospital.ocorrencias_count}
                <span className="text-sm font-normal text-muted-foreground ml-1">
                  {hospital.ocorrencias_count === 1 ? 'ocorrencia' : 'ocorrencias'}
                </span>
              </p>
            </div>
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
                <User className="h-3.5 w-3.5" />
                Operador de Plantao
              </div>
              <p className="text-sm font-medium truncate">
                {hospital.operador_plantao?.nome || 'Nao definido'}
              </p>
            </div>
          </div>

          {/* Urgency Status */}
          {hospital.ocorrencias_count > 0 && (
            <div className="flex items-center justify-between rounded-lg border p-3 bg-muted/30">
              <span className="text-sm text-muted-foreground">Status de Urgencia</span>
              <Badge
                variant="outline"
                className={cn('text-xs', getUrgencyBadgeClasses(hospital.urgencia_maxima))}
              >
                {getUrgencyLabel(hospital.urgencia_maxima)}
              </Badge>
            </div>
          )}

          {/* Occurrences List */}
          <div className="space-y-3">
            <h3 className="font-semibold text-sm">
              Ocorrencias
              {(hospital.ocorrencias?.length ?? 0) > 0 && (
                <span className="text-muted-foreground font-normal ml-2">
                  ({hospital.ocorrencias?.length})
                </span>
              )}
            </h3>

            {!hospital.ocorrencias || hospital.ocorrencias.length === 0 ? (
              <div className="rounded-lg border border-dashed p-6 text-center">
                <AlertCircle className="h-8 w-8 mx-auto text-muted-foreground/50 mb-2" />
                <p className="text-sm text-muted-foreground">
                  Nenhuma ocorrencia ativa neste hospital
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {hospital.ocorrencias.map((occurrence) => (
                  <OccurrenceItem
                    key={occurrence.id}
                    occurrence={occurrence}
                    onViewDetails={handleViewDetails}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
