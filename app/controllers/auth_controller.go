package controllers

import (
	"os"
	"strconv"
	"time"

	"Komentory/auth/app/models"
	"Komentory/auth/pkg/helpers"
	"Komentory/auth/platform/database"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserSignUp method to create a new user.
func UserSignUp(c *fiber.Ctx) error {
	// Create a new user auth struct.
	signUp := &models.SignUp{}

	// Checking received data from JSON body.
	if err := c.BodyParser(signUp); err != nil {
		// Return status 400 and error message.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Create a new validator for a User model.
	validate := utilities.NewValidator()

	// Validate sign up fields.
	if err := validate.Struct(signUp); err != nil {
		// Return, if some fields are not valid.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   utilities.ValidatorErrors(err),
		})
	}

	// Create database connection.
	db, errOpenDBConnection := database.OpenDBConnection()
	if errOpenDBConnection != nil {
		// Return status 500 and database connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errOpenDBConnection.Error(),
		})
	}

	// Create a new user struct.
	user := &models.User{}

	// Generate a new username with nanoID.
	randomUsername, errGenerateNewNanoID := utilities.GenerateNewNanoID("", 18)
	if errGenerateNewNanoID != nil {
		// Return status 500 and username generation error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errGenerateNewNanoID.Error(),
		})
	}

	// Set initialized default data for user:
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.Email = signUp.Email
	user.PasswordHash = utilities.GeneratePassword(signUp.Password)
	user.Username = randomUsername
	user.UserStatus = 0 // 0 == unconfirmed, 1 == active, 2 == blocked
	user.UserRole = utilities.RoleNameUser
	user.UserAttrs.FirstName = signUp.UserAttrs.FirstName
	user.UserSettings.TransactionalEmailSubscription = true

	// Set optional user data:
	if signUp.UserAttrs.LastName != "" {
		user.UserAttrs.LastName = signUp.UserAttrs.LastName
	}

	// Set optional user settings:
	if signUp.UserSettings.MarketingEmailSubscription {
		signUp.UserSettings.MarketingEmailSubscription = true
	}

	// Validate user fields.
	if err := validate.Struct(user); err != nil {
		// Return, if some fields are not valid.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   utilities.ValidatorErrors(err),
		})
	}

	// Create a new user with validated data.
	if err := db.CreateUser(user); err != nil {
		// Return status 500 and create user process error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Return status 201 created.
	return c.SendStatus(fiber.StatusCreated)
}

// UserSignIn method to auth user and return access and refresh tokens.
func UserSignIn(c *fiber.Ctx) error {
	// Create a new user auth struct.
	signIn := &models.SignIn{}

	// Checking received data from JSON body.
	if err := c.BodyParser(signIn); err != nil {
		// Return status 400 and error message.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Create database connection.
	db, errOpenDBConnection := database.OpenDBConnection()
	if errOpenDBConnection != nil {
		// Return status 500 and database connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errOpenDBConnection.Error(),
		})
	}

	// Get user by given email.
	foundedUser, status, errGetUserByEmail := db.GetUserByEmail(signIn.Email)
	if errGetUserByEmail != nil {
		// Return status and error message.
		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   errGetUserByEmail.Error(),
		})
	}

	// Compare given user password with stored in found user.
	compareUserPassword := utilities.ComparePasswords(foundedUser.PasswordHash, signIn.Password)
	if !compareUserPassword {
		// Return status 403, if password is not compare to stored in database.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": true,
			"msg":   utilities.GenerateErrorMessage(403, "auth", "email or password"),
		})
	}

	// Generate a new pair of access and refresh tokens.
	tokens, errGenerateNewTokens := helpers.GenerateNewTokens(foundedUser.ID.String(), foundedUser.UserRole)
	if errGenerateNewTokens != nil {
		// Return status 500 and token generation error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errGenerateNewTokens.Error(),
		})
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
		// Return status 500 and Redis connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Set expires hours count for refresh key from .env file.
	hoursCount, err := strconv.Atoi(os.Getenv("JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT"))
	if err != nil {
		// Return status 500 and Redis connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
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

	// Return status 200 OK.
	return c.JSON(fiber.Map{
		"error": false,
		"jwt": fiber.Map{
			"expire": time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
			"token":  tokens.Access,
		},
	})
}

// UserSignOut method to de-authorize user and delete refresh token from Redis.
func UserSignOut(c *fiber.Ctx) error {
	// Get claims from JWT.
	_, errExtractTokenMetaData := utilities.ExtractTokenMetaData(c)
	if errExtractTokenMetaData != nil {
		// Return status 500 and JWT parse error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errExtractTokenMetaData.Error(),
		})
	}

	// Define user ID.
	// userID := claims.UserID.String()

	// Create a new Redis connection.
	// connRedis, errRedisConnection := cache.RedisConnection()
	// if errRedisConnection != nil {
	// 	// Return status 500 and Redis connection error.
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": true,
	// 		"msg":   errRedisConnection.Error(),
	// 	})
	// }

	// // Delete user token from Redis.
	// errDelFromRedis := connRedis.Del(context.Background(), userID).Err()
	// if errDelFromRedis != nil {
	// 	// Return status 400 and bad request error.
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": true,
	// 		"msg":   errDelFromRedis.Error(),
	// 	})
	// }

	// Clear refresh token cookie.
	c.ClearCookie("refresh_token")

	// Return status 204 no content.
	return c.SendStatus(fiber.StatusNoContent)
}
