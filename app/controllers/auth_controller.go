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
		return utilities.CheckForError(c, err, 400, "sign up", err.Error())
	}

	// Create a new validator for a User model.
	validate := utilities.NewValidator()

	// Validate sign up fields.
	if err := validate.Struct(signUp); err != nil {
		return utilities.CheckForError(c, err, 400, "sign up", fmt.Sprintf("validation error, %v", err))
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Check for user is already sign up by given email.
	// If status is 404, user is not signed up.
	foundedUser, status, err := db.GetUserByEmail(signUp.Email)
	if err != nil && status != 404 {
		return utilities.CheckForErrorWithStatusCode(c, err, status, "user", err.Error())
	}

	// If user with given email is already sign up, return error.
	if foundedUser.Email == signUp.Email {
		return utilities.ThrowJSONErrorWithStatusCode(c, 400, "user", "already signed up")
	}

	// Create a new user struct.
	user := &models.User{}

	// Generate a new username with nanoID.
	randomUsername, err := utilities.GenerateNewNanoID("", 18)
	if err != nil {
		return utilities.CheckForError(c, err, 500, "nanoid", err.Error())
	}

	// Create a new variable for timestamp, because time fields in User model are pointers.
	now := time.Now()

	// Set user data:
	user.ID = uuid.New()
	user.CreatedAt = &now
	user.Email = signUp.Email
	user.PasswordHash = utilities.GeneratePassword(signUp.Password)
	user.Username = randomUsername
	user.UserStatus = 0 // 0 == unconfirmed, 1 == active, 2 == blocked
	user.UserRole = utilities.RoleNameUser
	user.UserAttrs.FirstName = signUp.UserAttrs.FirstName
	user.UserSettings.EmailSubscriptions.Transactional = true

	// Set optional user data:
	if signUp.UserAttrs.LastName != "" {
		user.UserAttrs.LastName = signUp.UserAttrs.LastName
	}

	// Set optional user settings:
	if signUp.UserSettings.MarketingEmailSubscription {
		user.UserSettings.EmailSubscriptions.Marketing = true
	}

	// Validate user fields.
	if err := validate.Struct(user); err != nil {
		return utilities.CheckForError(c, err, 400, "user", fmt.Sprintf("validation error, %v", err))
	}

	// Create a new user with validated data.
	if err := db.CreateUser(user); err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 400, "user", err.Error())
	}

	// Generate a new activation code with nanoID.
	randomActivationCode, err := utilities.GenerateNewNanoID(os.Getenv("RESET_CODES_CHARS_STRING"), 14)
	if err != nil {
		return utilities.CheckForError(c, err, 500, "nanoid", err.Error())
	}

	// Create a new ResetCode struct for activation code.
	activationCode := &models.ActivationCode{}

	// Set data for activation code:
	activationCode.Code = randomActivationCode
	activationCode.ExpireAt = user.CreatedAt.Add(time.Hour * 24) // set 24 hour expiration time
	activationCode.UserID = user.ID

	// Validate activation code fields.
	if err := validate.Struct(activationCode); err != nil {
		return utilities.CheckForError(
			c, err, 400, "activation code", fmt.Sprintf("validation error, %v", err),
		)
	}

	// Create a new activation code with validated data.
	if err := db.CreateNewActivationCode(activationCode); err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "activation code", err.Error())
	}

	// Return status 201 created.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":          fiber.StatusCreated,
		"activation_code": randomActivationCode,
	})
}

// UserSignIn method to auth user and return access and refresh tokens.
func UserSignIn(c *fiber.Ctx) error {
	// Create a new user auth struct.
	signIn := &models.SignIn{}

	// Checking received data from JSON body.
	if err := c.BodyParser(signIn); err != nil {
		return utilities.CheckForError(c, err, 400, "sign in", err.Error())
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get user by given email.
	foundedUser, status, err := db.GetUserByEmail(signIn.Email)
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, status, "user", err.Error())
	}

	// Compare given user password with stored in found user.
	compareUserPassword := utilities.ComparePasswords(foundedUser.PasswordHash, signIn.Password)
	if !compareUserPassword {
		return utilities.ThrowJSONErrorWithStatusCode(c, 403, "sign in", "email or password")
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
	return c.JSON(fiber.Map{
		"status": fiber.StatusOK,
		"jwt": fiber.Map{
			"expire": time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
			"token":  tokens.Access,
		},
		"user": foundedUser,
	})
}

// UserSignOut method to de-authorize user and delete refresh token from Redis.
func UserSignOut(c *fiber.Ctx) error {
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
