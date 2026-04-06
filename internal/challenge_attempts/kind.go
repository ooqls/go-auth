package challengeattempts

import (
	"time"

	"github.com/google/uuid"
)

const (
	Kind = "challenge_attempt"
)

type ChallengeAttempt struct {
	ID          uuid.UUID `json:"id"`
	ChallengeID uuid.UUID `json:"challenge_id"`
	UserID      uuid.UUID `json:"user_id"`
	Success     bool      `json:"success"`
	CreatedAt   time.Time `json:"created_at"`
}
