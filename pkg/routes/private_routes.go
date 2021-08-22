package routes

import (
	"Komentory/auth/app/controllers"
	"Komentory/auth/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App) {
	// Create routes group.
	route := a.Group("/v1")

	// Routes for POST method:
	route.Post("/sign/out", middleware.JWTProtected(), controllers.UserSignOut) // de-authorization user
}
