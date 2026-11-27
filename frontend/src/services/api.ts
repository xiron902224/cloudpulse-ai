import axios, { AxiosError } from 'axios';

import type {
  Metric,
  MetricsResponse,
  CreateMetricPayload,
  Alert,
  AlertsResponse,
  CreateAlertPayload,
  UpdateAlertPayload,
  HealthResponse,
  ApiError,
  MetricsFilters,
  AlertsFilters,
} from '../types';

// ============================================
// Cliente Axios configurado
// ============================================

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const axiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// ============================================
// Interceptores
// ============================================

// Response interceptor: manejo centralizado de errores
axiosInstance.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiError>) => {
    if (error.response) {
      // Error de la API (4xx, 5xx)
      const apiError = error.response.data;
      console.error(
        `[API Error] ${error.response.status}:`,
        apiError?.error || error.message
      );
    } else if (error.request) {
      // No hubo respuesta (red, timeout)
      console.error('[API Error] No response:', error.message);
    } else {
      // Error configurando la petición
      console.error('[API Error]', error.message);
    }
    return Promise.reject(error);
  }
);

// ============================================
// Helpers
// ============================================

/**
 * Extrae el mensaje de error de un error de Axios
 */
function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const apiError = error.response?.data as ApiError | undefined;
    if (apiError?.error) {
      return apiError.error;
    }
    if (error.code === 'ECONNABORTED') {
      return 'La petición tardó demasiado';
    }
    if (!error.response) {
      return 'No se pudo conectar con el servidor';
    }
    return `Error ${error.response.status}: ${error.response.statusText}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Error desconocido';
}

// ============================================
// API Service
// ============================================

export const api = {
  // ---------- Health ----------
  async getHealth(): Promise<HealthResponse> {
    const { data } = await axiosInstance.get<HealthResponse>('/health');
    return data;
  },

  // ---------- Métricas ----------
  async getMetrics(filters: MetricsFilters = {}): Promise<MetricsResponse> {
    const params = new URLSearchParams();
    if (filters.limit !== undefined) params.append('limit', String(filters.limit));
    if (filters.offset !== undefined) params.append('offset', String(filters.offset));
    if (filters.from !== undefined) params.append('from', String(filters.from));
    if (filters.to !== undefined) params.append('to', String(filters.to));

    const { data } = await axiosInstance.get<MetricsResponse>(
      `/api/v1/metrics?${params.toString()}`
    );
    return data;
  },

  async getMetric(id: string): Promise<Metric> {
    const { data } = await axiosInstance.get<Metric>(`/api/v1/metrics/${id}`);
    return data;
  },

  async createMetric(payload: CreateMetricPayload): Promise<Metric> {
    const { data } = await axiosInstance.post<Metric>('/api/v1/metrics', payload);
    return data;
  },

  async deleteMetric(id: string): Promise<void> {
    await axiosInstance.delete(`/api/v1/metrics/${id}`);
  },

  // ---------- Alertas ----------
  async getAlerts(filters: AlertsFilters = {}): Promise<AlertsResponse> {
    const params = new URLSearchParams();
    if (filters.severity) params.append('severity', filters.severity);
    if (filters.status) params.append('status', filters.status);

    const { data } = await axiosInstance.get<AlertsResponse>(
      `/api/v1/alerts?${params.toString()}`
    );
    return data;
  },

  async createAlert(payload: CreateAlertPayload): Promise<Alert> {
    const { data } = await axiosInstance.post<Alert>('/api/v1/alerts', payload);
    return data;
  },

  async updateAlert(id: string, payload: UpdateAlertPayload): Promise<Alert> {
    const { data } = await axiosInstance.put<Alert>(
      `/api/v1/alerts/${id}`,
      payload
    );
    return data;
  },
};

// ============================================
// Exportaciones adicionales
// ============================================

export { getErrorMessage };
export { axiosInstance };