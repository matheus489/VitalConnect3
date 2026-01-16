import type { UrgencyLevel } from '@/types';

/**
 * Calcula o nivel de urgencia baseado no tempo restante em minutos
 *
 * Logica de classificacao:
 * - Verde (green): tempo restante > 4 horas (> 240 minutos)
 * - Amarelo (yellow): tempo restante entre 2 e 4 horas (120-240 minutos)
 * - Vermelho (red): tempo restante < 2 horas (< 120 minutos)
 *
 * @param tempoRestanteMinutos - Tempo restante em minutos
 * @returns Nivel de urgencia ('green' | 'yellow' | 'red')
 *
 * @example
 * ```ts
 * calculateUrgencyLevel(300) // 'green' (5 horas)
 * calculateUrgencyLevel(180) // 'yellow' (3 horas)
 * calculateUrgencyLevel(60)  // 'red' (1 hora)
 * ```
 */
export function calculateUrgencyLevel(tempoRestanteMinutos: number): UrgencyLevel {
  if (tempoRestanteMinutos <= 0) {
    return 'red';
  }
  if (tempoRestanteMinutos < 120) {
    // < 2 horas
    return 'red';
  }
  if (tempoRestanteMinutos < 240) {
    // 2-4 horas
    return 'yellow';
  }
  return 'green'; // > 4 horas
}

/**
 * Retorna a cor CSS/hex para o nivel de urgencia
 *
 * Cores utilizadas (Tailwind CSS palette):
 * - green: #22c55e (green-500)
 * - yellow: #eab308 (yellow-500)
 * - red: #ef4444 (red-500)
 * - none: #6b7280 (gray-500)
 *
 * @param level - Nivel de urgencia
 * @returns Cor em formato hexadecimal
 *
 * @example
 * ```ts
 * getUrgencyColor('red') // '#ef4444'
 * ```
 */
export function getUrgencyColor(level: UrgencyLevel): string {
  switch (level) {
    case 'green':
      return '#22c55e'; // green-500
    case 'yellow':
      return '#eab308'; // yellow-500
    case 'red':
      return '#ef4444'; // red-500
    case 'none':
      return '#6b7280'; // gray-500
    default:
      return '#6b7280';
  }
}

/**
 * Retorna o texto descritivo em portugues para o nivel de urgencia
 *
 * Labels:
 * - green: "Normal"
 * - yellow: "Atencao"
 * - red: "Critico"
 * - none: "Sem ocorrencias"
 *
 * @param level - Nivel de urgencia
 * @returns Label em portugues
 *
 * @example
 * ```ts
 * getUrgencyLabel('red') // 'Critico'
 * ```
 */
export function getUrgencyLabel(level: UrgencyLevel): string {
  switch (level) {
    case 'green':
      return 'Normal';
    case 'yellow':
      return 'Atencao';
    case 'red':
      return 'Critico';
    case 'none':
      return 'Sem ocorrencias';
    default:
      return 'Desconhecido';
  }
}

/**
 * Formata o tempo restante em minutos para uma string legivel
 *
 * Formato de saida:
 * - Para >= 60 minutos: "Xh Ym" (ex: "2h 30m")
 * - Para < 60 minutos: "Xm" (ex: "45m")
 * - Para 0 minutos: "0m"
 * - Para valores negativos: "Expirado"
 *
 * @param minutos - Tempo restante em minutos
 * @returns String formatada do tempo restante
 *
 * @example
 * ```ts
 * formatTimeRemaining(150) // '2h 30m'
 * formatTimeRemaining(45)  // '45m'
 * formatTimeRemaining(-5)  // 'Expirado'
 * ```
 */
export function formatTimeRemaining(minutos: number): string {
  if (minutos < 0) {
    return 'Expirado';
  }

  if (minutos === 0) {
    return '0m';
  }

  const horas = Math.floor(minutos / 60);
  const mins = minutos % 60;

  if (horas > 0) {
    return `${horas}h ${mins}m`;
  }

  return `${mins}m`;
}

/**
 * Retorna a classe CSS Tailwind para o background do nivel de urgencia
 *
 * @param level - Nivel de urgencia
 * @returns Classes CSS Tailwind para background e texto
 *
 * @example
 * ```tsx
 * <div className={getUrgencyBadgeClasses('red')}>Critico</div>
 * ```
 */
export function getUrgencyBadgeClasses(level: UrgencyLevel): string {
  switch (level) {
    case 'green':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
    case 'yellow':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
    case 'red':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
    case 'none':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
  }
}

/**
 * Constantes para os bounds do mapa (estado de Goias)
 */
export const GOIAS_BOUNDS = {
  // Centro aproximado de Goiania
  center: {
    lat: -16.6799,
    lng: -49.255,
  },
  // Bounds do estado de Goias para zoom inicial
  bounds: {
    north: -12.39,
    south: -19.5,
    east: -45.9,
    west: -53.25,
  },
  // Zoom padrao para o mapa
  defaultZoom: 7,
  // Zoom maximo permitido
  maxZoom: 18,
  // Zoom minimo permitido
  minZoom: 5,
};
