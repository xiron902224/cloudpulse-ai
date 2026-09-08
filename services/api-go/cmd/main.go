package main

import (
    "log"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"

    "github.com/xiron902224/cloudpulse-ai/api/internal/cache"
    "github.com/xiron902224/cloudpulse-ai/api/internal/config"
    "github.com/xiron902224/cloudpulse-ai/api/internal/database"
    "github.com/xiron902224/cloudpulse-ai/api/internal/handlers"

    _ "github.com/xiron902224/cloudpulse-ai/api/docs" // Swagger docs
)

// @title           CloudPulse AI API
// @version         1.0
// @description     API de monitoreo cloud con IA
// @host            localhost:8080
// @BasePath        /
func main() {
    // 1. Cargar configuración
    cfg := config.Load()
    log.Println("✅ Configuración cargada")

    // 2. Conectar a PostgreSQL
    pool := database.Connect(cfg.DatabaseURL)
    defer pool.Close()
    log.Println("✅ PostgreSQL conectado")

    // 3. Crear tablas
    if err := database.CreateTables(pool); err != nil {
        log.Fatalf("Error creando tablas: %v", err)
    }
    log.Println("✅ Tablas listas")

    // 4. Conectar a Redis
    redisClient := cache.Connect(cfg.RedisURL)
    log.Println("✅ Redis conectado")

    // 5. Crear app Fiber
    app := fiber.New(fiber.Config{
        AppName:      "CloudPulse AI API",
        ServerHeader: "CloudPulse",
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError
            if e, ok := err.(*fiber.Error); ok {
                code = e.Code
            }
            return c.Status(code).JSON(fiber.Map{
                "error": err.Error(),
            })
        },
    })

    // 6. Middlewares
    app.Use(logger.New())
    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins: "*",
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
        AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
    }))

    // 7. Crear handlers
    metricsHandler := &handlers.MetricsHandler{
        DB:    pool,
        Cache: redisClient,
    }

    alertsHandler := &handlers.AlertsHandler{
        DB:    pool,
        Cache: redisClient,
    }

    // 8. Rutas
    app.Get("/health", handlers.HealthCheck)

    // API v1
    v1 := app.Group("/api/v1")

    // Métricas
    v1.Post("/metrics", metricsHandler.CreateMetric)
    v1.Get("/metrics", metricsHandler.ListMetrics)
    v1.Get("/metrics/:id", metricsHandler.GetMetric)
    v1.Delete("/metrics/:id", metricsHandler.DeleteMetric)

    // Alertas
    v1.Post("/alerts", alertsHandler.CreateAlert)
    v1.Get("/alerts", alertsHandler.ListAlerts)
    v1.Put("/alerts/:id", alertsHandler.UpdateAlert)

    // 9. Swagger
    handlers.SetupSwagger(app)

    // 10. Iniciar servidor
    log.Printf("🚀 API corriendo en http://localhost:%s", cfg.Port)
    log.Printf("📚 Swagger UI en http://localhost:%s/swagger/index.html", cfg.Port)

    if err := app.Listen(":" + cfg.Port); err != nil {
        log.Fatalf("Error iniciando servidor: %v", err)
    }
}