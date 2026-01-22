'use client';

import { Activity, Database, Radio, Zap, Clock, RefreshCw, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useSystemHealth } from '@/hooks';
import type { ServiceStatus, ServiceHealth } from '@/types';

const statusConfig: Record<
  string,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; icon: typeof CheckCircle; className: string }
> = {
  healthy: {
    label: 'Saudavel',
    variant: 'default',
    icon: CheckCircle,
    className: 'text-green-500',
  },
  up: {
    label: 'Online',
    variant: 'default',
    icon: CheckCircle,
    className: 'text-green-500',
  },
  degraded: {
    label: 'Degradado',
    variant: 'secondary',
    icon: AlertTriangle,
    className: 'text-yellow-500',
  },
  unhealthy: {
    label: 'Indisponivel',
    variant: 'destructive',
    icon: XCircle,
    className: 'text-red-500',
  },
  down: {
    label: 'Offline',
    variant: 'destructive',
    icon: XCircle,
    className: 'text-red-500',
  },
  unknown: {
    label: 'Desconhecido',
    variant: 'outline',
    icon: AlertTriangle,
    className: 'text-gray-500',
  },
};

const serviceIcons: Record<string, typeof Database> = {
  database: Database,
  redis: Database,
  listener: Radio,
  triagem: Zap,
  sse: Activity,
};

function ServiceCard({ name, service }: { name: string; service: ServiceHealth }) {
  const config = statusConfig[service.status] || statusConfig.unknown;
  const Icon = serviceIcons[name] || Activity;
  const StatusIcon = config.icon;

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Icon className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-base">{service.name}</CardTitle>
          </div>
          <Badge variant={config.variant} className="gap-1">
            <StatusIcon className={`h-3 w-3 ${config.className}`} />
            {config.label}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-1 text-sm">
          {service.latency_ms !== undefined && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Latencia</span>
              <span className="font-mono">{service.latency_ms}ms</span>
            </div>
          )}
          {service.last_check && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Ultima verificacao</span>
              <span>
                {new Date(service.last_check).toLocaleTimeString('pt-BR')}
              </span>
            </div>
          )}
          {service.error && (
            <p className="text-sm text-destructive mt-2">{service.error}</p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  const parts = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);

  return parts.join(' ') || '< 1m';
}

export default function StatusPage() {
  const { data: health, isLoading, refetch, isFetching } = useSystemHealth();

  const overallConfig = health ? (statusConfig[health.status] || statusConfig.unknown) : statusConfig.unknown;
  const OverallIcon = overallConfig.icon;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Status do Sistema</h1>
          <p className="text-muted-foreground">
            Monitoramento de saude dos servicos do SIDOT
          </p>
        </div>
        <Button
          variant="outline"
          onClick={() => refetch()}
          disabled={isFetching}
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${isFetching ? 'animate-spin' : ''}`} />
          Atualizar
        </Button>
      </div>

      {isLoading ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
          </CardContent>
        </Card>
      ) : (
        <>
          {/* Overall Status */}
          <Card
            className={`border-2 ${
              health?.status === 'healthy'
                ? 'border-green-200 bg-green-50'
                : health?.status === 'degraded'
                ? 'border-yellow-200 bg-yellow-50'
                : 'border-red-200 bg-red-50'
            }`}
          >
            <CardHeader>
              <div className="flex items-center gap-3">
                <OverallIcon className={`h-8 w-8 ${overallConfig.className}`} />
                <div>
                  <CardTitle>Status Geral: {overallConfig.label}</CardTitle>
                  <CardDescription>
                    {health?.timestamp &&
                      `Atualizado em ${new Date(health.timestamp).toLocaleString('pt-BR')}`}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            {health && (
              <CardContent>
                <div className="flex items-center gap-2 text-sm">
                  <Clock className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Uptime:</span>
                  <span className="font-medium">{formatUptime(health.uptime_seconds)}</span>
                </div>
              </CardContent>
            )}
          </Card>

          {/* Individual Services */}
          {health?.services && (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {Object.entries(health.services).map(([key, service]) => (
                <ServiceCard key={key} name={key} service={service} />
              ))}
            </div>
          )}

          {/* Info Card */}
          <Card className="bg-muted/50">
            <CardContent className="flex items-start gap-3 pt-6">
              <Activity className="h-5 w-5 text-primary mt-0.5" />
              <div className="text-sm">
                <p className="font-medium">Sobre o monitoramento</p>
                <ul className="mt-2 space-y-1 text-muted-foreground">
                  <li><strong>Database:</strong> Conexao com PostgreSQL para persistencia de dados</li>
                  <li><strong>Redis:</strong> Cache e filas de mensagens para processamento assincrono</li>
                  <li><strong>Listener:</strong> Servico de escuta de novos obitos nos hospitais</li>
                  <li><strong>Triagem:</strong> Motor de regras para elegibilidade de doacao</li>
                  <li><strong>SSE:</strong> Notificacoes em tempo real para o dashboard</li>
                </ul>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
