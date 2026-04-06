package authenticationv1

import (
	"crypto/rand"
	"encoding/base64"
)

func generateRandomChallenge() string {
	challenge := make([]byte, 64)
	rand.Read(challenge)
	return base64.StdEncoding.EncodeToString(challenge)
}

func generateRandomSalt() string {
	salt := make([]byte, 32)
	rand.Read(salt)
	return base64.StdEncoding.EncodeToString(salt)
}
