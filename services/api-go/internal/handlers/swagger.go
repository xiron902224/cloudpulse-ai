package handlers

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// SetupSwagger configures the Swagger UI route.
func SetupSwagger(app *fiber.App) {
	// Swagger UI route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Alternative route (some users prefer /docs)
	// app.Get("/docs/*", fiberSwagger.WrapHandler)
}
