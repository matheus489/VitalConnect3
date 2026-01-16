// VitalConnect Type Definitions

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
  endereco: string;
  config_conexao: Record<string, unknown>;
  ativo: boolean;
  created_at: string;
  updated_at: string;
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
  type: 'new_occurrence' | 'status_update';
  occurrence_id: string;
  hospital: string;
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
