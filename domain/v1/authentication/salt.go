package authentication

import (
	"crypto/sha256"

	"github.com/ooqls/go-crypto/crypto"
)

func GenerateSalt(username string) []byte {
	usernameHash := sha256.Sum256([]byte(username))
	return usernameHash[:crypto.SALT_SIZE]
}
