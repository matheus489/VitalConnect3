'use client';

import { useEffect, useMemo, useRef } from 'react';
import { MapContainer as LeafletMapContainer, TileLayer, useMap } from 'react-leaflet';
import { Skeleton } from '@/components/ui/skeleton';
import type { MapHospital } from '@/types';
import { GOIAS_BOUNDS } from '@/lib/map-utils';
import { HospitalMarker } from './HospitalMarker';

// Component to fit bounds when hospitals change
function MapBoundsController({ hospitals }: { hospitals: MapHospital[] }) {
  const map = useMap();

  useEffect(() => {
    if (hospitals.length > 0) {
      // Import leaflet dynamically
      import('leaflet').then((L) => {
        const bounds = L.latLngBounds(
          hospitals.map((h) => [h.latitude, h.longitude] as [number, number])
        );
        if (bounds.isValid()) {
          map.fitBounds(bounds, { padding: [50, 50], maxZoom: 12 });
        }
      });
    } else {
      // Default to Goias bounds when no hospitals
      map.setView(
        [GOIAS_BOUNDS.center.lat, GOIAS_BOUNDS.center.lng],
        GOIAS_BOUNDS.defaultZoom
      );
    }
  }, [hospitals, map]);

  return null;
}

interface MapContainerProps {
  /**
   * Lista de hospitais para exibir no mapa
   */
  hospitals: MapHospital[];
  /**
   * Callback executado ao clicar em um hospital
   */
  onHospitalClick: (hospital: MapHospital) => void;
  /**
   * Se true, exibe skeleton de loading
   */
  isLoading?: boolean;
}

/**
 * Container do mapa Leaflet com hospitais marcados
 *
 * Usa OpenStreetMap como tile layer (custo zero).
 * Inicializa com zoom no estado de Goias e ajusta automaticamente
 * para enquadrar todos os hospitais quando os dados carregam.
 */
export function MapContainer({ hospitals, onHospitalClick, isLoading = false }: MapContainerProps) {
  // Memoize the hospital markers to avoid unnecessary re-renders
  const markers = useMemo(() => {
    return hospitals.map((hospital) => (
      <HospitalMarker
        key={hospital.id}
        hospital={hospital}
        onClick={onHospitalClick}
      />
    ));
  }, [hospitals, onHospitalClick]);

  if (isLoading) {
    return (
      <div
        data-testid="map-loading"
        className="w-full h-full min-h-[400px] bg-muted rounded-lg flex items-center justify-center relative"
      >
        <Skeleton className="w-full h-full absolute inset-0 rounded-lg" />
        <div className="relative z-10 flex flex-col items-center gap-2">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">Carregando mapa...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full min-h-[400px] rounded-lg overflow-hidden border">
      <LeafletMapContainer
        center={[GOIAS_BOUNDS.center.lat, GOIAS_BOUNDS.center.lng]}
        zoom={GOIAS_BOUNDS.defaultZoom}
        minZoom={GOIAS_BOUNDS.minZoom}
        maxZoom={GOIAS_BOUNDS.maxZoom}
        className="w-full h-full min-h-[400px]"
        scrollWheelZoom={true}
        style={{ height: '100%', width: '100%' }}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapBoundsController hospitals={hospitals} />
        {markers}
      </LeafletMapContainer>
    </div>
  );
}
