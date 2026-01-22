// SIDOT Type Definitions

// =============================================================================
// User & Authentication
// =============================================================================

export type UserRole = 'operador' | 'gestor' | 'admin';

export interface User {
  id: string;
  email: string;
  nome: string;
  role: UserRole;
  hospital_id?: string;
  tenant_id?: string;
  is_super_admin?: boolean;
  ativo?: boolean;
  mobile_phone?: string;
  email_notifications?: boolean;
  hospitals?: { id: string; nome: string }[];
  created_at: string;
  updated_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  tokens: AuthTokens;
}

// =============================================================================
// Hospital
// =============================================================================

export interface Hospital {
  id: string;
  nome: string;
  codigo: string;
  endereco?: string;
  telefone?: string;
  latitude?: number;
  longitude?: number;
  config_conexao?: Record<string, unknown>;
  ativo: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * Input for creating a new hospital
 * Coordinates and address are required for map integration
 */
export interface CreateHospitalInput {
  nome: string;
  codigo: string;
  endereco: string;
  telefone?: string;
  latitude: number;
  longitude: number;
  ativo?: boolean;
}

/**
 * Input for updating an existing hospital
 * All fields are optional for partial updates
 */
export interface UpdateHospitalInput {
  nome?: string;
  codigo?: string;
  endereco?: string;
  telefone?: string;
  latitude?: number;
  longitude?: number;
  ativo?: boolean;
}

// =============================================================================
// Occurrence
// =============================================================================

export type OccurrenceStatus =
  | 'PENDENTE'
  | 'EM_ANDAMENTO'
  | 'ACEITA'
  | 'RECUSADA'
  | 'CANCELADA'
  | 'CONCLUIDA';

export type OutcomeType =
  | 'sucesso_captacao'
  | 'familia_recusou'
  | 'contraindicacao_medica'
  | 'tempo_excedido'
  | 'outro';

export interface Occurrence {
  id: string;
  obito_id: string;
  hospital_id: string;
  hospital?: Hospital;
  status: OccurrenceStatus;
  score_priorizacao: number;
  nome_paciente_mascarado: string;
  dados_completos?: ObitoData;
  notificado_em?: string;
  created_at: string;
  updated_at: string;
}

export interface OccurrenceDetail extends Occurrence {
  dados_completos: ObitoData;
  history: OccurrenceHistoryItem[];
}

export interface OccurrenceHistoryItem {
  id: string;
  occurrence_id: string;
  user_id: string;
  user?: User;
  acao: string;
  status_anterior?: OccurrenceStatus;
  status_novo?: OccurrenceStatus;
  observacoes?: string;
  desfecho?: OutcomeType;
  created_at: string;
}

export interface ObitoData {
  nome_paciente: string;
  data_nascimento: string;
  data_obito: string;
  causa_mortis: string;
  prontuario: string;
  setor: string;
  leito: string;
  identificacao_desconhecida: boolean;
  idade: number;
}

// =============================================================================
// Triagem Rules
// =============================================================================

export interface TriagemRule {
  id: string;
  nome: string;
  descricao: string;
  regras: TriagemRuleConfig;
  ativo: boolean;
  prioridade: number;
  created_at: string;
  updated_at: string;
}

export interface TriagemRuleConfig {
  idade_maxima?: number;
  causas_excludentes?: string[];
  janela_horas?: number;
  identificacao_desconhecida_inelegivel?: boolean;
  setores_prioridade?: Record<string, number>;
}

export interface CreateTriagemRuleInput {
  nome: string;
  descricao?: string;
  regras: TriagemRuleConfig;
  prioridade: number;
}

export interface UpdateTriagemRuleInput {
  nome?: string;
  descricao?: string;
  regras?: TriagemRuleConfig;
  ativo?: boolean;
  prioridade?: number;
}

// =============================================================================
// Metrics
// =============================================================================

export interface DashboardMetrics {
  obitos_elegiveis_hoje: number;
  tempo_medio_notificacao_segundos: number;
  corneas_potenciais: number;
}

// =============================================================================
// API Response Types
// =============================================================================

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export interface ApiError {
  error: string;
  message?: string;
  details?: Record<string, string[]>;
}

// =============================================================================
// Notification Events
// =============================================================================

export interface SSENotificationEvent {
  type: 'new_occurrence' | 'status_update' | 'map_update';
  occurrence_id: string;
  hospital: string;
  hospital_id?: string;
  setor: string;
  tempo_restante_minutos: number;
  timestamp: string;
}

// =============================================================================
// Filter & Sort Options
// =============================================================================

export interface OccurrenceFilters {
  status?: OccurrenceStatus;
  hospital_id?: string;
  date_from?: string;
  date_to?: string;
}

export type SortField = 'created_at' | 'score_priorizacao' | 'tempo_restante';
export type SortOrder = 'asc' | 'desc';

export interface SortOptions {
  field: SortField;
  order: SortOrder;
}

// =============================================================================
// Shifts (Escalas/Plantoes)
// =============================================================================

export type DayOfWeek = 0 | 1 | 2 | 3 | 4 | 5 | 6;

export const DayNames: Record<DayOfWeek, string> = {
  0: 'Domingo',
  1: 'Segunda-feira',
  2: 'Terca-feira',
  3: 'Quarta-feira',
  4: 'Quinta-feira',
  5: 'Sexta-feira',
  6: 'Sabado',
};

export interface Shift {
  id: string;
  hospital_id: string;
  user_id: string;
  day_of_week: DayOfWeek;
  day_name: string;
  start_time: string;
  end_time: string;
  is_night: boolean;
  user?: User;
  created_at: string;
  updated_at: string;
}

export interface TodayShift extends Shift {
  is_active: boolean;
}

export interface CreateShiftInput {
  hospital_id: string;
  user_id: string;
  day_of_week: DayOfWeek;
  start_time: string;
  end_time: string;
}

export interface CoverageGap {
  day_of_week: DayOfWeek;
  day_name: string;
  start_time: string;
  end_time: string;
}

export interface CoverageAnalysis {
  hospital_id: string;
  total_shifts: number;
  gaps: CoverageGap[];
  has_gaps: boolean;
}

// =============================================================================
// Audit Logs
// =============================================================================

export type Severity = 'INFO' | 'WARN' | 'CRITICAL';

export interface AuditLog {
  id: string;
  timestamp: string;
  usuario_id?: string;
  actor_name: string;
  acao: string;
  entidade_tipo: string;
  entidade_id: string;
  hospital_id?: string;
  hospital_name?: string;
  severity: Severity;
  detalhes?: Record<string, unknown>;
  ip_address?: string;
}

export interface AuditLogFilters {
  data_inicio?: string;
  data_fim?: string;
  usuario_id?: string;
  acao?: string;
  entidade_tipo?: string;
  severity?: Severity;
  hospital_id?: string;
  page?: number;
  page_size?: number;
}

// =============================================================================
// Reports
// =============================================================================

export interface ReportFilters {
  date_from?: string;
  date_to?: string;
  hospital_id?: string;
  desfecho?: string[];
}

// =============================================================================
// Health Check
// =============================================================================

export type ServiceStatus = 'healthy' | 'degraded' | 'unhealthy' | 'unknown';

export interface ServiceHealth {
  name: string;
  status: ServiceStatus;
  latency_ms?: number;
  last_check?: string;
  error?: string;
}

export interface SystemHealth {
  status: ServiceStatus;
  services: {
    database: ServiceHealth;
    redis: ServiceHealth;
    listener: ServiceHealth;
    triagem: ServiceHealth;
    sse: ServiceHealth;
  };
  uptime_seconds: number;
  timestamp: string;
}

// =============================================================================
// Map (Dashboard Geografico)
// =============================================================================

/**
 * Nivel de urgencia baseado no tempo restante da janela de isquemia
 * - green: > 4 horas (janela folgada)
 * - yellow: 2-4 horas (atencao)
 * - red: < 2 horas (critico)
 * - none: sem ocorrencias ativas
 */
export type UrgencyLevel = 'green' | 'yellow' | 'red' | 'none';

/**
 * Dados de um operador de plantao para exibicao no mapa
 */
export interface MapOperator {
  id: string;
  nome: string;
}

/**
 * Dados de uma ocorrencia para exibicao no mapa
 */
export interface MapOccurrence {
  id: string;
  nome_mascarado: string;
  setor: string;
  tempo_restante: string;
  tempo_restante_minutos: number;
  status: OccurrenceStatus;
  urgencia: UrgencyLevel;
}

/**
 * Dados de um hospital para renderizacao no mapa
 * Inclui coordenadas geograficas, ocorrencias ativas e operador de plantao
 */
export interface MapHospital {
  id: string;
  nome: string;
  codigo: string;
  latitude: number;
  longitude: number;
  ativo: boolean;
  urgencia_maxima: UrgencyLevel;
  ocorrencias_count: number;
  ocorrencias: MapOccurrence[];
  operador_plantao: MapOperator | null;
}

/**
 * Resposta do endpoint GET /api/v1/map/hospitals
 */
export interface MapDataResponse {
  hospitals: MapHospital[];
  total: number;
}

// =============================================================================
// Geocoding (Nominatim)
// =============================================================================

/**
 * Result from Nominatim geocoding search
 */
export interface NominatimResult {
  place_id: number;
  licence: string;
  osm_type: string;
  osm_id: number;
  lat: string;
  lon: string;
  display_name: string;
  address?: {
    road?: string;
    suburb?: string;
    city?: string;
    state?: string;
    postcode?: string;
    country?: string;
  };
  boundingbox?: string[];
}
