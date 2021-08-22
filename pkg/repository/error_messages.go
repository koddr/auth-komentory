package repository

const (
	// User messages:
	NotFoundUserWithID       string = "no user with the specified ID was found"
	NotFoundUserWithEmail    string = "no user with the specified email was found"
	NotFoundUserWithUsername string = "no user with the specified username was found"
	WrongUserEmailOrPassword string = "wrong user email address or password"
	PasswordsDoesNotMatch    string = "passwords does not match"

	// Token messages:
	UnauthorizedCredentials    string = "unauthorized, check credentials of your token"
	UnauthorizedSessionEnded   string = "unauthorized, check expiration time of your refresh token"
	UnauthorizedExpirationTime string = "unauthorized, check expiration time of your access token"
)
