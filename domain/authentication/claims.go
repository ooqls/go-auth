package authentication

import "github.com/ooqls/go-auth/records/users"

type UserClaims struct {
	UserID users.UserId `json:"user_id"`
}


