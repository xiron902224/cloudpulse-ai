package middleware

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitGlobal limita todas las peticiones (permisivo)
// 1000 req/min por IP → suficiente para desarrollo y uso normal
func RateLimitGlobal() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        1000,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "Demasiadas peticiones. Intenta de nuevo en 1 minuto.",
                "retry_after": 60,
            })
        },
    })
}

// RateLimitStrict para endpoints de escritura (POST, PUT, DELETE)
// 30 req/min por IP → protege contra abuso
func RateLimitStrict() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        30,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "Demasiadas peticiones de escritura. Espera 1 minuto.",
                "retry_after": 60,
            })
        },
    })
}

// RateLimitCollector para el collector (identificado por header)
// 300 req/min → el collector envía métricas frecuentemente
func RateLimitCollector() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        300,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            // Si el collector envía el header, usa eso como key
            // Si no, usa la IP (fallback)
            collectorID := c.Get("X-Collector-ID")
            if collectorID != "" {
                return "collector:" + collectorID
            }
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error":       "Collector rate limit exceeded",
                "retry_after": 60,
            })
        },
    })
}

// RateLimit (compatibilidad - usa el global)
// Deprecated: usa RateLimitGlobal() en su lugar
func RateLimit() fiber.Handler {
    return RateLimitGlobal()
}