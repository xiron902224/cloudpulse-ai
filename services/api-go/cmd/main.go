package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
        "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/gofiber/contrib/websocket"
	"github.com/xiron902224/cloudpulse-ai/api/internal/cache"
	"github.com/xiron902224/cloudpulse-ai/api/internal/config"
	"github.com/xiron902224/cloudpulse-ai/api/internal/database"
	"github.com/xiron902224/cloudpulse-ai/api/internal/handlers"
	"github.com/xiron902224/cloudpulse-ai/api/internal/middleware"

	_ "github.com/xiron902224/cloudpulse-ai/api/docs" // Swagger docs
)

// @title           CloudPulse AI API
// @version         1.0
// @description     Cloud monitoring API with AI
// @host            localhost:8080
// @BasePath        /
func main() {
	// 1. Load configuration
	cfg := config.Load()
	log.Println("✅ Configuration loaded")

	// 2. Connect to PostgreSQL
	pool := database.Connect(cfg.DatabaseURL)
	defer pool.Close()
	log.Println("✅ PostgreSQL connected")

	// 3. Create tables
	if err := database.CreateTables(pool); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}
	log.Println("✅ Tables ready")

	// 4. Connect to Redis
	redisClient := cache.Connect(cfg.RedisURL)
	log.Println("✅ Redis connected")

	// 5. Create Fiber app
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

	 handlers.InitWebSocketHub()
	// 6. Middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.RateLimit())
	app.Use("/ws", handlers.WebSocketUpgrade)
    app.Get("/ws", websocket.New(handlers.WebSocketHandler))
     app.Use(middleware.PrometheusMetrics())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

       // Endpoint de métricas para Prometheus
    app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// 7. Create handlers
	metricsHandler := &handlers.MetricsHandler{
		DB:    pool,
		Cache: redisClient,
	}

	alertsHandler := &handlers.AlertsHandler{
		DB:    pool,
		Cache: redisClient,
	}

	// 8. Routes
	app.Get("/health", handlers.HealthCheck)

	// API v1
	v1 := app.Group("/api/v1")

	// Metrics
	v1.Post("/metrics", metricsHandler.CreateMetric)
	v1.Get("/metrics", metricsHandler.ListMetrics)
	v1.Get("/metrics/:id", metricsHandler.GetMetric)
	v1.Delete("/metrics/:id", metricsHandler.DeleteMetric)

	// Alerts
	v1.Post("/alerts", alertsHandler.CreateAlert)
	v1.Get("/alerts", alertsHandler.ListAlerts)
	v1.Put("/alerts/:id", alertsHandler.UpdateAlert)

	// 9. Swagger
	handlers.SetupSwagger(app)

	// 10. Start server
	log.Printf("🚀 API running at http://localhost:%s", cfg.Port)
	log.Printf("📚 Swagger UI at http://localhost:%s/swagger/index.html", cfg.Port)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
