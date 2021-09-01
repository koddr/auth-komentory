package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// User struct to describe User object.
type User struct {
	ID           uuid.UUID `db:"id" json:"id" validate:"required,uuid"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	Email        string    `db:"email" json:"email" validate:"required,email,lte=255"`
	PasswordHash string    `db:"password_hash" json:"password_hash,omitempty" validate:"required,lte=64"`
	Username     string    `db:"username" json:"username" validate:"required,lte=18"`
	UserStatus   int       `db:"user_status" json:"user_status" validate:"int"`
	UserRole     string    `db:"user_role" json:"user_role" validate:"required,lte=32"`
	UserAttrs    UserAttrs `db:"user_attrs" json:"user_attrs" validate:"required,dive"`
}

// UserAttrs struct to describe user attributes.
type UserAttrs struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AboutMe   string `json:"about_me"`
	Picture   string `json:"picture"`
}

// Value make the UserAttrs struct implement the driver.Valuer interface.
// This method simply returns the JSON-encoded representation of the struct.
func (b UserAttrs) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan make the UserAttrs struct implement the sql.Scanner interface.
// This method simply decodes a JSON-encoded value into the struct fields.
func (b *UserAttrs) Scan(value interface{}) error {
	j, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(j, &b)
}
