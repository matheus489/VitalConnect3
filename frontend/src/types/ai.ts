// SIDOT AI Assistant Type Definitions

// =============================================================================
// AI Chat Types
// =============================================================================

export type AIMessageRole = 'user' | 'assistant' | 'system';

export interface AIMessage {
  id: string;
  role: AIMessageRole;
  content: string;
  tool_calls?: AIToolCall[];
  metadata?: AIMessageMetadata;
  timestamp: string;
}

export interface AIMessageMetadata {
  thinking?: boolean;
  current_step?: string;
  confirmation_required?: AIConfirmationRequest;
}

export interface AIToolCall {
  id: string;
  tool_name: string;
  input_params: Record<string, unknown>;
  output_result?: unknown;
  status: 'pending' | 'executing' | 'success' | 'failed';
}

// =============================================================================
// Confirmation Flow Types
// =============================================================================

export interface AIConfirmationRequest {
  action_id: string;
  action_type: string;
  tool_name: string;
  description: string;
  details: Record<string, unknown>;
  severity: 'INFO' | 'WARN' | 'CRITICAL';
}

export interface AIConfirmationResponse {
  action_id: string;
  confirmed: boolean;
  result?: unknown;
  error?: string;
}

// =============================================================================
// Generative UI Types
// =============================================================================

export type GenerativeUIComponentType =
  | 'occurrence_card'
  | 'occurrence_table'
  | 'confirmation_dialog'
  | 'action_buttons'
  | 'text';

export interface GenerativeUIComponent {
  type: GenerativeUIComponentType;
  data: unknown;
}

// Occurrence Card Data
export interface OccurrenceCardData {
  id: string;
  hospital_nome: string;
  hospital_id: string;
  status: string;
  nome_paciente_mascarado: string;
  tempo_restante: string;
  tempo_restante_minutos: number;
  setor?: string;
  urgencia?: 'green' | 'yellow' | 'red';
}

// Occurrence Table Data
export interface OccurrenceTableData {
  occurrences: OccurrenceCardData[];
  total: number;
  sortable?: boolean;
}

// Action Button Data
export interface ActionButtonData {
  id: string;
  label: string;
  variant: 'primary' | 'secondary' | 'link' | 'destructive';
  action: string;
  action_params?: Record<string, unknown>;
  disabled?: boolean;
}

// =============================================================================
// Parsed Message Content
// =============================================================================

export interface ParsedMessageContent {
  segments: MessageContentSegment[];
}

export type MessageContentSegment =
  | { type: 'text'; content: string }
  | { type: 'occurrence_card'; data: OccurrenceCardData }
  | { type: 'occurrence_table'; data: OccurrenceTableData }
  | { type: 'confirmation_dialog'; data: AIConfirmationRequest }
  | { type: 'action_buttons'; data: ActionButtonData[] };

// =============================================================================
// Chat API Types
// =============================================================================

export interface AIChatRequest {
  message: string;
  session_id?: string;
}

export interface AIChatResponse {
  response: string;
  session_id: string;
  tool_calls?: AIToolCall[];
  confirmation_required?: AIConfirmationRequest;
  generative_ui?: GenerativeUIComponent[];
}

export interface AIConversation {
  session_id: string;
  messages: AIMessage[];
  created_at: string;
  updated_at: string;
}

// =============================================================================
// SSE Event Types for AI
// =============================================================================

export type AISSEEventType =
  | 'thinking'
  | 'tool_call'
  | 'response_chunk'
  | 'done'
  | 'error';

export interface AISSEEvent {
  type: AISSEEventType;
  data: unknown;
}

export interface AIThinkingEvent {
  type: 'thinking';
  data: {
    step: string;
    tool_name?: string;
  };
}

export interface AIResponseChunkEvent {
  type: 'response_chunk';
  data: {
    content: string;
    index: number;
  };
}

export interface AIDoneEvent {
  type: 'done';
  data: {
    session_id: string;
    message_id: string;
    confirmation_required?: AIConfirmationRequest;
  };
}
