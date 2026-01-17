'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { MapContainer as LeafletMapContainer, TileLayer, Marker, useMap, useMapEvents } from 'react-leaflet';
import { GOIAS_BOUNDS } from '@/lib/map-utils';
import type { LatLng, LeafletMouseEvent, Icon } from 'leaflet';

interface LocationPickerMapProps {
  /**
   * Initial position for the map center and marker
   * If not provided, defaults to GOIAS_BOUNDS center
   */
  initialPosition?: {
    lat: number;
    lng: number;
  };
  /**
   * Current marker position (controlled mode)
   * When provided, the marker will be placed at this position
   */
  markerPosition?: {
    lat: number;
    lng: number;
  } | null;
  /**
   * Callback fired when location changes (click or drag)
   */
  onLocationChange: (lat: number, lng: number) => void;
  /**
   * Height of the map container (default: '250px')
   */
  height?: string;
  /**
   * Whether the map is disabled/read-only
   */
  disabled?: boolean;
}

/**
 * Component to handle map events and marker placement
 */
function MapEventHandler({
  onLocationChange,
  disabled,
}: {
  onLocationChange: (lat: number, lng: number) => void;
  disabled?: boolean;
}) {
  useMapEvents({
    click: (e: LeafletMouseEvent) => {
      if (disabled) return;
      onLocationChange(e.latlng.lat, e.latlng.lng);
    },
  });

  return null;
}

/**
 * Component to fly to a new position when marker position changes
 */
function MapCenterController({
  position,
}: {
  position: { lat: number; lng: number } | null;
}) {
  const map = useMap();
  const previousPosition = useRef<{ lat: number; lng: number } | null>(null);

  useEffect(() => {
    if (position && position.lat && position.lng) {
      const hasChanged =
        !previousPosition.current ||
        previousPosition.current.lat !== position.lat ||
        previousPosition.current.lng !== position.lng;

      if (hasChanged) {
        map.flyTo([position.lat, position.lng], 15, {
          duration: 0.5,
        });
        previousPosition.current = position;
      }
    }
  }, [map, position]);

  return null;
}

/**
 * Create marker icon HTML (blue pin)
 */
function createMarkerIconHtml(): string {
  return `
    <svg width="32" height="40" viewBox="0 0 32 40" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M16 0C7.16344 0 0 7.16344 0 16C0 28 16 40 16 40C16 40 32 28 32 16C32 7.16344 24.8366 0 16 0Z" fill="#2563eb"/>
      <circle cx="16" cy="16" r="8" fill="white"/>
      <circle cx="16" cy="16" r="4" fill="#2563eb"/>
    </svg>
  `;
}

/**
 * Draggable marker component
 */
function DraggableMarker({
  position,
  onDragEnd,
  disabled,
}: {
  position: { lat: number; lng: number };
  onDragEnd: (lat: number, lng: number) => void;
  disabled?: boolean;
}) {
  const markerRef = useRef<L.Marker>(null);
  const [icon, setIcon] = useState<L.DivIcon | null>(null);

  // Create custom icon on mount (client-side only)
  useEffect(() => {
    import('leaflet').then((L) => {
      const customIcon = L.divIcon({
        html: createMarkerIconHtml(),
        className: 'custom-location-marker',
        iconSize: [32, 40],
        iconAnchor: [16, 40],
      });
      setIcon(customIcon);
    });
  }, []);

  const handleDragEnd = useCallback(() => {
    const marker = markerRef.current;
    if (marker) {
      const latlng = marker.getLatLng();
      onDragEnd(latlng.lat, latlng.lng);
    }
  }, [onDragEnd]);

  // Don't render until icon is ready
  if (!icon) {
    return null;
  }

  return (
    <Marker
      ref={markerRef}
      position={[position.lat, position.lng]}
      draggable={!disabled}
      icon={icon}
      eventHandlers={{
        dragend: handleDragEnd,
      }}
    />
  );
}

/**
 * Interactive map component for selecting a location
 * Supports click-to-place and drag-to-adjust marker functionality
 *
 * Features:
 * - Click anywhere on map to place/move marker
 * - Drag marker for fine-tuning position
 * - Coordinates update in real-time
 * - Smooth animation when position changes programmatically
 *
 * @example
 * ```tsx
 * const [coords, setCoords] = useState<{ lat: number; lng: number } | null>(null);
 *
 * <LocationPickerMap
 *   markerPosition={coords}
 *   onLocationChange={(lat, lng) => setCoords({ lat, lng })}
 * />
 * ```
 */
export function LocationPickerMap({
  initialPosition,
  markerPosition,
  onLocationChange,
  height = '250px',
  disabled = false,
}: LocationPickerMapProps) {
  const [isReady, setIsReady] = useState(false);
  const center = initialPosition || GOIAS_BOUNDS.center;
  const zoom = initialPosition ? 15 : GOIAS_BOUNDS.defaultZoom;

  // Ensure we only render on client after a small delay
  // This allows the drawer animation to complete before mounting Leaflet
  useEffect(() => {
    const timer = setTimeout(() => {
      setIsReady(true);
    }, 100);
    return () => clearTimeout(timer);
  }, []);

  if (!isReady) {
    return (
      <div
        className="w-full rounded-lg overflow-hidden border bg-muted flex items-center justify-center"
        style={{ height }}
      >
        <div className="flex flex-col items-center gap-2">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">Carregando mapa...</p>
        </div>
      </div>
    );
  }

  return (
    <div
      className="w-full rounded-lg overflow-hidden border"
      style={{ height }}
    >
      <LeafletMapContainer
        center={[center.lat, center.lng]}
        zoom={zoom}
        minZoom={GOIAS_BOUNDS.minZoom}
        maxZoom={GOIAS_BOUNDS.maxZoom}
        className="w-full h-full"
        scrollWheelZoom={true}
        style={{ height: '100%', width: '100%' }}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapEventHandler onLocationChange={onLocationChange} disabled={disabled} />
        <MapCenterController position={markerPosition ?? null} />
        {markerPosition && (
          <DraggableMarker
            position={markerPosition}
            onDragEnd={onLocationChange}
            disabled={disabled}
          />
        )}
      </LeafletMapContainer>
    </div>
  );
}

export default LocationPickerMap;
