package handlers

import (
    "github.com/gofiber/fiber/v2"
    fiberSwagger "github.com/swaggo/fiber-swagger"
)

// SetupSwagger configura la ruta de Swagger UI
func SetupSwagger(app *fiber.App) {
    // Ruta para Swagger UI
    app.Get("/swagger/*", fiberSwagger.WrapHandler)
    
    // Ruta alternativa (algunos prefieren /docs)
    // app.Get("/docs/*", fiberSwagger.WrapHandler)
}