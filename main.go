package main

import (
	"Komentory/auth/pkg/configs"
	"Komentory/auth/pkg/middleware"
	"Komentory/auth/pkg/routes"
	"os"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"

	_ "github.com/joho/godotenv/autoload" // load .env file automatically
)

func main() {
	// Define Fiber config.
	config := configs.FiberConfig()

	// Define a new Fiber app with config.
	app := fiber.New(config)

	// Middlewares.
	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.

	// Routes.
	routes.PublicRoutes(app)  // Register a public routes for app.
	routes.PrivateRoutes(app) // Register a private routes for app.
	routes.NotFoundRoute(app) // Register route for 404 Error.

	// Start server (with or without graceful shutdown).
	if os.Getenv("STAGE_STATUS") == "dev" {
		utilities.StartServer(app)
	} else {
		utilities.StartServerWithGracefulShutdown(app)
	}
}
