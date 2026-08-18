package main

import (
	"context"
	"flag"
	"log"

	"github.com/ooqls/go-auth/v1/authentication/api/gen_authentication"
	"github.com/ooqls/go-auth/v1/authentication/authenticationcli"
)

var username string
var pw string

func init() {
	flag.StringVar(&username, "username", "", "Username to login")
	flag.StringVar(&pw, "password", "", "password to use when logging in")
	flag.Parse()
}

func main() {
	c, err := gen_authentication.NewClient("http://localhost:8080/authentication/api/v1")
	if err != nil {
		log.Fatal(err)
	}

	authCli := authenticationcli.NewAuthenticationClient(*c)

	uid, okey, rkey, err := authCli.Login(context.Background(), username, pw)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("uid: %s", *uid)
	log.Printf("okey: %s", *okey)
	log.Printf("rkey: %s", *rkey)

}
