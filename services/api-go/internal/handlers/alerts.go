package handlers

import (
    "context"
    "log"
    "time"
	"fmt"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/xiron902224/cloudpulse-ai/api/internal/cache"
    "github.com/xiron902224/cloudpulse-ai/api/internal/models"
)

// AlertsHandler maneja las peticiones de alertas
type AlertsHandler struct {
    DB    *pgxpool.Pool
    Cache *cache.RedisClient
}

// CreateAlert godoc
// @Summary     Crear alerta
// @Description Crea una nueva alerta del sistema
// @Tags        alerts
// @Accept      json
// @Produce     json
// @Param       alert body models.Alert true "Datos de la alerta"
// @Success     201 {object} models.Alert
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/alerts [post]
func (h *AlertsHandler) CreateAlert(c *fiber.Ctx) error {
    var alert models.Alert

    // Parsear body JSON
    if err := c.BodyParser(&alert); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Datos inválidos",
        })
    }

    // Validar severidad
    if alert.Severity != models.SeverityCritical &&
        alert.Severity != models.SeverityWarning &&
        alert.Severity != models.SeverityInfo {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "severity debe ser: critical, warning o info",
        })
    }

    // Validar mensaje
    if alert.Message == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "message es requerido",
        })
    }

    // Si no hay status, usar "active"
    if alert.Status == "" {
        alert.Status = models.StatusActive
    }

    // Validar status
    if alert.Status != models.StatusActive &&
        alert.Status != models.StatusResolved &&
        alert.Status != models.StatusAcknowledged {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "status debe ser: active, resolved o acknowledged",
        })
    }

    // Generar ID y fecha
    alert.ID = uuid.New().String()
    alert.CreatedAt = time.Now()

    // Guardar en PostgreSQL
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
        log.Printf("Error guardando alerta: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error guardando alerta",
        })
    }

    // Invalidar caché
    h.Cache.Delete(ctx, "alerts:latest")

    return c.Status(fiber.StatusCreated).JSON(alert)
}

// ListAlerts godoc
// @Summary     Listar alertas
// @Description Obtiene todas las alertas
// @Tags        alerts
// @Produce     json
// @Param       severity query string false "Filtrar por severidad"
// @Param       status query string false "Filtrar por estado"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /api/v1/alerts [get]
func (h *AlertsHandler) ListAlerts(c *fiber.Ctx) error {
    // Obtener filtros
    severity := c.Query("severity", "")
    status := c.Query("status", "")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Construir query dinámica
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
        log.Printf("Error consultando alertas: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error consultando alertas",
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
            log.Printf("Error escaneando alerta: %v", err)
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
// @Summary     Actualizar alerta
// @Description Actualiza una alerta existente
// @Tags        alerts
// @Accept      json
// @Produce     json
// @Param       id path string true "ID de la alerta"
// @Param       alert body models.Alert true "Datos a actualizar"
// @Success     200 {object} models.Alert
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Router      /api/v1/alerts/{id} [put]
func (h *AlertsHandler) UpdateAlert(c *fiber.Ctx) error {
    id := c.Params("id")

    var input models.Alert
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Datos inválidos",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Verificar que la alerta existe
    var existing models.Alert
    checkQuery := `SELECT id FROM alerts WHERE id = $1`
    err := h.DB.QueryRow(ctx, checkQuery, id).Scan(&existing.ID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Alerta no encontrada",
        })
    }

    // Actualizar solo los campos proporcionados
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
        log.Printf("Error actualizando alerta: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error actualizando alerta",
        })
    }

    // Invalidar caché
    h.Cache.Delete(ctx, "alerts:latest")

    // Obtener alerta actualizada
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
        log.Printf("Error obteniendo alerta actualizada: %v", err)
    }

    return c.JSON(updated)
}