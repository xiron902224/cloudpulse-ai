// ============================================
// Tipos de Métricas
// ============================================

/**
 * Métrica individual del sistema
 * Coincide con el modelo Metric del backend Go
 */
export interface Metric {
  id: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  network_in: number;
  network_out: number;
  timestamp: number;
  created_at: string;
}

/**
 * Respuesta del endpoint GET /api/v1/metrics
 */
export interface MetricsResponse {
  data: Metric[];
  total: number;
  limit: number;
  offset: number;
  cached?: boolean;
}

/**
 * Payload para crear una métrica (POST /api/v1/metrics)
 * Sin id ni created_at (los genera el backend)
 */
export interface CreateMetricPayload {
  cpu_usage: number;
  memory_usage: number;
  disk_usage?: number;
  network_in?: number;
  network_out?: number;
  timestamp?: number;
}

// ============================================
// Tipos de Alertas
// ============================================

/**
 * Severidad de una alerta
 */
export type AlertSeverity = 'critical' | 'warning' | 'info';

/**
 * Estado de una alerta
 */
export type AlertStatus = 'active' | 'resolved' | 'acknowledged';

/**
 * Alerta del sistema
 * Coincide con el modelo Alert del backend Go
 */
export interface Alert {
  id: string;
  severity: AlertSeverity;
  message: string;
  status: AlertStatus;
  created_at: string;
  updated_at?: string;
}

/**
 * Respuesta del endpoint GET /api/v1/alerts
 */
export interface AlertsResponse {
  data: Alert[];
  total: number;
}

/**
 * Payload para crear una alerta (POST /api/v1/alerts)
 */
export interface CreateAlertPayload {
  severity: AlertSeverity;
  message: string;
  status?: AlertStatus;
}

/**
 * Payload para actualizar una alerta (PUT /api/v1/alerts/:id)
 */
export interface UpdateAlertPayload {
  severity?: AlertSeverity;
  message?: string;
  status?: AlertStatus;
}

// ============================================
// Tipos de Health Check
// ============================================

/**
 * Respuesta del endpoint GET /health
 */
export interface HealthResponse {
  status: 'healthy' | 'unhealthy';
  service: string;
  version: string;
  timestamp: string;
}

// ============================================
// Tipos de Estadísticas
// ============================================

/**
 * Estadísticas calculadas de métricas
 * Usadas en las StatsCards del dashboard
 */
export interface MetricStats {
  current: number;
  average: number;
  max: number;
  min: number;
  trend: 'up' | 'down' | 'stable';
  trendValue: number;
}

// ============================================
// Tipos de Errores de API
// ============================================

/**
 * Respuesta de error de la API
 */
export interface ApiError {
  error: string;
}

// ============================================
// Tipos de Filtros
// ============================================

/**
 * Filtros para listar métricas
 */
export interface MetricsFilters {
  limit?: number;
  offset?: number;
  from?: number;
  to?: number;
}

/**
 * Filtros para listar alertas
 */
export interface AlertsFilters {
  severity?: AlertSeverity;
  status?: AlertStatus;
}