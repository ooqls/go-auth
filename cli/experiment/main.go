package main

import (
	"crypto/sha256"
	"encoding/base64"
	"log"

	"github.com/ooqls/go-crypto/crypto"
)

func GenerateSalt(username string) []byte {
	usernameHash := sha256.Sum256([]byte(username))
	return usernameHash[:crypto.SALT_SIZE]
}

func main() {
	salt := GenerateSalt("sam")
	log.Println(salt)
	s := "6vSA"
	decrypted, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("decrypted: %v", decrypted)

}
