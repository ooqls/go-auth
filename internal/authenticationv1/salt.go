package authenticationv1

import (
	"crypto/sha256"

	"github.com/ooqls/getset/crypto/crypto"
)

func GenerateSalt(username string) [crypto.SALT_SIZE]byte {
	usernameHash := sha256.Sum256([]byte(username))
	var salt [crypto.SALT_SIZE]byte
	copy(salt[:], usernameHash[:crypto.SALT_SIZE])
	return salt
}
