package authentication

import (
	"encoding/base64"
	"crypto/rand"
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
