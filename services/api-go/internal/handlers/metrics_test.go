package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestCreateMetric_ValidInput prueba crear una métrica con datos válidos
func TestCreateMetric_ValidInput(t *testing.T) {
	app := fiber.New()

	// Handler de prueba (mock)
	app.Post("/metrics", func(c *fiber.Ctx) error {
		var metric map[string]interface{}
		if err := c.BodyParser(&metric); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Datos inválidos",
			})
		}

		// Validación de CPU
		cpu, ok := metric["cpu_usage"].(float64)
		if !ok || cpu < 0 || cpu > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cpu_usage debe estar entre 0 y 100",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(metric)
	})

	// Body de prueba
	body := []byte(`{"cpu_usage": 45.5, "memory_usage": 62.3}`)

	req := httptest.NewRequest("POST", "/metrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error ejecutando test: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Status esperado: %d, obtenido: %d", fiber.StatusCreated, resp.StatusCode)
	}
}

// TestCreateMetric_InvalidCPU prueba validación de CPU
func TestCreateMetric_InvalidCPU(t *testing.T) {
	// Tabla de casos (table-driven tests)
	tests := []struct {
		name           string
		cpu            float64
		expectedStatus int
	}{
		// ↑ Esto es un slice de structs anónimos
		// ↑ Patrón común en tests de Go
		{"CPU negativa", -10, fiber.StatusBadRequest},
		{"CPU mayor a 100", 150, fiber.StatusBadRequest},
		{"CPU válida", 50, fiber.StatusCreated},
		{"CPU en límite 0", 0, fiber.StatusCreated},
		{"CPU en límite 100", 100, fiber.StatusCreated},
	}

	for _, tt := range tests {
		// ↑ tt = cada caso de la tabla
		// ↑ _ = índice (no lo usamos)
		t.Run(tt.name, func(t *testing.T) {
			// ↑ t.Run ejecuta un subtest con nombre
			app := fiber.New()

			app.Post("/metrics", func(c *fiber.Ctx) error {
				var metric map[string]interface{}
				c.BodyParser(&metric)

				cpu, _ := metric["cpu_usage"].(float64)
				if cpu < 0 || cpu > 100 {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "cpu_usage debe estar entre 0 y 100",
					})
				}

				return c.Status(fiber.StatusCreated).JSON(metric)
			})

			body := []byte(`{"cpu_usage": ` + formatFloat(tt.cpu) + `}`)
			req := httptest.NewRequest("POST", "/metrics", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Caso '%s': status esperado %d, obtenido %d",
					tt.name, tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestListMetrics_Pagination prueba paginación
func TestListMetrics_Pagination(t *testing.T) {
	tests := []struct {
		name          string
		queryString   string
		expectedLimit int
	}{
		{"Sin parámetros", "", 50},
		{"Limit 10", "?limit=10", 10},
		{"Limit 200 (excede max)", "?limit=200", 100},
		{"Limit negativo", "?limit=-5", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/metrics", func(c *fiber.Ctx) error {
				limit := c.QueryInt("limit", 50)
				if limit <= 0 {
					limit = 50
				}
				if limit > 100 {
					limit = 100
				}
				return c.JSON(fiber.Map{"limit": limit})
			})

			req := httptest.NewRequest("GET", "/metrics"+tt.queryString, nil)
			resp, _ := app.Test(req)

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			limit := int(result["limit"].(float64))
			if limit != tt.expectedLimit {
				t.Errorf("Caso '%s': limit esperado %d, obtenido %d",
					tt.name, tt.expectedLimit, limit)
			}
		})
	}
}

// formatFloat convierte float a string sin notación científica
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
