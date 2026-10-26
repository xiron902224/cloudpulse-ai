import type { MetricStats } from '../types';
import './StatsCard.css';

interface StatsCardProps {
  title: string;
  icon: string;
  unit?: string;
  stats: MetricStats;
  color?: 'primary' | 'success' | 'warning' | 'danger';
  isLoading?: boolean;
}

/**
 * Tarjeta de estadísticas del sistema.
 *
 * @example
 * ```tsx
 * <StatsCard
 *   title="CPU"
 *   icon="💻"
 *   unit="%"
 *   stats={{ current: 45.5, average: 42.3, max: 78.9, min: 12.5, trend: 'up', trendValue: 2.1 }}
 * />
 * ```
 */
export function StatsCard({
  title,
  icon,
  unit = '%',
  stats,
  color = 'primary',
  isLoading = false,
}: StatsCardProps) {
  // Formatear el número (ej: 45.5%)
  const formatValue = (value: number): string => {
    return `${value.toFixed(1)}${unit}`;
  };

  // Determinar el símbolo de tendencia
  const getTrendIcon = (): string => {
    switch (stats.trend) {
      case 'up':
        return '↑';
      case 'down':
        return '↓';
      case 'stable':
        return '→';
    }
  };

  // Clase de color según tendencia
  const trendClass = `trend trend--${stats.trend}`;

  if (isLoading) {
    return (
      <div className={`stats-card stats-card--${color}`}>
        <div className="stats-card__skeleton" />
      </div>
    );
  }

  return (
    <div className={`stats-card stats-card--${color}`}>
      {/* Header */}
      <div className="stats-card__header">
        <span className="stats-card__icon">{icon}</span>
        <span className="stats-card__title">{title}</span>
      </div>

      {/* Valor principal */}
      <div className="stats-card__value">
        {formatValue(stats.current)}
      </div>

      {/* Tendencia */}
      <div className={trendClass}>
        <span className="trend__icon">{getTrendIcon()}</span>
        <span className="trend__value">{stats.trendValue.toFixed(1)}%</span>
      </div>

      {/* Detalles */}
      <div className="stats-card__details">
        <div className="stats-card__detail">
          <span className="stats-card__detail-label">Promedio</span>
          <span className="stats-card__detail-value">
            {formatValue(stats.average)}
          </span>
        </div>
        <div className="stats-card__detail">
          <span className="stats-card__detail-label">Máx</span>
          <span className="stats-card__detail-value">{formatValue(stats.max)}</span>
        </div>
        <div className="stats-card__detail">
          <span className="stats-card__detail-label">Mín</span>
          <span className="stats-card__detail-value">{formatValue(stats.min)}</span>
        </div>
      </div>
    </div>
  );
}