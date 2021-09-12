package database

import (
	"Komentory/auth/app/queries"

	"github.com/Komentory/utilities/database"
)

// Queries struct for collect all app queries.
type Queries struct {
	*queries.UserQueries           // load queries from User model
	*queries.ActivationCodeQueries // load queries from ActivationCode model
	*queries.ResetCodeQueries      // load queries from ResetCode model
}

// OpenDBConnection func for opening database connection.
func OpenDBConnection() (*Queries, error) {
	// Define a new PostgreSQL connection.
	db, err := database.PostgreSQLConnection()
	if err != nil {
		return nil, err
	}

	return &Queries{
		// Set queries from models:
		UserQueries:           &queries.UserQueries{DB: db},           // from User model
		ActivationCodeQueries: &queries.ActivationCodeQueries{DB: db}, // from ActivationCode model
		ResetCodeQueries:      &queries.ResetCodeQueries{DB: db},      // from ResetCode model
	}, nil
}
