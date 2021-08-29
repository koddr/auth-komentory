package models

// SignUp struct to describe register a new user.
type SignUp struct {
	Email     string          `json:"email" validate:"required,email,lte=255"`
	Password  string          `json:"password" validate:"required,lte=255"`
	UserAttrs SignUpUserAttrs `json:"user_attrs" validate:"required,dive"`
}

// SignUpUserAttrs struct to describe user attributes.
type SignUpUserAttrs struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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
