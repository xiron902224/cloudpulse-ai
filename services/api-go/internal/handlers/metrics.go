package handlers

import (
    "context"
    "log"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/xiron902224/cloudpulse-ai/api/internal/cache"
    "github.com/xiron902224/cloudpulse-ai/api/internal/models"
)

// MetricsHandler maneja las peticiones de métricas
type MetricsHandler struct {
    DB    *pgxpool.Pool
    Cache *cache.RedisClient
}

// CreateMetric godoc
// @Summary     Crear métrica
// @Description Crea una nueva métrica de monitoreo
// @Tags        metrics
// @Accept      json
// @Produce     json
// @Param       metric body models.Metric true "Datos de la métrica"
// @Success     201 {object} models.Metric
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/metrics [post]
func (h *MetricsHandler) CreateMetric(c *fiber.Ctx) error {
    var metric models.Metric

    // Parsear body JSON
    if err := c.BodyParser(&metric); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Datos inválidos",
        })
    }

    // Validar datos
    if metric.CPUUsage < 0 || metric.CPUUsage > 100 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "cpu_usage debe estar entre 0 y 100",
        })
    }

    if metric.MemoryUsage < 0 || metric.MemoryUsage > 100 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "memory_usage debe estar entre 0 y 100",
        })
    }

    // Generar ID
    metric.ID = uuid.New().String()
    metric.CreatedAt = time.Now()

    // Si no hay timestamp, usar ahora
    if metric.Timestamp == 0 {
        metric.Timestamp = time.Now().Unix()
    }

    // Guardar en PostgreSQL
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    query := `
        INSERT INTO metrics (id, cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

    _, err := h.DB.Exec(ctx, query,
        metric.ID,
        metric.CPUUsage,
        metric.MemoryUsage,
        metric.DiskUsage,
        metric.NetworkIn,
        metric.NetworkOut,
        metric.Timestamp,
        metric.CreatedAt,
    )

    if err != nil {
        log.Printf("Error guardando métrica: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error guardando métrica",
        })
    }

    // Invalidar caché
    h.Cache.Delete(ctx, "metrics:latest")

    return c.Status(fiber.StatusCreated).JSON(metric)
}

// ListMetrics godoc
// @Summary     Listar métricas
// @Description Obtiene todas las métricas
// @Tags        metrics
// @Produce     json
// @Param       limit query int false "Máximo de resultados"
// @Param       offset query int false "Saltar N resultados"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/metrics [get]
func (h *MetricsHandler) ListMetrics(c *fiber.Ctx) error {
    // Obtener parámetros de paginación
    limit := c.QueryInt("limit", 50)
    offset := c.QueryInt("offset", 0)

    // Validar límites
    if limit <= 0 {
        limit = 50
    }
    if limit > 100 {
        limit = 100
    }
    if offset < 0 {
        offset = 0
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    query := `
        SELECT id, cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp, created_at
        FROM metrics
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

    rows, err := h.DB.Query(ctx, query, limit, offset)
    if err != nil {
        log.Printf("Error consultando métricas: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error consultando métricas",
        })
    }
    defer rows.Close()

    metrics := make([]models.Metric, 0)

    for rows.Next() {
        var m models.Metric
        err := rows.Scan(
            &m.ID,
            &m.CPUUsage,
            &m.MemoryUsage,
            &m.DiskUsage,
            &m.NetworkIn,
            &m.NetworkOut,
            &m.Timestamp,
            &m.CreatedAt,
        )
        if err != nil {
            log.Printf("Error escaneando métrica: %v", err)
            continue
        }
        metrics = append(metrics, m)
    }

    return c.JSON(fiber.Map{
        "data":   metrics,
        "total":  len(metrics),
        "limit":  limit,
        "offset": offset,
    })
}

// GetMetric godoc
// @Summary     Obtener métrica
// @Description Obtiene una métrica por ID
// @Tags        metrics
// @Produce     json
// @Param       id path string true "ID de la métrica"
// @Success     200 {object} models.Metric
// @Failure     404 {object} map[string]interface{}
// @Router      /api/v1/metrics/{id} [get]
func (h *MetricsHandler) GetMetric(c *fiber.Ctx) error {
    id := c.Params("id")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    query := `
        SELECT id, cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp, created_at
        FROM metrics
        WHERE id = $1
    `

    var metric models.Metric

    err := h.DB.QueryRow(ctx, query, id).Scan(
        &metric.ID,
        &metric.CPUUsage,
        &metric.MemoryUsage,
        &metric.DiskUsage,
        &metric.NetworkIn,
        &metric.NetworkOut,
        &metric.Timestamp,
        &metric.CreatedAt,
    )

    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Métrica no encontrada",
        })
    }

    return c.JSON(metric)
}

// DeleteMetric godoc
// @Summary     Eliminar métrica
// @Description Elimina una métrica por ID
// @Tags        metrics
// @Param       id path string true "ID de la métrica"
// @Success     204 "No Content"
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/metrics/{id} [delete]
func (h *MetricsHandler) DeleteMetric(c *fiber.Ctx) error {
    id := c.Params("id")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    query := `DELETE FROM metrics WHERE id = $1`

    _, err := h.DB.Exec(ctx, query, id)
    if err != nil {
        log.Printf("Error eliminando métrica: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error eliminando métrica",
        })
    }

    // Invalidar caché
    h.Cache.Delete(ctx, "metrics:latest")

    return c.SendStatus(fiber.StatusNoContent)
}