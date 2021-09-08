package models

import (
	"time"

	"github.com/google/uuid"
)

// ResetCode struct to describe reset codes object.
type ResetCode struct {
	Code     string    `db:"code" json:"code" validate:"required"`
	ExpireAt time.Time `db:"expire_at" json:"expire_at" validate:"required"`
	UserID   uuid.UUID `db:"user_id" json:"user_id" validate:"required,uuid"`
}

// ActivationCode struct to describe activation code object.
type ActivationCode struct {
	Code string `json:"code"`
}
