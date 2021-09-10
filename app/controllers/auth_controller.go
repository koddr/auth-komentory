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
	"github.com/google/uuid"
)

// UserSignUp method to create a new user.
func UserSignUp(c *fiber.Ctx) error {
	// Create a new user auth struct.
	signUp := &models.SignUp{}

	// Checking received data from JSON body.
	if err := c.BodyParser(signUp); err != nil {
		return utilities.CheckForError(c, err, 400, "sign up", "wrong incoming data")
	}

	// Create a new validator for a User model.
	validate := utilities.NewValidator()

	// Validate sign up fields.
	if err := validate.Struct(signUp); err != nil {
		return utilities.CheckForError(c, err, 400, "sign up", fmt.Sprintf("data is not valid, %v", err))
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForError(c, err, 500, "database", "no connection")
	}

	// Check for user is already sign up by given email.
	// If status is 404, user is not signed up.
	foundedUser, status, err := db.GetUserByEmail(signUp.Email)
	if err != nil && status != 404 {
		return utilities.CheckForError(c, err, status, "user", err.Error())
	}

	// If user with given email is already sign up, return error.
	if foundedUser.Email == signUp.Email {
		return utilities.ThrowJSONError(c, 400, "user", "already signed up")
	}

	// Create a new user struct.
	user := &models.User{}

	// Generate a new username with nanoID.
	randomUsername, err := utilities.GenerateNewNanoID("", 18)
	if err != nil {
		return utilities.CheckForError(c, err, 500, "nanoid", "fail generation")
	}

	// Set user data:
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
		user.UserSettings.MarketingEmailSubscription = true
	}

	// Validate user fields.
	if err := validate.Struct(user); err != nil {
		return utilities.CheckForError(c, err, 400, "user", fmt.Sprintf("data is not valid, %v", err))
	}

	// Create a new user with validated data.
	if err := db.CreateUser(user); err != nil {
		return utilities.CheckForError(c, err, 500, "user", fmt.Sprintf("wrong database inserting, %v", err))
	}

	// Generate a new activation code with nanoID.
	randomActivationCode, err := utilities.GenerateNewNanoID(os.Getenv("RESET_CODES_CHARS_STRING"), 14)
	if err != nil {
		return utilities.CheckForError(c, err, 500, "nanoid", "fail generation")
	}

	// Create a new ResetCode struct for activation code.
	activationCode := &models.ResetCode{}

	// Set data for activation code:
	activationCode.Code = randomActivationCode
	activationCode.ExpireAt = user.CreatedAt.Add(time.Hour * 24) // set 24 hour expiration time
	activationCode.UserID = user.ID

	// Validate activation code fields.
	if err := validate.Struct(activationCode); err != nil {
		return utilities.CheckForError(
			c, err, 400, "activation code", fmt.Sprintf("data is not valid, %v", err),
		)
	}

	// Create a new activation code with validated data.
	if err := db.CreateResetCode(activationCode); err != nil {
		return utilities.CheckForError(
			c, err, 500, "activation code", fmt.Sprintf("wrong database inserting, %v", err),
		)
	}

	// Return status 201 created.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":           false,
		"activation_code": randomActivationCode,
	})
}

// UserSignIn method to auth user and return access and refresh tokens.
func UserSignIn(c *fiber.Ctx) error {
	// Create a new user auth struct.
	signIn := &models.SignIn{}

	// Checking received data from JSON body.
	if err := c.BodyParser(signIn); err != nil {
		return utilities.CheckForError(c, err, 400, "sign in", "wrong incoming data")
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForError(c, err, 500, "database", "no connection")
	}

	// Get user by given email.
	foundedUser, status, err := db.GetUserByEmail(signIn.Email)
	if err != nil {
		return utilities.CheckForError(c, err, status, "user", err.Error())
	}

	// Compare given user password with stored in found user.
	compareUserPassword := utilities.ComparePasswords(foundedUser.PasswordHash, signIn.Password)
	if !compareUserPassword {
		return utilities.ThrowJSONError(c, 403, "auth", "email or password")
	}

	// Generate a new pair of access and refresh tokens.
	tokens, err := helpers.GenerateNewTokens(foundedUser.ID.String(), foundedUser.UserRole)
	if err != nil {
		return utilities.CheckForError(c, err, 400, "tokens", "failed to generate tokens")
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
	// Check data from JWT.
	_, err := utilities.TokenValidateExpireTime(c)
	if err != nil {
		return utilities.CheckForError(c, err, 401, "jwt", err.Error())
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
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now(),
		SameSite: os.Getenv("COOKIE_SAME_SITE"),
		Secure:   true,
		HTTPOnly: true,
	})

	// Return status 204 no content.
	return c.SendStatus(fiber.StatusNoContent)
}
