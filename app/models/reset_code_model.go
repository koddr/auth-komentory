package models

import "time"

// ResetCode struct to describe reset codes object.
type ResetCode struct {
	Code     string    `db:"code" json:"code" validate:"required"`
	ExpireAt time.Time `db:"expire_at" json:"expire_at" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
}

// ResetPassword struct to describe forgot password object.
type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPassword struct to describe reset password object.
type ResetPassword struct {
	Code string `json:"code" validate:"required"`
}
