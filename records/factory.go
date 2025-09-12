package records

import challengeattemps "github.com/ooqls/go-auth/records/challengeattempts"

type Factory interface {
	NewChallengeAttemptsReader() challengeattemps.Reader
}
