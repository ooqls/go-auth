package orn

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/v1/users"
)

func GetUserORN(user *users.User) string {
	return fmt.Sprintf("user:%s", user.ID.String())
}

func GetChallengeORN(challengeId uuid.UUID) string {
	return fmt.Sprintf("challenge:%s", challengeId.String())
}
