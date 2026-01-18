'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import {
  Save,
  RotateCcw,
  Command,
  Eye,
  Split,
  Maximize2,
  Minimize2,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { CommandPalette } from './CommandPalette';
import { ThemePreview } from './ThemePreview';
import { applyCommandToTheme } from './CommandParser';
import type {
  ThemeConfig,
  CommandAction,
  CommandHistoryEntry,
} from '@/types/theme';
import { DEFAULT_THEME_CONFIG } from '@/types/theme';
import { cn } from '@/lib/utils';

interface TenantThemeEditorProps {
  tenantId: string;
  tenantName: string;
  initialThemeConfig: ThemeConfig;
  logoUrl?: string;
  faviconUrl?: string;
  onSave: (themeConfig: ThemeConfig) => Promise<void>;
  onUploadLogo?: () => void;
  onUploadFavicon?: () => void;
  isSaving?: boolean;
}

type ViewMode = 'split' | 'preview-only' | 'editor-only';

export function TenantThemeEditor({
  tenantId,
  tenantName,
  initialThemeConfig,
  logoUrl,
  faviconUrl,
  onSave,
  onUploadLogo,
  onUploadFavicon,
  isSaving = false,
}: TenantThemeEditorProps) {
  // Theme state - merge with defaults to ensure all fields exist
  const [themeConfig, setThemeConfig] = useState<ThemeConfig>(() => ({
    ...DEFAULT_THEME_CONFIG,
    ...initialThemeConfig,
    theme: {
      ...DEFAULT_THEME_CONFIG.theme,
      ...initialThemeConfig.theme,
      colors: {
        ...DEFAULT_THEME_CONFIG.theme?.colors,
        ...initialThemeConfig.theme?.colors,
      },
      fonts: {
        ...DEFAULT_THEME_CONFIG.theme?.fonts,
        ...initialThemeConfig.theme?.fonts,
      },
    },
    layout: {
      ...DEFAULT_THEME_CONFIG.layout,
      ...initialThemeConfig.layout,
    },
  }));

  // Store the last saved config to enable reset
  const [lastSavedConfig, setLastSavedConfig] = useState<ThemeConfig>(themeConfig);

  // Command palette state
  const [isPaletteOpen, setIsPaletteOpen] = useState(false);
  const [commandHistory, setCommandHistory] = useState<CommandHistoryEntry[]>([]);

  // View mode state
  const [viewMode, setViewMode] = useState<ViewMode>('split');

  // Track if there are unsaved changes
  const hasUnsavedChanges = JSON.stringify(themeConfig) !== JSON.stringify(lastSavedConfig);

  // Handle keyboard shortcut to open command palette
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        setIsPaletteOpen(true);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Handle command execution
  const handleCommandExecute = useCallback((action: CommandAction) => {
    const newTheme = applyCommandToTheme(action, themeConfig);
    setThemeConfig(newTheme);

    // Add to history
    setCommandHistory((prev) => [
      {
        command: action.raw,
        timestamp: new Date(),
        success: action.isValid,
        result: action.isValid ? 'Aplicado com sucesso' : action.error,
      },
      ...prev.slice(0, 49), // Keep last 50 commands
    ]);
  }, [themeConfig]);

  // Handle save
  const handleSave = useCallback(async () => {
    try {
      await onSave(themeConfig);
      setLastSavedConfig(themeConfig);
    } catch (error) {
      console.error('Failed to save theme:', error);
    }
  }, [themeConfig, onSave]);

  // Handle reset
  const handleReset = useCallback(() => {
    setThemeConfig(lastSavedConfig);
  }, [lastSavedConfig]);

  // Get view mode icon
  const getViewModeIcon = () => {
    switch (viewMode) {
      case 'preview-only':
        return <Eye className="h-4 w-4" />;
      case 'editor-only':
        return <Command className="h-4 w-4" />;
      default:
        return <Split className="h-4 w-4" />;
    }
  };

  // Cycle through view modes
  const cycleViewMode = () => {
    setViewMode((prev) => {
      switch (prev) {
        case 'split':
          return 'preview-only';
        case 'preview-only':
          return 'editor-only';
        default:
          return 'split';
      }
    });
  };

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsPaletteOpen(true)}
            className="gap-2 bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700 hover:text-white"
          >
            <Command className="h-4 w-4" />
            <span>Abrir Command Palette</span>
            <kbd className="ml-2 px-1.5 py-0.5 rounded bg-slate-700 text-slate-400 text-xs font-mono">
              Ctrl+K
            </kbd>
          </Button>
        </div>

        <div className="flex items-center gap-2">
          {/* View Mode Toggle */}
          <Button
            variant="ghost"
            size="sm"
            onClick={cycleViewMode}
            className="gap-2 text-slate-400 hover:text-white hover:bg-slate-800"
            title={`Modo: ${viewMode === 'split' ? 'Dividido' : viewMode === 'preview-only' ? 'Preview' : 'Editor'}`}
          >
            {getViewModeIcon()}
          </Button>

          {/* Reset Button */}
          <Button
            variant="outline"
            size="sm"
            onClick={handleReset}
            disabled={!hasUnsavedChanges || isSaving}
            className="gap-2 bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700 hover:text-white disabled:opacity-50"
          >
            <RotateCcw className="h-4 w-4" />
            Resetar
          </Button>

          {/* Save Button */}
          <Button
            size="sm"
            onClick={handleSave}
            disabled={!hasUnsavedChanges || isSaving}
            className="gap-2 bg-violet-600 hover:bg-violet-700 text-white disabled:opacity-50"
          >
            <Save className="h-4 w-4" />
            {isSaving ? 'Salvando...' : 'Salvar Alteracoes'}
          </Button>
        </div>
      </div>

      {/* Unsaved Changes Indicator */}
      {hasUnsavedChanges && (
        <div className="mb-4 px-3 py-2 rounded-lg bg-amber-900/20 border border-amber-800/30 text-amber-400 text-sm flex-shrink-0">
          Voce tem alteracoes nao salvas. Clique em &quot;Salvar Alteracoes&quot; para persistir as mudancas.
        </div>
      )}

      {/* Main Content - Split Pane */}
      <div className={cn('flex-1 flex gap-4 min-h-0', viewMode === 'editor-only' && 'flex-col')}>
        {/* Command History Panel (Editor Side) */}
        {viewMode !== 'preview-only' && (
          <Card className={cn(
            'bg-slate-800 border-slate-700 flex flex-col',
            viewMode === 'split' ? 'w-1/2' : 'flex-1'
          )}>
            <CardHeader className="flex-shrink-0 pb-3">
              <CardTitle className="text-white text-base flex items-center gap-2">
                <Command className="h-4 w-4 text-violet-400" />
                Command Editor
              </CardTitle>
              <CardDescription className="text-slate-400 text-sm">
                Use o Command Palette (Ctrl+K) para executar comandos de personalizacao
              </CardDescription>
            </CardHeader>
            <CardContent className="flex-1 overflow-hidden flex flex-col">
              {/* Quick Actions */}
              <div className="flex flex-wrap gap-2 mb-4 flex-shrink-0">
                <button
                  onClick={() => setIsPaletteOpen(true)}
                  className="px-3 py-1.5 rounded-lg bg-violet-600/20 text-violet-400 text-xs font-medium hover:bg-violet-600/30 transition-colors"
                >
                  + Nova Cor
                </button>
                <button
                  onClick={() => setIsPaletteOpen(true)}
                  className="px-3 py-1.5 rounded-lg bg-emerald-600/20 text-emerald-400 text-xs font-medium hover:bg-emerald-600/30 transition-colors"
                >
                  + Item no Menu
                </button>
                <button
                  onClick={() => setIsPaletteOpen(true)}
                  className="px-3 py-1.5 rounded-lg bg-blue-600/20 text-blue-400 text-xs font-medium hover:bg-blue-600/30 transition-colors"
                >
                  + Widget
                </button>
                {onUploadLogo && (
                  <button
                    onClick={onUploadLogo}
                    className="px-3 py-1.5 rounded-lg bg-amber-600/20 text-amber-400 text-xs font-medium hover:bg-amber-600/30 transition-colors"
                  >
                    Upload Logo
                  </button>
                )}
              </div>

              {/* Command History */}
              <div className="flex-1 overflow-y-auto">
                <div className="text-xs font-medium text-slate-500 uppercase tracking-wider mb-2">
                  Historico de Comandos
                </div>
                {commandHistory.length === 0 ? (
                  <div className="text-center py-8">
                    <Command className="h-8 w-8 text-slate-600 mx-auto mb-2" />
                    <p className="text-slate-500 text-sm">Nenhum comando executado</p>
                    <p className="text-slate-600 text-xs mt-1">
                      Pressione Ctrl+K para abrir o Command Palette
                    </p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {commandHistory.slice(0, 20).map((entry, index) => (
                      <div
                        key={index}
                        className={cn(
                          'px-3 py-2 rounded-lg text-sm font-mono',
                          entry.success
                            ? 'bg-slate-900/50 text-slate-300 border border-slate-700/50'
                            : 'bg-red-900/20 text-red-400 border border-red-800/30'
                        )}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <code className="text-xs break-all">{entry.command}</code>
                          <span className="text-[10px] text-slate-500 whitespace-nowrap">
                            {new Date(entry.timestamp).toLocaleTimeString('pt-BR')}
                          </span>
                        </div>
                        {entry.result && (
                          <p className="text-[10px] text-slate-500 mt-1">{entry.result}</p>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Preview Panel */}
        {viewMode !== 'editor-only' && (
          <div className={cn(
            'flex flex-col',
            viewMode === 'split' ? 'w-1/2' : 'flex-1'
          )}>
            <div className="flex items-center justify-between mb-2 flex-shrink-0">
              <span className="text-sm font-medium text-slate-300 flex items-center gap-2">
                <Eye className="h-4 w-4 text-violet-400" />
                Preview em Tempo Real
              </span>
            </div>
            <div className="flex-1 min-h-0">
              <ThemePreview
                themeConfig={themeConfig}
                logoUrl={logoUrl}
                faviconUrl={faviconUrl}
                tenantName={tenantName}
              />
            </div>
          </div>
        )}
      </div>

      {/* Current Theme Summary */}
      <div className="mt-4 p-3 rounded-lg bg-slate-800/50 border border-slate-700/50 flex-shrink-0">
        <div className="flex flex-wrap items-center gap-4 text-xs text-slate-400">
          <span className="flex items-center gap-2">
            <span className="font-medium text-slate-300">Primary:</span>
            <span
              className="h-4 w-4 rounded border border-slate-600"
              style={{ backgroundColor: themeConfig.theme?.colors?.primary || '#0EA5E9' }}
            />
            <code>{themeConfig.theme?.colors?.primary || '#0EA5E9'}</code>
          </span>
          <span className="flex items-center gap-2">
            <span className="font-medium text-slate-300">Background:</span>
            <span
              className="h-4 w-4 rounded border border-slate-600"
              style={{ backgroundColor: themeConfig.theme?.colors?.background || '#1F2937' }}
            />
            <code>{themeConfig.theme?.colors?.background || '#1F2937'}</code>
          </span>
          <span className="flex items-center gap-2">
            <span className="font-medium text-slate-300">Font:</span>
            <code>{themeConfig.theme?.fonts?.body || 'Inter'}</code>
          </span>
          <span className="flex items-center gap-2">
            <span className="font-medium text-slate-300">Sidebar Items:</span>
            <code>{themeConfig.layout?.sidebar?.length || 0}</code>
          </span>
          <span className="flex items-center gap-2">
            <span className="font-medium text-slate-300">Widgets:</span>
            <code>{themeConfig.layout?.dashboard_widgets?.length || 0}</code>
          </span>
        </div>
      </div>

      {/* Command Palette Modal */}
      <CommandPalette
        isOpen={isPaletteOpen}
        onClose={() => setIsPaletteOpen(false)}
        onCommandExecute={handleCommandExecute}
        onUploadLogo={onUploadLogo}
        onUploadFavicon={onUploadFavicon}
        commandHistory={commandHistory}
      />
    </div>
  );
}

export default TenantThemeEditor;
