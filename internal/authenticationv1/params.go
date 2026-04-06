package authenticationv1

import "github.com/google/uuid"

type RegistrationParam struct {
	Username string
	Key      []byte
	Email    string
}

type TokenResponse struct {
	AuthToken    string    `json:"auth_token"`
	RefreshToken string    `json:"refresh_token"`
	UserId       uuid.UUID `json:"user_id"`
}
