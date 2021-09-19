package models

import "time"

// ResetCode struct to describe reset codes object.
type ResetCode struct {
	Code     string    `db:"code" json:"code" validate:"required,lte=14"`
	ExpireAt time.Time `db:"expire_at" json:"expire_at" validate:"required"`
	Email    string    `json:"email" validate:"required,email,lte=255"`
}

// NewResetCode struct to describe creation of a reset code for the given email.
type NewResetCode struct {
	Email string `json:"email" validate:"required,email,lte=255"`
}

// ApplyResetCode struct to describe applying of a given reset code.
type ApplyResetCode struct {
	Code string `json:"code" validate:"required,lte=14"`
}
