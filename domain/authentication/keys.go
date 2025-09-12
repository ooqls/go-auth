package authentication

import (
	"crypto/rand"
	"fmt"
)

type SupportedAlgorithm string

const (
	SupportedAlgorithmAESGCM SupportedAlgorithm = "AESGCM"
)

func GenerateKey(key SupportedAlgorithm) (string, error) {
	switch key {
	case SupportedAlgorithmAESGCM:
		return generateAESGCMKey()
	}
	return "", fmt.Errorf("unsupported algorithm: %s", key)
}

func generateAESGCMKey() (string, error) {
	var key [32]byte
	_, err := rand.Read(key[:])
	if err != nil {
		return "", err
	}

	return string(key[:]), nil
}
