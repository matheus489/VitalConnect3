'use client';

import { useState } from 'react';
import { Flag } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { OutcomeType } from '@/types';

interface OutcomeModalProps {
  occurrenceId: string | null;
  open: boolean;
  onClose: () => void;
  onConfirm: (id: string, outcome: OutcomeType, observacoes: string) => void;
  isLoading?: boolean;
}

const outcomeOptions: { value: OutcomeType; label: string; description: string }[] = [
  {
    value: 'sucesso_captacao',
    label: 'Sucesso na Captacao',
    description: 'A captacao das corneas foi realizada com sucesso',
  },
  {
    value: 'familia_recusou',
    label: 'Familia Recusou',
    description: 'A familia do paciente nao autorizou a doacao',
  },
  {
    value: 'contraindicacao_medica',
    label: 'Contraindicacao Medica',
    description: 'Foram identificadas contraindicacoes medicas para a doacao',
  },
  {
    value: 'tempo_excedido',
    label: 'Tempo Excedido',
    description: 'A janela de 6 horas para captacao foi excedida',
  },
  {
    value: 'outro',
    label: 'Outro',
    description: 'Outro motivo nao listado',
  },
];

export function OutcomeModal({
  occurrenceId,
  open,
  onClose,
  onConfirm,
  isLoading,
}: OutcomeModalProps) {
  const [outcome, setOutcome] = useState<OutcomeType | ''>('');
  const [observacoes, setObservacoes] = useState('');
  const [error, setError] = useState('');

  const handleConfirm = () => {
    if (!outcome) {
      setError('Selecione um desfecho');
      return;
    }

    if (!occurrenceId) return;

    onConfirm(occurrenceId, outcome, observacoes);
    handleClose();
  };

  const handleClose = () => {
    setOutcome('');
    setObservacoes('');
    setError('');
    onClose();
  };

  const selectedOutcome = outcomeOptions.find((o) => o.value === outcome);

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Flag className="h-5 w-5 text-primary" />
            Registrar Desfecho
          </DialogTitle>
          <DialogDescription>
            Selecione o resultado da ocorrencia e adicione observacoes se necessario.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Outcome Select */}
          <div className="space-y-2">
            <Label htmlFor="outcome">
              Desfecho <span className="text-destructive">*</span>
            </Label>
            <Select value={outcome} onValueChange={(value) => {
              setOutcome(value as OutcomeType);
              setError('');
            }}>
              <SelectTrigger id="outcome" aria-invalid={!!error}>
                <SelectValue placeholder="Selecione o desfecho" />
              </SelectTrigger>
              <SelectContent>
                {outcomeOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {error && <p className="text-sm text-destructive">{error}</p>}
            {selectedOutcome && (
              <p className="text-sm text-muted-foreground">{selectedOutcome.description}</p>
            )}
          </div>

          {/* Observations */}
          <div className="space-y-2">
            <Label htmlFor="observacoes">Observacoes</Label>
            <textarea
              id="observacoes"
              value={observacoes}
              onChange={(e) => setObservacoes(e.target.value)}
              placeholder="Adicione informacoes adicionais sobre o desfecho..."
              className="min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={handleClose} disabled={isLoading}>
            Cancelar
          </Button>
          <Button onClick={handleConfirm} disabled={isLoading}>
            {isLoading ? 'Salvando...' : 'Confirmar Desfecho'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
