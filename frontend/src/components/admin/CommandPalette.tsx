'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { Command } from 'cmdk';
import {
  Palette,
  Layout,
  LayoutDashboard,
  Upload,
  Type,
  Search,
  History,
  ChevronRight,
  X,
  AlertCircle,
  CheckCircle,
  Play,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  parseCommand,
  getCommandSuggestions,
  formatCommandResult,
  COMMAND_SUGGESTIONS,
} from './CommandParser';
import type { CommandAction, CommandSuggestion, CommandHistoryEntry } from '@/types/theme';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  onCommandExecute: (action: CommandAction) => void;
  onUploadLogo?: () => void;
  onUploadFavicon?: () => void;
  commandHistory?: CommandHistoryEntry[];
}

const categoryIcons: Record<string, React.ElementType> = {
  colors: Palette,
  sidebar: Layout,
  dashboard: LayoutDashboard,
  assets: Upload,
  fonts: Type,
};

const categoryLabels: Record<string, string> = {
  colors: 'Cores',
  sidebar: 'Menu Lateral',
  dashboard: 'Dashboard',
  assets: 'Assets',
  fonts: 'Fontes',
};

export function CommandPalette({
  isOpen,
  onClose,
  onCommandExecute,
  onUploadLogo,
  onUploadFavicon,
  commandHistory = [],
}: CommandPaletteProps) {
  const [inputValue, setInputValue] = useState('');
  const [suggestions, setSuggestions] = useState<CommandSuggestion[]>(COMMAND_SUGGESTIONS);
  const [lastResult, setLastResult] = useState<{ message: string; success: boolean } | null>(null);
  const [showHistory, setShowHistory] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  // Check if current input is a valid command
  const currentAction = inputValue.trim() ? parseCommand(inputValue) : null;
  const isValidCommand = currentAction?.isValid ?? false;

  // Update suggestions based on input
  useEffect(() => {
    const filtered = getCommandSuggestions(inputValue);
    setSuggestions(filtered);
  }, [inputValue]);

  // Focus input when opened
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
    if (isOpen) {
      setLastResult(null);
      setShowHistory(false);
    }
  }, [isOpen]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        if (!isOpen) {
          // Parent component should handle opening
        }
      }
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  const handleCommandSubmit = useCallback(() => {
    if (!inputValue.trim()) return;

    const action = parseCommand(inputValue);

    // Handle upload commands specially
    if (action.type === 'UPLOAD_LOGO' && action.isValid) {
      onUploadLogo?.();
      setLastResult({ message: 'Selecione um arquivo para o logo', success: true });
      setInputValue('');
      return;
    }

    if (action.type === 'UPLOAD_FAVICON' && action.isValid) {
      onUploadFavicon?.();
      setLastResult({ message: 'Selecione um arquivo para o favicon', success: true });
      setInputValue('');
      return;
    }

    const resultMessage = formatCommandResult(action);
    setLastResult({ message: resultMessage, success: action.isValid });

    if (action.isValid) {
      onCommandExecute(action);
    }

    setInputValue('');
  }, [inputValue, onCommandExecute, onUploadLogo, onUploadFavicon]);

  const handleSuggestionSelect = useCallback((suggestion: CommandSuggestion) => {
    // If suggestion has an example, use it to fill in, otherwise use the command template
    if (suggestion.example) {
      setInputValue(suggestion.example);
    } else {
      setInputValue(suggestion.command);
    }
    inputRef.current?.focus();
  }, []);

  // Execute a suggestion directly (for examples that are complete commands)
  const handleSuggestionExecute = useCallback((suggestion: CommandSuggestion, e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent triggering the parent select

    const commandToExecute = suggestion.example || suggestion.command;
    const action = parseCommand(commandToExecute);

    if (action.isValid) {
      // Handle upload commands specially
      if (action.type === 'UPLOAD_LOGO') {
        onUploadLogo?.();
        setLastResult({ message: 'Selecione um arquivo para o logo', success: true });
        return;
      }
      if (action.type === 'UPLOAD_FAVICON') {
        onUploadFavicon?.();
        setLastResult({ message: 'Selecione um arquivo para o favicon', success: true });
        return;
      }

      onCommandExecute(action);
      const resultMessage = formatCommandResult(action);
      setLastResult({ message: resultMessage, success: true });
    } else {
      // Fill the input with the suggestion for editing
      setInputValue(commandToExecute);
      inputRef.current?.focus();
    }
  }, [onCommandExecute, onUploadLogo, onUploadFavicon]);

  const handleHistorySelect = useCallback((entry: CommandHistoryEntry) => {
    setInputValue(entry.command);
    setShowHistory(false);
    inputRef.current?.focus();
  }, []);

  // Group suggestions by category
  const groupedSuggestions = suggestions.reduce(
    (acc, suggestion) => {
      if (!acc[suggestion.category]) {
        acc[suggestion.category] = [];
      }
      acc[suggestion.category].push(suggestion);
      return acc;
    },
    {} as Record<string, CommandSuggestion[]>
  );

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Command Palette */}
      <Command
        className="relative w-full max-w-2xl bg-slate-900 rounded-xl border border-slate-700 shadow-2xl overflow-hidden"
        shouldFilter={false}
      >
        {/* Header */}
        <div className="flex items-center border-b border-slate-700 px-4 py-3">
          <Search className="h-5 w-5 text-slate-400 mr-3" />
          <Command.Input
            ref={inputRef}
            value={inputValue}
            onValueChange={setInputValue}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                handleCommandSubmit();
              }
            }}
            placeholder="Digite um comando... (ex: Set Primary Color #0EA5E9)"
            className="flex-1 bg-transparent border-0 outline-none text-white placeholder:text-slate-500 text-sm"
          />
          <div className="flex items-center gap-2">
            {/* Valid command indicator */}
            {inputValue.trim() && (
              <span className={cn(
                'text-xs px-2 py-1 rounded',
                isValidCommand
                  ? 'bg-emerald-900/30 text-emerald-400'
                  : 'bg-amber-900/30 text-amber-400'
              )}>
                {isValidCommand ? '✓ Comando válido' : '? Comando incompleto'}
              </span>
            )}
            {/* Execute Button */}
            <button
              type="button"
              onClick={handleCommandSubmit}
              disabled={!inputValue.trim()}
              className={cn(
                'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
                isValidCommand
                  ? 'bg-emerald-600 text-white hover:bg-emerald-700'
                  : inputValue.trim()
                    ? 'bg-violet-600 text-white hover:bg-violet-700'
                    : 'bg-slate-700 text-slate-500 cursor-not-allowed'
              )}
              title="Executar comando (Enter)"
            >
              Executar
            </button>
            <button
              type="button"
              onClick={() => setShowHistory(!showHistory)}
              className={cn(
                'p-1.5 rounded-md transition-colors',
                showHistory
                  ? 'bg-violet-600/20 text-violet-400'
                  : 'text-slate-400 hover:text-slate-300 hover:bg-slate-800'
              )}
              title="Historico de comandos"
            >
              <History className="h-4 w-4" />
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-md text-slate-400 hover:text-slate-300 hover:bg-slate-800 transition-colors"
              title="Fechar (Esc)"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Result message */}
        {lastResult && (
          <div
            className={cn(
              'flex items-center gap-2 px-4 py-2 text-sm border-b',
              lastResult.success
                ? 'bg-emerald-900/20 text-emerald-400 border-emerald-800/30'
                : 'bg-red-900/20 text-red-400 border-red-800/30'
            )}
          >
            {lastResult.success ? (
              <CheckCircle className="h-4 w-4" />
            ) : (
              <AlertCircle className="h-4 w-4" />
            )}
            {lastResult.message}
          </div>
        )}

        {/* Command List */}
        <Command.List className="max-h-[400px] overflow-y-auto p-2">
          {showHistory ? (
            // History view
            <div>
              <div className="px-2 py-1.5 text-xs font-medium text-slate-500 uppercase tracking-wider">
                Historico de Comandos
              </div>
              {commandHistory.length === 0 ? (
                <div className="px-4 py-8 text-center text-slate-500 text-sm">
                  Nenhum comando no historico
                </div>
              ) : (
                commandHistory.slice(0, 10).map((entry, index) => (
                  <Command.Item
                    key={index}
                    value={entry.command}
                    onSelect={() => handleHistorySelect(entry)}
                    className="flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer text-slate-300 hover:bg-slate-800 transition-colors"
                  >
                    <History className="h-4 w-4 text-slate-500" />
                    <div className="flex-1">
                      <p className="text-sm font-mono">{entry.command}</p>
                      <p className="text-xs text-slate-500">
                        {new Date(entry.timestamp).toLocaleString('pt-BR')}
                      </p>
                    </div>
                    {entry.success ? (
                      <CheckCircle className="h-4 w-4 text-emerald-500" />
                    ) : (
                      <AlertCircle className="h-4 w-4 text-red-500" />
                    )}
                  </Command.Item>
                ))
              )}
            </div>
          ) : (
            // Suggestions view
            <>
              {Object.entries(groupedSuggestions).map(([category, items]) => {
                const CategoryIcon = categoryIcons[category] || Layout;
                return (
                  <Command.Group key={category} heading="">
                    <div className="flex items-center gap-2 px-2 py-1.5 text-xs font-medium text-slate-500 uppercase tracking-wider">
                      <CategoryIcon className="h-3.5 w-3.5" />
                      {categoryLabels[category] || category}
                    </div>
                    {items.map((suggestion, index) => {
                      // Check if this suggestion has a valid example that can be executed directly
                      const canExecuteDirectly = suggestion.example
                        ? parseCommand(suggestion.example).isValid
                        : false;

                      return (
                        <Command.Item
                          key={`${category}-${index}`}
                          value={suggestion.command}
                          onSelect={() => handleSuggestionSelect(suggestion)}
                          className="flex items-center justify-between gap-3 px-3 py-2 rounded-lg cursor-pointer text-slate-300 hover:bg-slate-800 transition-colors group"
                        >
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-mono truncate">{suggestion.command}</p>
                            <p className="text-xs text-slate-500 truncate">
                              {suggestion.description}
                            </p>
                          </div>
                          <div className="flex items-center gap-1">
                            {canExecuteDirectly && (
                              <button
                                type="button"
                                onClick={(e) => handleSuggestionExecute(suggestion, e)}
                                className="p-1 rounded bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600/40 transition-colors opacity-0 group-hover:opacity-100"
                                title="Executar exemplo agora"
                              >
                                <Play className="h-3 w-3" />
                              </button>
                            )}
                            <ChevronRight className="h-4 w-4 text-slate-600 group-hover:text-slate-400 transition-colors" />
                          </div>
                        </Command.Item>
                      );
                    })}
                  </Command.Group>
                );
              })}

              {suggestions.length === 0 && (
                <div className="px-4 py-8 text-center">
                  <AlertCircle className="h-8 w-8 text-slate-600 mx-auto mb-2" />
                  <p className="text-slate-500 text-sm">
                    Nenhum comando encontrado para &quot;{inputValue}&quot;
                  </p>
                  <p className="text-slate-600 text-xs mt-1">
                    Pressione Enter para executar o comando digitado
                  </p>
                </div>
              )}
            </>
          )}
        </Command.List>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-slate-700 px-4 py-2 text-xs text-slate-500">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">
                Enter
              </kbd>
              <span>executar</span>
            </span>
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">
                Esc
              </kbd>
              <span>fechar</span>
            </span>
          </div>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">
              Ctrl
            </kbd>
            <span>+</span>
            <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">K</kbd>
            <span>abrir palette</span>
          </span>
        </div>
      </Command>
    </div>
  );
}

export default CommandPalette;
