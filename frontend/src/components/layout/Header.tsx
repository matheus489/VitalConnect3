'use client';

import { Bell, BellOff, LogOut, Menu, Volume2, VolumeX, User } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useAuth } from '@/hooks/useAuth';
import { useSSE } from '@/hooks/useSSE';
import { cn } from '@/lib/utils';

interface HeaderProps {
  onMenuToggle: () => void;
  showMenuButton?: boolean;
}

export function Header({ onMenuToggle, showMenuButton = false }: HeaderProps) {
  const { user, logout } = useAuth();
  const { isConnected, pendingCount, soundEnabled, toggleSound } = useSSE();
  const queryClient = useQueryClient();

  const handleLogout = () => {
    // Clear all React Query cache before logout to prevent stale data on next login
    queryClient.clear();
    logout();
  };

  const getRoleLabel = (role: string) => {
    switch (role) {
      case 'admin':
        return 'Administrador';
      case 'gestor':
        return 'Gestor';
      case 'operador':
        return 'Operador';
      default:
        return role;
    }
  };

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b bg-background px-4 lg:px-6">
      <div className="flex items-center gap-4">
        {showMenuButton && (
          <Button
            variant="ghost"
            size="icon"
            onClick={onMenuToggle}
            className="lg:hidden"
            aria-label="Abrir menu"
          >
            <Menu className="h-5 w-5" />
          </Button>
        )}

        {/* Connection Status */}
        <div className="hidden items-center gap-2 sm:flex">
          <div
            className={cn(
              'h-2 w-2 rounded-full',
              isConnected ? 'bg-emerald-500' : 'bg-muted-foreground'
            )}
          />
          <span className="text-xs text-muted-foreground">
            {isConnected ? 'Conectado' : 'Desconectado'}
          </span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Notification Badge */}
        <div className="relative">
          <Button variant="ghost" size="icon" aria-label="Notificacoes">
            {pendingCount > 0 ? (
              <Bell className="h-5 w-5" />
            ) : (
              <BellOff className="h-5 w-5 text-muted-foreground" />
            )}
          </Button>
          {pendingCount > 0 && (
            <Badge
              variant="destructive"
              className={cn(
                'absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center px-1',
                'animate-pulse-alert'
              )}
            >
              {pendingCount > 99 ? '99+' : pendingCount}
            </Badge>
          )}
        </div>

        {/* Sound Toggle */}
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleSound}
          aria-label={soundEnabled ? 'Desativar som' : 'Ativar som'}
        >
          {soundEnabled ? (
            <Volume2 className="h-5 w-5" />
          ) : (
            <VolumeX className="h-5 w-5 text-muted-foreground" />
          )}
        </Button>

        {/* User Info */}
        <div className="hidden items-center gap-3 border-l pl-4 sm:flex">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary">
            <User className="h-4 w-4 text-primary-foreground" />
          </div>
          <div className="hidden flex-col md:flex">
            <span className="text-sm font-medium">{user?.nome || 'Usuario'}</span>
            <span className="text-xs text-muted-foreground">
              {user ? getRoleLabel(user.role) : ''}
            </span>
          </div>
        </div>

        {/* Logout */}
        <Button
          variant="ghost"
          size="icon"
          onClick={handleLogout}
          aria-label="Sair do sistema"
          className="text-muted-foreground hover:text-foreground"
        >
          <LogOut className="h-5 w-5" />
        </Button>
      </div>
    </header>
  );
}
