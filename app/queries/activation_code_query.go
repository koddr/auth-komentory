package queries

import (
	"Komentory/auth/app/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// ResetCodeQueries struct for queries from User model.
type ActivationCodeQueries struct {
	*sqlx.DB
}

// GetActivationCode query for getting activation code by given string.
func (q *ActivationCodeQueries) GetActivationCode(code string) (models.ActivationCode, int, error) {
	// Define activationCode variable.
	activationCode := models.ActivationCode{}

	// Define query string.
	query := `
	SELECT * 
	FROM activation_codes 
	WHERE code = $1::varchar
	`

	// Send query to database.
	err := q.Get(&activationCode, query, code)

	// Get query result.
	switch err {
	case nil:
		// Return object and 200 OK.
		return activationCode, fiber.StatusOK, nil
	case sql.ErrNoRows:
		// Return empty object and 404 error.
		return activationCode, fiber.StatusNotFound, err
	default:
		// Return empty object and 400 error.
		return activationCode, fiber.StatusBadRequest, err
	}
}

// CreateNewActivationCode query for creating a new activation code for a new user.
func (q *ActivationCodeQueries) CreateNewActivationCode(ac *models.ActivationCode) error {
	// Define query string.
	query := `
	INSERT INTO activation_codes 
	VALUES (
		$1::varchar, $2::timestamp, $3::uuid
	)
	`

	// Send query to database.
	_, err := q.Exec(
		query,
		ac.Code, ac.ExpireAt, ac.UserID,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// DeleteActivationCode query for deleting activation code.
func (q *ActivationCodeQueries) DeleteActivationCode(code string) error {
	// Define query string.
	query := `
	DELETE FROM activation_codes 
	WHERE code = $1::varchar
	`

	// Send query to database.
	_, err := q.Exec(query, code)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
