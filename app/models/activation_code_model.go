package models

import (
	"time"

	"github.com/google/uuid"
)

// ActivationCode struct to describe activation code object.
type ActivationCode struct {
	Code     string    `db:"code" json:"code" validate:"required,lte=14"`
	ExpireAt time.Time `db:"expire_at" json:"expire_at" validate:"required"`
	UserID   uuid.UUID `db:"user_id" json:"user_id" validate:"required,uuid"`
}

// ApplyActivationCode struct to describe applying activation code.
type ApplyActivationCode struct {
	Code string `json:"code" validate:"required,lte=14"`
}
