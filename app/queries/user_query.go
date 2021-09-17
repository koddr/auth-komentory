package queries

import (
	"Komentory/auth/app/models"
	"database/sql"

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
		return user, fiber.StatusNotFound, err
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
		return user, fiber.StatusNotFound, err
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
		$7::int, $8::int, $9::jsonb, 
		$10::jsonb
	)
	`

	// Send query to database.
	_, err := q.Exec(
		query,
		u.ID, u.CreatedAt, u.UpdatedAt,
		u.Email, u.PasswordHash, u.Username,
		u.UserStatus, u.UserRole, u.UserAttrs,
		u.UserSettings,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateUserStatus query for updating user status by given user ID.
func (q *UserQueries) UpdateUserStatus(id uuid.UUID) error {
	// Define query string.
	query := `
	UPDATE users 
	SET user_status = 1 
	WHERE id = $1::uuid
	`

	// Send query to database.
	_, err := q.Exec(query, id)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateUserPassword query for updating user password by given user ID.
func (q *UserQueries) UpdateUserPassword(id uuid.UUID, password_hash string) error {
	// Define query string.
	query := `
	UPDATE users 
	SET password_hash = $2::varchar 
	WHERE id = $1::uuid
	`

	// Send query to database.
	_, err := q.Exec(query, id, password_hash)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateUserAttrs query for updating user attrs by given user ID.
func (q *UserQueries) UpdateUserAttrs(id uuid.UUID, u *models.UserAttrs) error {
	// Define query string.
	query := `
	UPDATE users 
	SET user_attrs = $2::jsonb 
	WHERE id = $1::uuid
	`

	// Send query to database.
	_, err := q.Exec(query, id, u)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
