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

	// Routes for PATCH method:
	route.Patch("/account/settings/password", middleware.JWTProtected(), controllers.UpdateUserPassword) // update user password
	route.Patch("/account/settings/attrs", middleware.JWTProtected(), controllers.UpdateUserAttrs)       // update user attributes
}
