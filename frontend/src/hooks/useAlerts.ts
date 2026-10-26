import { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { api } from '../services/api';
import type {
  Alert,
  AlertSeverity,
  AlertStatus,
  AlertsFilters,
} from '../types';

// ============================================
// Tipos
// ============================================

type SeverityFilter = AlertSeverity | 'all';
type StatusFilter = AlertStatus | 'all';

interface AlertCounts {
  total: number;
  critical: number;
  warning: number;
  info: number;
  active: number;
  resolved: number;
  acknowledged: number;
}

interface UseAlertsOptions {
  /** Refrescar cada X ms (default: 30000) */
  refetchInterval?: number;

  /** Límite inicial de alertas a mostrar */
  limit?: number;

  /** Habilitar la query */
  enabled?: boolean;
}

interface UseAlertsReturn {
  // Datos
  alerts: Alert[];
  allAlerts: Alert[];

  // Estados de React Query
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  isFetching: boolean;
  refetch: () => void;

  // Filtros
  severityFilter: SeverityFilter;
  statusFilter: StatusFilter;
  setSeverityFilter: (filter: SeverityFilter) => void;
  setStatusFilter: (filter: StatusFilter) => void;
  clearFilters: () => void;
  hasActiveFilters: boolean;

  // Estadísticas
  counts: AlertCounts;

  // Acciones
  resolveAlert: (id: string) => Promise<void>;
  acknowledgeAlert: (id: string) => Promise<void>;
  isUpdating: boolean;
}

// ============================================
// Hook principal
// ============================================

/**
 * Hook para gestionar alertas con filtros, contadores y mutaciones.
 *
 * @example
 * ```tsx
 * const {
 *   alerts,
 *   counts,
 *   severityFilter,
 *   setSeverityFilter,
 *   resolveAlert,
 * } = useAlerts();
 * ```
 */
export function useAlerts(options: UseAlertsOptions = {}): UseAlertsReturn {
  const {
    refetchInterval = 30_000,
    limit,
    enabled = true,
  } = options;

  const queryClient = useQueryClient();

  // ---------- Filtros (estado local) ----------
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');

  // ---------- Query principal (sin filtros, para stats) ----------
  const allAlertsQuery = useQuery({
    queryKey: ['alerts', 'all'],
    queryFn: () => api.getAlerts(),
    refetchInterval,
    enabled,
    staleTime: 10_000,
  });

  // ---------- Query filtrada (para la lista) ----------
  const filters: AlertsFilters = {
    severity: severityFilter === 'all' ? undefined : severityFilter,
    status: statusFilter === 'all' ? undefined : statusFilter,
  };

  const filteredQuery = useQuery({
    queryKey: ['alerts', 'filtered', filters],
    queryFn: () => api.getAlerts(filters),
    refetchInterval,
    enabled,
    staleTime: 10_000,
  });

  // ---------- Lista final de alertas ----------
  const allAlerts = useMemo(() => {
    const data = allAlertsQuery.data?.data ?? [];
    // Ordenar por severidad (critical > warning > info) y luego por fecha
    return [...data].sort((a, b) => {
      const severityOrder: Record<AlertSeverity, number> = {
        critical: 0,
        warning: 1,
        info: 2,
      };
      const severityDiff = severityOrder[a.severity] - severityOrder[b.severity];
      if (severityDiff !== 0) return severityDiff;
      // Mismo severity → ordenar por fecha descendente
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });
  }, [allAlertsQuery.data]);

  const alerts = useMemo(() => {
    const data = filteredQuery.data?.data ?? [];
    const limited = limit ? data.slice(0, limit) : data;
    // Mismo ordenamiento que allAlerts
    return [...limited].sort((a, b) => {
      const severityOrder: Record<AlertSeverity, number> = {
        critical: 0,
        warning: 1,
        info: 2,
      };
      const severityDiff = severityOrder[a.severity] - severityOrder[b.severity];
      if (severityDiff !== 0) return severityDiff;
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });
  }, [filteredQuery.data, limit]);

  // ---------- Contadores ----------
  const counts = useMemo<AlertCounts>(() => {
    const initial: AlertCounts = {
      total: 0,
      critical: 0,
      warning: 0,
      info: 0,
      active: 0,
      resolved: 0,
      acknowledged: 0,
    };

    return allAlerts.reduce((acc, alert) => {
      acc.total += 1;
      acc[alert.severity] += 1;
      acc[alert.status] += 1;
      return acc;
    }, initial);
  }, [allAlerts]);

  // ---------- Mutación: resolver alerta ----------
  const resolveMutation = useMutation({
    mutationFn: (id: string) => api.updateAlert(id, { status: 'resolved' }),

    onMutate: async (id) => {
      // Cancelar queries en curso
      await queryClient.cancelQueries({ queryKey: ['alerts'] });

      // Snapshot para rollback
      const previousAll = queryClient.getQueryData(['alerts', 'all']);

      // Actualización optimista
      queryClient.setQueryData(
        ['alerts', 'all'],
        (old: { data: Alert[] } | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((a) =>
              a.id === id ? { ...a, status: 'resolved' as AlertStatus } : a
            ),
          };
        }
      );

      return { previousAll };
    },

    onError: (_err, _id, context) => {
      // Rollback en caso de error
      if (context?.previousAll) {
        queryClient.setQueryData(['alerts', 'all'], context.previousAll);
      }
    },

    onSettled: () => {
      // Refrescar datos reales
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
    },
  });

  // ---------- Mutación: acknowledge alerta ----------
  const acknowledgeMutation = useMutation({
    mutationFn: (id: string) => api.updateAlert(id, { status: 'acknowledged' }),

    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ['alerts'] });

      const previousAll = queryClient.getQueryData(['alerts', 'all']);

      queryClient.setQueryData(
        ['alerts', 'all'],
        (old: { data: Alert[] } | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((a) =>
              a.id === id ? { ...a, status: 'acknowledged' as AlertStatus } : a
            ),
          };
        }
      );

      return { previousAll };
    },

    onError: (_err, _id, context) => {
      if (context?.previousAll) {
        queryClient.setQueryData(['alerts', 'all'], context.previousAll);
      }
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
    },
  });

  // ---------- Acciones ----------
  const resolveAlert = useCallback(
    async (id: string) => {
      await resolveMutation.mutateAsync(id);
    },
    [resolveMutation]
  );

  const acknowledgeAlert = useCallback(
    async (id: string) => {
      await acknowledgeMutation.mutateAsync(id);
    },
    [acknowledgeMutation]
  );

  // ---------- Limpiar filtros ----------
  const clearFilters = useCallback(() => {
    setSeverityFilter('all');
    setStatusFilter('all');
  }, []);

  const hasActiveFilters = severityFilter !== 'all' || statusFilter !== 'all';

  // ---------- Refetch combinado ----------
  const refetch = useCallback(() => {
    allAlertsQuery.refetch();
    filteredQuery.refetch();
  }, [allAlertsQuery, filteredQuery]);

  // ---------- Return ----------
  return {
    alerts,
    allAlerts,

    isLoading: filteredQuery.isLoading,
    isError: filteredQuery.isError,
    error: filteredQuery.error as Error | null,
    isFetching: filteredQuery.isFetching,
    refetch,

    severityFilter,
    statusFilter,
    setSeverityFilter,
    setStatusFilter,
    clearFilters,
    hasActiveFilters,

    counts,

    resolveAlert,
    acknowledgeAlert,
    isUpdating: resolveMutation.isPending || acknowledgeMutation.isPending,
  };
}