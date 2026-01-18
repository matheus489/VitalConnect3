'use client';

import { User, Bot } from 'lucide-react';
import { cn } from '@/lib/utils';
import { MessageContent } from './MessageContent';
import type { AIChatMessage } from '@/lib/api';
import type { OccurrenceCardData, AIConfirmationRequest } from '@/types/ai';

interface MessageBubbleProps {
  message: AIChatMessage;
  onOccurrenceClick?: (occurrence: OccurrenceCardData) => void;
  onActionClick?: (action: string, params?: Record<string, unknown>) => void;
  onConfirmationRequired?: (request: AIConfirmationRequest) => void;
  loadingActionId?: string;
  className?: string;
}

function formatTimestamp(isoString: string): string {
  try {
    const date = new Date(isoString);
    return date.toLocaleTimeString('pt-BR', {
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return '';
  }
}

/**
 * Chat message bubble component
 * Displays user and assistant messages with appropriate styling
 * Supports generative UI components in assistant messages
 */
export function MessageBubble({
  message,
  onOccurrenceClick,
  onActionClick,
  onConfirmationRequired,
  loadingActionId,
  className,
}: MessageBubbleProps) {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';

  // System messages are displayed differently
  if (isSystem) {
    return (
      <div
        data-slot="message-bubble"
        className={cn(
          'mx-auto max-w-[90%] rounded-lg bg-muted/50 px-4 py-2 text-center text-sm text-muted-foreground',
          className
        )}
      >
        {message.content}
      </div>
    );
  }

  return (
    <div
      data-slot="message-bubble"
      className={cn(
        'flex gap-3',
        isUser ? 'flex-row-reverse' : 'flex-row',
        className
      )}
    >
      {/* Avatar */}
      <div
        className={cn(
          'flex size-8 shrink-0 items-center justify-center rounded-full',
          isUser
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-muted-foreground'
        )}
      >
        {isUser ? <User className="size-4" /> : <Bot className="size-4" />}
      </div>

      {/* Message content */}
      <div
        className={cn(
          'flex max-w-[85%] flex-col gap-1',
          isUser ? 'items-end' : 'items-start'
        )}
      >
        <div
          className={cn(
            'rounded-2xl px-4 py-2',
            isUser
              ? 'rounded-tr-sm bg-primary text-primary-foreground'
              : 'rounded-tl-sm bg-muted text-foreground'
          )}
        >
          {isUser ? (
            // User messages are plain text
            <p className="whitespace-pre-wrap break-words text-sm">
              {message.content}
            </p>
          ) : (
            // Assistant messages support generative UI components
            <MessageContent
              content={message.content}
              onOccurrenceClick={onOccurrenceClick}
              onActionClick={onActionClick}
              onConfirmationRequired={onConfirmationRequired}
              loadingActionId={loadingActionId}
              className="text-sm"
            />
          )}
        </div>

        {/* Timestamp */}
        <span className="px-1 text-xs text-muted-foreground">
          {formatTimestamp(message.created_at)}
        </span>
      </div>
    </div>
  );
}

export default MessageBubble;
