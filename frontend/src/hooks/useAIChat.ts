'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import {
  aiApi,
  AIChatMessage,
  AIStreamEvent,
  AIConfirmationRequest,
  getAccessToken,
} from '@/lib/api';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

/**
 * Generate a unique ID using crypto.randomUUID or fallback
 */
function generateId(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  // Fallback for older environments
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export interface UseAIChatOptions {
  onConfirmationRequired?: (confirmation: AIConfirmationRequest) => void;
  onError?: (error: string) => void;
}

export interface UseAIChatReturn {
  messages: AIChatMessage[];
  isLoading: boolean;
  isStreaming: boolean;
  thinkingStep: string | null;
  currentTool: string | null;
  sessionId: string | null;
  pendingConfirmation: AIConfirmationRequest | null;
  unreadCount: number;
  sendMessage: (content: string) => Promise<void>;
  confirmAction: (confirmed: boolean) => Promise<void>;
  clearMessages: () => void;
  loadHistory: (sessionId: string) => Promise<void>;
  startNewSession: () => void;
  markAsRead: () => void;
}

const SESSION_KEY = 'sidot_ai_session_id';

/**
 * Hook to manage AI chat state and communication
 * Handles SSE streaming, message history, and confirmation flows
 */
export function useAIChat(options: UseAIChatOptions = {}): UseAIChatReturn {
  const { onConfirmationRequired, onError } = options;

  const [messages, setMessages] = useState<AIChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isStreaming, setIsStreaming] = useState(false);
  const [thinkingStep, setThinkingStep] = useState<string | null>(null);
  const [currentTool, setCurrentTool] = useState<string | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [pendingConfirmation, setPendingConfirmation] = useState<AIConfirmationRequest | null>(null);
  const [unreadCount, setUnreadCount] = useState(0);

  const eventSourceRef = useRef<EventSource | null>(null);
  const streamingMessageRef = useRef<string>('');

  // Load session ID from localStorage on mount
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const storedSessionId = localStorage.getItem(SESSION_KEY);
      if (storedSessionId) {
        setSessionId(storedSessionId);
      }
    }
  }, []);

  // Save session ID to localStorage when it changes
  useEffect(() => {
    if (typeof window !== 'undefined' && sessionId) {
      localStorage.setItem(SESSION_KEY, sessionId);
    }
  }, [sessionId]);

  // Cleanup SSE connection on unmount
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, []);

  /**
   * Handle SSE streaming events
   */
  const handleStreamEvent = useCallback((event: AIStreamEvent) => {
    switch (event.type) {
      case 'thinking':
        setThinkingStep(event.data.step || 'Processando...');
        break;

      case 'tool_call':
        setCurrentTool(event.data.tool_name || null);
        setThinkingStep(`Executando: ${event.data.tool_name}`);
        break;

      case 'response_chunk':
        if (event.data.chunk) {
          streamingMessageRef.current += event.data.chunk;
          setMessages((prev) => {
            const lastMessage = prev[prev.length - 1];
            if (lastMessage?.role === 'assistant' && !lastMessage.id.startsWith('msg_')) {
              // Update streaming message
              return [
                ...prev.slice(0, -1),
                { ...lastMessage, content: streamingMessageRef.current },
              ];
            }
            return prev;
          });
        }
        break;

      case 'done':
        setIsStreaming(false);
        setThinkingStep(null);
        setCurrentTool(null);

        if (event.data.full_response) {
          setMessages((prev) => {
            const lastMessage = prev[prev.length - 1];
            if (lastMessage?.role === 'assistant') {
              return [
                ...prev.slice(0, -1),
                {
                  ...lastMessage,
                  id: `msg_${generateId()}`,
                  content: event.data.full_response || '',
                },
              ];
            }
            return prev;
          });
        }

        if (event.data.session_id) {
          setSessionId(event.data.session_id);
        }

        if (event.data.confirmation_required) {
          setPendingConfirmation(event.data.confirmation_required);
          onConfirmationRequired?.(event.data.confirmation_required);
        }

        // Increment unread count for new assistant messages
        setUnreadCount((prev) => prev + 1);

        streamingMessageRef.current = '';
        break;

      case 'error':
        setIsStreaming(false);
        setIsLoading(false);
        setThinkingStep(null);
        setCurrentTool(null);
        onError?.(event.data.error || 'Erro desconhecido');
        streamingMessageRef.current = '';
        break;
    }
  }, [onConfirmationRequired, onError]);

  /**
   * Connect to SSE stream for receiving responses
   */
  const connectToStream = useCallback((currentSessionId: string) => {
    const token = getAccessToken();
    if (!token) {
      console.error('[AI Chat] No access token available');
      return;
    }

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const url = `${API_URL}/ai/chat/stream?session_id=${encodeURIComponent(currentSessionId)}&token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as AIStreamEvent;
        handleStreamEvent(data);
      } catch (error) {
        console.error('[AI Chat] Failed to parse SSE event:', error);
      }
    };

    eventSource.onerror = () => {
      console.error('[AI Chat] SSE connection error');
      setIsStreaming(false);
    };
  }, [handleStreamEvent]);

  /**
   * Send a message to the AI assistant
   */
  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim() || isLoading) return;

    const currentSessionId = sessionId || generateId();
    if (!sessionId) {
      setSessionId(currentSessionId);
    }

    // Add user message to state
    const userMessage: AIChatMessage = {
      id: `msg_${generateId()}`,
      role: 'user',
      content: content.trim(),
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, userMessage]);

    // Add placeholder for assistant response
    const assistantPlaceholder: AIChatMessage = {
      id: `streaming_${generateId()}`,
      role: 'assistant',
      content: '',
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, assistantPlaceholder]);

    setIsLoading(true);
    setThinkingStep('Analisando mensagem...');

    try {
      // Send the message (non-streaming for now)
      const response = await aiApi.sendMessage({
        message: content.trim(),
        session_id: currentSessionId,
      });

      // Update with full response
      setMessages((prev) => {
        const lastMessage = prev[prev.length - 1];
        if (lastMessage?.role === 'assistant') {
          return [
            ...prev.slice(0, -1),
            {
              ...lastMessage,
              id: `msg_${generateId()}`,
              content: response.response,
              tool_calls: response.tool_calls,
            },
          ];
        }
        return prev;
      });

      if (response.confirmation_required) {
        setPendingConfirmation(response.confirmation_required);
        onConfirmationRequired?.(response.confirmation_required);
      }

      setUnreadCount((prev) => prev + 1);

      if (response.session_id) {
        setSessionId(response.session_id);
      }
    } catch (error) {
      console.error('[AI Chat] Failed to send message:', error);

      // Remove placeholder message on error
      setMessages((prev) => prev.filter((m) => m.id !== assistantPlaceholder.id));

      // Add error message
      const errorMessage: AIChatMessage = {
        id: `msg_${generateId()}`,
        role: 'assistant',
        content: 'Desculpe, ocorreu um erro ao processar sua mensagem. Por favor, tente novamente.',
        created_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, errorMessage]);

      onError?.('Falha ao enviar mensagem');
    } finally {
      setIsLoading(false);
      setIsStreaming(false);
      setThinkingStep(null);
      setCurrentTool(null);
    }
  }, [sessionId, isLoading, onConfirmationRequired, onError]);

  /**
   * Confirm or cancel a pending action
   */
  const confirmAction = useCallback(async (confirmed: boolean) => {
    if (!pendingConfirmation) return;

    setIsLoading(true);

    try {
      const response = await aiApi.confirmAction(pendingConfirmation.action_id, {
        confirmed,
      });

      // Add confirmation result as assistant message
      const resultMessage: AIChatMessage = {
        id: `msg_${generateId()}`,
        role: 'assistant',
        content: confirmed
          ? response.message || 'Acao executada com sucesso.'
          : 'Acao cancelada pelo usuario.',
        created_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, resultMessage]);
      setUnreadCount((prev) => prev + 1);

      setPendingConfirmation(null);
    } catch (error) {
      console.error('[AI Chat] Failed to confirm action:', error);
      onError?.('Falha ao confirmar acao');
    } finally {
      setIsLoading(false);
    }
  }, [pendingConfirmation, onError]);

  /**
   * Clear all messages and start fresh
   */
  const clearMessages = useCallback(() => {
    setMessages([]);
    setUnreadCount(0);
    streamingMessageRef.current = '';
  }, []);

  /**
   * Load conversation history for a session
   */
  const loadHistory = useCallback(async (historySessionId: string) => {
    setIsLoading(true);

    try {
      const history = await aiApi.getHistory(historySessionId);
      setMessages(history.messages);
      setSessionId(historySessionId);
    } catch (error) {
      console.error('[AI Chat] Failed to load history:', error);
      onError?.('Falha ao carregar historico');
    } finally {
      setIsLoading(false);
    }
  }, [onError]);

  /**
   * Start a new conversation session
   */
  const startNewSession = useCallback(() => {
    // Close existing SSE connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    // Clear state
    setMessages([]);
    setSessionId(null);
    setPendingConfirmation(null);
    setUnreadCount(0);
    streamingMessageRef.current = '';

    // Remove from localStorage
    if (typeof window !== 'undefined') {
      localStorage.removeItem(SESSION_KEY);
    }
  }, []);

  /**
   * Mark all messages as read
   */
  const markAsRead = useCallback(() => {
    setUnreadCount(0);
  }, []);

  return {
    messages,
    isLoading,
    isStreaming,
    thinkingStep,
    currentTool,
    sessionId,
    pendingConfirmation,
    unreadCount,
    sendMessage,
    confirmAction,
    clearMessages,
    loadHistory,
    startNewSession,
    markAsRead,
  };
}
