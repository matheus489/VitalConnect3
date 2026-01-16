'use client';

import { useMemo, useEffect, useState } from 'react';
import { Marker, Popup } from 'react-leaflet';
import type { DivIcon } from 'leaflet';
import type { MapHospital, UrgencyLevel } from '@/types';
import { getUrgencyColor } from '@/lib/map-utils';

interface HospitalMarkerProps {
  /**
   * Dados do hospital
   */
  hospital: MapHospital;
  /**
   * Callback executado ao clicar no marcador
   */
  onClick: (hospital: MapHospital) => void;
}

/**
 * Retorna as classes CSS para a animacao pulsante baseada na urgencia
 */
function getPulseClasses(urgency: UrgencyLevel): string {
  if (urgency === 'red') {
    return 'animate-pulse';
  }
  return '';
}

/**
 * Cria o HTML do icone customizado do marcador
 */
function createMarkerIconHtml(hospital: MapHospital): string {
  const color = getUrgencyColor(hospital.urgencia_maxima);
  const pulseClass = getPulseClasses(hospital.urgencia_maxima);
  const showBadge = hospital.ocorrencias_count > 1;

  // SVG marker pin with urgency color
  const svgMarker = `
    <svg width="32" height="40" viewBox="0 0 32 40" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M16 0C7.16344 0 0 7.16344 0 16C0 28 16 40 16 40C16 40 32 28 32 16C32 7.16344 24.8366 0 16 0Z" fill="${color}"/>
      <circle cx="16" cy="16" r="8" fill="white"/>
      <path d="M14 12H18V16H22V20H18V24H14V20H10V16H14V12Z" fill="${color}"/>
    </svg>
  `;

  // Badge for multiple occurrences
  const badge = showBadge
    ? `<span class="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-red-600 text-[10px] font-bold text-white shadow-sm">${hospital.ocorrencias_count}</span>`
    : '';

  return `
    <div class="relative ${pulseClass}" style="width: 32px; height: 40px;">
      ${svgMarker}
      ${badge}
    </div>
  `;
}

/**
 * Marcador de hospital no mapa
 *
 * Exibe um pino customizado com cor baseada na urgencia maxima:
 * - Cinza: sem ocorrencias ativas
 * - Verde: tempo restante > 4 horas
 * - Amarelo: tempo restante entre 2 e 4 horas
 * - Vermelho: tempo restante < 2 horas (com animacao pulsante)
 *
 * Inclui badge numerico quando ha multiplas ocorrencias.
 */
export function HospitalMarker({ hospital, onClick }: HospitalMarkerProps) {
  const [leafletIcon, setLeafletIcon] = useState<DivIcon | null>(null);

  // Create custom DivIcon - must be done client-side
  useEffect(() => {
    // Dynamic import of leaflet to avoid SSR issues
    import('leaflet').then((L) => {
      const icon = L.divIcon({
        html: createMarkerIconHtml(hospital),
        className: 'custom-marker-icon',
        iconSize: [32, 40],
        iconAnchor: [16, 40],
        popupAnchor: [0, -40],
      });
      setLeafletIcon(icon);
    });
  }, [hospital]);

  // Wait for icon to be created
  if (!leafletIcon) {
    return null;
  }

  return (
    <Marker
      position={[hospital.latitude, hospital.longitude]}
      icon={leafletIcon}
      eventHandlers={{
        click: () => onClick(hospital),
      }}
    >
      <Popup>
        <div className="p-1">
          <strong className="text-sm">{hospital.nome}</strong>
          <p className="text-xs text-muted-foreground">
            {hospital.ocorrencias_count} ocorrencia(s) ativa(s)
          </p>
        </div>
      </Popup>
    </Marker>
  );
}
