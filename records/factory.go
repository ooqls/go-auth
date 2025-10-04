package records

import challengeattemps "github.com/ooqls/go-auth/records/v1/challengeattempts"

type Factory interface {
	NewChallengeAttemptsReader() challengeattemps.Reader
}
