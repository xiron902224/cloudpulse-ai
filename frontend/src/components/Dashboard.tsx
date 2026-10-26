import {  useState } from 'react';

import { useMetrics, useMetricStats } from '../hooks/useMetrics';
import { useAlerts } from '../hooks/useAlerts';
import { useWebSocket } from '../hooks/useWebSocket';
import { getErrorMessage } from '../services/api';
import { StatsCard } from './StatsCard';
import { MetricsChart } from './MetricsChart';
import { AlertList } from './AlertList';
import { AlertFilters } from './AlertFilters';
import { AlertBadge } from './AlertBadge';

import type { WebSocketMessage } from '../services/websocket';
import type { Alert } from '../types';
import './Dashboard.css';

// ============================================
// Tipos
// ============================================

type Tab = 'overview' | 'alerts';

// ============================================
// Componente principal
// ============================================

/**
 * Dashboard principal de CloudPulse AI.
 *
 * Features:
 * - Estadísticas de métricas en tiempo real
 * - Gráfico de evolución
 * - Lista de alertas con filtros
 * - Conexión WebSocket
 * - Notificaciones toast
 */
export function Dashboard() {
  // ---------- Estado local ----------
  const [activeTab, setActiveTab] = useState<Tab>('overview');
  const [toasts, setToasts] = useState<Toast[]>([]);

  // ---------- Métricas (React Query) ----------
  const {
    data: metricsData,
    isLoading: metricsLoading,
    isError: metricsError,
    error: metricsErrorObj,
    refetch: refetchMetrics,
    isFetching: metricsFetching,
    dataUpdatedAt,
  } = useMetrics({ limit: 50, refetchInterval: 30_000 });

  const metrics = metricsData?.data ?? [];

  // ---------- Estadísticas ----------
  const cpuStats = useMetricStats(metrics, 'cpu_usage');
  const memoryStats = useMetricStats(metrics, 'memory_usage');
  const diskStats = useMetricStats(metrics, 'disk_usage');

  // ---------- Alertas (React Query) ----------
  const {
    alerts,
    counts,
    severityFilter,
    statusFilter,
    setSeverityFilter,
    setStatusFilter,
    clearFilters,
    hasActiveFilters,
    isLoading: alertsLoading,
    resolveAlert,
    acknowledgeAlert,
    isUpdating,
    refetch: refetchAlerts,
  } = useAlerts({ limit: 20 });

  // ---------- WebSocket ----------
  const { status: wsStatus, isConnected: wsConnected } = useWebSocket({
    onMessage: (message: WebSocketMessage) => {
      // Manejar alerta nueva
      if (message.type === 'alert') {
        const alert = message.data as Alert;
        if (alert.severity === 'critical') {
          addToast({
            id: alert.id,
            type: 'critical',
            message: alert.message,
          });
        }
        // Refrescar alertas para incluir la nueva
        refetchAlerts();
      }

      // Manejar métrica nueva (solo log por ahora)
      if (message.type === 'metric') {
        console.debug('📊 Nueva métrica en tiempo real:', message.data);
      }
    },
  });

  // ---------- Helpers de toast ----------
  const addToast = (toast: Toast) => {
    setToasts((prev) => [...prev, toast]);

    // Auto-remover después de 5s
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== toast.id));
    }, 5000);
  };

  const removeToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  // ---------- Refrescar todo ----------
  const handleRefreshAll = () => {
    refetchMetrics();
    refetchAlerts();
  };

  // ---------- Formatear última actualización ----------
  const lastUpdate = dataUpdatedAt
    ? new Date(dataUpdatedAt).toLocaleTimeString('es-ES')
    : 'Nunca';

  const anyFetching = metricsFetching;

  // ---------- Estado: Error ----------
  if (metricsError) {
    return (
      <div className="dashboard">
        <div className="dashboard__error">
          <h2>⚠️ Error al cargar datos</h2>
          <p>{getErrorMessage(metricsErrorObj)}</p>
          <button
            className="dashboard__retry"
            onClick={handleRefreshAll}
            type="button"
          >
            Reintentar
          </button>
        </div>
      </div>
    );
  }

  // ---------- Estado: Cargando inicial ----------
  if (metricsLoading) {
    return <DashboardSkeleton />;
  }

  // ---------- Render principal ----------
  return (
    <div className="dashboard">
      {/* Header con estado y tabs */}
      <div className="dashboard__header">
        <div className="dashboard__status-group">
          {/* Indicador WebSocket */}
          <div className="dashboard__ws-status">
            <span
              className={`dashboard__indicator dashboard__indicator--${wsStatus}`}
              aria-label={getWsStatusLabel(wsStatus)}
              role="status"
            />
            <span className="dashboard__ws-label">
              {getWsStatusLabel(wsStatus)}
            </span>
          </div>

          {/* Última actualización */}
          <span className="dashboard__last-update">
            Actualizado: <strong>{lastUpdate}</strong>
          </span>
        </div>

        <button
          className="dashboard__refresh"
          onClick={handleRefreshAll}
          disabled={anyFetching}
          type="button"
          aria-label="Refrescar datos"
        >
          {anyFetching ? '⟳ Actualizando...' : '⟳ Refrescar'}
        </button>
      </div>

      {/* Tabs */}
      <nav className="dashboard__tabs" role="tablist">
        <button
          type="button"
          role="tab"
          className={`dashboard__tab ${activeTab === 'overview' ? 'dashboard__tab--active' : ''}`}
          onClick={() => setActiveTab('overview')}
          aria-selected={activeTab === 'overview'}
        >
          📊 Resumen
        </button>
        <button
          type="button"
          role="tab"
          className={`dashboard__tab ${activeTab === 'alerts' ? 'dashboard__tab--active' : ''}`}
          onClick={() => setActiveTab('alerts')}
          aria-selected={activeTab === 'alerts'}
        >
          🔔 Alertas
          {counts.active > 0 && (
            <AlertBadge
              count={counts.active}
              variant={counts.critical > 0 ? 'critical' : 'warning'}
              size="sm"
              pulse={counts.critical > 0}
              hideWhenZero
            />
          )}
        </button>
      </nav>

      {/* ---------- TAB: Overview ---------- */}
      {activeTab === 'overview' && (
        <>
          {/* Grid de StatsCards */}
          <div className="dashboard__grid">
            <StatsCard
              title="CPU"
              icon="💻"
              stats={cpuStats}
              color="primary"
            />
            <StatsCard
              title="Memoria"
              icon="🧠"
              stats={memoryStats}
              color="warning"
            />
            <StatsCard
              title="Disco"
              icon="💾"
              stats={diskStats}
              color="success"
            />
            <StatsCard
              title="Métricas"
              icon="📊"
              unit=""
              stats={{
                current: metrics.length,
                average: metrics.length,
                max: metrics.length,
                min: metrics.length,
                trend: 'stable',
                trendValue: 0,
              }}
              color="primary"
            />
          </div>

          {/* Gráfico */}
          <div className="dashboard__chart">
            <h2 className="dashboard__chart-title">
              Evolución en tiempo real
              {wsConnected && (
                <span className="dashboard__live-badge" title="Actualización en vivo">
                  ● LIVE
                </span>
              )}
            </h2>
            <MetricsChart
              metrics={metrics}
              height={320}
              metricsToShow={['cpu_usage', 'memory_usage', 'disk_usage']}
            />
          </div>

          {/* Tabla de últimas métricas */}
          <div className="dashboard__table">
            <h2 className="dashboard__table-title">Últimas métricas</h2>
            <div className="dashboard__table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Hora</th>
                    <th>CPU</th>
                    <th>Memoria</th>
                    <th>Disco</th>
                    <th>Red In</th>
                    <th>Red Out</th>
                  </tr>
                </thead>
                <tbody>
                  {metrics.slice(0, 10).map((metric) => (
                    <tr key={metric.id}>
                      <td>
                        {new Date(metric.timestamp * 1000).toLocaleTimeString('es-ES')}
                      </td>
                      <td>{metric.cpu_usage.toFixed(1)}%</td>
                      <td>{metric.memory_usage.toFixed(1)}%</td>
                      <td>{metric.disk_usage.toFixed(1)}%</td>
                      <td>{metric.network_in.toFixed(1)} KB/s</td>
                      <td>{metric.network_out.toFixed(1)} KB/s</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* ---------- TAB: Alerts ---------- */}
      {activeTab === 'alerts' && (
        <div className="dashboard__alerts">
          {/* Filtros */}
          <AlertFilters
            severity={severityFilter}
            status={statusFilter}
            onSeverityChange={setSeverityFilter}
            onStatusChange={setStatusFilter}
            onClear={hasActiveFilters ? clearFilters : undefined}
            counts={counts}
            visibleCount={alerts.length}
            totalCount={counts.total}
            disabled={alertsLoading || isUpdating}
          />

          {/* Lista de alertas */}
          <AlertList
            alerts={alerts}
            isLoading={alertsLoading}
            onResolve={resolveAlert}
            onAcknowledge={acknowledgeAlert}
            disabled={isUpdating}
            emptyMessage={
              hasActiveFilters
                ? 'No hay alertas que coincidan con los filtros'
                : 'No hay alertas para mostrar'
            }
          />
        </div>
      )}

      {/* Toasts */}
      <div className="dashboard__toasts" aria-live="polite" aria-atomic="false">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onClose={() => removeToast(toast.id)} />
        ))}
      </div>
    </div>
  );
}

// ============================================
// Sub-componente: DashboardSkeleton
// ============================================

function DashboardSkeleton() {
  return (
    <div className="dashboard">
      <div className="dashboard__grid">
        <StatsCard title="CPU" icon="💻" stats={emptyStats} isLoading />
        <StatsCard title="Memoria" icon="🧠" stats={emptyStats} isLoading />
        <StatsCard title="Disco" icon="💾" stats={emptyStats} isLoading />
        <StatsCard title="Métricas" icon="📊" stats={emptyStats} isLoading />
      </div>
      <div className="dashboard__chart-skeleton" />
    </div>
  );
}

// ============================================
// Sub-componente: Toast
// ============================================

interface Toast {
  id: string;
  type: 'critical' | 'warning' | 'info' | 'success';
  message: string;
}

interface ToastItemProps {
  toast: Toast;
  onClose: () => void;
}

function ToastItem({ toast, onClose }: ToastItemProps) {
  const icons: Record<Toast['type'], string> = {
    critical: '🚨',
    warning: '⚠️',
    info: 'ℹ️',
    success: '✅',
  };

  return (
    <div className={`toast toast--${toast.type}`} role="alert">
      <span className="toast__icon" aria-hidden="true">
        {icons[toast.type]}
      </span>
      <p className="toast__message">{toast.message}</p>
      <button
        type="button"
        className="toast__close"
        onClick={onClose}
        aria-label="Cerrar notificación"
      >
        ✕
      </button>
    </div>
  );
}

// ============================================
// Helpers
// ============================================

const emptyStats = {
  current: 0,
  average: 0,
  max: 0,
  min: 0,
  trend: 'stable' as const,
  trendValue: 0,
};

function getWsStatusLabel(status: string): string {
  switch (status) {
    case 'connected':
      return 'En vivo';
    case 'connecting':
      return 'Conectando';
    case 'disconnected':
      return 'Desconectado';
    case 'error':
      return 'Error de conexión';
    default:
      return 'Desconocido';
  }
}