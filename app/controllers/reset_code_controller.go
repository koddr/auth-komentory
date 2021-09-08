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
		// Return status 400 and error message.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		// Return status 500 and database connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get code by given string.
	foundedCode, status, errGetResetCode := db.GetResetCode(activationCode.Code)
	if errGetResetCode != nil {
		// Return status and error message.
		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   errGetResetCode.Error(),
		})
	}

	// Checking, if now time greather than activation code expiration time.
	if now < foundedCode.ExpireAt.Unix() {
		// Get user by given ID.
		foundedUser, status, errGetUserByID := db.GetUserByID(foundedCode.UserID)
		if errGetUserByID != nil {
			// Return status and error message.
			return c.Status(status).JSON(fiber.Map{
				"error": true,
				"msg":   errGetUserByID.Error(),
			})
		}

		// Update user status to 1 (active).
		if errUpdateUserStatus := db.UpdateUserStatus(foundedUser.ID); errUpdateUserStatus != nil {
			// Return status 400 and bad request error message.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   errUpdateUserStatus.Error(),
			})
		}

		// Delete activation code.
		if errDeleteResetCode := db.DeleteResetCode(activationCode.Code); errDeleteResetCode != nil {
			// Return status 400 and bad request error message.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   errDeleteResetCode.Error(),
			})
		}

		// Return status 204 no content.
		return c.SendStatus(fiber.StatusNoContent)
	} else {
		// Return status 400 and bad request error message.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   utilities.GenerateErrorMessage(400, "code", "activation code is expire"),
		})
	}
}
