package controllers

import (
	"os"
	"strconv"
	"time"

	"Komentory/auth/pkg/helpers"
	"Komentory/auth/platform/database"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
)

// RenewTokens method for renew access and refresh tokens.
func RenewTokens(c *fiber.Ctx) error {
	// Get old refresh token from client.
	oldRefreshToken := c.Cookies("refresh_token", "")

	// If no refresh token in request.
	if oldRefreshToken == "" {
		// Return status 401 and unauthorized error message.
		return utilities.ThrowJSONError(c, 401, "token", "refresh token is missing")
	}

	// Get now time.
	now := time.Now().Unix()

	// Set expiration time from Refresh token of current user.
	userID, expires, err := helpers.ParseRefreshToken(oldRefreshToken)
	if err != nil {
		return utilities.CheckForError(c, err, 400, "refresh token", "wrong incoming data")
	}

	// Checking, if now time greather than Refresh token expiration time.
	if now < expires {
		// Create database connection.
		db, err := database.OpenDBConnection()
		if err != nil {
			return utilities.CheckForError(c, err, 500, "database", "no connection")
		}

		// Get user by ID.
		foundedUser, status, err := db.GetUserByID(userID)
		if err != nil {
			return utilities.CheckForError(c, err, status, "user", err.Error())
		}

		// Generate JWT Access & Refresh tokens.
		tokens, err := helpers.GenerateNewTokens(userID.String(), foundedUser.UserRole)
		if err != nil {
			return utilities.CheckForError(c, err, 500, "jwt", err.Error())
		}

		// Create a new Redis connection.
		// connRedis, err := cache.RedisConnection()
		// if err != nil {
		// 	// Return status 500 and Redis connection error.
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"error": true,
		// 		"msg":   err.Error(),
		// 	})
		// }

		//
		// _, err = connRedis.Get(context.Background(), userID.String()).Result()
		// if err == redis.Nil {
		// 	// Return status 401 and unauthorized error message.
		// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		// 		"error": true,
		// 		"msg":   repository.UnauthorizedSessionEnded,
		// 	})
		// }

		// Save refresh token to Redis.
		// errRedis := connRedis.Set(context.Background(), userID.String(), tokens.Refresh, 0).Err()
		// if errRedis != nil {
		// 	// Return status 500 and Redis connection error.
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"error": true,
		// 		"msg":   errRedis.Error(),
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

		// Return status 200 OK and new access token with expiration time.
		return c.JSON(fiber.Map{
			"error": false,
			"jwt": fiber.Map{
				"expire": time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
				"token":  tokens.Access,
			},
		})
	} else {
		// Return status 401 and unauthorized error message.
		return utilities.ThrowJSONError(c, 401, "token", "refresh token is expire")
	}
}
