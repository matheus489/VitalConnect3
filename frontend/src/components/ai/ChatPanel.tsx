'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  Minimize2,
  X,
  Send,
  Trash2,
  RotateCcw,
} from 'lucide-react';
import { MessageBubble } from './MessageBubble';
import { ThinkingIndicator } from './ThinkingIndicator';
import { ConfirmationDialog } from './ConfirmationDialog';
import type { AIChatMessage, AIConfirmationRequest } from '@/lib/api';
import type { OccurrenceCardData } from '@/types/ai';

interface ChatPanelProps {
  isOpen: boolean;
  messages: AIChatMessage[];
  isLoading: boolean;
  isStreaming: boolean;
  thinkingStep: string | null;
  currentTool: string | null;
  pendingConfirmation: AIConfirmationRequest | null;
  onClose: () => void;
  onMinimize: () => void;
  onSendMessage: (message: string) => void;
  onConfirmAction: (confirmed: boolean) => void;
  onClearMessages: () => void;
  onNewSession: () => void;
  onOccurrenceClick?: (occurrence: OccurrenceCardData) => void;
  onActionClick?: (action: string, params?: Record<string, unknown>) => void;
  className?: string;
}

/**
 * Side panel component for AI chat
 * Slides from right, 400px width, does not block main content
 * Supports generative UI components in messages
 */
export function ChatPanel({
  isOpen,
  messages,
  isLoading,
  isStreaming,
  thinkingStep,
  currentTool,
  pendingConfirmation,
  onClose,
  onMinimize,
  onSendMessage,
  onConfirmAction,
  onClearMessages,
  onNewSession,
  onOccurrenceClick,
  onActionClick,
  className,
}: ChatPanelProps) {
  const [inputValue, setInputValue] = React.useState('');
  const [confirmDialogOpen, setConfirmDialogOpen] = React.useState(false);
  const [isConfirming, setIsConfirming] = React.useState(false);
  const [loadingActionId, setLoadingActionId] = React.useState<string | undefined>();
  const messagesEndRef = React.useRef<HTMLDivElement>(null);
  const inputRef = React.useRef<HTMLTextAreaElement>(null);

  // Auto-scroll to bottom when new messages arrive
  React.useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, isStreaming, thinkingStep]);

  // Focus input when panel opens
  React.useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  // Open confirmation dialog when pendingConfirmation changes
  React.useEffect(() => {
    if (pendingConfirmation) {
      setConfirmDialogOpen(true);
    }
  }, [pendingConfirmation]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim() && !isLoading) {
      onSendMessage(inputValue.trim());
      setInputValue('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  const handleConfirmAction = async () => {
    if (!pendingConfirmation) return;
    setIsConfirming(true);
    try {
      onConfirmAction(true);
      setConfirmDialogOpen(false);
    } finally {
      setIsConfirming(false);
    }
  };

  const handleCancelAction = () => {
    onConfirmAction(false);
    setConfirmDialogOpen(false);
  };

  const handleOccurrenceClick = (occurrence: OccurrenceCardData) => {
    onOccurrenceClick?.(occurrence);
    // Default behavior: navigate to occurrence details
    if (!onOccurrenceClick) {
      window.location.href = `/ocorrencias/${occurrence.id}`;
    }
  };

  const handleActionClick = (action: string, params?: Record<string, unknown>) => {
    if (onActionClick) {
      setLoadingActionId(action);
      onActionClick(action, params);
      // Reset loading state after a timeout (in real implementation, this would be controlled by the action completion)
      setTimeout(() => setLoadingActionId(undefined), 2000);
    }
  };

  const handleInlineConfirmationRequired = (request: AIConfirmationRequest) => {
    // If we receive a confirmation request from message content parsing,
    // we can either show the dialog or let the parent handle it
    setConfirmDialogOpen(true);
  };

  return (
    <>
      <div
        id="ai-chat-panel"
        className={cn(
          'fixed inset-y-0 right-0 z-40 flex w-full flex-col border-l bg-background shadow-xl transition-transform duration-300 ease-in-out sm:w-[400px]',
          isOpen ? 'translate-x-0' : 'translate-x-full',
          className
        )}
        role="dialog"
        aria-label="Assistente IA"
        aria-hidden={!isOpen}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div className="flex items-center gap-2">
            <div className="size-2 rounded-full bg-green-500" />
            <h2 className="font-semibold">Assistente VitalConnect</h2>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={onNewSession}
              title="Nova conversa"
              disabled={isLoading}
            >
              <RotateCcw className="size-4" />
              <span className="sr-only">Nova conversa</span>
            </Button>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={onClearMessages}
              title="Limpar mensagens"
              disabled={isLoading || messages.length === 0}
            >
              <Trash2 className="size-4" />
              <span className="sr-only">Limpar mensagens</span>
            </Button>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={onMinimize}
              title="Minimizar"
            >
              <Minimize2 className="size-4" />
              <span className="sr-only">Minimizar</span>
            </Button>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={onClose}
              title="Fechar"
            >
              <X className="size-4" />
              <span className="sr-only">Fechar</span>
            </Button>
          </div>
        </div>

        {/* Messages area */}
        <div className="flex-1 overflow-y-auto px-4 py-4">
          {messages.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center gap-4 text-center text-muted-foreground">
              <div className="rounded-full bg-muted p-4">
                <svg
                  className="size-8"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                  />
                </svg>
              </div>
              <div>
                <p className="font-medium">Como posso ajudar?</p>
                <p className="mt-1 text-sm">
                  Pergunte sobre ocorrencias, documentacao ou solicite acoes.
                </p>
              </div>
              <div className="mt-4 flex flex-wrap justify-center gap-2">
                <SuggestionChip
                  onClick={() => onSendMessage('Quais ocorrencias estao pendentes?')}
                >
                  Ocorrencias pendentes
                </SuggestionChip>
                <SuggestionChip
                  onClick={() => onSendMessage('Como funciona o processo de captacao?')}
                >
                  Processo de captacao
                </SuggestionChip>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-4">
              {messages.map((message) => (
                <MessageBubble
                  key={message.id}
                  message={message}
                  onOccurrenceClick={handleOccurrenceClick}
                  onActionClick={handleActionClick}
                  onConfirmationRequired={handleInlineConfirmationRequired}
                  loadingActionId={loadingActionId}
                />
              ))}

              {/* Thinking indicator */}
              {(isLoading || isStreaming) && (
                <ThinkingIndicator step={thinkingStep} toolName={currentTool} />
              )}

              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Input area */}
        <form onSubmit={handleSubmit} className="border-t p-4">
          <div className="flex gap-2">
            <Textarea
              ref={inputRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Digite sua mensagem..."
              className="min-h-[44px] max-h-[120px] resize-none"
              disabled={isLoading}
              rows={1}
            />
            <Button
              type="submit"
              size="icon"
              disabled={!inputValue.trim() || isLoading}
            >
              <Send className="size-4" />
              <span className="sr-only">Enviar</span>
            </Button>
          </div>
          <p className="mt-2 text-center text-xs text-muted-foreground">
            Pressione Enter para enviar, Shift+Enter para nova linha
          </p>
        </form>
      </div>

      {/* Confirmation Dialog - rendered outside the panel for proper z-index */}
      {pendingConfirmation && (
        <ConfirmationDialog
          open={confirmDialogOpen}
          data={{
            action_id: pendingConfirmation.action_id,
            action_type: pendingConfirmation.action_type,
            tool_name: pendingConfirmation.action_type,
            description: pendingConfirmation.description,
            details: pendingConfirmation.details,
            severity: 'WARN',
          }}
          onConfirm={handleConfirmAction}
          onCancel={handleCancelAction}
          isLoading={isConfirming}
        />
      )}
    </>
  );
}

interface SuggestionChipProps {
  children: React.ReactNode;
  onClick: () => void;
}

function SuggestionChip({ children, onClick }: SuggestionChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="rounded-full border bg-background px-3 py-1.5 text-xs font-medium transition-colors hover:bg-muted"
    >
      {children}
    </button>
  );
}

export default ChatPanel;
