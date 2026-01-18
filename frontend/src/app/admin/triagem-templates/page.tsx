'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  FileSliders,
  Plus,
  MoreHorizontal,
  Eye,
  Edit,
  Copy,
  Power,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  fetchAdminTriagemTemplates,
  fetchAdminTenants,
  createAdminTriagemTemplate,
  updateAdminTriagemTemplate,
  cloneTriagemTemplateToTenants,
  fetchTriagemTemplateUsage,
  type AdminTriagemTemplate,
  type AdminTenant,
  type PaginatedAdminResponse,
} from '@/lib/api/admin';
import { CloneTemplateDialog } from '@/components/admin/CloneTemplateDialog';
import { toast } from 'sonner';

// Pagination Component
interface PaginationProps {
  currentPage: number;
  totalPages: number;
  totalItems: number;
  perPage: number;
  onPageChange: (page: number) => void;
}

function Pagination({
  currentPage,
  totalPages,
  totalItems,
  perPage,
  onPageChange,
}: PaginationProps) {
  const startItem = (currentPage - 1) * perPage + 1;
  const endItem = Math.min(currentPage * perPage, totalItems);

  return (
    <div className="flex items-center justify-between px-2 py-4">
      <p className="text-sm text-slate-400">
        Mostrando {startItem} a {endItem} de {totalItems} resultados
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage <= 1}
          className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700 disabled:opacity-50"
        >
          Anterior
        </Button>
        <span className="text-sm text-slate-400">
          Pagina {currentPage} de {totalPages}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages}
          className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700 disabled:opacity-50"
        >
          Proxima
        </Button>
      </div>
    </div>
  );
}

// Create/Edit Template Dialog
interface TemplateFormDialogProps {
  open: boolean;
  onClose: () => void;
  template: AdminTriagemTemplate | null;
  onSave: (data: {
    nome: string;
    tipo: string;
    condicao: Record<string, unknown>;
    descricao?: string;
  }) => void;
  isLoading?: boolean;
}

function TemplateFormDialog({
  open,
  onClose,
  template,
  onSave,
  isLoading,
}: TemplateFormDialogProps) {
  const [nome, setNome] = useState('');
  const [tipo, setTipo] = useState('');
  const [condicaoJson, setCondicaoJson] = useState('{}');
  const [descricao, setDescricao] = useState('');
  const [jsonError, setJsonError] = useState<string | null>(null);

  useEffect(() => {
    if (template) {
      setNome(template.nome);
      setTipo(template.tipo);
      setCondicaoJson(JSON.stringify(template.condicao, null, 2));
      setDescricao(template.descricao || '');
    } else {
      setNome('');
      setTipo('');
      setCondicaoJson('{}');
      setDescricao('');
    }
    setJsonError(null);
  }, [template, open]);

  const validateJson = (json: string): boolean => {
    try {
      JSON.parse(json);
      setJsonError(null);
      return true;
    } catch {
      setJsonError('JSON invalido');
      return false;
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!nome.trim() || !tipo.trim()) {
      toast.error('Preencha todos os campos obrigatorios');
      return;
    }
    if (!validateJson(condicaoJson)) {
      return;
    }
    onSave({
      nome: nome.trim(),
      tipo: tipo.trim(),
      condicao: JSON.parse(condicaoJson),
      descricao: descricao.trim() || undefined,
    });
  };

  const isEdit = !!template;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? 'Editar Template' : 'Criar Template de Triagem'}
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            {isEdit
              ? 'Altere os dados do template'
              : 'Crie uma regra de triagem que pode ser clonada para tenants'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Nome <span className="text-red-400">*</span>
            </label>
            <Input
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              placeholder="Ex: Prioridade Alta - Pediatria"
              className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Tipo <span className="text-red-400">*</span>
            </label>
            <Select value={tipo} onValueChange={setTipo}>
              <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                <SelectValue placeholder="Selecione o tipo" />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                <SelectItem value="prioridade" className="text-white hover:bg-slate-700">
                  Prioridade
                </SelectItem>
                <SelectItem value="classificacao" className="text-white hover:bg-slate-700">
                  Classificacao
                </SelectItem>
                <SelectItem value="encaminhamento" className="text-white hover:bg-slate-700">
                  Encaminhamento
                </SelectItem>
                <SelectItem value="alerta" className="text-white hover:bg-slate-700">
                  Alerta
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Descricao</label>
            <Input
              value={descricao}
              onChange={(e) => setDescricao(e.target.value)}
              placeholder="Descricao opcional..."
              className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Condicao (JSON) <span className="text-red-400">*</span>
            </label>
            <Textarea
              value={condicaoJson}
              onChange={(e) => {
                setCondicaoJson(e.target.value);
                if (jsonError) validateJson(e.target.value);
              }}
              onBlur={() => validateJson(condicaoJson)}
              placeholder='{"campo": "valor", "operador": "igual"}'
              rows={6}
              className={`bg-slate-900 border-slate-700 text-white placeholder:text-slate-500 font-mono text-sm ${
                jsonError ? 'border-red-500' : ''
              }`}
            />
            {jsonError && (
              <p className="text-xs text-red-400">{jsonError}</p>
            )}
            <p className="text-xs text-slate-500">
              Defina as condicoes da regra em formato JSON
            </p>
          </div>
          <DialogFooter className="pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={onClose}
              className="text-slate-400 hover:bg-slate-700"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={isLoading}
              className="bg-violet-600 hover:bg-violet-700"
            >
              {isLoading ? 'Salvando...' : isEdit ? 'Atualizar' : 'Criar Template'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// View Template Dialog
interface ViewTemplateDialogProps {
  open: boolean;
  onClose: () => void;
  template: AdminTriagemTemplate | null;
  tenantsUsing: { id: string; name: string }[];
  isLoadingUsage?: boolean;
}

function ViewTemplateDialog({
  open,
  onClose,
  template,
  tenantsUsing,
  isLoadingUsage,
}: ViewTemplateDialogProps) {
  if (!template) return null;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Detalhes do Template</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-lg bg-violet-600/20 flex items-center justify-center">
              <FileSliders className="h-6 w-6 text-violet-400" />
            </div>
            <div>
              <h3 className="text-lg font-medium text-white">{template.nome}</h3>
              <div className="flex items-center gap-2 mt-1">
                <Badge
                  variant="outline"
                  className="bg-slate-700/50 border-slate-600 text-slate-300"
                >
                  {template.tipo}
                </Badge>
                {template.ativo ? (
                  <Badge className="bg-emerald-400/10 text-emerald-400 border-emerald-400/20">
                    Ativo
                  </Badge>
                ) : (
                  <Badge className="bg-red-400/10 text-red-400 border-red-400/20">
                    Inativo
                  </Badge>
                )}
              </div>
            </div>
          </div>

          {template.descricao && (
            <div className="p-3 bg-slate-900 rounded-lg">
              <p className="text-sm text-slate-300">{template.descricao}</p>
            </div>
          )}

          <div className="space-y-2">
            <label className="text-xs text-slate-500 uppercase tracking-wide">
              Condicao
            </label>
            <pre className="p-3 bg-slate-900 rounded-lg overflow-x-auto text-xs text-slate-300 font-mono">
              {JSON.stringify(template.condicao, null, 2)}
            </pre>
          </div>

          <div className="space-y-2">
            <label className="text-xs text-slate-500 uppercase tracking-wide">
              Tenants usando este template
            </label>
            {isLoadingUsage ? (
              <div className="flex items-center justify-center py-4">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
              </div>
            ) : tenantsUsing.length === 0 ? (
              <p className="text-sm text-slate-400 py-2">
                Nenhum tenant esta usando este template
              </p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {tenantsUsing.map((tenant) => (
                  <Badge
                    key={tenant.id}
                    variant="outline"
                    className="bg-slate-700/50 border-slate-600 text-slate-300"
                  >
                    {tenant.name}
                  </Badge>
                ))}
              </div>
            )}
          </div>
        </div>
        <DialogFooter>
          <Button onClick={onClose} className="bg-violet-600 hover:bg-violet-700">
            Fechar
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function TriagemTemplatesPage() {
  const [templates, setTemplates] = useState<AdminTriagemTemplate[]>([]);
  const [tenants, setTenants] = useState<AdminTenant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [tipoFilter, setTipoFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [perPage] = useState(10);

  // Dialog states
  const [formDialogOpen, setFormDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [cloneDialogOpen, setCloneDialogOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<AdminTriagemTemplate | null>(null);
  const [tenantsUsing, setTenantsUsing] = useState<{ id: string; name: string }[]>([]);
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [isLoadingUsage, setIsLoadingUsage] = useState(false);

  // Load tenants on mount
  useEffect(() => {
    async function loadTenants() {
      try {
        const response = await fetchAdminTenants({ per_page: 100 });
        setTenants(response.data);
      } catch (err) {
        console.error('Failed to fetch tenants:', err);
      }
    }
    loadTenants();
  }, []);

  const loadTemplates = useCallback(async () => {
    try {
      setIsLoading(true);
      const response: PaginatedAdminResponse<AdminTriagemTemplate> =
        await fetchAdminTriagemTemplates({
          page,
          per_page: perPage,
          tipo: tipoFilter !== 'all' ? tipoFilter : undefined,
        });
      setTemplates(response.data);
      setTotalPages(response.meta.total_pages);
      setTotalItems(response.meta.total);
    } catch (err) {
      console.error('Failed to fetch templates:', err);
      toast.error('Falha ao carregar templates');
      setTemplates([]);
      setTotalPages(1);
      setTotalItems(0);
    } finally {
      setIsLoading(false);
    }
  }, [page, perPage, tipoFilter]);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  const handleTipoFilter = (value: string) => {
    setTipoFilter(value);
    setPage(1);
  };

  // Actions
  const handleView = async (template: AdminTriagemTemplate) => {
    setSelectedTemplate(template);
    setViewDialogOpen(true);
    setTenantsUsing([]);

    try {
      setIsLoadingUsage(true);
      const result = await fetchTriagemTemplateUsage(template.id);
      setTenantsUsing(result.tenants);
    } catch (err) {
      console.error('Failed to fetch template usage:', err);
    } finally {
      setIsLoadingUsage(false);
    }
  };

  const handleCreate = () => {
    setSelectedTemplate(null);
    setFormDialogOpen(true);
  };

  const handleEdit = (template: AdminTriagemTemplate) => {
    setSelectedTemplate(template);
    setFormDialogOpen(true);
  };

  const handleClone = (template: AdminTriagemTemplate) => {
    setSelectedTemplate(template);
    setCloneDialogOpen(true);
  };

  const handleSaveTemplate = async (data: {
    nome: string;
    tipo: string;
    condicao: Record<string, unknown>;
    descricao?: string;
  }) => {
    try {
      setIsActionLoading(true);
      if (selectedTemplate) {
        await updateAdminTriagemTemplate(selectedTemplate.id, data);
        toast.success('Template atualizado com sucesso');
      } else {
        await createAdminTriagemTemplate(data);
        toast.success('Template criado com sucesso');
      }
      setFormDialogOpen(false);
      loadTemplates();
    } catch (err) {
      console.error('Failed to save template:', err);
      toast.error('Falha ao salvar template');
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleToggleStatus = async (template: AdminTriagemTemplate) => {
    try {
      await updateAdminTriagemTemplate(template.id, { ativo: !template.ativo });
      toast.success(
        template.ativo ? 'Template desativado' : 'Template ativado'
      );
      loadTemplates();
    } catch (err) {
      console.error('Failed to toggle template status:', err);
      toast.error('Falha ao alterar status do template');
    }
  };

  const handleCloneToTenants = async (templateId: string, tenantIds: string[]) => {
    try {
      setIsActionLoading(true);
      const result = await cloneTriagemTemplateToTenants(templateId, tenantIds);
      toast.success(`Template clonado para ${result.cloned_count} tenant(s)`);
      setCloneDialogOpen(false);
    } catch (err) {
      console.error('Failed to clone template:', err);
      toast.error('Falha ao clonar template');
    } finally {
      setIsActionLoading(false);
    }
  };

  const getTipoBadgeColor = (tipo: string) => {
    switch (tipo) {
      case 'prioridade':
        return 'bg-red-400/10 text-red-400 border-red-400/20';
      case 'classificacao':
        return 'bg-blue-400/10 text-blue-400 border-blue-400/20';
      case 'encaminhamento':
        return 'bg-amber-400/10 text-amber-400 border-amber-400/20';
      case 'alerta':
        return 'bg-violet-400/10 text-violet-400 border-violet-400/20';
      default:
        return 'bg-slate-600/20 text-slate-400 border-slate-600/20';
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Templates de Triagem
          </h1>
          <p className="text-slate-400">
            Gerencie regras de triagem globais e clone para tenants
          </p>
        </div>
        <Button
          onClick={handleCreate}
          className="gap-2 bg-violet-600 hover:bg-violet-700"
        >
          <Plus className="h-4 w-4" />
          Novo Template
        </Button>
      </div>

      {/* Filters */}
      <Card className="bg-slate-800 border-slate-700">
        <CardContent className="pt-6">
          <div className="flex flex-col gap-4 sm:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Buscar por nome..."
                className="pl-10 bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
              />
            </div>
            <Select value={tipoFilter} onValueChange={handleTipoFilter}>
              <SelectTrigger className="w-full sm:w-[180px] bg-slate-900 border-slate-700 text-white">
                <SelectValue placeholder="Filtrar por tipo" />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                <SelectItem value="all" className="text-white hover:bg-slate-700">
                  Todos os tipos
                </SelectItem>
                <SelectItem value="prioridade" className="text-white hover:bg-slate-700">
                  Prioridade
                </SelectItem>
                <SelectItem value="classificacao" className="text-white hover:bg-slate-700">
                  Classificacao
                </SelectItem>
                <SelectItem value="encaminhamento" className="text-white hover:bg-slate-700">
                  Encaminhamento
                </SelectItem>
                <SelectItem value="alerta" className="text-white hover:bg-slate-700">
                  Alerta
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Lista de Templates</CardTitle>
          <CardDescription className="text-slate-400">
            {totalItems} template(s) encontrado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
            </div>
          ) : templates.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <FileSliders className="h-12 w-12 text-slate-600 mb-4" />
              <h3 className="text-lg font-medium text-slate-300">
                Nenhum template encontrado
              </h3>
              <p className="text-sm text-slate-500 mt-1">
                Crie seu primeiro template de triagem
              </p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-slate-700 hover:bg-slate-800">
                      <TableHead className="text-slate-400">Nome</TableHead>
                      <TableHead className="text-slate-400">Tipo</TableHead>
                      <TableHead className="text-slate-400">Status</TableHead>
                      <TableHead className="text-slate-400">Tenants</TableHead>
                      <TableHead className="text-slate-400 text-right">Acoes</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {templates.map((template) => (
                      <TableRow
                        key={template.id}
                        className="border-slate-700 hover:bg-slate-800/50"
                      >
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-lg bg-violet-600/20 flex items-center justify-center">
                              <FileSliders className="h-4 w-4 text-violet-400" />
                            </div>
                            <div>
                              <p className="font-medium text-white">{template.nome}</p>
                              {template.descricao && (
                                <p className="text-xs text-slate-400 truncate max-w-xs">
                                  {template.descricao}
                                </p>
                              )}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant="outline"
                            className={getTipoBadgeColor(template.tipo)}
                          >
                            {template.tipo}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {template.ativo ? (
                            <Badge className="bg-emerald-400/10 text-emerald-400 border-emerald-400/20">
                              Ativo
                            </Badge>
                          ) : (
                            <Badge className="bg-slate-600/20 text-slate-400 border-slate-600/20">
                              Inativo
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {template.tenant_usage_count ?? 0}
                        </TableCell>
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-slate-400 hover:text-white hover:bg-slate-700"
                              >
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                              align="end"
                              className="bg-slate-800 border-slate-700"
                            >
                              <DropdownMenuItem
                                onClick={() => handleView(template)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <Eye className="h-4 w-4 mr-2" />
                                Ver detalhes
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleEdit(template)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <Edit className="h-4 w-4 mr-2" />
                                Editar
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleClone(template)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <Copy className="h-4 w-4 mr-2" />
                                Clonar para tenants
                              </DropdownMenuItem>
                              <DropdownMenuSeparator className="bg-slate-700" />
                              <DropdownMenuItem
                                onClick={() => handleToggleStatus(template)}
                                className={
                                  template.ativo
                                    ? 'text-amber-400 hover:bg-slate-700 hover:text-amber-300 cursor-pointer'
                                    : 'text-emerald-400 hover:bg-slate-700 hover:text-emerald-300 cursor-pointer'
                                }
                              >
                                <Power className="h-4 w-4 mr-2" />
                                {template.ativo ? 'Desativar' : 'Ativar'}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <Pagination
                currentPage={page}
                totalPages={totalPages}
                totalItems={totalItems}
                perPage={perPage}
                onPageChange={setPage}
              />
            </>
          )}
        </CardContent>
      </Card>

      {/* Dialogs */}
      <TemplateFormDialog
        open={formDialogOpen}
        onClose={() => setFormDialogOpen(false)}
        template={selectedTemplate}
        onSave={handleSaveTemplate}
        isLoading={isActionLoading}
      />
      <ViewTemplateDialog
        open={viewDialogOpen}
        onClose={() => setViewDialogOpen(false)}
        template={selectedTemplate}
        tenantsUsing={tenantsUsing}
        isLoadingUsage={isLoadingUsage}
      />
      <CloneTemplateDialog
        open={cloneDialogOpen}
        onClose={() => setCloneDialogOpen(false)}
        template={selectedTemplate}
        tenants={tenants}
        onClone={handleCloneToTenants}
        isLoading={isActionLoading}
      />
    </div>
  );
}
