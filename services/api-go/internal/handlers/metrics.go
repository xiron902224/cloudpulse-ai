package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xiron902224/cloudpulse-ai/api/internal/cache"
	"github.com/xiron902224/cloudpulse-ai/api/internal/models"
)

// MetricsHandler handles metric requests.
type MetricsHandler struct {
	DB    *pgxpool.Pool
	Cache *cache.RedisClient
}

// CreateMetric godoc
// @Summary     Create metric
// @Description Creates a new monitoring metric
// @Tags        metrics
// @Accept      json
// @Produce     json
// @Param       metric body models.Metric true "Metric data"
// @Success     201 {object} models.Metric
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/metrics [post]
func (h *MetricsHandler) CreateMetric(c *fiber.Ctx) error {
	var metric models.Metric

	// Parse JSON body
	if err := c.BodyParser(&metric); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid data",
		})
	}

	// Validate data
	if metric.CPUUsage < 0 || metric.CPUUsage > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cpu_usage must be between 0 and 100",
		})
	}

	if metric.MemoryUsage < 0 || metric.MemoryUsage > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "memory_usage must be between 0 and 100",
		})
	}

	// Generate ID
	metric.ID = uuid.New().String()
	metric.CreatedAt = time.Now()

	// Use the current time when timestamp is not provided
	if metric.Timestamp == 0 {
		metric.Timestamp = time.Now().Unix()
	}

	// Save to PostgreSQL
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
		log.Printf("Error saving metric: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error saving metric",
		})
	}

	// Invalidate cache
	h.Cache.Delete(ctx, "metrics:latest")

	return c.Status(fiber.StatusCreated).JSON(metric)
}

// ListMetrics godoc
// @Summary     List metrics
// @Description Gets all metrics
// @Tags        metrics
// @Produce     json
// @Param       limit query int false "Maximum number of results"
// @Param       offset query int false "Skip N results"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/metrics [get]
func (h *MetricsHandler) ListMetrics(c *fiber.Ctx) error {
	// Get pagination parameters
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	// Validate limits
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	cacheKey := fmt.Sprintf("metrics:list:%d:%d", limit, offset)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cached, err := h.Cache.Get(ctx, cacheKey); err == nil {
		var cachedMetrics []models.Metric
		if err := json.Unmarshal([]byte(cached), &cachedMetrics); err == nil {
			return c.JSON(fiber.Map{
				"data":   cachedMetrics,
				"total":  len(cachedMetrics),
				"limit":  limit,
				"offset": offset,
				"cached": true,
			})
		}
	}
	query := `
        SELECT id, cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp, created_at
        FROM metrics
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := h.DB.Query(ctx, query, limit, offset)
	if err != nil {
		log.Printf("Error querying metrics: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error querying metrics",
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
			log.Printf("Error scanning metric: %v", err)
			continue
		}
		metrics = append(metrics, m)
	}

	if jsonData, err := json.Marshal(metrics); err == nil {
		h.Cache.Set(ctx, cacheKey, jsonData, 1*time.Minute)
	}
	return c.JSON(fiber.Map{
		"data":   metrics,
		"total":  len(metrics),
		"limit":  limit,
		"offset": offset,
		"cached": false,
	})
}

// GetMetric godoc
// @Summary     Get metric
// @Description Gets a metric by ID
// @Tags        metrics
// @Produce     json
// @Param       id path string true "Metric ID"
// @Success     200 {object} models.Metric
// @Failure     404 {object} map[string]interface{}
// @Router      /api/v1/metrics/{id} [get]
func (h *MetricsHandler) GetMetric(c *fiber.Ctx) error {
	id := c.Params("id")
	cacheKey := fmt.Sprintf("metric:%s", id)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cached, err := h.Cache.Get(ctx, cacheKey); err == nil {
		var metric models.Metric
		if err := json.Unmarshal([]byte(cached), &metric); err == nil {
			return c.JSON(metric)
		}
	}
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
			"error": "Metric not found",
		})
	}
	if jsonData, err := json.Marshal(metric); err == nil {
		h.Cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
	}
	return c.JSON(metric)
}

// DeleteMetric godoc
// @Summary     Delete metric
// @Description Deletes a metric by ID
// @Tags        metrics
// @Param       id path string true "Metric ID"
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
		log.Printf("Error deleting metric: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error deleting metric",
		})
	}

	// Invalidate cache
	h.Cache.Delete(ctx, "metrics:latest")

	return c.SendStatus(fiber.StatusNoContent)
}
