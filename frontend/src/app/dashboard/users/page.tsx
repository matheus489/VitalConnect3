'use client';

import { useState, useEffect } from 'react';
import {
  Users,
  Mail,
  Shield,
  Plus,
  Pencil,
  Trash2,
  Search,
  Phone,
  Building2,
  Loader2,
  X,
  Eye,
  EyeOff,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Switch } from '@/components/ui/switch';
import { toast } from 'sonner';
import { useAuth } from '@/hooks/useAuth';
import {
  listUsers,
  createUser,
  updateUser,
  deleteUser,
  type User,
  type UserRole,
  type CreateUserInput,
  type UpdateUserInput,
} from '@/lib/api/users';

function getRoleLabel(role: string) {
  switch (role) {
    case 'admin':
      return { label: 'Administrador', variant: 'destructive' as const };
    case 'gestor':
      return { label: 'Gestor', variant: 'default' as const };
    case 'operador':
      return { label: 'Operador', variant: 'secondary' as const };
    default:
      return { label: role, variant: 'outline' as const };
  }
}

interface UserFormData {
  email: string;
  password: string;
  nome: string;
  role: UserRole;
  mobile_phone: string;
  email_notifications: boolean;
}

const initialFormData: UserFormData = {
  email: '',
  password: '',
  nome: '',
  role: 'operador',
  mobile_phone: '',
  email_notifications: true,
};

export default function UsersPage() {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');

  // Dialog states
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [formData, setFormData] = useState<UserFormData>(initialFormData);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Load users
  useEffect(() => {
    loadUsers();
  }, []);

  async function loadUsers() {
    try {
      setIsLoading(true);
      const response = await listUsers();
      setUsers(response.data);
    } catch (error) {
      console.error('Failed to load users:', error);
      toast.error('Erro ao carregar usuarios');
    } finally {
      setIsLoading(false);
    }
  }

  // Filter users
  const filteredUsers = users.filter((user) => {
    const matchesSearch =
      user.nome.toLowerCase().includes(searchTerm.toLowerCase()) ||
      user.email.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesRole = roleFilter === 'all' || user.role === roleFilter;
    return matchesSearch && matchesRole;
  });

  // Handle create user
  async function handleCreateUser() {
    if (!formData.email || !formData.password || !formData.nome) {
      toast.error('Preencha todos os campos obrigatorios');
      return;
    }

    try {
      setIsSubmitting(true);
      const input: CreateUserInput = {
        email: formData.email,
        password: formData.password,
        nome: formData.nome,
        role: formData.role,
        email_notifications: formData.email_notifications,
      };

      if (formData.mobile_phone) {
        input.mobile_phone = formData.mobile_phone;
      }

      await createUser(input);
      toast.success('Usuario criado com sucesso');
      setIsCreateDialogOpen(false);
      setFormData(initialFormData);
      loadUsers();
    } catch (error: unknown) {
      console.error('Failed to create user:', error);
      const err = error as { response?: { data?: { error?: string } } };
      toast.error(err.response?.data?.error || 'Erro ao criar usuario');
    } finally {
      setIsSubmitting(false);
    }
  }

  // Handle edit user
  function openEditDialog(user: User) {
    setSelectedUser(user);
    setFormData({
      email: user.email,
      password: '',
      nome: user.nome,
      role: user.role,
      mobile_phone: user.mobile_phone || '',
      email_notifications: user.email_notifications,
    });
    setIsEditDialogOpen(true);
  }

  async function handleUpdateUser() {
    if (!selectedUser) return;

    try {
      setIsSubmitting(true);
      const input: UpdateUserInput = {
        nome: formData.nome,
        role: formData.role,
        email_notifications: formData.email_notifications,
      };

      if (formData.password) {
        input.password = formData.password;
      }
      if (formData.mobile_phone) {
        input.mobile_phone = formData.mobile_phone;
      }

      await updateUser(selectedUser.id, input);
      toast.success('Usuario atualizado com sucesso');
      setIsEditDialogOpen(false);
      setSelectedUser(null);
      setFormData(initialFormData);
      loadUsers();
    } catch (error: unknown) {
      console.error('Failed to update user:', error);
      const err = error as { response?: { data?: { error?: string } } };
      toast.error(err.response?.data?.error || 'Erro ao atualizar usuario');
    } finally {
      setIsSubmitting(false);
    }
  }

  // Handle delete user
  function openDeleteDialog(user: User) {
    setSelectedUser(user);
    setIsDeleteDialogOpen(true);
  }

  async function handleDeleteUser() {
    if (!selectedUser) return;

    try {
      setIsSubmitting(true);
      await deleteUser(selectedUser.id);
      toast.success('Usuario removido com sucesso');
      setIsDeleteDialogOpen(false);
      setSelectedUser(null);
      loadUsers();
    } catch (error: unknown) {
      console.error('Failed to delete user:', error);
      const err = error as { response?: { data?: { error?: string } } };
      toast.error(err.response?.data?.error || 'Erro ao remover usuario');
    } finally {
      setIsSubmitting(false);
    }
  }

  // Check if current user is admin
  const isAdmin = currentUser?.role === 'admin';

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Usuarios</h1>
          <p className="text-muted-foreground">
            Gerencie os usuarios do seu tenant
          </p>
        </div>
        {isAdmin && (
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Novo Usuario
          </Button>
        )}
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 sm:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Buscar por nome ou email..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select value={roleFilter} onValueChange={setRoleFilter}>
          <SelectTrigger className="w-full sm:w-[180px]">
            <SelectValue placeholder="Filtrar por perfil" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Todos os perfis</SelectItem>
            <SelectItem value="admin">Administrador</SelectItem>
            <SelectItem value="gestor">Gestor</SelectItem>
            <SelectItem value="operador">Operador</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Users Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : filteredUsers.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Users className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">
              {searchTerm || roleFilter !== 'all'
                ? 'Nenhum usuario encontrado com os filtros aplicados'
                : 'Nenhum usuario cadastrado'}
            </p>
            {isAdmin && !searchTerm && roleFilter === 'all' && (
              <Button
                variant="outline"
                className="mt-4"
                onClick={() => setIsCreateDialogOpen(true)}
              >
                <Plus className="mr-2 h-4 w-4" />
                Cadastrar primeiro usuario
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredUsers.map((user) => {
            const roleInfo = getRoleLabel(user.role);
            const isSelf = currentUser?.id === user.id;

            return (
              <Card key={user.id} className={!user.ativo ? 'opacity-60' : ''}>
                <CardHeader className="flex flex-row items-start justify-between gap-3">
                  <div className="flex items-center gap-3">
                    <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                      <Users className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                      <CardTitle className="text-base flex items-center gap-2">
                        {user.nome}
                        {!user.ativo && (
                          <Badge variant="outline" className="text-xs">
                            Inativo
                          </Badge>
                        )}
                      </CardTitle>
                      <Badge variant={roleInfo.variant} className="mt-1">
                        {roleInfo.label}
                      </Badge>
                    </div>
                  </div>
                  {isAdmin && !isSelf && (
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => openEditDialog(user)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:text-destructive"
                        onClick={() => openDeleteDialog(user)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  )}
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Mail className="h-4 w-4" />
                    {user.email}
                  </div>
                  {user.mobile_phone && (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Phone className="h-4 w-4" />
                      {user.mobile_phone}
                    </div>
                  )}
                  {user.hospitals && user.hospitals.length > 0 && (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Building2 className="h-4 w-4" />
                      {user.hospitals.length} hospital(is)
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Info Card */}
      <Card className="bg-muted/50">
        <CardContent className="flex items-start gap-3 pt-6">
          <Shield className="h-5 w-5 text-primary mt-0.5" />
          <div className="text-sm">
            <p className="font-medium">Sobre os perfis de acesso</p>
            <ul className="mt-2 space-y-1 text-muted-foreground">
              <li>
                <strong>Administrador:</strong> Acesso completo ao sistema,
                incluindo gestao de usuarios e hospitais
              </li>
              <li>
                <strong>Gestor:</strong> Acesso as configuracoes de triagem e
                visualizacao de metricas
              </li>
              <li>
                <strong>Operador:</strong> Acesso a gestao de ocorrencias e
                acoes de captacao
              </li>
            </ul>
          </div>
        </CardContent>
      </Card>

      {/* Create User Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Novo Usuario</DialogTitle>
            <DialogDescription>
              Cadastre um novo usuario para o seu tenant.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="nome">Nome *</Label>
              <Input
                id="nome"
                value={formData.nome}
                onChange={(e) =>
                  setFormData({ ...formData, nome: e.target.value })
                }
                placeholder="Nome completo"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="email">Email *</Label>
              <Input
                id="email"
                type="email"
                value={formData.email}
                onChange={(e) =>
                  setFormData({ ...formData, email: e.target.value })
                }
                placeholder="email@exemplo.com"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Senha *</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={(e) =>
                    setFormData({ ...formData, password: e.target.value })
                  }
                  placeholder="Minimo 8 caracteres"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="role">Perfil *</Label>
              <Select
                value={formData.role}
                onValueChange={(value: UserRole) =>
                  setFormData({ ...formData, role: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="operador">Operador</SelectItem>
                  <SelectItem value="gestor">Gestor</SelectItem>
                  <SelectItem value="admin">Administrador</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="mobile_phone">Telefone</Label>
              <Input
                id="mobile_phone"
                type="tel"
                value={formData.mobile_phone}
                onChange={(e) =>
                  setFormData({ ...formData, mobile_phone: e.target.value })
                }
                placeholder="+5511999999999"
              />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="email_notifications">
                Receber notificacoes por email
              </Label>
              <Switch
                id="email_notifications"
                checked={formData.email_notifications}
                onCheckedChange={(checked) =>
                  setFormData({ ...formData, email_notifications: checked })
                }
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsCreateDialogOpen(false)}
            >
              Cancelar
            </Button>
            <Button onClick={handleCreateUser} disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Criar Usuario
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit User Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Editar Usuario</DialogTitle>
            <DialogDescription>
              Atualize as informacoes do usuario.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="edit-nome">Nome</Label>
              <Input
                id="edit-nome"
                value={formData.nome}
                onChange={(e) =>
                  setFormData({ ...formData, nome: e.target.value })
                }
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-email">Email</Label>
              <Input id="edit-email" value={formData.email} disabled />
              <p className="text-xs text-muted-foreground">
                O email nao pode ser alterado
              </p>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-password">Nova Senha</Label>
              <div className="relative">
                <Input
                  id="edit-password"
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={(e) =>
                    setFormData({ ...formData, password: e.target.value })
                  }
                  placeholder="Deixe em branco para manter a atual"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="absolute right-0 top-0 h-full px-3"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-role">Perfil</Label>
              <Select
                value={formData.role}
                onValueChange={(value: UserRole) =>
                  setFormData({ ...formData, role: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="operador">Operador</SelectItem>
                  <SelectItem value="gestor">Gestor</SelectItem>
                  <SelectItem value="admin">Administrador</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="edit-mobile_phone">Telefone</Label>
              <Input
                id="edit-mobile_phone"
                type="tel"
                value={formData.mobile_phone}
                onChange={(e) =>
                  setFormData({ ...formData, mobile_phone: e.target.value })
                }
                placeholder="+5511999999999"
              />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="edit-email_notifications">
                Receber notificacoes por email
              </Label>
              <Switch
                id="edit-email_notifications"
                checked={formData.email_notifications}
                onCheckedChange={(checked) =>
                  setFormData({ ...formData, email_notifications: checked })
                }
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsEditDialogOpen(false)}
            >
              Cancelar
            </Button>
            <Button onClick={handleUpdateUser} disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Salvar Alteracoes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remover Usuario</AlertDialogTitle>
            <AlertDialogDescription>
              Tem certeza que deseja remover o usuario{' '}
              <strong>{selectedUser?.nome}</strong>? Esta acao nao pode ser
              desfeita.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteUser}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Remover
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
