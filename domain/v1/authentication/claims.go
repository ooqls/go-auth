package authentication

import "github.com/ooqls/go-auth/records/v1/users"

type UserClaims struct {
	UserID users.UserId `json:"user_id"`
}


