package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/helmet/v2"
)

// FiberMiddleware provide Fiber's built-in middlewares.
// See: https://docs.gofiber.io/api/middleware
func FiberMiddleware(a *fiber.App) {
	// Add middlewares.
	a.Use(
		// Add helmet.
		helmet.New(),
		// Add CORS to each route.
		cors.New(cors.Config{
			AllowOrigins:     os.Getenv("CORS_ALLOW_ORIGINS"),
			AllowCredentials: true,
		}),
		// Add compression.
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed, // 1
		}),
		// Add encrypt cookies.
		encryptcookie.New(encryptcookie.Config{
			Key: encryptcookie.GenerateKey(),
		}),
		// Add func for skip favicon from logs.
		favicon.New(),
		// Add simple logger.
		logger.New(),
	)
}
