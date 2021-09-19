package routes

import (
	"Komentory/auth/app/controllers"
	"Komentory/auth/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App) {
	// Create routes group.
	route := a.Group("/v1", middleware.JWTProtected())

	// Routes for PATCH method:
	route.Patch("/user/update/attrs", controllers.UpdateUserAttrs)       // update user attributes
	route.Patch("/user/update/settings", controllers.UpdateUserSettings) // update user settings
	route.Patch("/user/update/password", controllers.UpdateUserPassword) // update user password
}
