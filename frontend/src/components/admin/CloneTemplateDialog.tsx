'use client';

import { useState, useEffect } from 'react';
import { Copy, Building, Check, X } from 'lucide-react';
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
import { Input } from '@/components/ui/input';
import type { AdminTriagemTemplate, AdminTenant } from '@/lib/api/admin';

interface CloneTemplateDialogProps {
  open: boolean;
  onClose: () => void;
  template: AdminTriagemTemplate | null;
  tenants: AdminTenant[];
  onClone: (templateId: string, tenantIds: string[]) => void;
  isLoading?: boolean;
}

export function CloneTemplateDialog({
  open,
  onClose,
  template,
  tenants,
  onClone,
  isLoading = false,
}: CloneTemplateDialogProps) {
  const [selectedTenants, setSelectedTenants] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState('');

  // Reset selection when dialog opens/closes
  useEffect(() => {
    if (!open) {
      setSelectedTenants([]);
      setSearchTerm('');
    }
  }, [open]);

  if (!template) return null;

  // Filter tenants by search term
  const filteredTenants = tenants.filter(
    (tenant) =>
      tenant.is_active &&
      (tenant.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tenant.slug.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  const toggleTenant = (tenantId: string) => {
    setSelectedTenants((prev) =>
      prev.includes(tenantId)
        ? prev.filter((id) => id !== tenantId)
        : [...prev, tenantId]
    );
  };

  const selectAll = () => {
    setSelectedTenants(filteredTenants.map((t) => t.id));
  };

  const clearSelection = () => {
    setSelectedTenants([]);
  };

  const handleConfirm = () => {
    if (selectedTenants.length > 0) {
      onClone(template.id, selectedTenants);
    }
  };

  // Get tenant names for preview
  const selectedTenantNames = tenants
    .filter((t) => selectedTenants.includes(t.id))
    .map((t) => t.name);

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Copy className="h-5 w-5 text-violet-400" />
            Clonar Template de Triagem
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            Selecione os tenants que receberao esta regra de triagem
          </DialogDescription>
        </DialogHeader>

        {/* Template Info */}
        <div className="p-4 bg-slate-900 rounded-lg space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-white">{template.nome}</span>
            <Badge
              variant="outline"
              className="bg-slate-700/50 border-slate-600 text-slate-300"
            >
              {template.tipo}
            </Badge>
          </div>
          {template.descricao && (
            <p className="text-xs text-slate-400">{template.descricao}</p>
          )}
        </div>

        {/* Tenant Selection */}
        <div className="space-y-4">
          {/* Search */}
          <Input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Buscar tenants..."
            className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
          />

          {/* Selection Actions */}
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">
              {selectedTenants.length} tenant(s) selecionado(s)
            </span>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={selectAll}
                className="text-xs text-slate-400 hover:text-slate-200"
              >
                Selecionar todos
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={clearSelection}
                className="text-xs text-slate-400 hover:text-slate-200"
              >
                Limpar
              </Button>
            </div>
          </div>

          {/* Tenant List */}
          <div className="max-h-60 overflow-y-auto space-y-1 border border-slate-700 rounded-lg p-2">
            {filteredTenants.length === 0 ? (
              <p className="text-sm text-slate-500 text-center py-4">
                Nenhum tenant encontrado
              </p>
            ) : (
              filteredTenants.map((tenant) => {
                const isSelected = selectedTenants.includes(tenant.id);
                return (
                  <button
                    key={tenant.id}
                    type="button"
                    onClick={() => toggleTenant(tenant.id)}
                    className={`w-full flex items-center gap-3 p-3 rounded-lg transition-colors ${
                      isSelected
                        ? 'bg-violet-600/20 border border-violet-500/30'
                        : 'bg-slate-900 border border-transparent hover:bg-slate-800'
                    }`}
                  >
                    <div
                      className={`h-5 w-5 rounded flex items-center justify-center border ${
                        isSelected
                          ? 'bg-violet-600 border-violet-500'
                          : 'bg-slate-800 border-slate-600'
                      }`}
                    >
                      {isSelected && <Check className="h-3 w-3 text-white" />}
                    </div>
                    <Building className="h-4 w-4 text-slate-500" />
                    <div className="flex-1 text-left">
                      <p className="text-sm font-medium text-white">{tenant.name}</p>
                      <p className="text-xs text-slate-500">{tenant.slug}</p>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* Preview */}
        {selectedTenants.length > 0 && (
          <div className="p-3 bg-emerald-400/10 border border-emerald-400/20 rounded-lg">
            <p className="text-sm font-medium text-emerald-400 mb-2">
              Tenants que receberao a regra:
            </p>
            <div className="flex flex-wrap gap-1">
              {selectedTenantNames.map((name) => (
                <Badge
                  key={name}
                  variant="outline"
                  className="text-xs bg-emerald-600/20 border-emerald-500/30 text-emerald-400"
                >
                  {name}
                </Badge>
              ))}
            </div>
          </div>
        )}

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
            disabled={isLoading || selectedTenants.length === 0}
            className="bg-violet-600 hover:bg-violet-700 disabled:opacity-50"
          >
            {isLoading ? (
              <>
                <div className="h-4 w-4 mr-2 animate-spin rounded-full border-2 border-white border-t-transparent" />
                Clonando...
              </>
            ) : (
              <>
                <Copy className="h-4 w-4 mr-2" />
                Clonar para {selectedTenants.length} tenant(s)
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default CloneTemplateDialog;
