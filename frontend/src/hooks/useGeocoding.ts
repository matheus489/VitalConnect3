'use client';

import { useQuery } from '@tanstack/react-query';
import { useDebounce } from './useDebounce';
import { searchAddress } from '@/services/nominatim';
import type { NominatimResult } from '@/types';

interface UseGeocodingOptions {
  /**
   * Debounce delay in milliseconds (default: 300ms)
   */
  debounceMs?: number;
  /**
   * Whether the query is enabled (default: true)
   */
  enabled?: boolean;
}

interface UseGeocodingResult {
  /**
   * Array of address suggestions from Nominatim
   */
  suggestions: NominatimResult[];
  /**
   * Whether the geocoding request is in progress
   */
  isLoading: boolean;
  /**
   * Error object if the request failed
   */
  error: Error | null;
  /**
   * Whether there are any suggestions available
   */
  hasSuggestions: boolean;
}

/**
 * Hook for geocoding address searches using Nominatim
 * Automatically debounces the search query to prevent excessive API calls
 *
 * @param query - The address search query
 * @param options - Configuration options
 * @returns Geocoding results with loading and error states
 *
 * @example
 * ```tsx
 * const [address, setAddress] = useState('');
 * const { suggestions, isLoading } = useGeocoding(address);
 *
 * return (
 *   <input value={address} onChange={(e) => setAddress(e.target.value)} />
 *   {suggestions.map((s) => (
 *     <div key={s.place_id}>{s.display_name}</div>
 *   ))}
 * );
 * ```
 */
export function useGeocoding(
  query: string,
  options: UseGeocodingOptions = {}
): UseGeocodingResult {
  const { debounceMs = 300, enabled = true } = options;

  const debouncedQuery = useDebounce(query, debounceMs);

  const shouldFetch = enabled && debouncedQuery.trim().length >= 3;

  const { data, isLoading, error } = useQuery({
    queryKey: ['geocoding', debouncedQuery],
    queryFn: () => searchAddress(debouncedQuery),
    enabled: shouldFetch,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });

  return {
    suggestions: data ?? [],
    isLoading: shouldFetch && isLoading,
    error: error as Error | null,
    hasSuggestions: (data?.length ?? 0) > 0,
  };
}
