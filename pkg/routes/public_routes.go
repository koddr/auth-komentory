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
	route.Post("/sign/in", controllers.UserSignIn)      // auth, return Access & Refresh tokens
	route.Post("/sign/out", controllers.UserSignOut)    // de-authorization user
	route.Post("/token/renew", controllers.RenewTokens) // renew Access & Refresh tokens

	// Routes for PUT method:
	route.Put("/sign/up", controllers.UserSignUp)                // create a new user & send activation code
	route.Put("/password/reset", controllers.CreateNewResetCode) // create a new reset code

	// Routes for PATCH method:
	route.Patch("/account/activate", controllers.ActivateAccount) // activate account by code
	route.Patch("/password/reset", controllers.ResetPassword)     // reset password by code
}
