'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Plus, Search, Building, MoreHorizontal, Eye, Power, Edit } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
  fetchAdminTenants,
  createAdminTenant,
  toggleAdminTenantStatus,
  type AdminTenant,
  type PaginatedAdminResponse,
} from '@/lib/api/admin';
import { toast } from 'sonner';

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

interface CreateTenantDialogProps {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}

function CreateTenantDialog({ open, onClose, onCreated }: CreateTenantDialogProps) {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !slug.trim()) {
      toast.error('Preencha todos os campos');
      return;
    }

    try {
      setIsSubmitting(true);
      await createAdminTenant({ name: name.trim(), slug: slug.trim() });
      toast.success('Tenant criado com sucesso');
      setName('');
      setSlug('');
      onCreated();
      onClose();
    } catch (err) {
      console.error('Failed to create tenant:', err);
      toast.error('Falha ao criar tenant');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Auto-generate slug from name
  const handleNameChange = (value: string) => {
    setName(value);
    const generatedSlug = value
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/(^-|-$)/g, '');
    setSlug(generatedSlug);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white">
        <DialogHeader>
          <DialogTitle>Criar Novo Tenant</DialogTitle>
          <DialogDescription className="text-slate-400">
            Adicione um novo tenant ao sistema VitalConnect
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Nome do Tenant
            </label>
            <Input
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Ex: Hospital Sao Paulo"
              className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Slug (identificador unico)
            </label>
            <Input
              value={slug}
              onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
              placeholder="Ex: hospital-sao-paulo"
              className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
            />
            <p className="text-xs text-slate-500">
              Usado na URL e identificacao. Use apenas letras minusculas, numeros e hifens.
            </p>
          </div>
          <DialogFooter>
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
              disabled={isSubmitting}
              className="bg-violet-600 hover:bg-violet-700"
            >
              {isSubmitting ? 'Criando...' : 'Criar Tenant'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function TenantsListPage() {
  const [tenants, setTenants] = useState<AdminTenant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [perPage] = useState(10);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  const loadTenants = useCallback(async () => {
    try {
      setIsLoading(true);
      const response: PaginatedAdminResponse<AdminTenant> = await fetchAdminTenants({
        page,
        per_page: perPage,
        search: search || undefined,
        status: statusFilter,
      });
      setTenants(response.data);
      setTotalPages(response.meta.total_pages);
      setTotalItems(response.meta.total);
    } catch (err) {
      console.error('Failed to fetch tenants:', err);
      toast.error('Falha ao carregar tenants');
      // Set empty data on error for development
      setTenants([]);
      setTotalPages(1);
      setTotalItems(0);
    } finally {
      setIsLoading(false);
    }
  }, [page, perPage, search, statusFilter]);

  useEffect(() => {
    loadTenants();
  }, [loadTenants]);

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handleStatusFilter = (value: string) => {
    setStatusFilter(value as 'all' | 'active' | 'inactive');
    setPage(1);
  };

  const handleToggleStatus = async (tenant: AdminTenant) => {
    try {
      await toggleAdminTenantStatus(tenant.id);
      toast.success(
        tenant.is_active
          ? `Tenant "${tenant.name}" desativado`
          : `Tenant "${tenant.name}" ativado`
      );
      loadTenants();
    } catch (err) {
      console.error('Failed to toggle tenant status:', err);
      toast.error('Falha ao alterar status do tenant');
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Gerenciar Tenants
          </h1>
          <p className="text-slate-400">
            Visualize e gerencie todos os tenants do sistema
          </p>
        </div>
        <Button
          onClick={() => setCreateDialogOpen(true)}
          className="gap-2 bg-violet-600 hover:bg-violet-700"
        >
          <Plus className="h-4 w-4" />
          Novo Tenant
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
                onChange={(e) => handleSearch(e.target.value)}
                placeholder="Buscar por nome ou slug..."
                className="pl-10 bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
              />
            </div>
            <Select value={statusFilter} onValueChange={handleStatusFilter}>
              <SelectTrigger className="w-full sm:w-[180px] bg-slate-900 border-slate-700 text-white">
                <SelectValue placeholder="Filtrar por status" />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                <SelectItem value="all" className="text-white hover:bg-slate-700">
                  Todos
                </SelectItem>
                <SelectItem value="active" className="text-white hover:bg-slate-700">
                  Ativos
                </SelectItem>
                <SelectItem value="inactive" className="text-white hover:bg-slate-700">
                  Inativos
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Lista de Tenants</CardTitle>
          <CardDescription className="text-slate-400">
            {totalItems} tenant(s) encontrado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
            </div>
          ) : tenants.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Building className="h-12 w-12 text-slate-600 mb-4" />
              <h3 className="text-lg font-medium text-slate-300">
                Nenhum tenant encontrado
              </h3>
              <p className="text-sm text-slate-500 mt-1">
                {search || statusFilter !== 'all'
                  ? 'Tente ajustar os filtros de busca'
                  : 'Comece criando seu primeiro tenant'}
              </p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-slate-700 hover:bg-slate-800">
                      <TableHead className="text-slate-400">Nome</TableHead>
                      <TableHead className="text-slate-400">Slug</TableHead>
                      <TableHead className="text-slate-400">Status</TableHead>
                      <TableHead className="text-slate-400">Criado em</TableHead>
                      <TableHead className="text-slate-400 text-right">Acoes</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tenants.map((tenant) => (
                      <TableRow
                        key={tenant.id}
                        className="border-slate-700 hover:bg-slate-800/50"
                      >
                        <TableCell className="font-medium text-white">
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-lg bg-violet-600/20 flex items-center justify-center">
                              <Building className="h-4 w-4 text-violet-400" />
                            </div>
                            <span>{tenant.name}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-slate-400 font-mono text-sm">
                          {tenant.slug}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant={tenant.is_active ? 'default' : 'secondary'}
                            className={
                              tenant.is_active
                                ? 'bg-emerald-400/10 text-emerald-400 border-emerald-400/20'
                                : 'bg-slate-600/20 text-slate-400 border-slate-600/20'
                            }
                          >
                            {tenant.is_active ? 'Ativo' : 'Inativo'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-slate-400">
                          {new Date(tenant.created_at).toLocaleDateString('pt-BR')}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Link href={`/admin/tenants/${tenant.id}`}>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-slate-400 hover:text-white hover:bg-slate-700"
                              >
                                <Eye className="h-4 w-4" />
                              </Button>
                            </Link>
                            <Link href={`/admin/tenants/${tenant.id}`}>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="text-slate-400 hover:text-white hover:bg-slate-700"
                              >
                                <Edit className="h-4 w-4" />
                              </Button>
                            </Link>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleToggleStatus(tenant)}
                              className={
                                tenant.is_active
                                  ? 'text-amber-400 hover:text-amber-300 hover:bg-slate-700'
                                  : 'text-emerald-400 hover:text-emerald-300 hover:bg-slate-700'
                              }
                            >
                              <Power className="h-4 w-4" />
                            </Button>
                          </div>
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

      {/* Create Dialog */}
      <CreateTenantDialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        onCreated={loadTenants}
      />
    </div>
  );
}
