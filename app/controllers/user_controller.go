package controllers

import (
	"Komentory/auth/app/models"
	"Komentory/auth/pkg/helpers"
	"Komentory/auth/platform/database"
	"os"
	"strconv"
	"time"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateNewUser method to create a new user.
func CreateNewUser(c *fiber.Ctx) error {
	// Create a new user creation struct.
	newUser := &models.CreateNewUser{}

	// Checking received data from JSON body.
	if err := c.BodyParser(newUser); err != nil {
		return utilities.CheckForError(c, err, 400, "create user", err.Error())
	}

	// Create a new validator for the struct.
	validate := utilities.NewValidator()

	// Validate fields.
	if err := validate.Struct(newUser); err != nil {
		return utilities.CheckForValidationError(c, err, 400, "create user")
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Check for user is already sign up by given email.
	// If status is 404, user is not signed up.
	foundedUser, status, err := db.GetUserByEmail(newUser.Email)
	if err != nil && status != 404 {
		return utilities.CheckForError(c, err, status, "user", err.Error())
	}

	// If user with given email is already sign up, return error.
	if foundedUser.Email == newUser.Email {
		return utilities.ThrowJSONError(c, 400, "user", "already signed up")
	}

	// Create a new user struct.
	user := &models.User{}

	// Create a new variable for timestamp, because time fields in User model are pointers.
	now := time.Now()

	// Set user data:
	user.ID = uuid.New()
	user.CreatedAt = &now
	user.Email = newUser.Email
	user.PasswordHash = utilities.GeneratePassword(newUser.Password)
	user.UserStatus = 0 // 0 == unconfirmed, 1 == active, 2 == blocked
	user.UserRole = utilities.RoleNameUser
	user.UserAttrs.FirstName = newUser.UserAttrs.FirstName
	user.UserSettings.EmailSubscriptions.Transactional = true

	// Set optional user data:
	if newUser.UserAttrs.LastName != "" {
		user.UserAttrs.LastName = newUser.UserAttrs.LastName
	}

	// Set optional user settings:
	if newUser.UserSettings.EmailSubscriptions.Marketing {
		user.UserSettings.EmailSubscriptions.Marketing = true
	}

	// Validate user fields.
	if err := validate.Struct(user); err != nil {
		return utilities.CheckForValidationError(c, err, 400, "user")
	}

	// Create a new user with validated data.
	if err := db.CreateNewUser(user); err != nil {
		return utilities.CheckForError(c, err, 400, "user", err.Error())
	}

	// Generate a new activation code with nanoID.
	randomActivationCode, err := utilities.GenerateNewNanoID(utilities.LowerCaseWithoutDashesChars, 14)
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "nanoid", err.Error())
	}

	// Create a new ResetCode struct for activation code.
	activationCode := &models.ActivationCode{}

	// Set data for activation code:
	activationCode.Code = randomActivationCode
	activationCode.ExpireAt = user.CreatedAt.Add(time.Hour * 24) // set 24 hour expiration time
	activationCode.UserID = user.ID

	// Validate activation code fields.
	if err := validate.Struct(activationCode); err != nil {
		return utilities.CheckForValidationError(c, err, 400, "activation code")
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

// UserLogin method to user login, return user model and JWT + refresh token.
func UserLogin(c *fiber.Ctx) error {
	// Create a new user auth struct.
	userLogin := &models.UserLogin{}

	// Checking received data from JSON body.
	if err := c.BodyParser(userLogin); err != nil {
		return utilities.CheckForError(c, err, 400, "user login", err.Error())
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get user by given email.
	foundedUser, status, err := db.GetUserByEmail(userLogin.Email)
	if err != nil {
		return utilities.CheckForError(c, err, status, "user", err.Error())
	}

	// Compare given user password with stored in found user.
	compareUserPassword := utilities.ComparePasswords(foundedUser.PasswordHash, userLogin.Password)
	if !compareUserPassword {
		return utilities.ThrowJSONError(c, 403, "user login", "email or password")
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
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "minutes count", err.Error())
	}

	// Set expires hours count for refresh key from .env file.
	hoursCount, err := strconv.Atoi(os.Getenv("JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT"))
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "hours count", err.Error())
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
		"user":   foundedUser,
		"jwt": fiber.Map{
			"expire": time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
			"token":  tokens.Access,
		},
	})
}

// UserLogout method to de-authorize user and clear refresh token.
func UserLogout(c *fiber.Ctx) error {
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

// UpdateUserAttrs method for update user attributes.
func UpdateUserAttrs(c *fiber.Ctx) error {
	// Set needed credentials.
	credentials := []string{
		utilities.GenerateCredential("user_attrs", "update", true),
	}

	// Validate JWT token.
	claims, err := utilities.TokenValidateExpireTimeAndCredentials(c, credentials)
	if err != nil {
		return utilities.CheckForError(c, err, 401, "update user attrs", err.Error())
	}

	// Create a new user auth struct.
	userAttrs := &models.UserAttrs{}

	// Checking received data from JSON body.
	if err := c.BodyParser(userAttrs); err != nil {
		return utilities.CheckForError(c, err, 400, "user attrs", err.Error())
	}

	// Create a new validator for the struct.
	validate := utilities.NewValidator()

	// Validate fields.
	if err := validate.Struct(userAttrs); err != nil {
		return utilities.CheckForValidationError(c, err, 400, "update user attrs")
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get user by ID from JWT.
	foundedUser, status, err := db.GetUserByID(claims.UserID)
	if err != nil {
		return utilities.CheckForError(c, err, status, "user", err.Error())
	}

	// Update user attributes.
	err = db.UpdateUserAttrs(foundedUser.ID, userAttrs)
	if err != nil {
		return utilities.CheckForError(c, err, 400, "user attrs", err.Error())
	}

	// Return status 204 no content.
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateUserSettings method for update user settings.
func UpdateUserSettings(c *fiber.Ctx) error {
	// Set needed credentials.
	credentials := []string{
		utilities.GenerateCredential("user_settings", "update", true),
	}

	// Validate JWT token.
	claims, err := utilities.TokenValidateExpireTimeAndCredentials(c, credentials)
	if err != nil {
		return utilities.CheckForError(c, err, 401, "update user settings", err.Error())
	}

	// Create a new UserSettings struct.
	userSettings := &models.UserSettings{}

	// Checking received data from JSON body.
	if err := c.BodyParser(userSettings); err != nil {
		return utilities.CheckForError(c, err, 400, "user settings", err.Error())
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Update user attributes.
	err = db.UpdateUserSettings(claims.UserID, userSettings)
	if err != nil {
		return utilities.CheckForError(c, err, 400, "user settings", err.Error())
	}

	// Return status 204 no content.
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateUserPassword method to update user password.
func UpdateUserPassword(c *fiber.Ctx) error {
	// Set needed credentials.
	credentials := []string{
		utilities.GenerateCredential("user_password", "update", true),
	}

	// Validate JWT token.
	claims, err := utilities.TokenValidateExpireTimeAndCredentials(c, credentials)
	if err != nil {
		return utilities.CheckForError(c, err, 401, "update user password", err.Error())
	}

	// Create a new UpdateUserPassword struct.
	updatePassword := &models.UpdateUserPassword{}

	// Checking received data from JSON body.
	if err := c.BodyParser(updatePassword); err != nil {
		return utilities.CheckForError(c, err, 400, "user password", err.Error())
	}

	// Create a new validator for a User model.
	validate := utilities.NewValidator()

	// Validate sign up fields.
	if err := validate.Struct(updatePassword); err != nil {
		return utilities.CheckForValidationError(c, err, 400, "task")
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		return utilities.CheckForErrorWithStatusCode(c, err, 500, "database", err.Error())
	}

	// Get user by given email.
	foundedUser, status, err := db.GetUserByID(claims.UserID)
	if err != nil {
		return utilities.CheckForError(c, err, status, "user", err.Error())
	}

	// Compare given user password with stored in found user.
	matchUserPasswords := utilities.ComparePasswords(foundedUser.PasswordHash, updatePassword.OldPassword)
	if !matchUserPasswords {
		return utilities.ThrowJSONError(c, 403, "user", "email or password")
	}

	// Set initialized default data for user:
	newPasswordHash := utilities.GeneratePassword(updatePassword.NewPassword)

	// Create a new user with validated data.
	if err := db.UpdateUserPassword(foundedUser.ID, newPasswordHash); err != nil {
		return utilities.CheckForError(c, err, 400, "user", err.Error())
	}

	// Return status 204 no content.
	return c.SendStatus(fiber.StatusNoContent)
}
