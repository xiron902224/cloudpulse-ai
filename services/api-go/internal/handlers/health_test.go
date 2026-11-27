package handlers


import (
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/gofiber/fiber/v2"
)

func TestHealthCheck(t *testing.T) {
    // Crear app Fiber para testing
    app := fiber.New()

    // Registrar la ruta
    app.Get("/health", HealthCheck)

    // Crear petición de prueba
    req := httptest.NewRequest("GET", "/health", nil)

    // Ejecutar la petición
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("Error ejecutando test: %v", err)
    }

    // Verificar código de estado
    if resp.StatusCode != fiber.StatusOK {
        t.Errorf("Status esperado: %d, obtenido: %d", fiber.StatusOK, resp.StatusCode)
    }

    // Decodificar respuesta
    var health HealthResponse
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        t.Fatalf("Error decodificando respuesta: %v", err)
    }

    // Verificar campos
    if health.Status != "healthy" {
        t.Errorf("Status esperado: healthy, obtenido: %s", health.Status)
    }

    if health.Service != "cloudpulse-api" {
        t.Errorf("Service esperado: cloudpulse-api, obtenido: %s", health.Service)
    }

    if health.Version != "0.1.0" {
        t.Errorf("Version esperada: 0.1.0, obtenida: %s", health.Version)
    }

    if health.Timestamp == "" {
        t.Error("Timestamp no debería estar vacío")
    }
}