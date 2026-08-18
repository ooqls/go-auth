package main

import (
	"encoding/base64"
	"fmt"

	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/go-auth/internal/authenticationv1"
)

func main() {
	salt := authenticationv1.GenerateSalt("alice")
	key, _ := crypto.DeriveAESGCMKey("hunter2", salt)
	fmt.Println(base64.StdEncoding.EncodeToString(salt[:]))
	fmt.Println(base64.StdEncoding.EncodeToString(key))
}
