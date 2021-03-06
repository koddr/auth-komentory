package helpers

import (
	"Komentory/auth/app/models"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Komentory/utilities"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// GenerateNewTokens func for generate a new Access & Refresh tokens.
func GenerateNewTokens(id string, role int) (*models.Tokens, error) {
	// Generate JWT Access token.
	accessToken, err := generateNewAccessToken(id, role)
	if err != nil {
		// Return token generation error.
		return nil, err
	}

	// Generate JWT Refresh token.
	refreshToken, err := generateNewRefreshToken(id)
	if err != nil {
		// Return token generation error.
		return nil, err
	}

	return &models.Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

// ParseRefreshToken func for parse second argument from refresh token.
func ParseRefreshToken(refreshToken string) (uuid.UUID, int64, error) {
	// Send error message, when refresh token is empty.
	if refreshToken == "" {
		return uuid.UUID{}, 0, fmt.Errorf("refresh token is empty or not valid")
	}

	// Parse user ID (UUID).
	userID, err := uuid.Parse(strings.Split(refreshToken, ".")[0])
	if err != nil {
		return uuid.UUID{}, 0, fmt.Errorf("user ID is empty or not valid")
	}

	// Parse timestamp (int64).
	token, err := strconv.ParseInt(strings.Split(refreshToken, ".")[1], 0, 64)
	if err != nil {
		return uuid.UUID{}, 0, fmt.Errorf("expire time is empty or not valid")
	}

	// Return user ID and timestamp of the given refresh token.
	return userID, token, nil
}

func generateNewAccessToken(id string, role int) (string, error) {
	// Set secret key from .env file.
	secret := os.Getenv("JWT_SECRET_KEY")

	// Set expires minutes count for secret key from .env file.
	minutesCount, err := strconv.Atoi(os.Getenv("JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT"))
	if err != nil {
		m := utilities.GenerateErrorMessage(400, "token", "invalid expiration minutes count")
		return "", fmt.Errorf(m)
	}

	// Create a new claims.
	claims := jwt.MapClaims{}

	// Set default role.
	if role == 0 {
		role = utilities.RoleNameUser
	}

	// Get credentials from role.
	credentials, err := utilities.GenerateCredentialsByRole(role)
	if err != nil {
		return "", err
	}

	// Set public claims:
	claims["id"] = id
	claims["expire"] = time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix()
	claims["credentials"] = credentials

	// Create a new JWT access token with claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate token.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		// Return error, it JWT token generation failed.
		return "", err
	}

	return t, nil
}

func generateNewRefreshToken(userID string) (string, error) {
	// Set expires hours count for refresh key from .env file.
	hoursCount, err := strconv.Atoi(os.Getenv("JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT"))
	if err != nil {
		m := utilities.GenerateErrorMessage(400, "refresh token", "invalid expiration hours count")
		return "", fmt.Errorf(m)
	}

	// Set expiration time.
	expireTime := time.Now().Add(time.Hour * time.Duration(hoursCount)).Unix()

	// Return a new refresh token (nanoID random string + user ID + expire time).
	return fmt.Sprintf("%s.%d", userID, expireTime), nil
}
