package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthResponse is the response from the /health endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// HealthCheck godoc
// @Summary     Check service health
// @Description Returns the API status
// @Tags        health
// @Produce     json
// @Success     200 {object} HealthResponse
// @Router      /health [get]
func HealthCheck(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "cloudpulse-api",
		Version:   "0.1.0",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
