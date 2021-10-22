package routes

import (
	"Komentory/auth/app/controllers"

	"github.com/gofiber/fiber/v2"
)

// PublicRoutes func for describe group of public routes.
func PublicRoutes(a *fiber.App) {
	// Create routes group.
	route := a.Group("/v1")

	// Routes for POST method:
	route.Post("/user/create", controllers.CreateNewUser)         // create a new user & send activation code
	route.Post("/user/login", controllers.UserLogin)              // auth, return Access & Refresh tokens
	route.Post("/token/renew", controllers.RenewTokens)           // renew Access & Refresh tokens
	route.Post("/password/reset", controllers.CreateNewResetCode) // create a new reset code

	// Routes for PATCH method:
	route.Patch("/user/activate", controllers.ActivateUser)    // activate user account by code
	route.Patch("/password/reset", controllers.ApplyResetCode) // apply code for reset password

	// Routes for DELETE method:
	route.Delete("/user/logout", controllers.UserLogout) // de-authorization user
}
