'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  Users,
  MoreHorizontal,
  Eye,
  UserCog,
  LogIn,
  Ban,
  KeyRound,
  CheckCircle,
  XCircle,
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
  fetchAdminUsers,
  fetchAdminTenants,
  impersonateUser,
  updateAdminUserRole,
  banAdminUser,
  resetAdminUserPassword,
  type AdminUser,
  type AdminTenant,
  type PaginatedAdminResponse,
} from '@/lib/api/admin';
import { ImpersonateDialog } from '@/components/admin/ImpersonateDialog';
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

// Edit Role Dialog
interface EditRoleDialogProps {
  open: boolean;
  onClose: () => void;
  user: AdminUser | null;
  onSave: (userId: string, role: string, isSuperAdmin: boolean) => void;
  isLoading?: boolean;
}

function EditRoleDialog({ open, onClose, user, onSave, isLoading }: EditRoleDialogProps) {
  const [role, setRole] = useState(user?.role || 'operador');
  const [isSuperAdmin, setIsSuperAdmin] = useState(user?.is_super_admin || false);

  useEffect(() => {
    if (user) {
      setRole(user.role);
      setIsSuperAdmin(user.is_super_admin);
    }
  }, [user]);

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white">
        <DialogHeader>
          <DialogTitle>Editar Permissoes</DialogTitle>
          <DialogDescription className="text-slate-400">
            Alterar o papel de {user.nome}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Role</label>
            <Select value={role} onValueChange={setRole}>
              <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-slate-800 border-slate-700">
                <SelectItem value="operador" className="text-white hover:bg-slate-700">
                  Operador
                </SelectItem>
                <SelectItem value="gestor" className="text-white hover:bg-slate-700">
                  Gestor
                </SelectItem>
                <SelectItem value="admin" className="text-white hover:bg-slate-700">
                  Administrador
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center gap-3 p-3 bg-slate-900 rounded-lg">
            <input
              type="checkbox"
              id="superadmin"
              checked={isSuperAdmin}
              onChange={(e) => setIsSuperAdmin(e.target.checked)}
              className="h-4 w-4 rounded border-slate-600 bg-slate-900 text-violet-600 focus:ring-violet-500"
            />
            <label htmlFor="superadmin" className="text-sm text-slate-300">
              Super Administrador (acesso global)
            </label>
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
            onClick={() => onSave(user.id, role, isSuperAdmin)}
            disabled={isLoading}
            className="bg-violet-600 hover:bg-violet-700"
          >
            {isLoading ? 'Salvando...' : 'Salvar'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// Ban User Dialog
interface BanDialogProps {
  open: boolean;
  onClose: () => void;
  user: AdminUser | null;
  onBan: (userId: string, reason: string) => void;
  isLoading?: boolean;
}

function BanDialog({ open, onClose, user, onBan, isLoading }: BanDialogProps) {
  const [reason, setReason] = useState('');

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white">
        <DialogHeader>
          <DialogTitle className="text-red-400">Banir Usuario</DialogTitle>
          <DialogDescription className="text-slate-400">
            Desativar a conta de {user.nome}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="p-4 bg-red-400/10 border border-red-400/20 rounded-lg">
            <p className="text-sm text-red-400">
              Esta acao impedira o usuario de acessar o sistema. A acao sera registrada no log de auditoria.
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">
              Motivo do ban (obrigatorio)
            </label>
            <Input
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Informe o motivo..."
              className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
            />
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
            onClick={() => onBan(user.id, reason)}
            disabled={isLoading || !reason.trim()}
            className="bg-red-600 hover:bg-red-700"
          >
            {isLoading ? 'Banindo...' : 'Confirmar Ban'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// View User Dialog
interface ViewUserDialogProps {
  open: boolean;
  onClose: () => void;
  user: AdminUser | null;
}

function ViewUserDialog({ open, onClose, user }: ViewUserDialogProps) {
  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white">
        <DialogHeader>
          <DialogTitle>Detalhes do Usuario</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="flex items-center gap-4">
            <div className="h-16 w-16 rounded-full bg-violet-600/20 flex items-center justify-center">
              <Users className="h-8 w-8 text-violet-400" />
            </div>
            <div>
              <h3 className="text-lg font-medium text-white">{user.nome}</h3>
              <p className="text-sm text-slate-400">{user.email}</p>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-slate-700">
            <div>
              <label className="text-xs text-slate-500">Role</label>
              <p className="text-sm text-white">{user.role}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Tenant</label>
              <p className="text-sm text-white">{user.tenant_name || '-'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Status</label>
              <div className="flex items-center gap-2">
                {user.is_active ? (
                  <>
                    <CheckCircle className="h-4 w-4 text-emerald-400" />
                    <span className="text-sm text-emerald-400">Ativo</span>
                  </>
                ) : (
                  <>
                    <XCircle className="h-4 w-4 text-red-400" />
                    <span className="text-sm text-red-400">Inativo</span>
                  </>
                )}
              </div>
            </div>
            <div>
              <label className="text-xs text-slate-500">Super Admin</label>
              <p className="text-sm text-white">{user.is_super_admin ? 'Sim' : 'Nao'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Criado em</label>
              <p className="text-sm text-white">
                {new Date(user.created_at).toLocaleDateString('pt-BR')}
              </p>
            </div>
            {user.banned_at && (
              <div className="col-span-2">
                <label className="text-xs text-slate-500">Ban</label>
                <p className="text-sm text-red-400">
                  Banido em {new Date(user.banned_at).toLocaleDateString('pt-BR')}
                  {user.ban_reason && ` - ${user.ban_reason}`}
                </p>
              </div>
            )}
          </div>
        </div>
        <DialogFooter>
          <Button
            onClick={onClose}
            className="bg-violet-600 hover:bg-violet-700"
          >
            Fechar
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function UsersListPage() {
  const [users, setUsers] = useState<AdminUser[]>([]);
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
  const [impersonateDialogOpen, setImpersonateDialogOpen] = useState(false);
  const [editRoleDialogOpen, setEditRoleDialogOpen] = useState(false);
  const [banDialogOpen, setBanDialogOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);
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

  const loadUsers = useCallback(async () => {
    try {
      setIsLoading(true);
      const response: PaginatedAdminResponse<AdminUser> = await fetchAdminUsers({
        page,
        per_page: perPage,
        search: search || undefined,
        tenant_id: tenantFilter !== 'all' ? tenantFilter : undefined,
      });
      setUsers(response.data);
      setTotalPages(response.meta.total_pages);
      setTotalItems(response.meta.total);
    } catch (err) {
      console.error('Failed to fetch users:', err);
      toast.error('Falha ao carregar usuarios');
      setUsers([]);
      setTotalPages(1);
      setTotalItems(0);
    } finally {
      setIsLoading(false);
    }
  }, [page, perPage, search, tenantFilter]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handleTenantFilter = (value: string) => {
    setTenantFilter(value);
    setPage(1);
  };

  // Actions
  const handleView = (user: AdminUser) => {
    setSelectedUser(user);
    setViewDialogOpen(true);
  };

  const handleImpersonate = async (userId: string) => {
    try {
      setIsActionLoading(true);
      const result = await impersonateUser(userId);

      // Open new tab with impersonation token
      const url = `/dashboard?impersonate_token=${result.access_token}`;
      window.open(url, '_blank');

      toast.success('Sessao de impersonalizacao iniciada em nova aba');
      setImpersonateDialogOpen(false);
    } catch (err) {
      console.error('Failed to impersonate user:', err);
      toast.error('Falha ao iniciar impersonalizacao');
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleEditRole = async (userId: string, role: string, isSuperAdmin: boolean) => {
    try {
      setIsActionLoading(true);
      await updateAdminUserRole(userId, { role, is_super_admin: isSuperAdmin });
      toast.success('Permissoes atualizadas com sucesso');
      setEditRoleDialogOpen(false);
      loadUsers();
    } catch (err) {
      console.error('Failed to update user role:', err);
      toast.error('Falha ao atualizar permissoes');
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleBan = async (userId: string, reason: string) => {
    try {
      setIsActionLoading(true);
      await banAdminUser(userId, { reason });
      toast.success('Usuario banido com sucesso');
      setBanDialogOpen(false);
      loadUsers();
    } catch (err) {
      console.error('Failed to ban user:', err);
      toast.error('Falha ao banir usuario');
    } finally {
      setIsActionLoading(false);
    }
  };

  const handleResetPassword = async (user: AdminUser) => {
    try {
      const result = await resetAdminUserPassword(user.id);
      toast.success(result.message || 'Senha resetada com sucesso');
      if (result.temporary_password) {
        toast.info(`Senha temporaria: ${result.temporary_password}`, {
          duration: 10000,
        });
      }
    } catch (err) {
      console.error('Failed to reset password:', err);
      toast.error('Falha ao resetar senha');
    }
  };

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'bg-violet-400/10 text-violet-400 border-violet-400/20';
      case 'gestor':
        return 'bg-blue-400/10 text-blue-400 border-blue-400/20';
      default:
        return 'bg-slate-600/20 text-slate-400 border-slate-600/20';
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">
          Gerenciar Usuarios
        </h1>
        <p className="text-slate-400">
          Visualize e gerencie usuarios de todos os tenants
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
                placeholder="Buscar por email ou nome..."
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
          <CardTitle className="text-white">Lista de Usuarios</CardTitle>
          <CardDescription className="text-slate-400">
            {totalItems} usuario(s) encontrado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
            </div>
          ) : users.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Users className="h-12 w-12 text-slate-600 mb-4" />
              <h3 className="text-lg font-medium text-slate-300">
                Nenhum usuario encontrado
              </h3>
              <p className="text-sm text-slate-500 mt-1">
                {search || tenantFilter !== 'all'
                  ? 'Tente ajustar os filtros de busca'
                  : 'Nao ha usuarios cadastrados'}
              </p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-slate-700 hover:bg-slate-800">
                      <TableHead className="text-slate-400">Usuario</TableHead>
                      <TableHead className="text-slate-400">Role</TableHead>
                      <TableHead className="text-slate-400">Tenant</TableHead>
                      <TableHead className="text-slate-400">Status</TableHead>
                      <TableHead className="text-slate-400 text-right">Acoes</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {users.map((user) => (
                      <TableRow
                        key={user.id}
                        className="border-slate-700 hover:bg-slate-800/50"
                      >
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-full bg-violet-600/20 flex items-center justify-center">
                              <Users className="h-4 w-4 text-violet-400" />
                            </div>
                            <div>
                              <p className="font-medium text-white">{user.nome}</p>
                              <p className="text-xs text-slate-400">{user.email}</p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className={getRoleBadgeColor(user.role)}
                            >
                              {user.role}
                            </Badge>
                            {user.is_super_admin && (
                              <Badge className="bg-amber-400/10 text-amber-400 border-amber-400/20 text-xs">
                                Super
                              </Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-slate-400">
                          {user.tenant_name || '-'}
                        </TableCell>
                        <TableCell>
                          {user.is_active ? (
                            <Badge className="bg-emerald-400/10 text-emerald-400 border-emerald-400/20">
                              Ativo
                            </Badge>
                          ) : (
                            <Badge className="bg-red-400/10 text-red-400 border-red-400/20">
                              Inativo
                            </Badge>
                          )}
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
                                onClick={() => handleView(user)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <Eye className="h-4 w-4 mr-2" />
                                Ver detalhes
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => {
                                  setSelectedUser(user);
                                  setImpersonateDialogOpen(true);
                                }}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <LogIn className="h-4 w-4 mr-2" />
                                Impersonalizar
                              </DropdownMenuItem>
                              <DropdownMenuSeparator className="bg-slate-700" />
                              <DropdownMenuItem
                                onClick={() => {
                                  setSelectedUser(user);
                                  setEditRoleDialogOpen(true);
                                }}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <UserCog className="h-4 w-4 mr-2" />
                                Editar permissoes
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleResetPassword(user)}
                                className="text-slate-300 hover:bg-slate-700 hover:text-white cursor-pointer"
                              >
                                <KeyRound className="h-4 w-4 mr-2" />
                                Resetar senha
                              </DropdownMenuItem>
                              <DropdownMenuSeparator className="bg-slate-700" />
                              <DropdownMenuItem
                                onClick={() => {
                                  setSelectedUser(user);
                                  setBanDialogOpen(true);
                                }}
                                disabled={!user.is_active}
                                className="text-red-400 hover:bg-slate-700 hover:text-red-300 cursor-pointer"
                              >
                                <Ban className="h-4 w-4 mr-2" />
                                Banir usuario
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
      <ViewUserDialog
        open={viewDialogOpen}
        onClose={() => setViewDialogOpen(false)}
        user={selectedUser}
      />
      <ImpersonateDialog
        open={impersonateDialogOpen}
        onClose={() => setImpersonateDialogOpen(false)}
        user={selectedUser}
        onImpersonate={handleImpersonate}
        isLoading={isActionLoading}
      />
      <EditRoleDialog
        open={editRoleDialogOpen}
        onClose={() => setEditRoleDialogOpen(false)}
        user={selectedUser}
        onSave={handleEditRole}
        isLoading={isActionLoading}
      />
      <BanDialog
        open={banDialogOpen}
        onClose={() => setBanDialogOpen(false)}
        user={selectedUser}
        onBan={handleBan}
        isLoading={isActionLoading}
      />
    </div>
  );
}
