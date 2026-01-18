'use client';

import { useMemo } from 'react';
import {
  LayoutDashboard,
  MapPin,
  ClipboardList,
  Building2,
  Users,
  Settings,
  FileText,
  History,
  Activity,
  Sliders,
  Calendar,
  Home,
  Bell,
  User,
  ChevronDown,
  BarChart3,
  TrendingUp,
  Clock,
  Eye,
  EyeOff,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ThemeConfig, SidebarItem, DashboardWidget } from '@/types/theme';

interface ThemePreviewProps {
  themeConfig: ThemeConfig;
  logoUrl?: string;
  faviconUrl?: string;
  tenantName?: string;
}

// Icon mapping for dynamic rendering
const iconMap: Record<string, React.ElementType> = {
  LayoutDashboard,
  MapPin,
  ClipboardList,
  Building2,
  Users,
  Settings,
  FileText,
  History,
  Activity,
  Sliders,
  Calendar,
  Home,
  Bell,
  User,
  BarChart3,
  TrendingUp,
  Clock,
};

function DynamicIcon({ name, className }: { name: string; className?: string }) {
  const Icon = iconMap[name] || LayoutDashboard;
  return <Icon className={className} />;
}

function PreviewSidebarItem({ item, colors }: { item: SidebarItem; colors: ThemeConfig['theme'] }) {
  const isActive = item.link === '/dashboard';

  return (
    <div
      className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-lg text-xs transition-colors cursor-default',
        isActive ? 'bg-white/10' : 'hover:bg-white/5'
      )}
      style={{
        color: colors?.colors?.sidebar_foreground || '#FFFFFF',
      }}
    >
      <DynamicIcon name={item.icon} className="h-4 w-4" />
      <span className="truncate">{item.label}</span>
    </div>
  );
}

function PreviewWidget({ widget }: { widget: DashboardWidget }) {
  if (!widget.visible) {
    return (
      <div className="bg-slate-800/50 rounded-lg p-4 border border-slate-700 border-dashed opacity-50">
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs font-medium text-slate-500">
            {widget.title || widget.type}
          </span>
          <EyeOff className="h-3 w-3 text-slate-600" />
        </div>
        <div className="text-xs text-slate-600">Widget oculto</div>
      </div>
    );
  }

  const renderWidgetContent = () => {
    switch (widget.type) {
      case 'stats_card':
        return (
          <div className="grid grid-cols-2 gap-2">
            <div className="bg-slate-700/50 rounded p-2">
              <div className="text-[10px] text-slate-400">Ocorrencias</div>
              <div className="text-sm font-semibold text-white">24</div>
            </div>
            <div className="bg-slate-700/50 rounded p-2">
              <div className="text-[10px] text-slate-400">Hospitais</div>
              <div className="text-sm font-semibold text-white">8</div>
            </div>
          </div>
        );
      case 'map_preview':
        return (
          <div className="bg-slate-700/30 rounded h-16 flex items-center justify-center">
            <MapPin className="h-6 w-6 text-slate-500" />
          </div>
        );
      case 'recent_occurrences':
        return (
          <div className="space-y-1">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center gap-2 text-[10px]">
                <div className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                <span className="text-slate-400">Ocorrencia #{i}</span>
              </div>
            ))}
          </div>
        );
      case 'chart':
        return (
          <div className="flex items-end gap-1 h-12">
            {[40, 60, 35, 80, 55, 70].map((h, i) => (
              <div
                key={i}
                className="flex-1 bg-violet-500/50 rounded-t"
                style={{ height: `${h}%` }}
              />
            ))}
          </div>
        );
      case 'activity_feed':
        return (
          <div className="space-y-1 text-[10px] text-slate-400">
            <div>Admin atualizou config</div>
            <div>Nova ocorrencia criada</div>
          </div>
        );
      case 'quick_actions':
        return (
          <div className="flex gap-1">
            <button className="flex-1 bg-violet-600/20 text-violet-400 rounded px-2 py-1 text-[10px]">
              Acao 1
            </button>
            <button className="flex-1 bg-slate-700 text-slate-300 rounded px-2 py-1 text-[10px]">
              Acao 2
            </button>
          </div>
        );
      default:
        return <div className="text-xs text-slate-500">Widget: {widget.type}</div>;
    }
  };

  return (
    <div className="bg-slate-800 rounded-lg p-3 border border-slate-700">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-slate-300">
          {widget.title || widget.type}
        </span>
        <Eye className="h-3 w-3 text-slate-500" />
      </div>
      {renderWidgetContent()}
    </div>
  );
}

export function ThemePreview({
  themeConfig,
  logoUrl,
  faviconUrl,
  tenantName = 'Tenant',
}: ThemePreviewProps) {
  const { theme, layout } = themeConfig;

  // Extract colors with fallbacks
  const colors = useMemo(
    () => ({
      primary: theme?.colors?.primary || '#0EA5E9',
      background: theme?.colors?.background || '#1F2937',
      foreground: theme?.colors?.foreground || '#FFFFFF',
      sidebar: theme?.colors?.sidebar || '#111827',
      sidebarForeground: theme?.colors?.sidebar_foreground || '#FFFFFF',
      sidebarAccent: theme?.colors?.sidebar_accent || '#374151',
      card: theme?.colors?.card || '#1F2937',
      cardForeground: theme?.colors?.card_foreground || '#FFFFFF',
    }),
    [theme?.colors]
  );

  const fontFamily = theme?.fonts?.body || 'Inter';

  // Get sidebar items with fallback
  const sidebarItems = layout?.sidebar || [
    { label: 'Dashboard', icon: 'LayoutDashboard', link: '/dashboard' },
    { label: 'Mapa', icon: 'MapPin', link: '/dashboard/map' },
    { label: 'Ocorrencias', icon: 'ClipboardList', link: '/dashboard/occurrences' },
  ];

  // Get dashboard widgets with fallback
  const dashboardWidgets = layout?.dashboard_widgets || [
    { id: 'stats', type: 'stats_card' as const, visible: true, order: 1, title: 'Estatisticas' },
    { id: 'chart', type: 'chart' as const, visible: true, order: 2, title: 'Grafico' },
  ];

  // Sort widgets by order
  const sortedWidgets = [...dashboardWidgets].sort((a, b) => a.order - b.order);

  return (
    <div
      className="rounded-xl overflow-hidden border border-slate-700 shadow-xl"
      style={{
        fontFamily: fontFamily,
        backgroundColor: colors.background,
        color: colors.foreground,
      }}
    >
      {/* Mini Browser Chrome */}
      <div className="flex items-center gap-2 px-3 py-2 bg-slate-900 border-b border-slate-700">
        <div className="flex gap-1.5">
          <div className="h-2.5 w-2.5 rounded-full bg-red-500/70" />
          <div className="h-2.5 w-2.5 rounded-full bg-yellow-500/70" />
          <div className="h-2.5 w-2.5 rounded-full bg-green-500/70" />
        </div>
        <div className="flex-1 flex items-center justify-center">
          <div className="flex items-center gap-2 bg-slate-800 rounded-full px-3 py-1 text-[10px] text-slate-400">
            {faviconUrl ? (
              <img src={faviconUrl} alt="Favicon" className="h-3 w-3 rounded" />
            ) : (
              <div
                className="h-3 w-3 rounded flex items-center justify-center text-[8px] font-bold"
                style={{ backgroundColor: colors.primary, color: '#FFFFFF' }}
              >
                V
              </div>
            )}
            <span>app.vitalconnect.com/{tenantName.toLowerCase().replace(/\s+/g, '-')}</span>
          </div>
        </div>
      </div>

      {/* Preview Content */}
      <div className="flex h-[400px]">
        {/* Sidebar */}
        <div
          className="w-44 flex flex-col border-r border-slate-700/50"
          style={{ backgroundColor: colors.sidebar }}
        >
          {/* Logo */}
          <div className="flex items-center gap-2 px-3 py-3 border-b border-white/10">
            {logoUrl ? (
              <img src={logoUrl} alt="Logo" className="h-6 w-6 rounded" />
            ) : (
              <div
                className="h-6 w-6 rounded flex items-center justify-center text-xs font-bold"
                style={{ backgroundColor: colors.primary, color: '#FFFFFF' }}
              >
                V
              </div>
            )}
            <span
              className="text-sm font-semibold truncate"
              style={{ color: colors.sidebarForeground }}
            >
              {tenantName}
            </span>
          </div>

          {/* Navigation */}
          <nav className="flex-1 p-2 space-y-0.5 overflow-y-auto">
            {sidebarItems.map((item, index) => (
              <PreviewSidebarItem key={index} item={item} colors={theme} />
            ))}
          </nav>

          {/* Sidebar Footer */}
          <div className="border-t border-white/10 p-2">
            <div className="flex items-center gap-2 px-3 py-2 text-xs" style={{ color: colors.sidebarForeground }}>
              <User className="h-4 w-4" />
              <span className="truncate">usuario@email.com</span>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="flex-1 flex flex-col" style={{ backgroundColor: colors.background }}>
          {/* Topbar */}
          <div
            className="flex items-center justify-between px-4 py-2 border-b"
            style={{
              backgroundColor: colors.card,
              borderColor: 'rgba(255,255,255,0.1)',
            }}
          >
            <h1 className="text-sm font-semibold" style={{ color: colors.cardForeground }}>
              Dashboard
            </h1>
            <div className="flex items-center gap-3">
              {layout?.topbar?.show_notifications !== false && (
                <Bell className="h-4 w-4 text-slate-400" />
              )}
              {layout?.topbar?.show_user_info !== false && (
                <div className="flex items-center gap-2 text-xs text-slate-400">
                  <User className="h-4 w-4" />
                  <ChevronDown className="h-3 w-3" />
                </div>
              )}
            </div>
          </div>

          {/* Dashboard Content */}
          <div className="flex-1 p-4 overflow-y-auto">
            <div className="grid grid-cols-2 gap-3">
              {sortedWidgets.map((widget, index) => (
                <PreviewWidget key={widget.id || `widget-${index}`} widget={widget} />
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Theme Info Footer */}
      <div className="flex items-center justify-between px-3 py-2 bg-slate-900/50 border-t border-slate-700 text-[10px] text-slate-500">
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1">
            <div
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: colors.primary }}
            />
            Primary: {colors.primary}
          </span>
          <span className="flex items-center gap-1">
            <div
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: colors.background }}
            />
            Background: {colors.background}
          </span>
        </div>
        <span>Font: {fontFamily}</span>
      </div>
    </div>
  );
}

export default ThemePreview;
