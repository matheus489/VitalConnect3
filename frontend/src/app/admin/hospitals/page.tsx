'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  Building2,
  MoreHorizontal,
  Eye,
  Edit,
  ArrowRightLeft,
  MapPin,
} from 'lucide-react';
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  fetchAdminHospitals,
  fetchAdminTenants,
  updateAdminHospital,
  reassignHospitalTenant,
  type AdminHospital,
  type AdminTenant,
  type PaginatedAdminResponse,
} from '@/lib/api/admin';
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

// View Hospital Dialog
interface ViewHospitalDialogProps {
  open: boolean;
  onClose: () => void;
  hospital: AdminHospital | null;
}

function ViewHospitalDialog({ open, onClose, hospital }: ViewHospitalDialogProps) {
  if (!hospital) return null;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white">
        <DialogHeader>
          <DialogTitle>Detalhes do Hospital</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="flex items-center gap-4">
            <div className="h-16 w-16 rounded-lg bg-emerald-600/20 flex items-center justify-center">
              <Building2 className="h-8 w-8 text-emerald-400" />
            </div>
            <div>
              <h3 className="text-lg font-medium text-white">{hospital.nome}</h3>
              <p className="text-sm text-slate-400">Codigo: {hospital.codigo}</p>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-slate-700">
            <div>
              <label className="text-xs text-slate-500">Cidade</label>
              <p className="text-sm text-white">{hospital.cidade || '-'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Estado</label>
              <p className="text-sm text-white">{hospital.estado || '-'}</p>
            </div>
            <div className="col-span-2">
              <label className="text-xs text-slate-500">Endereco</label>
              <p className="text-sm text-white">{hospital.endereco || '-'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Tenant</label>
              <p className="text-sm text-white">{hospital.tenant_name || '-'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Status</label>
              <p className="text-sm text-white">
                {hospital.ativo ? (
                  <span className="text-emerald-400">Ativo</span>
                ) : (
                  <span className="text-red-400">Inativo</span>
                )}
              </p>
            </div>
            {hospital.latitude && hospital.longitude && (
              <div className="col-span-2">
                <label className="text-xs text-slate-500">Coordenadas</label>
                <p className="text-sm text-white flex items-center gap-2">
                  <MapPin className="h-4 w-4 text-slate-400" />
                  {hospital.latitude.toFixed(6)}, {hospital.longitude.toFixed(6)}
                </p>
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

// Edit Hospital Dialog
interface EditHospitalDialogProps {
  open: boolean;
  onClose: () => void;
  hospital: AdminHospital | null;
  onSave: (id: string, data: Partial<AdminHospital>) => void;
  isLoading?: boolean;
}

function EditHospitalDialog({
  open,
  onClose,
  hospital,
  onSave,
  isLoading,
}: EditHospitalDialogProps) {
  const [formData, setFormData] = useState({
    nome: '',
    codigo: '',
    endereco: '',
    cidade: '',
    estado: '',
    latitude: '',
    longitude: '',
  });

  useEffect(() => {
    if (hospital) {
      setFormData({
        nome: hospital.nome || '',
        codigo: hospital.codigo || '',
        endereco: hospital.endereco || '',
        cidade: hospital.cidade || '',
        estado: hospital.estado || '',
        latitude: hospital.latitude?.toString() || '',
        longitude: hospital.longitude?.toString() || '',
      });
    }
  }, [hospital]);

  if (!hospital) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(hospital.id, {
      nome: formData.nome,
      codigo: formData.codigo,
      endereco: formData.endereco || undefined,
      cidade: formData.cidade || undefined,
      estado: formData.estado || undefined,
      latitude: formData.latitude ? parseFloat(formData.latitude) : undefined,
      longitude: formData.longitude ? parseFloat(formData.longitude) : undefined,
    });
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Editar Hospital</DialogTitle>
          <DialogDescription className="text-slate-400">
            Altere os dados do hospital
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Nome</label>
            <Input
              value={formData.nome}
              onChange={(e) => setFormData({ ...formData, nome: e.target.value })}
              className="bg-slate-900 border-slate-700 text-white"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Codigo</label>
            <Input
              value={formData.codigo}
              onChange={(e) => setFormData({ ...formData, codigo: e.target.value })}
              className="bg-slate-900 border-slate-700 text-white"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Endereco</label>
            <Input
              value={formData.endereco}
              onChange={(e) => setFormData({ ...formData, endereco: e.target.value })}
              className="bg-slate-900 border-slate-700 text-white"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Cidade</label>
              <Input
                value={formData.cidade}
                onChange={(e) => setFormData({ ...formData, cidade: e.target.value })}
                className="bg-slate-900 border-slate-700 text-white"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Estado</label>
              <Input
                value={formData.estado}
                onChange={(e) => setFormData({ ...formData, estado: e.target.value })}
                maxLength={2}
                className="bg-slate-900 border-slate-700 text-white"
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Latitude</label>
              <Input
                type="number"
                step="any"
                value={formData.latitude}
                onChange={(e) => setFormData({ ...formData, latitude: e.target.value })}
                className="bg-slate-900 border-slate-700 text-white"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Longitude</label>
              <Input
                type="number"
                step="any"
                value={formData.longitude}
                onChange={(e) => setFormData({ ...formData, longitude: e.target.value })}
                className="bg-slate-900 border-slate-700 text-white"
              />
            </div>
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
              {isLoading ? 'Salvando...' : 'Salvar'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Reassign Tenant Dialog
interface ReassignTenantDialogProps {
  open: boolean;
  onClose: () => void;
  hospital: AdminHospital | null;
  tenants: AdminTenant[];
  onReassign: (hospitalId: string, tenantId: string) => void;
  isLoading?: boolean;
}

function ReassignTenantDialog({
  open,
  onClose,
  hospital,
  tenants,
  onReassign,
  isLoading,
}: ReassignTenantDialogProps) {
  const [selectedTenant, setSelectedTenant] = useState<string>('');

  useEffect(() => {
    if (hospital) {
      setSelectedTenant(hospital.tenant_id);
    }
  }, [hospital]);

  if (!hospital) return null;

  const availableTenants = tenants.filter(
    (t) => t.is_active && t.id !== hospital.tenant_id
  );

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ArrowRightLeft className="h-5 w-5 text-violet-400" />
            Reassignar Tenant
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            Transferir {hospital.nome} para outro tenant
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="p-4 bg-amber-400/10 border border-amber-400/20 rounded-lg">
            <p className="text-sm text-amber-400">
              <strong>Atencao:</strong> Esta acao transferira o hospital e todos os
              dados associados para o tenant selecionado.
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Tenant atual: <span className="text-white">{hospital.tenant_name}</span>
            </label>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Novo tenant
            </label>
            <Select value={selectedTenant} onValueChange={setSelectedTenant}>
              <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                <SelectValue placeholder="Selecione o tenant de destino" />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                {availableTenants.map((tenant) => (
                  <SelectItem
                    key={tenant.id}
                    value={tenant.id}
                    className="text-white hover:bg-slate-700"
                  >
                    {tenant.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="ghost"
            onClick={onClose}
            className="text-slate-400 hover:bg-slate-700"
          >
            Cancelar
          </Button>
          <Button
            onClick={() => onReassign(hospital.id, selectedTenant)}
            disabled={isLoading || !selectedTenant || selectedTenant === hospital.tenant_id}
            className="bg-violet-600 hover:bg-violet-700"
          >
            {isLoading ? 'Transferindo...' : 'Confirmar Transferencia'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function HospitalsListPage() {
  const [hospitals, setHospitals] = useState<AdminHospital[]>([]);
  const [tenants, setTenants] = useState<AdminTenant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [tenantFilter, setTenantFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [perPage] = useState(10);

  // Dialog states
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [reassignDialogOpen, setReassignDialogOpen] = useState(false);
  const [selectedHospital, setSelectedHospital] = useState<AdminHospital | null>(null);
  const [isActionLoading, setIsActionLoading] = useState(false);

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

  const loadHospitals = useCallback(async () => {
    try {
      setIsLoading(true);
      const response: PaginatedAdminResponse<AdminHospital> = await fetchAdminHospitals({
        page,
        per_page: perPage,
        search: search || undefined,
        tenant_id: tenantFilter !== 'all' ? tenantFilter : undefined,
      });
      setHospitals(response.data);
      setTotalPages(response.meta.total_pages);
      setTotalItems(response.meta.total);
    } catch (err) {
      console.error('Failed to fetch hospitals:', err);
      toast.error('Falha ao carregar hospitais');
      setHospitals([]);
      setTotalPages(1);
      setTotalItems(0);
    } finally {
      setIsLoading(false);
    }
  }, [page, perPage, search, tenantFilter]);

  useEffect(() => {
    loadHospitals();
  }, [loadHospitals]);

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handleTenantFilter = (value: string) => {
    setTenantFilter(value);
    setPage(1);
  };

  // Actions
  const handleView = (hospital: AdminHospital) => {
    setSelectedHospital(hospital);
    setViewDialogOpen(true);
  };

  const handleEdit = (hospital: AdminHospital) => {
    setSelectedHospital(hospital);
    setEditDialogOpen(true);
  };

  const handleReassignClick = (hospital: AdminHospital) => {
    setSelectedHospital(hospital);
    setReassignDialogOpen(true);
  };

  const handleSave = async (id: string, data: Partial<AdminHospital>) => {
    try {
      setIsActionLoading(true);
      await updateAdminHospital(id, data);
      toast.success('Hospital atualizado com sucesso');
      setEditDialogOpen(false);
      loadHospitals();
    } catch (err) {
      console.error('Failed to update hospital:', err);
      toast.error('Falha ao atualizar hospital');
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleReassign = async (hospitalId: string, tenantId: string) => {
    try {
      setIsActionLoading(true);
      await reassignHospitalTenant(hospitalId, tenantId);
      toast.success('Hospital transferido com sucesso');
      setReassignDialogOpen(false);
      loadHospitals();
    } catch (err) {
      console.error('Failed to reassign hospital:', err);
      toast.error('Falha ao transferir hospital');
    } finally {
      setIsActionLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">
          Gerenciar Hospitais
        </h1>
        <p className="text-slate-400">
          Visualize e gerencie hospitais de todos os tenants
        </p>
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
                placeholder="Buscar por nome ou cidade..."
                className="pl-10 bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
              />
            </div>
            <Select value={tenantFilter} onValueChange={handleTenantFilter}>
              <SelectTrigger className="w-full sm:w-[220px] bg-slate-900 border-slate-700 text-white">
                <SelectValue placeholder="Filtrar por tenant" />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                <SelectItem value="all" className="text-white hover:bg-slate-700">
                  Todos os tenants
                </SelectItem>
                {tenants.map((tenant) => (
                  <SelectItem
                    key={tenant.id}
                    value={tenant.id}
                    className="text-white hover:bg-slate-700"
                  >
                    {tenant.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Lista de Hospitais</CardTitle>
          <CardDescription className="text-slate-400">
            {totalItems} hospital(is) encontrado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
            </div>
          ) : hospitals.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Building2 className="h-12 w-12 text-slate-600 mb-4" />
              <h3 className="text-lg font-medium text-slate-300">
                Nenhum hospital encontrado
              </h3>
              <p className="text-sm text-slate-500 mt-1">
                {search || tenantFilter !== 'all'
                  ? 'Tente ajustar os filtros de busca'
                  : 'Nao ha hospitais cadastrados'}
              </p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-slate-700 hover:bg-slate-800">
                      <TableHead className="text-slate-400">Hospital</TableHead>
                      <TableHead className="text-slate-400">Cidade</TableHead>
                      <TableHead className="text-slate-400">Estado</TableHead>
                      <TableHead className="text-slate-400">Tenant</TableHead>
                      <TableHead className="text-slate-400 text-right">Acoes</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {hospitals.map((hospital) => (
                      <TableRow
                        key={hospital.id}
                        className="border-slate-700 hover:bg-slate-800/50"
                      >
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-lg bg-emerald-600/20 flex items-center justify-center">
                              <Building2 className="h-4 w-4 text-emerald-400" />
                            </div>
                            <div>
                              <p className="font-medium text-white">{hospital.nome}</p>
                              <p className="text-xs text-slate-400">{hospital.codigo}</p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {hospital.cidade || '-'}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {hospital.estado || '-'}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant="outline"
                            className="bg-slate-700/50 border-slate-600 text-slate-300"
                          >
                            {hospital.tenant_name || '-'}
                          </Badge>
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
                                onClick={() => handleView(hospital)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <Eye className="h-4 w-4 mr-2" />
                                Ver detalhes
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleEdit(hospital)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <Edit className="h-4 w-4 mr-2" />
                                Editar
                              </DropdownMenuItem>
                              <DropdownMenuSeparator className="bg-slate-700" />
                              <DropdownMenuItem
                                onClick={() => handleReassignClick(hospital)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <ArrowRightLeft className="h-4 w-4 mr-2" />
                                Reassignar tenant
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
      <ViewHospitalDialog
        open={viewDialogOpen}
        onClose={() => setViewDialogOpen(false)}
        hospital={selectedHospital}
      />
      <EditHospitalDialog
        open={editDialogOpen}
        onClose={() => setEditDialogOpen(false)}
        hospital={selectedHospital}
        onSave={handleSave}
        isLoading={isActionLoading}
      />
      <ReassignTenantDialog
        open={reassignDialogOpen}
        onClose={() => setReassignDialogOpen(false)}
        hospital={selectedHospital}
        tenants={tenants}
        onReassign={handleReassign}
        isLoading={isActionLoading}
      />
    </div>
  );
}
