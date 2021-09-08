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
	ID           uuid.UUID    `db:"id" json:"id" validate:"required,uuid"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
	Email        string       `db:"email" json:"email" validate:"required,email,lte=255"`
	PasswordHash string       `db:"password_hash" json:"password_hash,omitempty" validate:"required,lte=64"`
	Username     string       `db:"username" json:"username" validate:"required,lte=18"`
	UserStatus   int          `db:"user_status" json:"user_status" validate:"int"`
	UserRole     int          `db:"user_role" json:"user_role" validate:"required,int"`
	UserAttrs    UserAttrs    `db:"user_attrs" json:"user_attrs" validate:"required,dive"`
	UserSettings UserSettings `db:"user_settings" json:"user_settings" validate:"required,dive"`
}

// UserAttrs struct to describe user attributes.
type UserAttrs struct {
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name"`
	AboutMe    string `json:"about_me"`
	Picture    string `json:"picture"`
	WebsiteURL string `json:"website_url"`
}

// UserSettings struct to describe user settings.
type UserSettings struct {
	TransactionalEmailSubscription bool `json:"transactional_email_subscription"` // like "forgot password"
	MarketingEmailSubscription     bool `json:"marketing_email_subscription"`     // like "invite friends and get X"
}

// ---
// This methods simply returns the JSON-encoded representation of the struct.
// ---

// Value make the UserAttrs struct implement the driver.Valuer interface.
func (u UserAttrs) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Value make the UserSettings struct implement the driver.Valuer interface.
func (u UserSettings) Value() (driver.Value, error) {
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
