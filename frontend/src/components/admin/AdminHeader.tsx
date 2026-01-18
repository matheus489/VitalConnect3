'use client';

import Link from 'next/link';
import { LogOut, Menu, User, ArrowLeft, Shield } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';

interface AdminHeaderProps {
  onMenuToggle: () => void;
  showMenuButton?: boolean;
}

export function AdminHeader({ onMenuToggle, showMenuButton = false }: AdminHeaderProps) {
  const { user, logout } = useAuth();
  const queryClient = useQueryClient();

  const handleLogout = () => {
    // Clear all React Query cache before logout to prevent stale data on next login
    queryClient.clear();
    logout();
  };

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b bg-slate-900 border-slate-700 px-4 lg:px-6">
      <div className="flex items-center gap-4">
        {showMenuButton && (
          <Button
            variant="ghost"
            size="icon"
            onClick={onMenuToggle}
            className="lg:hidden text-slate-400 hover:bg-slate-800 hover:text-slate-200"
            aria-label="Abrir menu"
          >
            <Menu className="h-5 w-5" />
          </Button>
        )}

        {/* Admin Indicator */}
        <div className="hidden items-center gap-2 sm:flex">
          <div className="flex h-6 w-6 items-center justify-center rounded bg-violet-600/20">
            <Shield className="h-3.5 w-3.5 text-violet-400" />
          </div>
          <span className="text-sm font-medium text-violet-400">
            Painel Administrativo
          </span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Back to App Link */}
        <Link href="/dashboard">
          <Button
            variant="ghost"
            size="sm"
            className="gap-2 text-slate-400 hover:bg-slate-800 hover:text-slate-200"
          >
            <ArrowLeft className="h-4 w-4" />
            <span className="hidden sm:inline">Voltar ao App</span>
          </Button>
        </Link>

        {/* User Info */}
        <div className="hidden items-center gap-3 border-l border-slate-700 pl-4 sm:flex">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-violet-600">
            <User className="h-4 w-4 text-white" />
          </div>
          <div className="hidden flex-col md:flex">
            <span className="text-sm font-medium text-slate-200">
              {user?.nome || 'Usuario'}
            </span>
            <span className="text-xs text-violet-400">
              Super Administrador
            </span>
          </div>
        </div>

        {/* Logout */}
        <Button
          variant="ghost"
          size="icon"
          onClick={handleLogout}
          aria-label="Sair do sistema"
          className="text-slate-400 hover:bg-slate-800 hover:text-slate-200"
        >
          <LogOut className="h-5 w-5" />
        </Button>
      </div>
    </header>
  );
}

export default AdminHeader;
