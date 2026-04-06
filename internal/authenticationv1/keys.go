package authenticationv1

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type SupportedAlgorithm string

const (
	SupportedAlgorithmAESGCM SupportedAlgorithm = "AESGCM"
)

const alphanumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateKey returns the raw key bytes (as a string) and the alphanumeric
// string used to derive them. The caller should store the key bytes and
// present the alphanumeric string to the user for safe keeping.
func GenerateKey(algorithm SupportedAlgorithm) (keyBytes []byte, alphanumeric string, err error) {
	switch algorithm {
	case SupportedAlgorithmAESGCM:
		return generateAESGCMKey()
	}
	return nil, "", fmt.Errorf("unsupported algorithm: %s", algorithm)
}

// generateAESGCMKey generates a 32-character random alphanumeric string,
// derives a 32-byte AES-256-GCM key from it via SHA-256, and returns both.
func generateAESGCMKey() ([]byte, string, error) {
	buf := make([]byte, 32)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumericChars))))
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate random alphanumeric string: %w", err)
		}
		buf[i] = alphanumericChars[n.Int64()]
	}

	keyBytes := sha256.Sum256(buf)
	return keyBytes[:], string(buf), nil
}
