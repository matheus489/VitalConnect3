'use client';

import { Users, Mail, Shield } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

// Mock data for demonstration
const users = [
  {
    id: '1',
    nome: 'Administrador',
    email: 'admin@vitalconnect.gov.br',
    role: 'admin',
  },
  {
    id: '2',
    nome: 'Gestor',
    email: 'gestor@vitalconnect.gov.br',
    role: 'gestor',
  },
  {
    id: '3',
    nome: 'Operador',
    email: 'operador@vitalconnect.gov.br',
    role: 'operador',
  },
];

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

export default function UsersPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Usuarios</h1>
        <p className="text-muted-foreground">
          Usuarios cadastrados no sistema
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {users.map((user) => {
          const roleInfo = getRoleLabel(user.role);

          return (
            <Card key={user.id}>
              <CardHeader className="flex flex-row items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                  <Users className="h-6 w-6 text-primary" />
                </div>
                <div className="flex-1">
                  <CardTitle className="text-base">{user.nome}</CardTitle>
                  <Badge variant={roleInfo.variant} className="mt-1">
                    {roleInfo.label}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Mail className="h-4 w-4" />
                  {user.email}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Info Card */}
      <Card className="bg-muted/50">
        <CardContent className="flex items-start gap-3 pt-6">
          <Shield className="h-5 w-5 text-primary mt-0.5" />
          <div className="text-sm">
            <p className="font-medium">Sobre os perfis de acesso</p>
            <ul className="mt-2 space-y-1 text-muted-foreground">
              <li><strong>Administrador:</strong> Acesso completo ao sistema, incluindo gestao de usuarios e hospitais</li>
              <li><strong>Gestor:</strong> Acesso as configuracoes de triagem e visualizacao de metricas</li>
              <li><strong>Operador:</strong> Acesso a gestao de ocorrencias e acoes de captacao</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
