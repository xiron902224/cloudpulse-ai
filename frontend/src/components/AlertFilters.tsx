import type { AlertSeverity, AlertStatus } from '../types';
import './AlertFilters.css';

// ============================================
// Tipos
// ============================================

type SeverityFilter = AlertSeverity | 'all';
type StatusFilter = AlertStatus | 'all';

interface FilterOption<T> {
  value: T;
  label: string;
  icon?: string;
}

interface AlertFiltersProps {
  /** Filtro actual de severidad */
  severity: SeverityFilter;

  /** Filtro actual de estado */
  status: StatusFilter;

  /** Callback al cambiar severidad */
  onSeverityChange: (severity: SeverityFilter) => void;

  /** Callback al cambiar estado */
  onStatusChange: (status: StatusFilter) => void;

  /** Callback al limpiar todos los filtros */
  onClear?: () => void;

  /** Contadores para mostrar en cada opción (opcional) */
  counts?: {
    critical: number;
    warning: number;
    info: number;
    active: number;
    resolved: number;
    acknowledged: number;
  };

  /** Número de alertas mostradas actualmente */
  visibleCount?: number;

  /** Número total de alertas */
  totalCount?: number;

  /** Deshabilitar todos los filtros */
  disabled?: boolean;
}

// ============================================
// Opciones
// ============================================

const SEVERITY_OPTIONS: FilterOption<SeverityFilter>[] = [
  { value: 'all', label: 'Todas' },
  { value: 'critical', label: 'Critical', icon: '🔴' },
  { value: 'warning', label: 'Warning', icon: '🟡' },
  { value: 'info', label: 'Info', icon: '🔵' },
];

const STATUS_OPTIONS: FilterOption<StatusFilter>[] = [
  { value: 'all', label: 'Todas' },
  { value: 'active', label: 'Activas', icon: '●' },
  { value: 'resolved', label: 'Resueltas', icon: '✓' },
  { value: 'acknowledged', label: 'Reconocidas', icon: '◐' },
];

// ============================================
// Componente
// ============================================

/**
 * Panel de filtros para alertas.
 *
 * @example
 * ```tsx
 * <AlertFilters
 *   severity={severityFilter}
 *   status={statusFilter}
 *   onSeverityChange={setSeverityFilter}
 *   onStatusChange={setStatusFilter}
 *   onClear={clearFilters}
 *   counts={counts}
 * />
 * ```
 */
export function AlertFilters({
  severity,
  status,
  onSeverityChange,
  onStatusChange,
  onClear,
  counts,
  visibleCount,
  totalCount,
  disabled = false,
}: AlertFiltersProps) {
  const hasActiveFilters = severity !== 'all' || status !== 'all';

  // ---------- Obtener contador para una opción ----------
  const getCount = (
    type: 'severity' | 'status',
    value: SeverityFilter | StatusFilter
  ): number | undefined => {
    if (!counts || value === 'all') return undefined;

    if (type === 'severity') {
      return counts[value as AlertSeverity];
    }
    return counts[value as AlertStatus];
  };

  // ---------- Render de un grupo de filtros ----------
  const renderFilterGroup = <T extends SeverityFilter | StatusFilter>(
    title: string,
    options: FilterOption<T>[],
    currentValue: T,
    onChange: (value: T) => void,
    type: 'severity' | 'status'
  ) => (
    <div className="alert-filters__group">
      <span className="alert-filters__group-label">{title}</span>
      <div className="alert-filters__options" role="group" aria-label={title}>
        {options.map((option) => {
          const isActive = currentValue === option.value;
          const optionCount = getCount(type, option.value);

          return (
            <button
              key={option.value}
              type="button"
              className={[
                'alert-filters__chip',
                `alert-filters__chip--${type}`,
                isActive && 'alert-filters__chip--active',
              ]
                .filter(Boolean)
                .join(' ')}
              onClick={() => onChange(option.value)}
              disabled={disabled}
              aria-pressed={isActive}
            >
              {option.icon && (
                <span
                  className="alert-filters__chip-icon"
                  aria-hidden="true"
                >
                  {option.icon}
                </span>
              )}
              <span className="alert-filters__chip-label">{option.label}</span>
              {optionCount !== undefined && optionCount > 0 && (
                <span className="alert-filters__chip-count">
                  {optionCount}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );

  return (
    <div className="alert-filters">
      {/* Severidad */}
      {renderFilterGroup(
        'Severidad',
        SEVERITY_OPTIONS,
        severity,
        onSeverityChange,
        'severity'
      )}

      {/* Estado */}
      {renderFilterGroup(
        'Estado',
        STATUS_OPTIONS,
        status,
        onStatusChange,
        'status'
      )}

      {/* Footer con contador y limpiar */}
      <div className="alert-filters__footer">
        <div className="alert-filters__info">
          {visibleCount !== undefined && totalCount !== undefined && (
            <span className="alert-filters__counter">
              Mostrando <strong>{visibleCount}</strong> de{' '}
              <strong>{totalCount}</strong>
            </span>
          )}
        </div>

        {hasActiveFilters && onClear && (
          <button
            type="button"
            className="alert-filters__clear"
            onClick={onClear}
            disabled={disabled}
          >
            <span aria-hidden="true">✕</span>
            <span>Limpiar filtros</span>
          </button>
        )}
      </div>
    </div>
  );
}