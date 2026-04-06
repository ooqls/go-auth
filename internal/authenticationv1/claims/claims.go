package claims

import "github.com/google/uuid"

type UserId = uuid.UUID

type UserClaims struct {
	UserID   UserId `json:"user_id"`
	UserData []byte `json:"user_data"`
}
