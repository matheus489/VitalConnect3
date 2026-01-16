'use client';

import { useState, useCallback, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { AlertCircle, Map as MapIcon, RefreshCw } from 'lucide-react';
import { HospitalDrawer } from '@/components/map/HospitalDrawer';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useMapHospitals, useMapSSEHandler } from '@/hooks/useMap';
import { useSSE } from '@/hooks/useSSE';
import { useAuth } from '@/hooks/useAuth';
import type { MapHospital } from '@/types';

// Dynamically import MapContainer to avoid SSR issues with Leaflet
const MapContainer = dynamic(
  () => import('@/components/map/MapContainer').then((mod) => mod.MapContainer),
  {
    ssr: false,
    loading: () => (
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
    ),
  }
);

/**
 * Pagina do Dashboard Geografico
 *
 * Exibe um mapa interativo com:
 * - Hospitais como marcadores com cores de urgencia
 * - Drawer lateral com detalhes ao clicar em um hospital
 * - Atualizacoes em tempo real via SSE
 */
export default function MapPage() {
  const [selectedHospital, setSelectedHospital] = useState<MapHospital | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const { isAuthenticated } = useAuth();
  const { data, isLoading, error, refetch } = useMapHospitals();
  const handleMapUpdate = useMapSSEHandler();

  // Connect to SSE for real-time updates
  useSSE({
    onNotification: handleMapUpdate,
    enabled: isAuthenticated,
  });

  const handleHospitalClick = useCallback((hospital: MapHospital) => {
    setSelectedHospital(hospital);
    setDrawerOpen(true);
  }, []);

  const handleDrawerClose = useCallback(() => {
    setDrawerOpen(false);
  }, []);

  // Update selected hospital data when map data changes
  useEffect(() => {
    if (selectedHospital && data?.hospitals) {
      const updatedHospital = data.hospitals.find((h) => h.id === selectedHospital.id);
      if (updatedHospital) {
        setSelectedHospital(updatedHospital);
      }
    }
  }, [data, selectedHospital]);

  // Error state
  if (error) {
    return (
      <div className="space-y-6">
        {/* Page Title */}
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <MapIcon className="h-6 w-6" />
            Mapa
          </h1>
          <p className="text-muted-foreground">
            Visualizacao geografica dos hospitais e ocorrencias ativas
          </p>
        </div>

        {/* Error Message */}
        <div className="flex flex-col items-center justify-center rounded-lg border border-destructive/30 bg-destructive/5 p-8">
          <AlertCircle className="h-12 w-12 text-destructive mb-4" />
          <h2 className="text-lg font-semibold mb-2">Erro ao carregar mapa</h2>
          <p className="text-sm text-muted-foreground text-center mb-4 max-w-md">
            Nao foi possivel carregar os dados do mapa. Verifique sua conexao e tente novamente.
          </p>
          <Button onClick={() => refetch()} variant="outline">
            <RefreshCw className="mr-2 h-4 w-4" />
            Tentar novamente
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 h-[calc(100vh-8rem)]">
      {/* Page Title */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <MapIcon className="h-6 w-6" />
            Mapa
          </h1>
          <p className="text-muted-foreground">
            Visualizacao geografica dos hospitais e ocorrencias ativas
          </p>
        </div>

        {/* Stats summary */}
        {data && (
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <span className="h-2.5 w-2.5 rounded-full bg-primary" />
              <span className="text-muted-foreground">
                {data.total} {data.total === 1 ? 'hospital' : 'hospitais'}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="h-2.5 w-2.5 rounded-full bg-red-500" />
              <span className="text-muted-foreground">
                {data.hospitals.filter((h) => h.urgencia_maxima === 'red').length} criticos
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Map Container - fills remaining space */}
      <div className="flex-1 h-[calc(100%-5rem)]">
        <MapContainer
          hospitals={data?.hospitals || []}
          onHospitalClick={handleHospitalClick}
          isLoading={isLoading}
        />
      </div>

      {/* Hospital Details Drawer */}
      <HospitalDrawer
        hospital={selectedHospital}
        open={drawerOpen}
        onClose={handleDrawerClose}
      />
    </div>
  );
}
