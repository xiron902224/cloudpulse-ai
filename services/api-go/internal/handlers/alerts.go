package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xiron902224/cloudpulse-ai/api/internal/cache"
	"github.com/xiron902224/cloudpulse-ai/api/internal/models"
)

// AlertsHandler handles alert requests.
type AlertsHandler struct {
	DB    *pgxpool.Pool
	Cache *cache.RedisClient
}

// CreateAlert godoc
// @Summary     Create alert
// @Description Creates a new system alert
// @Tags        alerts
// @Accept      json
// @Produce     json
// @Param       alert body models.Alert true "Alert data"
// @Success     201 {object} models.Alert
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/alerts [post]
func (h *AlertsHandler) CreateAlert(c *fiber.Ctx) error {
	var alert models.Alert

	// Parse JSON body
	if err := c.BodyParser(&alert); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid data",
		})
	}

	// Validate severity
	if alert.Severity != models.SeverityCritical &&
		alert.Severity != models.SeverityWarning &&
		alert.Severity != models.SeverityInfo {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "severity must be critical, warning, or info",
		})
	}

	// Validate message
	if alert.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message is required",
		})
	}

	// Use "active" when status is not provided
	if alert.Status == "" {
		alert.Status = models.StatusActive
	}

	// Validate status
	if alert.Status != models.StatusActive &&
		alert.Status != models.StatusResolved &&
		alert.Status != models.StatusAcknowledged {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "status must be active, resolved, or acknowledged",
		})
	}

	// Generate ID and timestamp
	alert.ID = uuid.New().String()
	alert.CreatedAt = time.Now()

	// Save to PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        INSERT INTO alerts (id, severity, message, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := h.DB.Exec(ctx, query,
		alert.ID,
		alert.Severity,
		alert.Message,
		alert.Status,
		alert.CreatedAt,
	)

	if err != nil {
		log.Printf("Error saving alert: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error saving alert",
		})
	}

	// Invalidate cache
	h.Cache.Delete(ctx, "alerts:latest")

	return c.Status(fiber.StatusCreated).JSON(alert)
}

// ListAlerts godoc
// @Summary     List alerts
// @Description Gets all alerts
// @Tags        alerts
// @Produce     json
// @Param       severity query string false "Filter by severity"
// @Param       status query string false "Filter by status"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/alerts [get]
func (h *AlertsHandler) ListAlerts(c *fiber.Ctx) error {
	// Get filters
	severity := c.Query("severity", "")
	status := c.Query("status", "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Build dynamic query
	query := `
        SELECT id, severity, message, status, created_at, updated_at
        FROM alerts
        WHERE 1=1
    `
	args := []interface{}{}
	argPos := 1

	if severity != "" {
		query += ` AND severity = $` + fmt.Sprint(argPos)
		args = append(args, severity)
		argPos++
	}

	if status != "" {
		query += ` AND status = $` + fmt.Sprint(argPos)
		args = append(args, status)
		argPos++
	}

	query += ` ORDER BY created_at DESC`

	rows, err := h.DB.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Error querying alerts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error querying alerts",
		})
	}
	defer rows.Close()

	alerts := make([]models.Alert, 0)

	for rows.Next() {
		var a models.Alert
		err := rows.Scan(
			&a.ID,
			&a.Severity,
			&a.Message,
			&a.Status,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning alert: %v", err)
			continue
		}
		alerts = append(alerts, a)
	}

	return c.JSON(fiber.Map{
		"data":  alerts,
		"total": len(alerts),
	})
}

// UpdateAlert godoc
// @Summary     Update alert
// @Description Updates an existing alert
// @Tags        alerts
// @Accept      json
// @Produce     json
// @Param       id path string true "Alert ID"
// @Param       alert body models.Alert true "Data to update"
// @Success     200 {object} models.Alert
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Router      /api/v1/alerts/{id} [put]
func (h *AlertsHandler) UpdateAlert(c *fiber.Ctx) error {
	id := c.Params("id")

	var input models.Alert
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid data",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify that the alert exists
	var existing models.Alert
	checkQuery := `SELECT id FROM alerts WHERE id = $1`
	err := h.DB.QueryRow(ctx, checkQuery, id).Scan(&existing.ID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Alert not found",
		})
	}

	// Update only the provided fields
	query := `UPDATE alerts SET updated_at = $1`
	args := []interface{}{time.Now()}
	argPos := 2

	if input.Status != "" {
		query += `, status = $` + fmt.Sprint(argPos)
		args = append(args, input.Status)
		argPos++
	}

	if input.Severity != "" {
		query += `, severity = $` + fmt.Sprint(argPos)
		args = append(args, input.Severity)
		argPos++
	}

	if input.Message != "" {
		query += `, message = $` + fmt.Sprint(argPos)
		args = append(args, input.Message)
		argPos++
	}

	query += ` WHERE id = $1`

	_, err = h.DB.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Error updating alert: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error updating alert",
		})
	}

	// Invalidate cache
	h.Cache.Delete(ctx, "alerts:latest")

	// Get updated alert
	var updated models.Alert
	selectQuery := `
        SELECT id, severity, message, status, created_at, updated_at
        FROM alerts WHERE id = $1
    `
	err = h.DB.QueryRow(ctx, selectQuery, id).Scan(
		&updated.ID,
		&updated.Severity,
		&updated.Message,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error retrieving updated alert: %v", err)
	}

	return c.JSON(updated)
}
