package models

// Tokens struct to describe tokens object.
type Tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

// Renew struct to describe refresh token object.
type Renew struct {
	RefreshToken string `json:"refresh_token"`
}
