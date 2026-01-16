'use client';

import React from 'react';
import {
  Database,
  HardDrive,
  Radio,
  Cpu,
  Wifi,
  Server,
  AlertCircle
} from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import {
  ComponentStatus,
  ServiceStatus,
  STATUS_COLORS,
  LATENCY_THRESHOLDS,
} from '@/types/health';
import { cn } from '@/lib/utils';

interface ServiceStatusCardProps {
  serviceKey: string;
  component: ComponentStatus;
  className?: string;
}

// Icon mapping
const ICON_MAP: Record<string, React.ComponentType<{ className?: string }>> = {
  Database: Database,
  HardDrive: HardDrive,
  Radio: Radio,
  Cpu: Cpu,
  Wifi: Wifi,
  Server: Server,
};

// Get icon based on service key
function getServiceIcon(serviceKey: string): React.ComponentType<{ className?: string }> {
  const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
    database: Database,
    redis: HardDrive,
    listener: Radio,
    triagem_motor: Cpu,
    sse_hub: Wifi,
    api: Server,
  };
  return iconMap[serviceKey] || Server;
}

// Format latency for display
function formatLatency(latencyMs: number): string {
  if (latencyMs < 1) {
    return '< 1ms';
  }
  if (latencyMs < 1000) {
    return `${Math.round(latencyMs)}ms`;
  }
  return `${(latencyMs / 1000).toFixed(2)}s`;
}

// Get latency status indicator
function getLatencyStatus(latencyMs: number): ServiceStatus {
  if (latencyMs < LATENCY_THRESHOLDS.OK) {
    return 'up';
  }
  if (latencyMs < LATENCY_THRESHOLDS.DEGRADED) {
    return 'degraded';
  }
  return 'down';
}

export function ServiceStatusCard({
  serviceKey,
  component,
  className
}: ServiceStatusCardProps) {
  const Icon = getServiceIcon(serviceKey);
  const statusConfig = STATUS_COLORS[component.status];
  const latencyStatus = getLatencyStatus(component.latency_ms);

  // Format last check time
  const lastCheckDate = new Date(component.last_check);
  const formattedLastCheck = lastCheckDate.toLocaleTimeString('pt-BR', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });

  return (
    <Card
      className={cn(
        'transition-all duration-200 hover:shadow-md',
        statusConfig.bg,
        statusConfig.border,
        className
      )}
    >
      <CardContent className="p-4">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={cn(
              'p-2 rounded-lg',
              component.status === 'up' ? 'bg-emerald-100' :
              component.status === 'degraded' ? 'bg-amber-100' : 'bg-red-100'
            )}>
              <Icon className={cn('h-5 w-5', statusConfig.text)} />
            </div>
            <div>
              <h3 className="font-medium text-gray-900">
                {component.name}
              </h3>
              <p className="text-sm text-gray-500">
                Ultima verificacao: {formattedLastCheck}
              </p>
            </div>
          </div>

          {/* Status Indicator */}
          <div className="flex flex-col items-end gap-1">
            <div className="flex items-center gap-2">
              <span className={cn(
                'h-3 w-3 rounded-full animate-pulse',
                statusConfig.indicator
              )} />
              <span className={cn('text-sm font-medium', statusConfig.text)}>
                {statusConfig.label}
              </span>
            </div>

            {/* Latency */}
            <span className={cn(
              'text-xs',
              latencyStatus === 'up' ? 'text-gray-500' :
              latencyStatus === 'degraded' ? 'text-amber-600' : 'text-red-600'
            )}>
              Latencia: {formatLatency(component.latency_ms)}
            </span>
          </div>
        </div>

        {/* Error Message */}
        {component.message && component.status !== 'up' && (
          <div className="mt-3 flex items-start gap-2 p-2 bg-white/50 rounded-md">
            <AlertCircle className="h-4 w-4 text-red-500 flex-shrink-0 mt-0.5" />
            <p className="text-sm text-red-700">
              {component.message}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default ServiceStatusCard;
