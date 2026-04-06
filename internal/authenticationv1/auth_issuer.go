package authenticationv1

import (
	"sync"

	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
)

var authenticationIssuer jwt.TokenIssuer[claims.UserClaims]
var authenticationIssuerM sync.Mutex

func AuthenticationIssuer() jwt.TokenIssuer[claims.UserClaims] {
	authenticationIssuerM.Lock()
	defer authenticationIssuerM.Unlock()

	if authenticationIssuer == nil {
		panic("authentication issuer not initialized")
	}
	return authenticationIssuer
}

func InitAuthenticationIssuer(issuer jwt.TokenIssuer[claims.UserClaims]) {
	authenticationIssuerM.Lock()
	defer authenticationIssuerM.Unlock()

	if issuer == nil {
		panic("authentication issuer is nil")
	}

	authenticationIssuer = issuer
}
