package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ooqls/go-auth/api/v1/gen/gen_authentication"
	"github.com/ooqls/go-auth/cli/authentication"
)

func unmarshal(resp *http.Response, target interface{}) error {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, target)
}

func main() {
	c, err := gen_authentication.NewClient("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	authCli := authentication.NewAuthenticationClient(*c)

	resp, err := authCli.Register(context.Background(), "test@test.com", "test", "test2")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("resp: %+v", resp)

	uid, okey, rkey, err := authCli.Login(context.Background(), "test2", "test")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("uid: %s", *uid)
	log.Printf("okey: %s", *okey)
	log.Printf("rkey: %s", *rkey)

}
