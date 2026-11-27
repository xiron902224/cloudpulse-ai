import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

import { formatTimestamp } from '../hooks/useMetrics';
import type { Metric } from '../types';
import './MetricsChart.css';

// ============================================
// Tipos
// ============================================

interface ChartDataPoint {
  timestamp: number;
  time: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
}

interface MetricsChartProps {
  metrics: Metric[];
  height?: number;
  showLegend?: boolean;
  showGrid?: boolean;
  metricsToShow?: Array<'cpu_usage' | 'memory_usage' | 'disk_usage'>;
}

// ============================================
// Configuración de líneas
// ============================================

const LINE_CONFIG = {
  cpu_usage: {
    name: 'CPU',
    color: '#667eea',
  },
  memory_usage: {
    name: 'Memoria',
    color: '#f59e0b',
  },
  disk_usage: {
    name: 'Disco',
    color: '#10b981',
  },
} as const;

// ============================================
// Componente
// ============================================

/**
 * Gráfico de líneas que muestra la evolución de métricas del sistema.
 *
 * @example
 * ```tsx
 * <MetricsChart
 *   metrics={metrics}
 *   height={300}
 *   metricsToShow={['cpu_usage', 'memory_usage']}
 * />
 * ```
 */
export function MetricsChart({
  metrics,
  height = 300,
  showLegend = true,
  showGrid = true,
  metricsToShow = ['cpu_usage', 'memory_usage', 'disk_usage'],
}: MetricsChartProps) {
  // Transformar métricas al formato de Recharts
  const chartData: ChartDataPoint[] = metrics
    .slice()
    .sort((a, b) => a.timestamp - b.timestamp) // Orden ascendente por tiempo
    .map((m) => ({
      timestamp: m.timestamp,
      time: formatTimestamp(m.timestamp),
      cpu_usage: m.cpu_usage,
      memory_usage: m.memory_usage,
      disk_usage: m.disk_usage,
    }));

  // Estado vacío
  if (chartData.length === 0) {
    return (
      <div className="metrics-chart metrics-chart--empty">
        <div className="metrics-chart__empty">
          <p>No hay datos disponibles</p>
          <small>Esperando métricas del collector...</small>
        </div>
      </div>
    );
  }

  return (
    <div className="metrics-chart">
      <ResponsiveContainer width="100%" height={height}>
        <LineChart
          data={chartData}
          margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
        >
          {/* Grid de fondo */}
          {showGrid && (
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="rgba(255, 255, 255, 0.1)"
            />
          )}

          {/* Eje X (tiempo) */}
          <XAxis
            dataKey="time"
            stroke="rgba(255, 255, 255, 0.5)"
            tick={{ fill: 'rgba(255, 255, 255, 0.7)', fontSize: 12 }}
            tickLine={{ stroke: 'rgba(255, 255, 255, 0.2)' }}
            axisLine={{ stroke: 'rgba(255, 255, 255, 0.2)' }}
          />

          {/* Eje Y (porcentaje) */}
          <YAxis
            domain={[0, 100]}
            stroke="rgba(255, 255, 255, 0.5)"
            tick={{ fill: 'rgba(255, 255, 255, 0.7)', fontSize: 12 }}
            tickLine={{ stroke: 'rgba(255, 255, 255, 0.2)' }}
            axisLine={{ stroke: 'rgba(255, 255, 255, 0.2)' }}
            tickFormatter={(value) => `${value}%`}
          />

          {/* Tooltip al pasar el mouse */}
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(15, 23, 42, 0.95)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: '0.5rem',
              color: '#ffffff',
            }}
            labelStyle={{ color: 'rgba(255, 255, 255, 0.7)' }}
            itemStyle={{ color: '#ffffff' }}
            formatter={(value: number) => `${value.toFixed(2)}%`}
          />

          {/* Leyenda */}
          {showLegend && (
            <Legend
              wrapperStyle={{
                color: 'rgba(255, 255, 255, 0.7)',
                fontSize: '0.875rem',
              }}
              iconType="line"
            />
          )}

          {/* Líneas */}
          {metricsToShow.map((key) => (
            <Line
              key={key}
              type="monotone"
              dataKey={key}
              name={LINE_CONFIG[key].name}
              stroke={LINE_CONFIG[key].color}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 6, strokeWidth: 0 }}
              animationDuration={500}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}