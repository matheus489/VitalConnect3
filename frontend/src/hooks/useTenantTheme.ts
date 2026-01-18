'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useAuth } from '@/hooks/useAuth';
import { DEFAULT_THEME_CONFIG, type ThemeConfig } from '@/types/theme';

/**
 * Response from the tenant theme API
 */
interface TenantThemeResponse {
  theme_config: ThemeConfig | null;
  logo_url?: string;
  favicon_url?: string;
}

/**
 * Fetch the current tenant's theme configuration
 */
async function fetchTenantTheme(): Promise<TenantThemeResponse> {
  try {
    const { data } = await api.get<TenantThemeResponse>('/tenants/current/theme');
    return data;
  } catch (error) {
    // If endpoint doesn't exist or fails, return empty response
    // The hook will use DEFAULT_THEME_CONFIG as fallback
    console.warn('Failed to fetch tenant theme, using defaults:', error);
    return { theme_config: null };
  }
}

/**
 * Result type for useTenantTheme hook
 */
export interface UseTenantThemeResult {
  /** The merged theme configuration (API data + defaults) */
  themeConfig: ThemeConfig;
  /** Raw theme config from API (may be null) */
  rawThemeConfig: ThemeConfig | null;
  /** Logo URL if configured */
  logoUrl?: string;
  /** Favicon URL if configured */
  faviconUrl?: string;
  /** Whether the theme is still loading */
  isLoading: boolean;
  /** Whether there was an error loading the theme */
  isError: boolean;
  /** Error object if any */
  error: Error | null;
  /** Refetch the theme configuration */
  refetch: () => void;
}

/**
 * Deep merge two objects, with source overriding target
 */
function deepMerge<T extends Record<string, unknown>>(
  target: T,
  source: Partial<T> | null | undefined
): T {
  if (!source) return target;

  const result = { ...target };

  for (const key in source) {
    if (Object.prototype.hasOwnProperty.call(source, key)) {
      const sourceValue = source[key];
      const targetValue = target[key];

      if (
        sourceValue &&
        typeof sourceValue === 'object' &&
        !Array.isArray(sourceValue) &&
        targetValue &&
        typeof targetValue === 'object' &&
        !Array.isArray(targetValue)
      ) {
        // Recursively merge nested objects
        (result as Record<string, unknown>)[key] = deepMerge(
          targetValue as Record<string, unknown>,
          sourceValue as Record<string, unknown>
        );
      } else if (sourceValue !== undefined) {
        // Direct assignment for primitives and arrays
        (result as Record<string, unknown>)[key] = sourceValue;
      }
    }
  }

  return result;
}

/**
 * Hook to fetch and cache the current tenant's theme configuration.
 *
 * Features:
 * - Fetches theme_config from the API
 * - Caches the result using React Query
 * - Provides fallback to DEFAULT_THEME_CONFIG when no config exists
 * - Merges API config with defaults for complete configuration
 *
 * @returns Object containing theme configuration and loading states
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const { themeConfig, isLoading } = useTenantTheme();
 *
 *   if (isLoading) return <Loading />;
 *
 *   return (
 *     <div style={{ color: themeConfig.theme.colors.primary }}>
 *       {themeConfig.layout.sidebar.map(item => ...)}
 *     </div>
 *   );
 * }
 * ```
 */
export function useTenantTheme(): UseTenantThemeResult {
  const { isAuthenticated } = useAuth();

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery<TenantThemeResponse, Error>({
    queryKey: ['tenantTheme'],
    queryFn: fetchTenantTheme,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 30 * 60 * 1000, // 30 minutes
    retry: 1,
    refetchOnWindowFocus: false,
  });

  // Merge API config with defaults
  const mergedConfig = deepMerge(
    DEFAULT_THEME_CONFIG as unknown as Record<string, unknown>,
    data?.theme_config as unknown as Record<string, unknown>
  ) as ThemeConfig;

  return {
    themeConfig: mergedConfig,
    rawThemeConfig: data?.theme_config ?? null,
    logoUrl: data?.logo_url,
    faviconUrl: data?.favicon_url,
    isLoading,
    isError,
    error: error ?? null,
    refetch,
  };
}

export default useTenantTheme;
