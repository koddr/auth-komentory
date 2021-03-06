package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ---
// Structures to describing user model.
// ---

// User struct to describe User object.
type User struct {
	ID           uuid.UUID    `db:"id" json:"id" validate:"required,uuid"`
	CreatedAt    *time.Time   `db:"created_at" json:"created_at,omitempty"` // pointer to time.Time for omitempty
	UpdatedAt    *time.Time   `db:"updated_at" json:"updated_at,omitempty"` // pointer to time.Time for omitempty
	Email        string       `db:"email" json:"email" validate:"required,email,lte=128"`
	PasswordHash string       `db:"password_hash" json:"password_hash,omitempty" validate:"required,lte=64"`
	UserStatus   int          `db:"user_status" json:"user_status" validate:"int"`
	UserRole     int          `db:"user_role" json:"user_role,omitempty" validate:"required,int"`
	UserAttrs    UserAttrs    `db:"user_attrs" json:"user_attrs" validate:"required,dive"`
	UserSettings UserSettings `db:"user_settings" json:"user_settings" validate:"required,dive"`
}

// UserAttrs struct to describe user attributes.
type UserAttrs struct {
	FirstName  string   `json:"first_name" validate:"required"`
	LastName   string   `json:"last_name"`
	AboutMe    string   `json:"about_me"`
	Picture    string   `json:"picture"`
	WebsiteURL string   `json:"website_url"`
	Abilities  []string `json:"abilities"`
}

// UserSettings struct to describe user settings.
type UserSettings struct {
	EmailSubscriptions EmailSubscriptions `json:"email_subscriptions" validate:"dive"`
}

// EmailSubscriptions struct to describe user settings > email subscriptions.
type EmailSubscriptions struct {
	Transactional bool `json:"transactional"` // like "forgot password"
	Marketing     bool `json:"marketing"`     // like "invite friends and get X"
}

// ---
// Structures to creating a new user.
// ---

// CreateNewUser struct to describe creation of a new user.
type CreateNewUser struct {
	Email        string       `json:"email" validate:"required,email,lte=255"`
	Password     string       `json:"password" validate:"required,lte=255"`
	UserAttrs    UserAttrs    `json:"user_attrs" validate:"required,dive"`
	UserSettings UserSettings `json:"user_settings" validate:"required,dive"`
}

// ---
// Structures to updating user attributes, settings and password.
// ---

// UpdateUserPassword struct to describe updating user password.
type UpdateUserPassword struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// ---
// Structures to authenticating user.
// ---

// UserLogin struct to describe user login.
type UserLogin struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}

// User struct to describe authenticated user object.
type AuthenticatedUser struct {
	ID         uuid.UUID    `json:"id"`
	Email      string       `json:"email"`
	FirstName  string       `json:"first_name"`
	LastName   string       `json:"last_name"`
	AboutMe    string       `json:"about_me"`
	Picture    string       `json:"picture"`
	WebsiteURL string       `json:"website_url"`
	Abilities  []string     `json:"abilities"`
	Status     int          `json:"status"`
	Settings   UserSettings `json:"settings"`
}

// ---
// This methods simply returns the JSON-encoded representation of the struct.
// ---

// Value make the UserAttrs struct implement the driver.Valuer interface.
func (u *UserAttrs) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Value make the UserSettings struct implement the driver.Valuer interface.
func (u *UserSettings) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// ---
// This methods simply decodes a JSON-encoded value into the struct fields.
// ---

// Scan make the UserAttrs struct implement the sql.Scanner interface.
func (u *UserAttrs) Scan(value interface{}) error {
	j, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(j, &u)
}

// Scan make the UserSettings struct implement the sql.Scanner interface.
func (u *UserSettings) Scan(value interface{}) error {
	j, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(j, &u)
}
