import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { ApiError } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// Create axios instance
export const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Token management
const TOKEN_KEY = 'sidot_access_token';
const REFRESH_TOKEN_KEY = 'sidot_refresh_token';

export const getAccessToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
};

export const getRefreshToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(REFRESH_TOKEN_KEY);
};

export const setTokens = (accessToken: string, refreshToken: string): void => {
  if (typeof window === 'undefined') return;
  localStorage.setItem(TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
};

export const clearTokens = (): void => {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
};

// Request interceptor - add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - handle token refresh
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value: unknown) => void;
  reject: (error: unknown) => void;
}> = [];

const processQueue = (error: Error | null, token: string | null = null) => {
  failedQueue.forEach((promise) => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve(token);
    }
  });
  failedQueue = [];
};

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // If error is 401 and we haven't retried yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return api(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = getRefreshToken();
      if (!refreshToken) {
        clearTokens();
        window.location.href = '/login';
        return Promise.reject(error);
      }

      try {
        const response = await axios.post(`${API_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        const { access_token, refresh_token } = response.data;
        setTokens(access_token, refresh_token);
        processQueue(null, access_token);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
        }
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError as Error, null);
        clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

// =============================================================================
// AI Assistant API
// =============================================================================

export interface AIChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  tool_calls?: AIToolCall[];
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AIToolCall {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  result?: unknown;
}

export interface AIChatRequest {
  message: string;
  session_id?: string;
}

export interface AIChatResponse {
  response: string;
  session_id: string;
  tool_calls?: AIToolCall[];
  confirmation_required?: AIConfirmationRequest;
}

export interface AIConfirmationRequest {
  action_id: string;
  action_type: string;
  description: string;
  details: Record<string, unknown>;
}

export interface AIConfirmRequest {
  confirmed: boolean;
}

export interface AIConfirmResponse {
  success: boolean;
  result?: unknown;
  message?: string;
}

export interface AIConversation {
  session_id: string;
  title?: string;
  last_message?: string;
  created_at: string;
  updated_at: string;
}

export interface AIConversationHistory {
  session_id: string;
  messages: AIChatMessage[];
}

/**
 * AI API client for chat operations
 */
export const aiApi = {
  /**
   * Send a message to the AI assistant
   */
  sendMessage: async (request: AIChatRequest): Promise<AIChatResponse> => {
    const response = await api.post<AIChatResponse>('/ai/chat', request);
    return response.data;
  },

  /**
   * Confirm or cancel a pending action
   */
  confirmAction: async (actionId: string, request: AIConfirmRequest): Promise<AIConfirmResponse> => {
    const response = await api.post<AIConfirmResponse>(`/ai/confirm/${actionId}`, request);
    return response.data;
  },

  /**
   * Get conversation history for a session
   */
  getHistory: async (sessionId: string): Promise<AIConversationHistory> => {
    const response = await api.get<AIConversationHistory>(`/ai/conversations/${sessionId}`);
    return response.data;
  },

  /**
   * List all user conversations
   */
  listConversations: async (): Promise<AIConversation[]> => {
    const response = await api.get<AIConversation[]>('/ai/conversations');
    return response.data;
  },

  /**
   * Delete a conversation
   */
  deleteConversation: async (sessionId: string): Promise<void> => {
    await api.delete(`/ai/conversations/${sessionId}`);
  },

  /**
   * Create SSE connection for streaming AI responses
   */
  createStreamConnection: (sessionId: string, onEvent: (event: AIStreamEvent) => void): () => void => {
    const token = getAccessToken();
    if (!token) {
      console.error('[AI] No access token available for streaming');
      return () => {};
    }

    const url = `${API_URL}/ai/chat/stream?session_id=${encodeURIComponent(sessionId)}&token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as AIStreamEvent;
        onEvent(data);
      } catch (error) {
        console.error('[AI] Failed to parse SSE event:', error);
      }
    };

    eventSource.onerror = () => {
      console.error('[AI] SSE connection error');
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  },
};

/**
 * Types of events received from SSE streaming
 */
export type AIStreamEventType = 'thinking' | 'tool_call' | 'response_chunk' | 'done' | 'error';

export interface AIStreamEvent {
  type: AIStreamEventType;
  data: {
    step?: string;
    tool_name?: string;
    chunk?: string;
    full_response?: string;
    session_id?: string;
    confirmation_required?: AIConfirmationRequest;
    error?: string;
  };
}

export default api;
