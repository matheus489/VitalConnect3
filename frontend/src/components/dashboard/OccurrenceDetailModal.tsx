'use client';

import { format, parseISO } from 'date-fns';
import { ptBR } from 'date-fns/locale';
import { Play, Check, X, Ban, Flag, Clock, User, Building2, Activity } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { StatusBadge, getStatusConfig } from './StatusBadge';
import { useOccurrenceDetail, useOccurrenceHistory } from '@/hooks/useOccurrences';
import type { OccurrenceStatus } from '@/types';
import { cn } from '@/lib/utils';

interface OccurrenceDetailModalProps {
  occurrenceId: string | null;
  open: boolean;
  onClose: () => void;
  onStatusChange: (id: string, status: OccurrenceStatus) => void;
  onComplete: (id: string) => void;
}

function formatDate(dateString: string): string {
  try {
    return format(parseISO(dateString), "dd/MM/yyyy 'as' HH:mm", { locale: ptBR });
  } catch {
    return dateString;
  }
}

function calculateAge(birthDate: string): number {
  const birth = parseISO(birthDate);
  const today = new Date();
  let age = today.getFullYear() - birth.getFullYear();
  const monthDiff = today.getMonth() - birth.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    age--;
  }
  return age;
}

export function OccurrenceDetailModal({
  occurrenceId,
  open,
  onClose,
  onStatusChange,
  onComplete,
}: OccurrenceDetailModalProps) {
  const { data: occurrence, isLoading } = useOccurrenceDetail(occurrenceId);
  const { data: history } = useOccurrenceHistory(occurrenceId);

  const getActionButtons = () => {
    if (!occurrence) return null;

    const buttons: Array<{
      status?: OccurrenceStatus;
      action?: 'complete';
      label: string;
      icon: React.ElementType;
      variant: 'default' | 'outline' | 'destructive' | 'secondary';
    }> = [];

    switch (occurrence.status) {
      case 'PENDENTE':
        buttons.push(
          { status: 'EM_ANDAMENTO', label: 'Assumir Ocorrencia', icon: Play, variant: 'default' },
          { status: 'CANCELADA', label: 'Cancelar', icon: Ban, variant: 'secondary' }
        );
        break;
      case 'EM_ANDAMENTO':
        buttons.push(
          { status: 'ACEITA', label: 'Aceitar', icon: Check, variant: 'default' },
          { status: 'RECUSADA', label: 'Recusar', icon: X, variant: 'secondary' },
          { status: 'CANCELADA', label: 'Cancelar', icon: Ban, variant: 'secondary' }
        );
        break;
      case 'ACEITA':
      case 'RECUSADA':
        buttons.push(
          { action: 'complete', label: 'Concluir', icon: Flag, variant: 'default' },
          { status: 'CANCELADA', label: 'Cancelar', icon: Ban, variant: 'secondary' }
        );
        break;
      default:
        break;
    }

    return buttons;
  };

  const actionButtons = getActionButtons();

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3">
            Detalhes da Ocorrencia
            {occurrence && <StatusBadge status={occurrence.status} />}
          </DialogTitle>
          <DialogDescription>
            Informacoes completas e historico de acoes
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-6 animate-pulse rounded bg-muted" />
            ))}
          </div>
        ) : occurrence?.dados_completos ? (
          <div className="space-y-6">
            {/* Patient Info */}
            <div className="rounded-lg border p-4">
              <h3 className="mb-3 flex items-center gap-2 font-semibold">
                <User className="h-4 w-4 text-primary" />
                Dados do Paciente
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <p className="text-sm text-muted-foreground">Nome Completo</p>
                  <p className="font-medium">{occurrence.dados_completos.nome_paciente}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Prontuario</p>
                  <p className="font-medium">{occurrence.dados_completos.prontuario}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Data de Nascimento</p>
                  <p className="font-medium">
                    {formatDate(occurrence.dados_completos.data_nascimento)} (
                    {calculateAge(occurrence.dados_completos.data_nascimento)} anos)
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Identificacao</p>
                  <p className="font-medium">
                    {occurrence.dados_completos.identificacao_desconhecida
                      ? 'Desconhecida'
                      : 'Identificado'}
                  </p>
                </div>
              </div>
            </div>

            {/* Death Info */}
            <div className="rounded-lg border p-4">
              <h3 className="mb-3 flex items-center gap-2 font-semibold">
                <Clock className="h-4 w-4 text-primary" />
                Dados do Obito
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <p className="text-sm text-muted-foreground">Data/Hora do Obito</p>
                  <p className="font-medium">
                    {formatDate(occurrence.dados_completos.data_obito)}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Causa Mortis</p>
                  <p className="font-medium">{occurrence.dados_completos.causa_mortis}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Setor</p>
                  <p className="font-medium">{occurrence.dados_completos.setor}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Leito</p>
                  <p className="font-medium">{occurrence.dados_completos.leito}</p>
                </div>
              </div>
            </div>

            {/* Hospital Info */}
            <div className="rounded-lg border p-4">
              <h3 className="mb-3 flex items-center gap-2 font-semibold">
                <Building2 className="h-4 w-4 text-primary" />
                Hospital
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <p className="text-sm text-muted-foreground">Nome</p>
                  <p className="font-medium">{occurrence.hospital?.nome || 'N/A'}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Codigo</p>
                  <p className="font-medium">{occurrence.hospital?.codigo || 'N/A'}</p>
                </div>
              </div>
            </div>

            {/* Occurrence Info */}
            <div className="rounded-lg border p-4">
              <h3 className="mb-3 flex items-center gap-2 font-semibold">
                <Activity className="h-4 w-4 text-primary" />
                Informacoes da Ocorrencia
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <p className="text-sm text-muted-foreground">Score de Priorizacao</p>
                  <p className="font-medium">{occurrence.score_priorizacao}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Notificado em</p>
                  <p className="font-medium">
                    {occurrence.notificado_em
                      ? formatDate(occurrence.notificado_em)
                      : 'Aguardando'}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Criado em</p>
                  <p className="font-medium">{formatDate(occurrence.created_at)}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Atualizado em</p>
                  <p className="font-medium">{formatDate(occurrence.updated_at)}</p>
                </div>
              </div>
            </div>

            {/* History */}
            {history && history.length > 0 && (
              <div className="rounded-lg border p-4">
                <h3 className="mb-3 font-semibold">Historico de Acoes</h3>
                <div className="space-y-3">
                  {history.map((item, index) => {
                    const statusConfig = item.status_novo
                      ? getStatusConfig(item.status_novo)
                      : null;

                    return (
                      <div
                        key={item.id}
                        className={cn(
                          'flex items-start gap-3 pb-3',
                          index < history.length - 1 && 'border-b'
                        )}
                      >
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
                          {index + 1}
                        </div>
                        <div className="flex-1 space-y-1">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="font-medium">{item.acao}</span>
                            {item.status_novo && statusConfig && (
                              <Badge variant="outline" className={cn('text-xs', statusConfig.className)}>
                                {statusConfig.label}
                              </Badge>
                            )}
                          </div>
                          {item.observacoes && (
                            <p className="text-sm text-muted-foreground">{item.observacoes}</p>
                          )}
                          <p className="text-xs text-muted-foreground">
                            {formatDate(item.created_at)}
                            {item.user && ` - por ${item.user.nome}`}
                          </p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Action Buttons */}
            {actionButtons && actionButtons.length > 0 && (
              <div className="flex flex-wrap justify-end gap-2 pt-4 border-t">
                {actionButtons.map((btn) => (
                  <Button
                    key={btn.status || btn.action}
                    variant={btn.variant}
                    onClick={() => {
                      if (btn.action === 'complete') {
                        onComplete(occurrence.id);
                        onClose();
                      } else if (btn.status) {
                        onStatusChange(occurrence.id, btn.status);
                        onClose();
                      }
                    }}
                  >
                    <btn.icon className="mr-2 h-4 w-4" />
                    {btn.label}
                  </Button>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-8">
            <p className="text-muted-foreground">Dados nao encontrados</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
