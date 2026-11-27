package handlers_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/xiron902224/cloudpulse-ai/api/internal/handlers"
    "github.com/xiron902224/cloudpulse-ai/api/internal/models"
    "github.com/xiron902224/cloudpulse-ai/api/internal/testutils"
)

// ============================================
// Setup
// ============================================

func setupAlertsTest(t *testing.T) (*fiber.App, *handlers.AlertsHandler, *pgxpool.Pool) {
    t.Helper()

    pool := testutils.SetupTestDB(t)

    handler := &handlers.AlertsHandler{
        DB:    pool,
        Cache: nil, // Los tests no necesitan cache por ahora
    }

    app := fiber.New()
    app.Post("/api/v1/alerts", handler.CreateAlert)
    app.Get("/api/v1/alerts", handler.ListAlerts)
    app.Put("/api/v1/alerts/:id", handler.UpdateAlert)

    return app, handler, pool
}

// ============================================
// Tests: CreateAlert
// ============================================

func TestCreateAlert_Success(t *testing.T) {
    app, _, pool := setupAlertsTest(t)
    ctx := context.Background()

    body := `{
        "severity": "critical",
        "message": "CPU usage above 90%",
        "status": "active"
    }`

    req := httptest.NewRequest("POST", "/api/v1/alerts", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error ejecutando test: %v", err)
    }

    if resp.StatusCode != 201 {
        t.Errorf("Esperado 201, obtenido %d", resp.StatusCode)
    }

    // Verificar que se guardó en la DB
    var count int
    err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM alerts").Scan(&count)
    if err != nil {
        t.Fatalf("Error consultando DB: %v", err)
    }

    if count != 1 {
        t.Errorf("Esperado 1 alerta, obtenido %d", count)
    }

    // Verificar contenido de la respuesta
    var alert models.Alert
    if err := json.NewDecoder(resp.Body).Decode(&alert); err != nil {
        t.Fatalf("Error decodificando respuesta: %v", err)
    }

    if alert.Severity != "critical" {
        t.Errorf("Esperado severity=critical, obtenido %s", alert.Severity)
    }

    if alert.Message != "CPU usage above 90%" {
        t.Errorf("Message incorrecto: %s", alert.Message)
    }

    if alert.Status != "active" {
        t.Errorf("Esperado status=active, obtenido %s", alert.Status)
    }

    if alert.ID == "" {
        t.Error("ID no debería estar vacío")
    }

    if alert.CreatedAt.IsZero() {
        t.Error("CreatedAt no debería estar vacío")
    }
}

func TestCreateAlert_ValidationErrors(t *testing.T) {
    app, _, _ := setupAlertsTest(t)

    tests := []struct {
        name           string
        body           string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "severity inválido",
            body:           `{"severity": "invalid", "message": "test", "status": "active"}`,
            expectedStatus: 400,
            expectedError:  "severity debe ser",
        },
        {
            name:           "message vacío",
            body:           `{"severity": "critical", "message": "", "status": "active"}`,
            expectedStatus: 400,
            expectedError:  "message es requerido",
        },
        {
            name:           "status inválido",
            body:           `{"severity": "critical", "message": "test", "status": "invalid"}`,
            expectedStatus: 400,
            expectedError:  "status debe ser",
        },
        {
            name:           "JSON malformado",
            body:           `{invalid json}`,
            expectedStatus: 400,
            expectedError:  "Datos inválidos",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/api/v1/alerts", bytes.NewBufferString(tt.body))
            req.Header.Set("Content-Type", "application/json")

            resp, err := app.Test(req)
            if err != nil {
                t.Fatalf("Error ejecutando test: %v", err)
            }

            if resp.StatusCode != tt.expectedStatus {
                t.Errorf("Esperado %d, obtenido %d", tt.expectedStatus, resp.StatusCode)
            }
        })
    }
}

func TestCreateAlert_DefaultStatus(t *testing.T) {
    app, _, _ := setupAlertsTest(t)

    // Sin status → debe usar "active"
    body := `{"severity": "warning", "message": "Test"}`

    req := httptest.NewRequest("POST", "/api/v1/alerts", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    var alert models.Alert
    if err := json.NewDecoder(resp.Body).Decode(&alert); err != nil {
        t.Fatalf("Error decodificando: %v", err)
    }

    if alert.Status != "active" {
        t.Errorf("Esperado status=active por defecto, obtenido %s", alert.Status)
    }
}

// ============================================
// Tests: ListAlerts
// ============================================

func TestListAlerts_Empty(t *testing.T) {
    app, _, _ := setupAlertsTest(t)

    req := httptest.NewRequest("GET", "/api/v1/alerts", nil)

    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    if resp.StatusCode != 200 {
        t.Errorf("Esperado 200, obtenido %d", resp.StatusCode)
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        t.Fatalf("Error decodificando: %v", err)
    }

    if result["total"].(float64) != 0 {
        t.Errorf("Esperado total=0, obtenido %v", result["total"])
    }
}

func TestListAlerts_WithData(t *testing.T) {
    app, handler, pool := setupAlertsTest(t)
    ctx := context.Background()

    // Insertar alertas de prueba
    testAlerts := []models.Alert{
        {
            ID:        "11111111-1111-1111-1111-111111111111",
            Severity:  "critical",
            Message:   "Critical alert",
            Status:    "active",
            CreatedAt: time.Now(),
        },
        {
            ID:        "22222222-2222-2222-2222-222222222222",
            Severity:  "warning",
            Message:   "Warning alert",
            Status:    "active",
            CreatedAt: time.Now(),
        },
        {
            ID:        "33333333-3333-3333-3333-333333333333",
            Severity:  "info",
            Message:   "Info alert",
            Status:    "resolved",
            CreatedAt: time.Now(),
        },
    }

    for _, alert := range testAlerts {
        _, err := pool.Exec(ctx,
            `INSERT INTO alerts (id, severity, message, status, created_at)
             VALUES ($1, $2, $3, $4, $5)`,
            alert.ID, alert.Severity, alert.Message, alert.Status, alert.CreatedAt,
        )
        if err != nil {
            t.Fatalf("Error insertando alerta: %v", err)
        }
    }
    _ = handler

    // Test: sin filtros → 3 alertas
    req := httptest.NewRequest("GET", "/api/v1/alerts", nil)
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    var result struct {
        Data  []models.Alert `json:"data"`
        Total int            `json:"total"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        t.Fatalf("Error decodificando: %v", err)
    }

    if result.Total != 3 {
        t.Errorf("Esperado total=3, obtenido %d", result.Total)
    }

    // Test: filtrar por severity
    req = httptest.NewRequest("GET", "/api/v1/alerts?severity=critical", nil)
    resp, err = app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    result.Data = nil
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        t.Fatalf("Error decodificando: %v", err)
    }

    if result.Total != 1 {
        t.Errorf("Esperado 1 crítica, obtenido %d", result.Total)
    }

    // Test: filtrar por status
    req = httptest.NewRequest("GET", "/api/v1/alerts?status=resolved", nil)
    resp, err = app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    result.Data = nil
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        t.Fatalf("Error decodificando: %v", err)
    }

    if result.Total != 1 {
        t.Errorf("Esperado 1 resuelta, obtenido %d", result.Total)
    }
}

// ============================================
// Tests: UpdateAlert
// ============================================

func TestUpdateAlert_Success(t *testing.T) {
    app, _, pool := setupAlertsTest(t)
    ctx := context.Background()

    // Insertar alerta
    alertID := "44444444-4444-4444-4444-444444444444"
    _, err := pool.Exec(ctx,
        `INSERT INTO alerts (id, severity, message, status, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
        alertID, "critical", "Original message", "active", time.Now(),
    )
    if err != nil {
        t.Fatalf("Error insertando alerta: %v", err)
    }

    // Actualizar
    body := `{"status": "resolved"}`
    req := httptest.NewRequest("PUT", "/api/v1/alerts/"+alertID, bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    if resp.StatusCode != 200 {
        t.Errorf("Esperado 200, obtenido %d", resp.StatusCode)
    }

    // Verificar en DB
    var status string
    err = pool.QueryRow(ctx, "SELECT status FROM alerts WHERE id = $1", alertID).Scan(&status)
    if err != nil {
        t.Fatalf("Error consultando: %v", err)
    }

    if status != "resolved" {
        t.Errorf("Esperado status=resolved, obtenido %s", status)
    }
}

func TestUpdateAlert_NotFound(t *testing.T) {
    app, _, _ := setupAlertsTest(t)

    body := `{"status": "resolved"}`
    req := httptest.NewRequest("PUT", "/api/v1/alerts/nonexistent-id", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }

    if resp.StatusCode != 404 {
        t.Errorf("Esperado 404, obtenido %d", resp.StatusCode)
    }
}