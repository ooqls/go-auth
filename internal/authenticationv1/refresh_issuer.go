package authenticationv1

import (
	"sync"

	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
)

var refreshIssuer jwt.TokenIssuer[claims.UserClaims]
var refreshIssuerM sync.Mutex

func RefreshIssuer() jwt.TokenIssuer[claims.UserClaims] {
	refreshIssuerM.Lock()
	defer refreshIssuerM.Unlock()
	if refreshIssuer == nil {
		panic("refresh issuer not initialized")
	}
	return refreshIssuer
}

func InitRefreshIssuer(issuer jwt.TokenIssuer[claims.UserClaims]) {
	refreshIssuerM.Lock()
	defer refreshIssuerM.Unlock()

	if issuer == nil {
		panic("refresh issuer is nil")
	}

	refreshIssuer = issuer
}
