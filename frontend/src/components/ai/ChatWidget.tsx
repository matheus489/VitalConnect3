'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { MessageCircle, X } from 'lucide-react';
import { ChatPanel } from './ChatPanel';
import { useAIChat } from '@/hooks/useAIChat';

interface ChatWidgetProps {
  className?: string;
}

/**
 * Floating action button chat widget with expandable side panel
 * Positioned at bottom-right corner of the screen
 */
export function ChatWidget({ className }: ChatWidgetProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const [isMinimized, setIsMinimized] = React.useState(false);

  const {
    messages,
    isLoading,
    isStreaming,
    thinkingStep,
    currentTool,
    pendingConfirmation,
    unreadCount,
    sendMessage,
    confirmAction,
    clearMessages,
    startNewSession,
    markAsRead,
  } = useAIChat({
    onError: (error) => {
      console.error('[ChatWidget] Error:', error);
    },
  });

  const handleOpen = React.useCallback(() => {
    setIsOpen(true);
    setIsMinimized(false);
    markAsRead();
  }, [markAsRead]);

  const handleClose = React.useCallback(() => {
    setIsOpen(false);
    setIsMinimized(false);
  }, []);

  const handleMinimize = React.useCallback(() => {
    setIsMinimized(true);
    setIsOpen(false);
  }, []);

  const handleToggle = React.useCallback(() => {
    if (isOpen) {
      handleClose();
    } else {
      handleOpen();
    }
  }, [isOpen, handleClose, handleOpen]);

  // Handle escape key to close panel
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        handleClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, handleClose]);

  return (
    <>
      {/* Floating Action Button */}
      <div
        className={cn(
          'fixed bottom-6 right-6 z-50',
          className
        )}
      >
        <Button
          onClick={handleToggle}
          size="icon-lg"
          className={cn(
            'relative size-14 rounded-full shadow-lg transition-all hover:scale-105 hover:shadow-xl',
            isOpen && 'bg-destructive hover:bg-destructive/90'
          )}
          aria-label={isOpen ? 'Fechar assistente' : 'Abrir assistente'}
          aria-expanded={isOpen}
          aria-controls="ai-chat-panel"
        >
          {isOpen ? (
            <X className="size-6" />
          ) : (
            <MessageCircle className="size-6" />
          )}

          {/* Unread badge */}
          {!isOpen && unreadCount > 0 && (
            <Badge
              variant="destructive"
              className="absolute -right-1 -top-1 flex size-5 items-center justify-center rounded-full p-0 text-xs"
            >
              {unreadCount > 9 ? '9+' : unreadCount}
            </Badge>
          )}
        </Button>

        {/* Minimized indicator */}
        {isMinimized && !isOpen && (
          <div className="absolute -top-2 left-1/2 -translate-x-1/2">
            <div className="flex gap-0.5">
              <span className="size-1.5 animate-bounce rounded-full bg-primary" style={{ animationDelay: '0ms' }} />
              <span className="size-1.5 animate-bounce rounded-full bg-primary" style={{ animationDelay: '150ms' }} />
              <span className="size-1.5 animate-bounce rounded-full bg-primary" style={{ animationDelay: '300ms' }} />
            </div>
          </div>
        )}
      </div>

      {/* Chat Panel */}
      <ChatPanel
        isOpen={isOpen}
        messages={messages}
        isLoading={isLoading}
        isStreaming={isStreaming}
        thinkingStep={thinkingStep}
        currentTool={currentTool}
        pendingConfirmation={pendingConfirmation}
        onClose={handleClose}
        onMinimize={handleMinimize}
        onSendMessage={sendMessage}
        onConfirmAction={confirmAction}
        onClearMessages={clearMessages}
        onNewSession={startNewSession}
      />

      {/* Backdrop for mobile */}
      {isOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/20 sm:hidden"
          onClick={handleClose}
          aria-hidden="true"
        />
      )}
    </>
  );
}
