package queries

import (
	"Komentory/auth/app/models"
	"database/sql"
	"fmt"

	"github.com/Komentory/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserQueries struct for queries from User model.
type UserQueries struct {
	*sqlx.DB
}

// GetUserByID query for getting one User by given ID.
func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, int, error) {
	// Define User variable.
	user := models.User{}

	// Define query string.
	query := `
	SELECT * 
	FROM users 
	WHERE id = $1::uuid
	`

	// Send query to database.
	err := q.Get(&user, query, id)

	// Get query result.
	switch err {
	case nil:
		// Return object and 200 OK.
		return user, fiber.StatusOK, nil
	case sql.ErrNoRows:
		// Return empty object and 404 error.
		return user, fiber.StatusNotFound, fmt.Errorf(repository.GenerateErrorMessage(404, "user", "id"))
	default:
		// Return empty object and 400 error.
		return user, fiber.StatusBadRequest, err
	}
}

// GetUserByEmail query for getting one User by given Email.
func (q *UserQueries) GetUserByEmail(email string) (models.User, int, error) {
	// Define User variable.
	user := models.User{}

	// Define query string.
	query := `
	SELECT * 
	FROM users 
	WHERE email = $1::varchar
	`

	// Send query to database.
	err := q.Get(&user, query, email)

	// Get query result.
	switch err {
	case nil:
		// Return object and 200 OK.
		return user, fiber.StatusOK, nil
	case sql.ErrNoRows:
		// Return empty object and 404 error.
		return user, fiber.StatusNotFound, fmt.Errorf(repository.GenerateErrorMessage(404, "user", "email"))
	default:
		// Return empty object and 400 error.
		return user, fiber.StatusBadRequest, err
	}
}

// CreateUser query for creating a new user by given email and password hash.
func (q *UserQueries) CreateUser(u *models.User) error {
	// Define query string.
	query := `
	INSERT INTO users 
	VALUES (
		$1::uuid, $2::timestamp, $3::timestamp, 
		$4::varchar, $5::varchar, $6::varchar, 
		$7::int, $8::varchar, $9::jsonb
	)
	`

	// Send query to database.
	_, err := q.Exec(
		query,
		u.ID, u.CreatedAt, u.UpdatedAt,
		u.Email, u.PasswordHash, u.Username,
		u.UserStatus, u.UserRole, u.UserAttrs,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
