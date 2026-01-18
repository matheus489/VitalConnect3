// Theme Configuration Types for Tenant Customization
// These types match the JSONB structure stored in the backend tenants.theme_config column

/**
 * Theme color configuration
 * Supports common CSS color values (hex, rgb, etc.)
 */
export interface ThemeColors {
  primary?: string;
  background?: string;
  foreground?: string;
  sidebar?: string;
  sidebar_foreground?: string;
  sidebar_accent?: string;
  accent?: string;
  muted?: string;
  card?: string;
  card_foreground?: string;
  border?: string;
  [key: string]: string | undefined;
}

/**
 * Font configuration for the theme
 */
export interface ThemeFonts {
  body?: string;
  heading?: string;
}

/**
 * Theme appearance settings (colors and fonts)
 */
export interface ThemeAppearance {
  colors?: ThemeColors;
  fonts?: ThemeFonts;
}

/**
 * Sidebar navigation item configuration
 */
export interface SidebarItem {
  id?: string;
  label: string;
  icon: string;
  link: string;
  roles?: string[];
  order?: number;
}

/**
 * Top navigation bar configuration
 */
export interface TopbarConfig {
  show_user_info?: boolean;
  show_tenant_logo?: boolean;
  show_notifications?: boolean;
}

/**
 * Dashboard widget types supported by the system
 */
export type DashboardWidgetType =
  | 'stats_card'
  | 'map_preview'
  | 'recent_occurrences'
  | 'chart'
  | 'activity_feed'
  | 'quick_actions';

/**
 * Dashboard widget configuration
 */
export interface DashboardWidget {
  id: string;
  type: DashboardWidgetType;
  visible: boolean;
  order: number;
  title?: string;
  config?: Record<string, unknown>;
}

/**
 * Layout configuration for the tenant UI
 */
export interface ThemeLayout {
  sidebar?: SidebarItem[];
  topbar?: TopbarConfig;
  dashboard_widgets?: DashboardWidget[];
}

/**
 * Complete theme configuration structure
 * This matches the JSONB structure in the database
 */
export interface ThemeConfig {
  theme?: ThemeAppearance;
  layout?: ThemeLayout;
}

/**
 * Default theme configuration used when tenant has no custom config
 */
export const DEFAULT_THEME_CONFIG: ThemeConfig = {
  theme: {
    colors: {
      primary: '#0EA5E9',
      background: '#FFFFFF',
      foreground: '#1F2937',
      sidebar: '#F9FAFB',
      sidebar_foreground: '#1F2937',
      sidebar_accent: '#E0F2FE',
      accent: '#E0F2FE',
      muted: '#F3F4F6',
      card: '#FFFFFF',
      card_foreground: '#1F2937',
      border: '#E5E7EB',
    },
    fonts: {
      body: 'Inter',
      heading: 'Inter',
    },
  },
  layout: {
    sidebar: [
      { label: 'Dashboard', icon: 'LayoutDashboard', link: '/dashboard', order: 1 },
      { label: 'Mapa', icon: 'MapPin', link: '/dashboard/map', order: 2 },
      { label: 'Ocorrencias', icon: 'ClipboardList', link: '/dashboard/occurrences', order: 3 },
      { label: 'Escalas', icon: 'Calendar', link: '/dashboard/shifts', order: 4 },
      { label: 'Regras', icon: 'Sliders', link: '/dashboard/rules', roles: ['admin', 'gestor'], order: 5 },
      { label: 'Hospitais', icon: 'Building2', link: '/dashboard/hospitals', roles: ['admin'], order: 6 },
      { label: 'Usuarios', icon: 'Users', link: '/dashboard/users', roles: ['admin'], order: 7 },
      { label: 'Relatorios', icon: 'FileText', link: '/dashboard/reports', roles: ['admin', 'gestor'], order: 8 },
      { label: 'Auditoria', icon: 'History', link: '/dashboard/audit-logs', roles: ['admin', 'gestor'], order: 9 },
      { label: 'Status', icon: 'Activity', link: '/dashboard/status', roles: ['admin'], order: 10 },
      { label: 'Configuracoes', icon: 'Settings', link: '/dashboard/settings', roles: ['admin', 'gestor'], order: 11 },
    ],
    topbar: {
      show_user_info: true,
      show_tenant_logo: true,
      show_notifications: true,
    },
    dashboard_widgets: [
      { id: 'stats', type: 'stats_card', visible: true, order: 1, title: 'Estatisticas' },
      { id: 'map', type: 'map_preview', visible: true, order: 2, title: 'Mapa' },
      { id: 'recent', type: 'recent_occurrences', visible: true, order: 3, title: 'Ocorrencias Recentes' },
      { id: 'chart', type: 'chart', visible: true, order: 4, title: 'Graficos' },
    ],
  },
};

/**
 * Command action types for the command palette
 */
export type CommandActionType =
  | 'SET_PRIMARY_COLOR'
  | 'SET_BACKGROUND_COLOR'
  | 'SET_SIDEBAR_COLOR'
  | 'SET_FONT'
  | 'SIDEBAR_ADD_ITEM'
  | 'SIDEBAR_REMOVE_ITEM'
  | 'SIDEBAR_MOVE_ITEM'
  | 'DASHBOARD_ADD_WIDGET'
  | 'DASHBOARD_HIDE_WIDGET'
  | 'DASHBOARD_SHOW_WIDGET'
  | 'UPLOAD_LOGO'
  | 'UPLOAD_FAVICON'
  | 'UNKNOWN';

/**
 * Parsed command action result
 */
export interface CommandAction {
  type: CommandActionType;
  payload: Record<string, string | number | boolean | string[] | undefined>;
  raw: string;
  isValid: boolean;
  error?: string;
}

/**
 * Command suggestion for auto-complete
 */
export interface CommandSuggestion {
  command: string;
  description: string;
  category: 'colors' | 'sidebar' | 'dashboard' | 'assets' | 'fonts';
  example?: string;
}

/**
 * Command history entry
 */
export interface CommandHistoryEntry {
  command: string;
  timestamp: Date;
  success: boolean;
  result?: string;
}
