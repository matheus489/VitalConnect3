'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Loader2, Search, Wrench, Brain, Database } from 'lucide-react';

interface ThinkingIndicatorProps {
  step?: string | null;
  toolName?: string | null;
  className?: string;
}

/**
 * Animated indicator showing AI processing state
 * Displays current step and tool being executed
 */
export function ThinkingIndicator({
  step,
  toolName,
  className,
}: ThinkingIndicatorProps) {
  const getToolIcon = () => {
    if (!toolName) {
      return <Brain className="size-4" />;
    }

    const lowerToolName = toolName.toLowerCase();

    if (lowerToolName.includes('search') || lowerToolName.includes('documentation')) {
      return <Search className="size-4" />;
    }
    if (lowerToolName.includes('occurrence') || lowerToolName.includes('list')) {
      return <Database className="size-4" />;
    }
    return <Wrench className="size-4" />;
  };

  const getDisplayText = () => {
    if (step) return step;
    if (toolName) return `Executando: ${toolName}`;
    return 'Pensando...';
  };

  return (
    <div
      className={cn(
        'flex items-center gap-3 rounded-lg bg-muted/50 px-4 py-3',
        className
      )}
      role="status"
      aria-live="polite"
      aria-label="Assistente processando"
    >
      <div className="relative flex items-center justify-center">
        <Loader2 className="size-5 animate-spin text-primary" />
        <span className="absolute inset-0 flex items-center justify-center text-primary">
          {getToolIcon()}
        </span>
      </div>

      <div className="flex flex-col gap-0.5">
        <span className="text-sm font-medium text-foreground">
          {getDisplayText()}
        </span>
        {toolName && step && (
          <span className="text-xs text-muted-foreground">
            Ferramenta: {toolName}
          </span>
        )}
      </div>

      <div className="ml-auto flex gap-1">
        <span
          className="size-1.5 animate-bounce rounded-full bg-primary"
          style={{ animationDelay: '0ms' }}
        />
        <span
          className="size-1.5 animate-bounce rounded-full bg-primary"
          style={{ animationDelay: '150ms' }}
        />
        <span
          className="size-1.5 animate-bounce rounded-full bg-primary"
          style={{ animationDelay: '300ms' }}
        />
      </div>
    </div>
  );
}
