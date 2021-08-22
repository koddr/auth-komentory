package models

// SignUp struct to describe register a new user.
type SignUp struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
	UserRole string `json:"user_role" validate:"lte=32"`
}

// SignIn struct to describe login user.
type SignIn struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}

// PasswordChange struct to describe change process of the user password.
type PasswordChange struct {
	NewPassword string `json:"new_password" validate:"required,lte=255"`
	OldPassword string `json:"old_password" validate:"required,lte=255"`
}
