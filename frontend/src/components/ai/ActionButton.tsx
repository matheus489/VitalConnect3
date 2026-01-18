'use client';

import { Loader2, ExternalLink } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button, buttonVariants } from '@/components/ui/button';
import type { ActionButtonData } from '@/types/ai';

interface ActionButtonProps {
  data: ActionButtonData;
  onClick?: (action: string, params?: Record<string, unknown>) => void;
  isLoading?: boolean;
  className?: string;
}

interface ActionButtonGroupProps {
  buttons: ActionButtonData[];
  onClick?: (action: string, params?: Record<string, unknown>) => void;
  loadingActionId?: string;
  className?: string;
}

const variantMapping: Record<
  ActionButtonData['variant'],
  'default' | 'secondary' | 'link' | 'destructive' | 'outline'
> = {
  primary: 'default',
  secondary: 'outline',
  link: 'link',
  destructive: 'destructive',
};

export function ActionButton({
  data,
  onClick,
  isLoading = false,
  className,
}: ActionButtonProps) {
  const buttonVariant = variantMapping[data.variant] || 'default';
  const isDisabled = data.disabled || isLoading;

  const handleClick = () => {
    if (!isDisabled) {
      onClick?.(data.action, data.action_params);
    }
  };

  return (
    <Button
      data-slot="action-button"
      variant={buttonVariant}
      size="sm"
      onClick={handleClick}
      disabled={isDisabled}
      className={cn(
        data.variant === 'link' && 'px-0 h-auto',
        className
      )}
    >
      {isLoading ? (
        <>
          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
          Processando...
        </>
      ) : (
        <>
          {data.label}
          {data.variant === 'link' && (
            <ExternalLink className="h-3 w-3 ml-1" />
          )}
        </>
      )}
    </Button>
  );
}

export function ActionButtonGroup({
  buttons,
  onClick,
  loadingActionId,
  className,
}: ActionButtonGroupProps) {
  if (buttons.length === 0) {
    return null;
  }

  return (
    <div
      data-slot="action-button-group"
      className={cn('flex flex-wrap items-center gap-2 mt-2', className)}
    >
      {buttons.map((button) => (
        <ActionButton
          key={button.id}
          data={button}
          onClick={onClick}
          isLoading={loadingActionId === button.id}
        />
      ))}
    </div>
  );
}

export default ActionButton;
