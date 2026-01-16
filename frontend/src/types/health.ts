// Health Check Type Definitions

export type ServiceStatus = 'up' | 'degraded' | 'down';

export interface ComponentStatus {
  name: string;
  status: ServiceStatus;
  latency_ms: number;
  last_check: string;
  message?: string;
}

export interface HealthSummary {
  status: ServiceStatus;
  timestamp: string;
  components: Record<string, ComponentStatus>;
}

// Service display configuration
export interface ServiceConfig {
  key: string;
  name: string;
  icon: string;
  priority: number;
}

// Available services with display configuration
export const SERVICE_CONFIG: ServiceConfig[] = [
  { key: 'database', name: 'Database (PostgreSQL)', icon: 'Database', priority: 1 },
  { key: 'redis', name: 'Redis', icon: 'HardDrive', priority: 2 },
  { key: 'listener', name: 'Obito Listener', icon: 'Radio', priority: 3 },
  { key: 'triagem_motor', name: 'Triagem Motor', icon: 'Cpu', priority: 4 },
  { key: 'sse_hub', name: 'SSE Hub', icon: 'Wifi', priority: 5 },
  { key: 'api', name: 'API', icon: 'Server', priority: 6 },
];

// Status color configuration
export interface StatusColorConfig {
  bg: string;
  border: string;
  text: string;
  indicator: string;
  label: string;
}

export const STATUS_COLORS: Record<ServiceStatus, StatusColorConfig> = {
  up: {
    bg: 'bg-emerald-50',
    border: 'border-emerald-200',
    text: 'text-emerald-700',
    indicator: 'bg-emerald-500',
    label: 'Operacional',
  },
  degraded: {
    bg: 'bg-amber-50',
    border: 'border-amber-200',
    text: 'text-amber-700',
    indicator: 'bg-amber-500',
    label: 'Latencia Alta',
  },
  down: {
    bg: 'bg-red-50',
    border: 'border-red-200',
    text: 'text-red-700',
    indicator: 'bg-red-500',
    label: 'Fora do Ar',
  },
};

// Latency thresholds in milliseconds
export const LATENCY_THRESHOLDS = {
  OK: 500,
  DEGRADED: 2000,
};
