package controllers

import (
	"context"
	"os"
	"strconv"
	"time"

	"Komentory/auth/app/models"
	"Komentory/auth/pkg/repository"
	"Komentory/auth/pkg/utils"
	"Komentory/auth/platform/cache"
	"Komentory/auth/platform/database"

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
	validate := utils.NewValidator()

	// Validate sign up fields.
	if err := validate.Struct(signUp); err != nil {
		// Return, if some fields are not valid.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   utils.ValidatorErrors(err),
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

	// Set initialized default data for user:
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.Email = signUp.Email
	user.PasswordHash = utils.GeneratePassword(signUp.Password)
	user.Username = user.ID.String()[:4] + user.ID.String()[24:]
	user.UserStatus = 0 // 0 == unconfirmed, 1 == active, 2 == blocked
	user.UserRole = repository.UserRoleName

	// Validate user fields.
	if err := validate.Struct(user); err != nil {
		// Return, if some fields are not valid.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   utils.ValidatorErrors(err),
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

	// Delete password hash field from JSON view.
	user.PasswordHash = ""

	// Return status 201 created.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"user":  user,
	})
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
			"user":  nil,
		})
	}

	// Compare given user password with stored in found user.
	compareUserPassword := utils.ComparePasswords(foundedUser.PasswordHash, signIn.Password)
	if !compareUserPassword {
		// Return status 403, if password is not compare to stored in database.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": true,
			"msg":   repository.WrongUserEmailOrPassword,
		})
	}

	// Generate a new pair of access and refresh tokens.
	tokens, errGenerateNewTokens := utils.GenerateNewTokens(foundedUser.ID.String(), foundedUser.UserRole)
	if errGenerateNewTokens != nil {
		// Return status 500 and token generation error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errGenerateNewTokens.Error(),
		})
	}

	// Define user ID.
	userID := foundedUser.ID.String()

	// Create a new Redis connection.
	connRedis, errRedisConnection := cache.RedisConnection()
	if errRedisConnection != nil {
		// Return status 500 and Redis connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errRedisConnection.Error(),
		})
	}

	// Set refresh token to Redis.
	errSetToRedis := connRedis.Set(context.Background(), userID, tokens.Refresh, 0).Err()
	if errSetToRedis != nil {
		// Return status 500 and Redis connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errSetToRedis.Error(),
		})
	}

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
		"msg":   nil,
		"jwt": fiber.Map{
			"expire": time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
			"token":  tokens.Access,
		},
	})
}

// UserSignOut method to de-authorize user and delete refresh token from Redis.
func UserSignOut(c *fiber.Ctx) error {
	// Get claims from JWT.
	claims, errExtractTokenMetaData := utils.ExtractTokenMetaData(c)
	if errExtractTokenMetaData != nil {
		// Return status 500 and JWT parse error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errExtractTokenMetaData.Error(),
		})
	}

	// Define user ID.
	userID := claims.UserID.String()

	// Create a new Redis connection.
	connRedis, errRedisConnection := cache.RedisConnection()
	if errRedisConnection != nil {
		// Return status 500 and Redis connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   errRedisConnection.Error(),
		})
	}

	// Delete user token from Redis.
	errDelFromRedis := connRedis.Del(context.Background(), userID).Err()
	if errDelFromRedis != nil {
		// Return status 400 and bad request error.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   errDelFromRedis.Error(),
		})
	}

	// Return status 204 no content.
	return c.SendStatus(fiber.StatusNoContent)
}
