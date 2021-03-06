package queries

import (
	"Komentory/auth/app/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// ResetCodeQueries struct for queries from User model.
type ResetCodeQueries struct {
	*sqlx.DB
}

// GetResetCode query for getting reset code by given string.
func (q *ResetCodeQueries) GetResetCode(code string) (models.ResetCode, int, error) {
	// Define ResetCode variable.
	resetCode := models.ResetCode{}

	// Define query string.
	query := `
	SELECT *
	FROM
		reset_codes
	WHERE
		code = $1::varchar
	LIMIT 1
	`

	// Send query to database.
	err := q.Get(&resetCode, query, code)

	// Get query result.
	switch err {
	case nil:
		// Return object and 200 OK.
		return resetCode, fiber.StatusOK, nil
	case sql.ErrNoRows:
		// Return empty object and 404 error.
		return resetCode, fiber.StatusNotFound, err
	default:
		// Return empty object and 400 error.
		return resetCode, fiber.StatusBadRequest, err
	}
}

// CreateNewResetCode query for creating a new reset code.
func (q *ResetCodeQueries) CreateNewResetCode(rc *models.ResetCode) error {
	// Define query string.
	query := `
	INSERT INTO reset_codes 
	VALUES (
		$1::varchar, $2::timestamp, $3::varchar
	)
	`

	// Send query to database.
	_, err := q.Exec(
		query,
		rc.Code, rc.ExpireAt, rc.Email,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// DeleteResetCode query for deleting reset code.
func (q *ResetCodeQueries) DeleteResetCode(code string) error {
	// Define query string.
	query := `
	DELETE FROM reset_codes 
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

// DeleteResetCodesByEmail query for deleting all reset codes for the given email.
func (q *ResetCodeQueries) DeleteResetCodesByEmail(email string) error {
	// Define query string.
	query := `
	DELETE FROM reset_codes 
	WHERE email = $1::varchar
	`

	// Send query to database.
	_, err := q.Exec(query, email)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
