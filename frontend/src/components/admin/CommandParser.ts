import type {
  CommandAction,
  CommandActionType,
  CommandSuggestion,
  ThemeConfig,
  SidebarItem,
  DashboardWidget,
  DashboardWidgetType,
} from '@/types/theme';

/**
 * Command patterns for parsing user input
 * Each pattern includes the regex and corresponding action type
 */
const COMMAND_PATTERNS: Array<{
  pattern: RegExp;
  type: CommandActionType;
  extractor: (match: RegExpMatchArray) => Record<string, string | number | boolean | string[] | undefined>;
}> = [
  // Color commands
  {
    pattern: /^set\s+primary\s+color\s+(#[0-9A-Fa-f]{6}|#[0-9A-Fa-f]{3})$/i,
    type: 'SET_PRIMARY_COLOR',
    extractor: (match) => ({ color: match[1] }),
  },
  {
    pattern: /^set\s+background\s+(#[0-9A-Fa-f]{6}|#[0-9A-Fa-f]{3})$/i,
    type: 'SET_BACKGROUND_COLOR',
    extractor: (match) => ({ color: match[1] }),
  },
  {
    pattern: /^set\s+sidebar\s+color\s+(#[0-9A-Fa-f]{6}|#[0-9A-Fa-f]{3})$/i,
    type: 'SET_SIDEBAR_COLOR',
    extractor: (match) => ({ color: match[1] }),
  },
  // Font command
  {
    pattern: /^set\s+font\s+"([^"]+)"$/i,
    type: 'SET_FONT',
    extractor: (match) => ({ fontFamily: match[1] }),
  },
  // Sidebar commands
  {
    pattern: /^sidebar:\s*add\s+item\s+"([^"]+)"\s+icon="([^"]+)"\s+link="([^"]+)"(?:\s+roles="([^"]+)")?$/i,
    type: 'SIDEBAR_ADD_ITEM',
    extractor: (match) => ({
      label: match[1],
      icon: match[2],
      link: match[3],
      roles: match[4] ? match[4].split(',').map((r) => r.trim()) : undefined,
    }),
  },
  {
    pattern: /^sidebar:\s*remove\s+"([^"]+)"$/i,
    type: 'SIDEBAR_REMOVE_ITEM',
    extractor: (match) => ({ label: match[1] }),
  },
  {
    pattern: /^sidebar:\s*move\s+"([^"]+)"\s+to\s+(top|bottom)$/i,
    type: 'SIDEBAR_MOVE_ITEM',
    extractor: (match) => ({ label: match[1], position: match[2].toLowerCase() }),
  },
  // Dashboard commands
  {
    pattern: /^dashboard:\s*add\s+widget\s+"([^"]+)"$/i,
    type: 'DASHBOARD_ADD_WIDGET',
    extractor: (match) => ({ widgetType: match[1] }),
  },
  {
    pattern: /^dashboard:\s*hide\s+"([^"]+)"$/i,
    type: 'DASHBOARD_HIDE_WIDGET',
    extractor: (match) => ({ widgetId: match[1] }),
  },
  {
    pattern: /^dashboard:\s*show\s+"([^"]+)"$/i,
    type: 'DASHBOARD_SHOW_WIDGET',
    extractor: (match) => ({ widgetId: match[1] }),
  },
  // Asset commands
  {
    pattern: /^upload\s+logo$/i,
    type: 'UPLOAD_LOGO',
    extractor: () => ({}),
  },
  {
    pattern: /^upload\s+favicon$/i,
    type: 'UPLOAD_FAVICON',
    extractor: () => ({}),
  },
];

/**
 * Available command suggestions for auto-complete
 */
export const COMMAND_SUGGESTIONS: CommandSuggestion[] = [
  // Color commands
  {
    command: 'Set Primary Color #',
    description: 'Define a cor primaria do tema',
    category: 'colors',
    example: 'Set Primary Color #0EA5E9',
  },
  {
    command: 'Set Background #',
    description: 'Define a cor de fundo',
    category: 'colors',
    example: 'Set Background #FFFFFF',
  },
  {
    command: 'Set Sidebar Color #',
    description: 'Define a cor da barra lateral',
    category: 'colors',
    example: 'Set Sidebar Color #1F2937',
  },
  // Font command
  {
    command: 'Set Font ""',
    description: 'Define a fonte do sistema',
    category: 'fonts',
    example: 'Set Font "Inter"',
  },
  // Sidebar commands
  {
    command: 'Sidebar: Add Item "" icon="" link=""',
    description: 'Adiciona um novo item ao menu lateral',
    category: 'sidebar',
    example: 'Sidebar: Add Item "Relatorios" icon="FileText" link="/dashboard/reports"',
  },
  {
    command: 'Sidebar: Remove ""',
    description: 'Remove um item do menu lateral',
    category: 'sidebar',
    example: 'Sidebar: Remove "Configuracoes"',
  },
  {
    command: 'Sidebar: Move "" to Top',
    description: 'Move um item para o topo do menu',
    category: 'sidebar',
    example: 'Sidebar: Move "Dashboard" to Top',
  },
  {
    command: 'Sidebar: Move "" to Bottom',
    description: 'Move um item para o final do menu',
    category: 'sidebar',
    example: 'Sidebar: Move "Configuracoes" to Bottom',
  },
  // Dashboard commands
  {
    command: 'Dashboard: Add Widget ""',
    description: 'Adiciona um widget ao dashboard',
    category: 'dashboard',
    example: 'Dashboard: Add Widget "stats_card"',
  },
  {
    command: 'Dashboard: Hide ""',
    description: 'Oculta um widget do dashboard',
    category: 'dashboard',
    example: 'Dashboard: Hide "chart"',
  },
  {
    command: 'Dashboard: Show ""',
    description: 'Exibe um widget oculto',
    category: 'dashboard',
    example: 'Dashboard: Show "map"',
  },
  // Asset commands
  {
    command: 'Upload Logo',
    description: 'Faz upload de um novo logo',
    category: 'assets',
  },
  {
    command: 'Upload Favicon',
    description: 'Faz upload de um novo favicon',
    category: 'assets',
  },
];

/**
 * Validates a hex color code
 */
function isValidHexColor(color: string): boolean {
  return /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(color);
}

/**
 * Validates a widget type
 */
function isValidWidgetType(type: string): type is DashboardWidgetType {
  const validTypes: DashboardWidgetType[] = [
    'stats_card',
    'map_preview',
    'recent_occurrences',
    'chart',
    'activity_feed',
    'quick_actions',
  ];
  return validTypes.includes(type as DashboardWidgetType);
}

/**
 * Parses a command string into a structured action
 */
export function parseCommand(command: string): CommandAction {
  const trimmedCommand = command.trim();

  if (!trimmedCommand) {
    return {
      type: 'UNKNOWN',
      payload: {},
      raw: command,
      isValid: false,
      error: 'Comando vazio',
    };
  }

  for (const { pattern, type, extractor } of COMMAND_PATTERNS) {
    const match = trimmedCommand.match(pattern);
    if (match) {
      const payload = extractor(match);

      // Validate specific command types
      if (type === 'SET_PRIMARY_COLOR' || type === 'SET_BACKGROUND_COLOR' || type === 'SET_SIDEBAR_COLOR') {
        const color = payload.color as string;
        if (!isValidHexColor(color)) {
          return {
            type,
            payload,
            raw: command,
            isValid: false,
            error: `Cor invalida: ${color}. Use formato hexadecimal (#RGB ou #RRGGBB)`,
          };
        }
      }

      if (type === 'DASHBOARD_ADD_WIDGET') {
        const widgetType = payload.widgetType as string;
        if (!isValidWidgetType(widgetType)) {
          return {
            type,
            payload,
            raw: command,
            isValid: false,
            error: `Tipo de widget invalido: ${widgetType}. Tipos validos: stats_card, map_preview, recent_occurrences, chart, activity_feed, quick_actions`,
          };
        }
      }

      return {
        type,
        payload,
        raw: command,
        isValid: true,
      };
    }
  }

  return {
    type: 'UNKNOWN',
    payload: {},
    raw: command,
    isValid: false,
    error: `Comando nao reconhecido: ${trimmedCommand}`,
  };
}

/**
 * Gets command suggestions based on partial input
 */
export function getCommandSuggestions(partialInput: string): CommandSuggestion[] {
  const input = partialInput.toLowerCase().trim();

  if (!input) {
    return COMMAND_SUGGESTIONS;
  }

  return COMMAND_SUGGESTIONS.filter(
    (suggestion) =>
      suggestion.command.toLowerCase().includes(input) ||
      suggestion.description.toLowerCase().includes(input) ||
      suggestion.category.toLowerCase().includes(input)
  );
}

/**
 * Applies a command action to a theme configuration
 * Returns a new theme config with the changes applied
 */
export function applyCommandToTheme(
  action: CommandAction,
  currentTheme: ThemeConfig
): ThemeConfig {
  if (!action.isValid) {
    return currentTheme;
  }

  const newTheme: ThemeConfig = JSON.parse(JSON.stringify(currentTheme));

  // Ensure nested objects exist
  if (!newTheme.theme) newTheme.theme = {};
  if (!newTheme.theme.colors) newTheme.theme.colors = {};
  if (!newTheme.theme.fonts) newTheme.theme.fonts = {};
  if (!newTheme.layout) newTheme.layout = {};
  if (!newTheme.layout.sidebar) newTheme.layout.sidebar = [];
  if (!newTheme.layout.dashboard_widgets) newTheme.layout.dashboard_widgets = [];

  switch (action.type) {
    case 'SET_PRIMARY_COLOR':
      newTheme.theme.colors.primary = action.payload.color as string;
      break;

    case 'SET_BACKGROUND_COLOR':
      newTheme.theme.colors.background = action.payload.color as string;
      break;

    case 'SET_SIDEBAR_COLOR':
      newTheme.theme.colors.sidebar = action.payload.color as string;
      break;

    case 'SET_FONT':
      newTheme.theme.fonts.body = action.payload.fontFamily as string;
      newTheme.theme.fonts.heading = action.payload.fontFamily as string;
      break;

    case 'SIDEBAR_ADD_ITEM': {
      const newItem: SidebarItem = {
        label: action.payload.label as string,
        icon: action.payload.icon as string,
        link: action.payload.link as string,
        roles: action.payload.roles as string[] | undefined,
        order: newTheme.layout.sidebar!.length + 1,
      };
      newTheme.layout.sidebar!.push(newItem);
      break;
    }

    case 'SIDEBAR_REMOVE_ITEM': {
      const labelToRemove = action.payload.label as string;
      newTheme.layout.sidebar = newTheme.layout.sidebar!.filter(
        (item) => item.label.toLowerCase() !== labelToRemove.toLowerCase()
      );
      break;
    }

    case 'SIDEBAR_MOVE_ITEM': {
      const labelToMove = action.payload.label as string;
      const position = action.payload.position as string;
      const itemIndex = newTheme.layout.sidebar!.findIndex(
        (item) => item.label.toLowerCase() === labelToMove.toLowerCase()
      );

      if (itemIndex !== -1) {
        const [item] = newTheme.layout.sidebar!.splice(itemIndex, 1);
        if (position === 'top') {
          newTheme.layout.sidebar!.unshift(item);
        } else {
          newTheme.layout.sidebar!.push(item);
        }
      }
      break;
    }

    case 'DASHBOARD_ADD_WIDGET': {
      const widgetType = action.payload.widgetType as DashboardWidgetType;
      const newWidget: DashboardWidget = {
        id: `widget_${Date.now()}`,
        type: widgetType,
        visible: true,
        order: newTheme.layout.dashboard_widgets!.length + 1,
        title: widgetType.replace('_', ' '),
      };
      newTheme.layout.dashboard_widgets!.push(newWidget);
      break;
    }

    case 'DASHBOARD_HIDE_WIDGET': {
      const widgetId = action.payload.widgetId as string;
      const widget = newTheme.layout.dashboard_widgets!.find(
        (w) => w.id === widgetId || w.type === widgetId
      );
      if (widget) {
        widget.visible = false;
      }
      break;
    }

    case 'DASHBOARD_SHOW_WIDGET': {
      const widgetId = action.payload.widgetId as string;
      const widget = newTheme.layout.dashboard_widgets!.find(
        (w) => w.id === widgetId || w.type === widgetId
      );
      if (widget) {
        widget.visible = true;
      }
      break;
    }

    // UPLOAD_LOGO and UPLOAD_FAVICON are handled separately (trigger file picker)
    default:
      break;
  }

  return newTheme;
}

/**
 * Formats a command action into a human-readable result message
 */
export function formatCommandResult(action: CommandAction): string {
  if (!action.isValid) {
    return action.error || 'Comando invalido';
  }

  switch (action.type) {
    case 'SET_PRIMARY_COLOR':
      return `Cor primaria alterada para ${action.payload.color}`;
    case 'SET_BACKGROUND_COLOR':
      return `Cor de fundo alterada para ${action.payload.color}`;
    case 'SET_SIDEBAR_COLOR':
      return `Cor da sidebar alterada para ${action.payload.color}`;
    case 'SET_FONT':
      return `Fonte alterada para "${action.payload.fontFamily}"`;
    case 'SIDEBAR_ADD_ITEM':
      return `Item "${action.payload.label}" adicionado ao menu`;
    case 'SIDEBAR_REMOVE_ITEM':
      return `Item "${action.payload.label}" removido do menu`;
    case 'SIDEBAR_MOVE_ITEM':
      return `Item "${action.payload.label}" movido para ${action.payload.position === 'top' ? 'o inicio' : 'o final'}`;
    case 'DASHBOARD_ADD_WIDGET':
      return `Widget "${action.payload.widgetType}" adicionado ao dashboard`;
    case 'DASHBOARD_HIDE_WIDGET':
      return `Widget "${action.payload.widgetId}" ocultado`;
    case 'DASHBOARD_SHOW_WIDGET':
      return `Widget "${action.payload.widgetId}" exibido`;
    case 'UPLOAD_LOGO':
      return 'Selecione um arquivo para o logo';
    case 'UPLOAD_FAVICON':
      return 'Selecione um arquivo para o favicon';
    default:
      return 'Comando executado';
  }
}
