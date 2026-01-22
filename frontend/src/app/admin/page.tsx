'use client';

import { useEffect, useState } from 'react';
import {
  Building,
  Users,
  Building2,
  Activity,
  AlertCircle,
  CheckCircle,
  TrendingUp,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { fetchAdminDashboardMetrics, type AdminDashboardMetrics } from '@/lib/api/admin';

interface MetricCardProps {
  title: string;
  value: string | number;
  description?: string;
  icon: React.ElementType;
  iconColor?: string;
  trend?: {
    value: number;
    isPositive: boolean;
  };
}

function MetricCard({
  title,
  value,
  description,
  icon: Icon,
  iconColor = 'text-violet-400',
  trend,
}: MetricCardProps) {
  return (
    <Card className="bg-slate-800 border-slate-700">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-slate-300">{title}</CardTitle>
        <Icon className={`h-5 w-5 ${iconColor}`} />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold text-white">{value}</div>
        {description && (
          <p className="text-xs text-slate-400 mt-1">{description}</p>
        )}
        {trend && (
          <div className="flex items-center gap-1 mt-2">
            <TrendingUp
              className={`h-3 w-3 ${
                trend.isPositive ? 'text-emerald-400' : 'text-red-400 rotate-180'
              }`}
            />
            <span
              className={`text-xs ${
                trend.isPositive ? 'text-emerald-400' : 'text-red-400'
              }`}
            >
              {trend.value}% vs. ultimo mes
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

interface SystemStatusCardProps {
  status: 'healthy' | 'degraded' | 'unhealthy';
  services: {
    name: string;
    status: 'healthy' | 'degraded' | 'unhealthy';
  }[];
}

function SystemStatusCard({ status, services }: SystemStatusCardProps) {
  const statusConfig = {
    healthy: {
      label: 'Todos os sistemas operacionais',
      icon: CheckCircle,
      color: 'text-emerald-400',
      bgColor: 'bg-emerald-400/10',
    },
    degraded: {
      label: 'Alguns sistemas com problemas',
      icon: AlertCircle,
      color: 'text-amber-400',
      bgColor: 'bg-amber-400/10',
    },
    unhealthy: {
      label: 'Sistema com problemas criticos',
      icon: AlertCircle,
      color: 'text-red-400',
      bgColor: 'bg-red-400/10',
    },
  };

  const config = statusConfig[status];
  const StatusIcon = config.icon;

  return (
    <Card className="bg-slate-800 border-slate-700">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-white">
          Status do Sistema
        </CardTitle>
        <CardDescription className="text-slate-400">
          Monitoramento em tempo real
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div
          className={`flex items-center gap-3 p-3 rounded-lg ${config.bgColor}`}
        >
          <StatusIcon className={`h-5 w-5 ${config.color}`} />
          <span className={`text-sm font-medium ${config.color}`}>
            {config.label}
          </span>
        </div>

        <div className="mt-4 space-y-2">
          {services.map((service) => (
            <div
              key={service.name}
              className="flex items-center justify-between py-2 border-b border-slate-700 last:border-0"
            >
              <span className="text-sm text-slate-300">{service.name}</span>
              <span
                className={`text-xs font-medium px-2 py-1 rounded-full ${
                  service.status === 'healthy'
                    ? 'bg-emerald-400/10 text-emerald-400'
                    : service.status === 'degraded'
                    ? 'bg-amber-400/10 text-amber-400'
                    : 'bg-red-400/10 text-red-400'
                }`}
              >
                {service.status === 'healthy'
                  ? 'Operacional'
                  : service.status === 'degraded'
                  ? 'Degradado'
                  : 'Indisponivel'}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

interface RecentActivityItem {
  id: string;
  action: string;
  description: string;
  tenant?: string;
  timestamp: string;
}

function RecentActivityCard({ activities }: { activities: RecentActivityItem[] }) {
  return (
    <Card className="bg-slate-800 border-slate-700">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-white">
          Atividade Recente
        </CardTitle>
        <CardDescription className="text-slate-400">
          Ultimas acoes no sistema
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {activities.length === 0 ? (
            <p className="text-sm text-slate-400 text-center py-4">
              Nenhuma atividade recente
            </p>
          ) : (
            activities.map((activity) => (
              <div
                key={activity.id}
                className="flex items-start gap-3 pb-3 border-b border-slate-700 last:border-0"
              >
                <div className="h-2 w-2 mt-2 rounded-full bg-violet-400 shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-slate-200">
                    {activity.action}
                  </p>
                  <p className="text-xs text-slate-400 truncate">
                    {activity.description}
                  </p>
                  {activity.tenant && (
                    <p className="text-xs text-violet-400 mt-1">
                      Tenant: {activity.tenant}
                    </p>
                  )}
                </div>
                <span className="text-xs text-slate-500 shrink-0">
                  {activity.timestamp}
                </span>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default function AdminDashboardPage() {
  const [metrics, setMetrics] = useState<AdminDashboardMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadMetrics() {
      try {
        setIsLoading(true);
        const data = await fetchAdminDashboardMetrics();
        setMetrics(data);
      } catch (err) {
        console.error('Failed to fetch admin metrics:', err);
        setError('Falha ao carregar metricas');
        // Set default metrics on error for development
        setMetrics({
          total_tenants: 0,
          active_tenants: 0,
          inactive_tenants: 0,
          total_users: 0,
          total_hospitals: 0,
          total_occurrences: 0,
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadMetrics();
  }, []);

  // Mock data for system status and recent activity
  // In production, these would come from API endpoints
  const systemServices = [
    { name: 'API Principal', status: 'healthy' as const },
    { name: 'Banco de Dados', status: 'healthy' as const },
    { name: 'Redis Cache', status: 'healthy' as const },
    { name: 'Servico de Email', status: 'healthy' as const },
    { name: 'SSE Notificacoes', status: 'healthy' as const },
  ];

  const recentActivities: RecentActivityItem[] = [
    {
      id: '1',
      action: 'Novo tenant criado',
      description: 'Hospital Sao Paulo foi adicionado ao sistema',
      tenant: 'Hospital Sao Paulo',
      timestamp: 'Agora',
    },
    {
      id: '2',
      action: 'Usuario promovido',
      description: 'joao@email.com foi promovido para gestor',
      tenant: 'Hospital Central',
      timestamp: '5 min',
    },
    {
      id: '3',
      action: 'Configuracao alterada',
      description: 'Configuracoes de SMTP atualizadas',
      timestamp: '15 min',
    },
    {
      id: '4',
      action: 'Template clonado',
      description: 'Regra de triagem copiada para 3 tenants',
      timestamp: '1 hora',
    },
  ];

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Dashboard Administrativo
          </h1>
          <p className="text-slate-400">
            Visao geral do sistema SIDOT
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="bg-slate-800 border-slate-700">
              <CardContent className="pt-6">
                <div className="h-20 animate-pulse bg-slate-700 rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Title */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">
          Dashboard Administrativo
        </h1>
        <p className="text-slate-400">
          Visao geral do sistema SIDOT
        </p>
      </div>

      {error && (
        <div className="bg-amber-400/10 border border-amber-400/20 rounded-lg p-4">
          <p className="text-sm text-amber-400">{error}</p>
        </div>
      )}

      {/* Metrics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total de Tenants"
          value={metrics?.total_tenants ?? 0}
          description={`${metrics?.active_tenants ?? 0} ativos / ${metrics?.inactive_tenants ?? 0} inativos`}
          icon={Building}
          iconColor="text-violet-400"
        />
        <MetricCard
          title="Total de Usuarios"
          value={metrics?.total_users ?? 0}
          description="Em todos os tenants"
          icon={Users}
          iconColor="text-blue-400"
        />
        <MetricCard
          title="Total de Hospitais"
          value={metrics?.total_hospitals ?? 0}
          description="Cadastrados no sistema"
          icon={Building2}
          iconColor="text-emerald-400"
        />
        <MetricCard
          title="Total de Ocorrencias"
          value={metrics?.total_occurrences ?? 0}
          description="Desde o inicio"
          icon={Activity}
          iconColor="text-amber-400"
        />
      </div>

      {/* Secondary Section: System Status and Recent Activity */}
      <div className="grid gap-6 md:grid-cols-2">
        <SystemStatusCard status="healthy" services={systemServices} />
        <RecentActivityCard activities={recentActivities} />
      </div>
    </div>
  );
}
