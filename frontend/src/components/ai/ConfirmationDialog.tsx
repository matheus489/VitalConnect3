'use client';

import { AlertTriangle, Info, AlertCircle, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import type { AIConfirmationRequest } from '@/types/ai';

interface ConfirmationDialogProps {
  open: boolean;
  data: AIConfirmationRequest;
  onConfirm: (actionId: string) => void;
  onCancel: () => void;
  isLoading?: boolean;
  className?: string;
}

const severityConfig: Record<
  string,
  { icon: typeof Info; color: string; bgColor: string }
> = {
  INFO: {
    icon: Info,
    color: 'text-blue-600 dark:text-blue-400',
    bgColor: 'bg-blue-50 dark:bg-blue-950/30',
  },
  WARN: {
    icon: AlertTriangle,
    color: 'text-yellow-600 dark:text-yellow-400',
    bgColor: 'bg-yellow-50 dark:bg-yellow-950/30',
  },
  CRITICAL: {
    icon: AlertCircle,
    color: 'text-red-600 dark:text-red-400',
    bgColor: 'bg-red-50 dark:bg-red-950/30',
  },
};

const actionTypeLabels: Record<string, string> = {
  update_occurrence_status: 'Atualizar Status',
  send_team_notification: 'Enviar Notificacao',
  generate_report: 'Gerar Relatorio',
};

export function ConfirmationDialog({
  open,
  data,
  onConfirm,
  onCancel,
  isLoading = false,
  className,
}: ConfirmationDialogProps) {
  const severity = severityConfig[data.severity] || severityConfig.INFO;
  const SeverityIcon = severity.icon;
  const actionLabel = actionTypeLabels[data.action_type] || data.action_type;

  const handleConfirm = () => {
    onConfirm(data.action_id);
  };

  const renderDetails = () => {
    if (!data.details || Object.keys(data.details).length === 0) {
      return null;
    }

    return (
      <div className="mt-4 space-y-2">
        <p className="text-sm font-medium text-foreground">Detalhes da acao:</p>
        <div className={cn('rounded-lg p-3 text-sm', severity.bgColor)}>
          <dl className="space-y-1">
            {Object.entries(data.details).map(([key, value]) => (
              <div key={key} className="flex justify-between gap-4">
                <dt className="text-muted-foreground capitalize">
                  {key.replace(/_/g, ' ')}:
                </dt>
                <dd className="font-medium text-right">
                  {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                </dd>
              </div>
            ))}
          </dl>
        </div>
      </div>
    );
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && !isLoading && onCancel()}>
      <DialogContent className={cn('sm:max-w-md', className)} showCloseButton={!isLoading}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className={cn('rounded-full p-2', severity.bgColor)}>
              <SeverityIcon className={cn('h-5 w-5', severity.color)} />
            </div>
            <DialogTitle>Confirmar Acao</DialogTitle>
          </div>
          <DialogDescription className="text-left pt-2">
            {data.description}
          </DialogDescription>
        </DialogHeader>

        <div className="py-2">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>Tipo de acao:</span>
            <span className="font-medium text-foreground">{actionLabel}</span>
          </div>
          {renderDetails()}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
          >
            Cancelar
          </Button>
          <Button
            type="button"
            onClick={handleConfirm}
            disabled={isLoading}
            className={cn(
              data.severity === 'CRITICAL' && 'bg-red-600 hover:bg-red-700'
            )}
          >
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Processando...
              </>
            ) : (
              'Confirmar'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default ConfirmationDialog;
