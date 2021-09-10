package controllers

import (
	"time"

	"Komentory/auth/app/models"
	"Komentory/auth/platform/database"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
)

// ActivateAccount method for activate user account by code.
func ActivateAccount(c *fiber.Ctx) error {
	// Get now time.
	now := time.Now().Unix()

	// Create a new activation code struct.
	activationCode := &models.ActivationCode{}

	// Checking received data from JSON body.
	if err := c.BodyParser(activationCode); err != nil {
		return utilities.CheckForError(c, err, 400, "activation code", err.Error())
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForError(c, err, 500, "database", err.Error())
	}

	// Get code by given string.
	foundedCode, status, err := db.GetResetCode(activationCode.Code)
	if err != nil {
		return utilities.CheckForError(c, err, status, "activation code", err.Error())
	}

	// Checking, if now time greather than activation code expiration time.
	if now < foundedCode.ExpireAt.Unix() {
		// Get user by given ID.
		foundedUser, status, errGetUserByID := db.GetUserByID(foundedCode.UserID)
		if errGetUserByID != nil {
			return utilities.CheckForError(c, err, status, "user", err.Error())
		}

		// Update user status to 1 (active).
		if err := db.UpdateUserStatus(foundedUser.ID); err != nil {
			return utilities.CheckForError(c, err, 400, "user", err.Error())
		}

		// Delete activation code.
		if err := db.DeleteResetCode(activationCode.Code); err != nil {
			return utilities.CheckForError(c, err, 400, "activation code", err.Error())
		}

		// Return status 204 no content.
		return c.SendStatus(fiber.StatusNoContent)
	} else {
		// Return status 400 and bad request error message.
		return utilities.ThrowJSONError(c, 403, "activation code", "was expired")
	}
}
