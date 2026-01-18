'use client';

import { useState } from 'react';
import { AlertTriangle, User, Shield, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import type { AdminUser } from '@/lib/api/admin';

interface ImpersonateDialogProps {
  open: boolean;
  onClose: () => void;
  user: AdminUser | null;
  onImpersonate: (userId: string) => void;
  isLoading?: boolean;
}

export function ImpersonateDialog({
  open,
  onClose,
  user,
  onImpersonate,
  isLoading = false,
}: ImpersonateDialogProps) {
  if (!user) return null;

  const handleConfirm = () => {
    onImpersonate(user.id);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-violet-400" />
            Impersonalizar Usuario
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            Voce esta prestes a fazer login como outro usuario
          </DialogDescription>
        </DialogHeader>

        {/* User Details */}
        <div className="space-y-4 py-4">
          <div className="flex items-center gap-4 p-4 bg-slate-900 rounded-lg">
            <div className="h-12 w-12 rounded-full bg-violet-600/20 flex items-center justify-center">
              <User className="h-6 w-6 text-violet-400" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="font-medium text-white truncate">{user.nome}</p>
              <p className="text-sm text-slate-400 truncate">{user.email}</p>
              <div className="flex items-center gap-2 mt-1">
                <Badge
                  variant="outline"
                  className="text-xs bg-slate-700/50 border-slate-600 text-slate-300"
                >
                  {user.role}
                </Badge>
                {user.tenant_name && (
                  <span className="text-xs text-slate-500">
                    {user.tenant_name}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Warning */}
          <div className="flex items-start gap-3 p-4 bg-amber-400/10 border border-amber-400/20 rounded-lg">
            <AlertTriangle className="h-5 w-5 text-amber-400 shrink-0 mt-0.5" />
            <div className="space-y-1">
              <p className="text-sm font-medium text-amber-400">
                Atencao: Esta acao sera registrada
              </p>
              <p className="text-xs text-amber-400/80">
                A acao de impersonalizacao sera registrada no log de auditoria, incluindo
                seu ID de administrador e todas as acoes realizadas durante a sessao.
              </p>
            </div>
          </div>

          {/* Info */}
          <div className="space-y-2 text-sm text-slate-400">
            <p className="flex items-center gap-2">
              <ExternalLink className="h-4 w-4" />
              Uma nova aba sera aberta com a sessao do usuario
            </p>
            <p>
              A sessao de impersonalizacao tem duracao limitada de 1 hora.
            </p>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            type="button"
            variant="ghost"
            onClick={onClose}
            disabled={isLoading}
            className="text-slate-400 hover:bg-slate-700 hover:text-slate-200"
          >
            Cancelar
          </Button>
          <Button
            type="button"
            onClick={handleConfirm}
            disabled={isLoading}
            className="bg-violet-600 hover:bg-violet-700"
          >
            {isLoading ? (
              <>
                <div className="h-4 w-4 mr-2 animate-spin rounded-full border-2 border-white border-t-transparent" />
                Gerando token...
              </>
            ) : (
              'Confirmar Impersonalizacao'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default ImpersonateDialog;
