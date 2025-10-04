package authentication

import "github.com/ooqls/go-auth/records"

type TokenResponse struct {
	AuthToken    string        `json:"auth_token"`
	RefreshToken string        `json:"refresh_token"`
	UserId       records.UserId `json:"user_id"`
}


