package middleware

import (
    "strconv"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Variables de métricas (globales)
var (
    // Contador de peticiones totales
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Número total de peticiones HTTP",
        },
        []string{"method", "path", "status"},
    )

    // Histograma de duración de peticiones
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duración de las peticiones HTTP",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    // Contador de errores
    httpErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_errors_total",
            Help: "Número total de errores HTTP",
        },
        []string{"method", "path", "status"},
    )
)

// PrometheusMetrics middleware que recolecta métricas
func PrometheusMetrics() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Registrar tiempo de inicio
        start := time.Now()

        // Ejecutar la petición
        err := c.Next()

        // Calcular duración
        duration := time.Since(start).Seconds()

        // Obtener método y path
        method := c.Method()
        path := c.Path()
        status := strconv.Itoa(c.Response().StatusCode())

        // Registrar métricas
        httpRequestsTotal.WithLabelValues(method, path, status).Inc()
        httpRequestDuration.WithLabelValues(method, path).Observe(duration)

        // Si es error (4xx o 5xx)
        if c.Response().StatusCode() >= 400 {
            httpErrorsTotal.WithLabelValues(method, path, status).Inc()
        }

        return err
    }
}