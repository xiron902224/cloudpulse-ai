import { useMemo } from 'react';

import type { Alert, AlertSeverity, AlertStatus } from '../types';
import './AlertList.css';

// ============================================
// Tipos
// ============================================

interface AlertListProps {
  /** Lista de alertas a mostrar */
  alerts: Alert[];

  /** ¿Está cargando? */
  isLoading?: boolean;

  /** Callback al resolver una alerta */
  onResolve?: (id: string) => void;

  /** Callback al reconocer una alerta */
  onAcknowledge?: (id: string) => void;

  /** Deshabilitar acciones */
  disabled?: boolean;

  /** Máximo de alertas a mostrar (default: sin límite) */
  maxItems?: number;

  /** Mensaje cuando está vacío */
  emptyMessage?: string;
}

// ============================================
// Configuración
// ============================================

const SEVERITY_CONFIG: Record<
  AlertSeverity,
  { icon: string; label: string; className: string }
> = {
  critical: {
    icon: '🔴',
    label: 'Critical',
    className: 'alert-item--critical',
  },
  warning: {
    icon: '🟡',
    label: 'Warning',
    className: 'alert-item--warning',
  },
  info: {
    icon: '🔵',
    label: 'Info',
    className: 'alert-item--info',
  },
};

const STATUS_CONFIG: Record<
  AlertStatus,
  { label: string; className: string; icon: string }
> = {
  active: {
    label: 'Activa',
    className: 'alert-status--active',
    icon: '●',
  },
  resolved: {
    label: 'Resuelta',
    className: 'alert-status--resolved',
    icon: '✓',
  },
  acknowledged: {
    label: 'Reconocida',
    className: 'alert-status--acknowledged',
    icon: '◐',
  },
};

// ============================================
// Helpers
// ============================================

/**
 * Formatea un timestamp como tiempo relativo.
 * Ej: "Hace 2 min", "Hace 1h", "Hace 3 días"
 */
function formatRelativeTime(isoString: string): string {
  const now = Date.now();
  const timestamp = new Date(isoString).getTime();
  const diffMs = now - timestamp;
  const diffSec = Math.floor(diffMs / 1000);

  if (diffSec < 60) {
    return 'Hace unos segundos';
  }

  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) {
    return `Hace ${diffMin} min`;
  }

  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) {
    return `Hace ${diffHour}h`;
  }

  const diffDay = Math.floor(diffHour / 24);
  if (diffDay < 30) {
    return `Hace ${diffDay} ${diffDay === 1 ? 'día' : 'días'}`;
  }

  const diffMonth = Math.floor(diffDay / 30);
  if (diffMonth < 12) {
    return `Hace ${diffMonth} ${diffMonth === 1 ? 'mes' : 'meses'}`;
  }

  const diffYear = Math.floor(diffMonth / 12);
  return `Hace ${diffYear} ${diffYear === 1 ? 'año' : 'años'}`;
}

// ============================================
// Componente
// ============================================

/**
 * Lista de alertas con acciones.
 *
 * @example
 * ```tsx
 * <AlertList
 *   alerts={alerts}
 *   onResolve={resolveAlert}
 *   onAcknowledge={acknowledgeAlert}
 * />
 * ```
 */
export function AlertList({
  alerts,
  isLoading = false,
  onResolve,
  onAcknowledge,
  disabled = false,
  maxItems,
  emptyMessage = 'No hay alertas para mostrar',
}: AlertListProps) {
  // ---------- Slice si maxItems ----------
  const visibleAlerts = useMemo(() => {
    if (maxItems && alerts.length > maxItems) {
      return alerts.slice(0, maxItems);
    }
    return alerts;
  }, [alerts, maxItems]);

  // ---------- Skeleton de carga ----------
  if (isLoading) {
    return (
      <div className="alert-list">
        {[1, 2, 3].map((i) => (
          <div key={i} className="alert-item alert-item--skeleton">
            <div className="alert-item__skeleton-line" />
            <div className="alert-item__skeleton-line alert-item__skeleton-line--short" />
          </div>
        ))}
      </div>
    );
  }

  // ---------- Estado vacío ----------
  if (visibleAlerts.length === 0) {
    return (
      <div className="alert-list alert-list--empty">
        <div className="alert-list__empty">
          <span className="alert-list__empty-icon" aria-hidden="true">
            ✅
          </span>
          <p className="alert-list__empty-text">{emptyMessage}</p>
          <small className="alert-list__empty-hint">
            Todo está funcionando correctamente
          </small>
        </div>
      </div>
    );
  }

  // ---------- Lista de alertas ----------
  return (
    <div className="alert-list">
      {visibleAlerts.map((alert) => (
        <AlertItem
          key={alert.id}
          alert={alert}
          onResolve={onResolve}
          onAcknowledge={onAcknowledge}
          disabled={disabled}
        />
      ))}

      {maxItems && alerts.length > maxItems && (
        <div className="alert-list__more">
          Y {alerts.length - maxItems} más...
        </div>
      )}
    </div>
  );
}

// ============================================
// Sub-componente: AlertItem
// ============================================

interface AlertItemProps {
  alert: Alert;
  onResolve?: (id: string) => void;
  onAcknowledge?: (id: string) => void;
  disabled?: boolean;
}

function AlertItem({
  alert,
  onResolve,
  onAcknowledge,
  disabled = false,
}: AlertItemProps) {
  const severityConfig = SEVERITY_CONFIG[alert.severity];
  const statusConfig = STATUS_CONFIG[alert.status];
  const relativeTime = formatRelativeTime(alert.created_at);

  const isActive = alert.status === 'active';
  const isResolved = alert.status === 'resolved';
  const isAcknowledged = alert.status === 'acknowledged';

  const className = [
    'alert-item',
    severityConfig.className,
    isResolved && 'alert-item--resolved',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <article className={className} role="article">
      {/* ---------- Header: severidad + mensaje + tiempo ---------- */}
      <header className="alert-item__header">
        <span
          className="alert-item__severity"
          aria-label={`Severidad: ${severityConfig.label}`}
        >
          <span aria-hidden="true">{severityConfig.icon}</span>
          <span className="alert-item__severity-label">
            {severityConfig.label}
          </span>
        </span>

        <time
          className="alert-item__time"
          dateTime={alert.created_at}
          title={new Date(alert.created_at).toLocaleString('es-ES')}
        >
          {relativeTime}
        </time>
      </header>

      {/* ---------- Mensaje ---------- */}
      <p className="alert-item__message">{alert.message}</p>

      {/* ---------- Footer: estado + acciones ---------- */}
      <footer className="alert-item__footer">
        <span
          className={`alert-status ${statusConfig.className}`}
          aria-label={`Estado: ${statusConfig.label}`}
        >
          <span aria-hidden="true">{statusConfig.icon}</span>
          <span>{statusConfig.label}</span>
        </span>

        {!isResolved && (onResolve || onAcknowledge) && (
          <div className="alert-item__actions">
            {isActive && onAcknowledge && (
              <button
                type="button"
                className="alert-item__action alert-item__action--acknowledge"
                onClick={() => onAcknowledge(alert.id)}
                disabled={disabled}
                aria-label={`Reconocer alerta: ${alert.message}`}
              >
                Reconocer
              </button>
            )}

            {onResolve && (
              <button
                type="button"
                className="alert-item__action alert-item__action--resolve"
                onClick={() => onResolve(alert.id)}
                disabled={disabled}
                aria-label={`Resolver alerta: ${alert.message}`}
              >
                {isAcknowledged ? 'Resolver' : 'Resolver'}
              </button>
            )}
          </div>
        )}
      </footer>
    </article>
  );
}