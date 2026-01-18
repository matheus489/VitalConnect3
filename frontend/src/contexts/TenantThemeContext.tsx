'use client';

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  type ReactNode,
} from 'react';
import { useTenantTheme as useTenantThemeHook } from '@/hooks/useTenantTheme';
import { DEFAULT_THEME_CONFIG, type ThemeConfig, type ThemeColors } from '@/types/theme';

/**
 * Context value type for TenantThemeContext
 */
interface TenantThemeContextValue {
  /** The complete theme configuration */
  themeConfig: ThemeConfig;
  /** Whether the theme is loading */
  isLoading: boolean;
  /** Logo URL for the tenant */
  logoUrl?: string;
  /** Favicon URL for the tenant */
  faviconUrl?: string;
  /** Refetch the theme from the API */
  refetch: () => void;
}

const TenantThemeContext = createContext<TenantThemeContextValue | undefined>(undefined);

/**
 * Convert a hex color to HSL for CSS variables
 */
function hexToHsl(hex: string): string {
  // Remove the hash if present
  hex = hex.replace(/^#/, '');

  // Parse hex values
  const r = parseInt(hex.substring(0, 2), 16) / 255;
  const g = parseInt(hex.substring(2, 4), 16) / 255;
  const b = parseInt(hex.substring(4, 6), 16) / 255;

  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const l = (max + min) / 2;

  let h = 0;
  let s = 0;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);

    switch (max) {
      case r:
        h = ((g - b) / d + (g < b ? 6 : 0)) / 6;
        break;
      case g:
        h = ((b - r) / d + 2) / 6;
        break;
      case b:
        h = ((r - g) / d + 4) / 6;
        break;
    }
  }

  return `${Math.round(h * 360)} ${Math.round(s * 100)}% ${Math.round(l * 100)}%`;
}

/**
 * Check if a string is a valid hex color
 */
function isHexColor(color: string): boolean {
  return /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(color);
}

/**
 * CSS variable names mapped from theme color keys
 */
const colorVariableMap: Record<keyof ThemeColors, string> = {
  primary: '--primary',
  background: '--background',
  foreground: '--foreground',
  sidebar: '--sidebar',
  sidebar_foreground: '--sidebar-foreground',
  sidebar_accent: '--sidebar-accent',
  accent: '--accent',
  muted: '--muted',
  card: '--card',
  card_foreground: '--card-foreground',
  border: '--border',
};

/**
 * Apply theme colors to the document root as CSS custom properties
 */
function applyThemeColors(colors: ThemeColors | undefined) {
  if (!colors || typeof document === 'undefined') return;

  const root = document.documentElement;

  Object.entries(colors).forEach(([key, value]) => {
    if (!value) return;

    const cssVar = colorVariableMap[key as keyof ThemeColors];
    if (!cssVar) return;

    try {
      // For hex colors, we can set them directly
      // The CSS variables in globals.css use OKLCH, but hex colors work too
      if (isHexColor(value)) {
        root.style.setProperty(cssVar, value);
      } else {
        // For other formats (rgb, hsl, oklch), set directly
        root.style.setProperty(cssVar, value);
      }
    } catch (error) {
      console.warn(`Failed to set CSS variable ${cssVar}:`, error);
    }
  });
}

/**
 * Apply theme fonts to the document
 */
function applyThemeFonts(fonts: { body?: string; heading?: string } | undefined) {
  if (!fonts || typeof document === 'undefined') return;

  const root = document.documentElement;

  if (fonts.body) {
    root.style.setProperty('--font-body', fonts.body);
  }

  if (fonts.heading) {
    root.style.setProperty('--font-heading', fonts.heading);
  }
}

/**
 * Props for TenantThemeProvider
 */
interface TenantThemeProviderProps {
  children: ReactNode;
  /** Optional initial config (useful for SSR or testing) */
  initialConfig?: ThemeConfig;
}

/**
 * Provider component that fetches and applies tenant theme configuration.
 *
 * This provider:
 * 1. Fetches the tenant's theme_config from the API
 * 2. Merges it with default configuration
 * 3. Injects CSS custom properties into the document root
 * 4. Provides theme values to child components via context
 *
 * @example
 * ```tsx
 * // In your app layout
 * export default function DashboardLayout({ children }: { children: ReactNode }) {
 *   return (
 *     <TenantThemeProvider>
 *       {children}
 *     </TenantThemeProvider>
 *   );
 * }
 *
 * // In child components
 * function MyComponent() {
 *   const { themeConfig } = useTenantTheme();
 *   const sidebarItems = themeConfig.layout?.sidebar || [];
 *   // ...
 * }
 * ```
 */
export function TenantThemeProvider({
  children,
  initialConfig,
}: TenantThemeProviderProps) {
  const {
    themeConfig: fetchedConfig,
    isLoading,
    logoUrl,
    faviconUrl,
    refetch,
  } = useTenantThemeHook();

  // Use initial config if provided, otherwise use fetched config
  const themeConfig = useMemo(() => {
    if (initialConfig) {
      return {
        ...DEFAULT_THEME_CONFIG,
        ...initialConfig,
        theme: {
          ...DEFAULT_THEME_CONFIG.theme,
          ...initialConfig.theme,
          colors: {
            ...DEFAULT_THEME_CONFIG.theme?.colors,
            ...initialConfig.theme?.colors,
          },
          fonts: {
            ...DEFAULT_THEME_CONFIG.theme?.fonts,
            ...initialConfig.theme?.fonts,
          },
        },
        layout: {
          ...DEFAULT_THEME_CONFIG.layout,
          ...initialConfig.layout,
        },
      };
    }
    return fetchedConfig;
  }, [initialConfig, fetchedConfig]);

  // Apply CSS variables when theme changes
  useEffect(() => {
    if (isLoading && !initialConfig) return;

    // Apply colors
    applyThemeColors(themeConfig.theme?.colors);

    // Apply fonts
    applyThemeFonts(themeConfig.theme?.fonts);

    // Update favicon if provided
    if (faviconUrl && typeof document !== 'undefined') {
      const existingFavicon = document.querySelector('link[rel="icon"]');
      if (existingFavicon) {
        existingFavicon.setAttribute('href', faviconUrl);
      } else {
        const link = document.createElement('link');
        link.rel = 'icon';
        link.href = faviconUrl;
        document.head.appendChild(link);
      }
    }
  }, [themeConfig, isLoading, faviconUrl, initialConfig]);

  const contextValue = useMemo<TenantThemeContextValue>(() => ({
    themeConfig,
    isLoading: initialConfig ? false : isLoading,
    logoUrl,
    faviconUrl,
    refetch,
  }), [themeConfig, isLoading, logoUrl, faviconUrl, refetch, initialConfig]);

  return (
    <TenantThemeContext.Provider value={contextValue}>
      {children}
    </TenantThemeContext.Provider>
  );
}

/**
 * Hook to access the tenant theme context.
 *
 * Must be used within a TenantThemeProvider.
 *
 * @returns The tenant theme context value
 * @throws Error if used outside of TenantThemeProvider
 *
 * @example
 * ```tsx
 * function SidebarComponent() {
 *   const { themeConfig, isLoading } = useTenantTheme();
 *
 *   if (isLoading) return <Skeleton />;
 *
 *   return (
 *     <nav>
 *       {themeConfig.layout?.sidebar?.map(item => (
 *         <Link key={item.link} href={item.link}>
 *           <DynamicIcon name={item.icon} />
 *           {item.label}
 *         </Link>
 *       ))}
 *     </nav>
 *   );
 * }
 * ```
 */
export function useTenantTheme(): TenantThemeContextValue {
  const context = useContext(TenantThemeContext);

  if (context === undefined) {
    throw new Error('useTenantTheme must be used within a TenantThemeProvider');
  }

  return context;
}

export default TenantThemeProvider;
