package controllers

import (
	"time"

	"Komentory/auth/app/models"
	"Komentory/auth/platform/database"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
)

// ActivateAccount method for activate user account by given code.
func ActivateAccount(c *fiber.Ctx) error {
	// Get now time.
	now := time.Now().Unix()

	// Create a new activation code struct.
	activateAccount := &models.ActivateAccount{}

	// Checking received data from JSON body.
	if err := c.BodyParser(activateAccount); err != nil {
		return utilities.CheckForError(c, err, 400, "activation code", err.Error())
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get code by given string.
	foundedCode, status, err := db.GetActivationCode(activateAccount.Code)
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, status, "activation code", err.Error())
	}

	// Checking, if now time greather than activation code expiration time.
	if now < foundedCode.ExpireAt.Unix() {
		// Get user by given ID.
		foundedUser, status, err := db.GetUserByID(foundedCode.UserID)
		if err != nil {
			return utilities.CheckForErrorWithStatusCode(c, err, status, "user", err.Error())
		}

		// Update user status to 1 (active).
		if err := db.UpdateUserStatus(foundedUser.ID); err != nil {
			return utilities.CheckForErrorWithStatusCode(c, err, 400, "user", err.Error())
		}

		// Delete activation code.
		if err := db.DeleteActivationCode(activateAccount.Code); err != nil {
			return utilities.CheckForErrorWithStatusCode(c, err, 400, "activation code", err.Error())
		}

		// Return status 200 OK.
		// User info returns for sending welcome email by Postmark.
		return c.JSON(fiber.Map{
			"status": fiber.StatusOK,
			"user": fiber.Map{
				"email":      foundedUser.Email,
				"first_name": foundedUser.UserAttrs.FirstName,
			},
		})
	} else {
		// Return status 403 and forbidden error message.
		return utilities.ThrowJSONErrorWithStatusCode(c, 403, "activation code", "was expired")
	}
}
