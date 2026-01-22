import type { NominatimResult } from '@/types';

const NOMINATIM_BASE_URL = 'https://nominatim.openstreetmap.org';

/**
 * Search for addresses using Nominatim (OpenStreetMap) geocoding service
 * Nominatim is free and does not require an API key
 *
 * @param query - The address search query
 * @param limit - Maximum number of results (default: 5)
 * @returns Array of geocoding results with coordinates
 *
 * @example
 * ```ts
 * const results = await searchAddress('Hospital das Clinicas, Goiania');
 * console.log(results[0].lat, results[0].lon, results[0].display_name);
 * ```
 */
export async function searchAddress(
  query: string,
  limit: number = 5
): Promise<NominatimResult[]> {
  if (!query || query.trim().length < 3) {
    return [];
  }

  const params = new URLSearchParams({
    format: 'json',
    q: query,
    limit: limit.toString(),
    addressdetails: '1',
    countrycodes: 'br',
  });

  try {
    const response = await fetch(`${NOMINATIM_BASE_URL}/search?${params}`, {
      headers: {
        'User-Agent': 'SIDOT/1.0 (contact@sidot.gov.br)',
        'Accept-Language': 'pt-BR,pt;q=0.9',
      },
    });

    if (!response.ok) {
      throw new Error(`Nominatim API error: ${response.status}`);
    }

    const data: NominatimResult[] = await response.json();
    return data;
  } catch (error) {
    console.error('Geocoding search failed:', error);
    return [];
  }
}

/**
 * Parse coordinates from a Nominatim result
 * Nominatim returns coordinates as strings, this helper converts them to numbers
 *
 * @param result - A Nominatim search result
 * @returns Object with numeric latitude and longitude
 */
export function parseCoordinates(result: NominatimResult): {
  latitude: number;
  longitude: number;
} {
  return {
    latitude: parseFloat(result.lat),
    longitude: parseFloat(result.lon),
  };
}

/**
 * Reverse geocode coordinates to get an address using Nominatim
 * This is used when the user clicks on the map to get the address
 *
 * @param lat - Latitude
 * @param lon - Longitude
 * @returns The address string or null if not found
 *
 * @example
 * ```ts
 * const address = await reverseGeocode(-16.6869, -49.2648);
 * console.log(address); // "Rua 1, Setor Central, Goiania, GO, 74000-000, Brasil"
 * ```
 */
export async function reverseGeocode(
  lat: number,
  lon: number
): Promise<string | null> {
  const params = new URLSearchParams({
    format: 'json',
    lat: lat.toString(),
    lon: lon.toString(),
    addressdetails: '1',
  });

  try {
    const response = await fetch(`${NOMINATIM_BASE_URL}/reverse?${params}`, {
      headers: {
        'User-Agent': 'SIDOT/1.0 (contact@sidot.gov.br)',
        'Accept-Language': 'pt-BR,pt;q=0.9',
      },
    });

    if (!response.ok) {
      throw new Error(`Nominatim API error: ${response.status}`);
    }

    const data = await response.json();
    return data.display_name || null;
  } catch (error) {
    console.error('Reverse geocoding failed:', error);
    return null;
  }
}
