import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { api } from '../services/api';
import type { Metric, MetricStats } from '../types';

// ============================================
// Hook principal: obtener métricas
// ============================================

interface UseMetricsOptions {
  limit?: number;
  refetchInterval?: number;
}

/**
 * Hook para obtener métricas de la API con React Query.
 *
 * @param options - Opciones de configuración
 * @returns Estado de la query (data, isLoading, error, refetch)
 *
 * @example
 * ```tsx
 * const { data, isLoading, error } = useMetrics({ limit: 50 });
 * ```
 */
export function useMetrics(options: UseMetricsOptions = {}) {
  const { limit = 50, refetchInterval = 10_000 } = options;

  return useQuery({
    queryKey: ['metrics', { limit }],
    queryFn: () => api.getMetrics({ limit }),
    refetchInterval,
    staleTime: 5_000,
  });
}

// ============================================
// Hook: estadísticas de una métrica
// ============================================

/**
 * Calcula estadísticas de una serie de valores.
 *
 * @param metrics - Lista de métricas
 * @param field - Campo a analizar (cpu_usage, memory_usage, etc.)
 * @returns Estadísticas calculadas
 */
export function useMetricStats(
  metrics: Metric[] | undefined,
  field: keyof Pick<Metric, 'cpu_usage' | 'memory_usage' | 'disk_usage'>
): MetricStats {
  return useMemo(() => {
    const empty: MetricStats = {
      current: 0,
      average: 0,
      max: 0,
      min: 0,
      trend: 'stable',
      trendValue: 0,
    };

    if (!metrics || metrics.length === 0) {
      return empty;
    }

    // Extraer valores del campo
    const values = metrics.map((m) => m[field]);
    const current = values[0] ?? 0;

    // Calcular estadísticas
    const sum = values.reduce((acc, val) => acc + val, 0);
    const average = sum / values.length;
    const max = Math.max(...values);
    const min = Math.min(...values);

    // Calcular tendencia (últimos 5 vs anteriores 5)
    const recentCount = Math.min(5, Math.floor(values.length / 2));
    const recent = values.slice(0, recentCount);
    const previous = values.slice(recentCount, recentCount * 2);

    const recentAvg = recent.length > 0
      ? recent.reduce((a, b) => a + b, 0) / recent.length
      : current;
    const previousAvg = previous.length > 0
      ? previous.reduce((a, b) => a + b, 0) / previous.length
      : recentAvg;

    const trendValue = recentAvg - previousAvg;
    const threshold = 2; // 2% de diferencia para considerar tendencia

    let trend: 'up' | 'down' | 'stable' = 'stable';
    if (trendValue > threshold) trend = 'up';
    else if (trendValue < -threshold) trend = 'down';

    return {
      current: round(current),
      average: round(average),
      max: round(max),
      min: round(min),
      trend,
      trendValue: round(Math.abs(trendValue)),
    };
  }, [metrics, field]);
}

// ============================================
// Hook: últimas N métricas
// ============================================

/**
 * Devuelve las últimas N métricas ordenadas por timestamp descendente.
 *
 * @param metrics - Lista completa
 * @param count - Cuántas devolver
 */
export function useLatestMetrics(metrics: Metric[] | undefined, count: number = 10): Metric[] {
  return useMemo(() => {
    if (!metrics) return [];
    return [...metrics]
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, count);
  }, [metrics, count]);
}

// ============================================
// Helpers
// ============================================

/**
 * Redondea a 2 decimales
 */
function round(value: number): number {
  return Math.round(value * 100) / 100;
}

/**
 * Formatea un timestamp Unix a hora local
 */
export function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp * 1000);
  return date.toLocaleTimeString('es-ES', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}