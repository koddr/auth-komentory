package controllers

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"Komentory/auth/app/models"
	"Komentory/auth/pkg/helpers"
	"Komentory/auth/platform/database"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
)

// CreateNewResetCode method to create a new request to reset user password by given email.
func CreateNewResetCode(c *fiber.Ctx) error {
	// Create a new user auth struct.
	forgotPassword := &models.ForgotPassword{}

	// Checking received data from JSON body.
	if err := c.BodyParser(forgotPassword); err != nil {
		return utilities.CheckForError(c, err, 400, "forgot password", err.Error())
	}

	// Create a new validator for a User model.
	validate := utilities.NewValidator()

	// Validate sign up fields.
	if err := validate.Struct(forgotPassword); err != nil {
		return utilities.CheckForError(
			c, err, 400, "forgot password", fmt.Sprintf("validation error, %v", err),
		)
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get user by email.
	foundedUser, status, err := db.GetUserByEmail(forgotPassword.Email)
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, status, "user", err.Error())
	}

	// Deleting all previously created reset codes for the given email.
	err = db.DeleteResetCodesByEmail(foundedUser.Email)
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 400, "reset code", err.Error())
	}

	// Generate a new reset code with nanoID.
	randomResetCode, err := utilities.GenerateNewNanoID(os.Getenv("RESET_CODES_CHARS_STRING"), 14)
	if err != nil {
		return utilities.CheckForError(c, err, 500, "nanoid", err.Error())
	}

	// Create a new ResetCode struct for reset code.
	resetCode := &models.ResetCode{}

	// Set data for reset code:
	resetCode.Code = randomResetCode
	resetCode.ExpireAt = time.Now().Add(time.Hour * 2) // set 2 hour expiration time
	resetCode.Email = foundedUser.Email

	// Validate reset code fields.
	if err := validate.Struct(resetCode); err != nil {
		return utilities.CheckForError(
			c, err, 400, "reset code", fmt.Sprintf("validation error, %v", err),
		)
	}

	// Create a new reset code with validated data.
	if err := db.CreateNewResetCode(resetCode); err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "reset code", err.Error())
	}

	// Return status 201 created.
	return c.SendStatus(fiber.StatusCreated)
}

// ResetPassword method for reset password by given code.
func ResetPassword(c *fiber.Ctx) error {
	// Get now time.
	now := time.Now().Unix()

	// Create a new reset password code struct.
	resetPassword := &models.ResetPassword{}

	// Checking received data from JSON body.
	if err := c.BodyParser(resetPassword); err != nil {
		return utilities.CheckForError(c, err, 400, "reset password code", err.Error())
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get code by given string.
	foundedCode, status, err := db.GetResetCode(resetPassword.Code)
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, status, "reset password code", err.Error())
	}

	// Checking, if now time greather than activation code expiration time.
	if now < foundedCode.ExpireAt.Unix() {
		// Get user by email.
		foundedUser, status, err := db.GetUserByEmail(foundedCode.Email)
		if err != nil {
			return utilities.CheckForErrorWithStatusCode(c, err, status, "user", err.Error())
		}

		// Generate a new random string for the password with nanoID.
		randomString, err := utilities.GenerateNewNanoID(os.Getenv("RESET_CODES_CHARS_STRING"), 14)
		if err != nil {
			return utilities.CheckForError(c, err, 500, "nanoid", err.Error())
		}

		// Create a new hash for the given password.
		newPassword := utilities.GeneratePassword(randomString)

		// Update user password to random generated by nanoID.
		if err := db.UpdateUserPassword(foundedUser.ID, newPassword); err != nil {
			return utilities.CheckForErrorWithStatusCode(c, err, 400, "user", err.Error())
		}

		// Delete activation code.
		if err := db.DeleteResetCode(resetPassword.Code); err != nil {
			return utilities.CheckForErrorWithStatusCode(c, err, 400, "reset password code", err.Error())
		}

		// Generate a new pair of access and refresh tokens.
		tokens, err := helpers.GenerateNewTokens(foundedUser.ID.String(), foundedUser.UserRole)
		if err != nil {
			return utilities.CheckForError(c, err, 400, "tokens", err.Error())
		}

		// Define user ID.
		// userID := foundedUser.ID.String()

		// Create a new Redis connection.
		// connRedis, errRedisConnection := cache.RedisConnection()
		// if errRedisConnection != nil {
		// 	// Return status 500 and Redis connection error.
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"error": true,
		// 		"msg":   errRedisConnection.Error(),
		// 	})
		// }

		// Set refresh token to Redis.
		// errSetToRedis := connRedis.Set(context.Background(), userID, tokens.Refresh, 0).Err()
		// if errSetToRedis != nil {
		// 	// Return status 500 and Redis connection error.
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"error": true,
		// 		"msg":   errSetToRedis.Error(),
		// 	})
		// }

		// Set expires minutes count for secret key from .env file.
		minutesCount, err := strconv.Atoi(os.Getenv("JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT"))
		if err != nil {
			return utilities.CheckForError(c, err, 500, "minutes count", err.Error())
		}

		// Set expires hours count for refresh key from .env file.
		hoursCount, err := strconv.Atoi(os.Getenv("JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT"))
		if err != nil {
			return utilities.CheckForError(c, err, 500, "hours count", err.Error())
		}

		// Set HttpOnly cookie with refresh token.
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    tokens.Refresh,
			Expires:  time.Now().Add(time.Hour * time.Duration(hoursCount)),
			SameSite: os.Getenv("COOKIE_SAME_SITE"),
			Secure:   true,
			HTTPOnly: true,
		})

		// Clear no needed fields from JSON output.
		foundedUser.CreatedAt = nil
		foundedUser.UpdatedAt = nil
		foundedUser.PasswordHash = ""
		foundedUser.UserRole = 0

		// Return status 200 OK.
		// User is authenticated automatically.
		return c.JSON(fiber.Map{
			"status": fiber.StatusOK,
			"jwt": fiber.Map{
				"expire": time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
				"token":  tokens.Access,
			},
			"user": foundedUser,
		})
	} else {
		// Return status 403 and forbidden error message.
		return utilities.ThrowJSONErrorWithStatusCode(c, 403, "reset password code", "was expired")
	}
}
