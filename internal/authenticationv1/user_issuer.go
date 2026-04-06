package authenticationv1

import (
	"sync"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/contexts"
)

var tokenIssuer *UserTokenIssuer
var m sync.Mutex

func GetUserTokenIssuer() UserTokenIssuer {
	m.Lock()
	defer m.Unlock()
	if tokenIssuer == nil {
		tokenIssuer = NewUserTokenIssuer(RefreshIssuer(), AuthenticationIssuer())
	}

	return *tokenIssuer
}

type UserTokenIssuer struct {
	refreshIssuer       jwt.TokenIssuer[claims.UserClaims]
	authorizationIssuer jwt.TokenIssuer[claims.UserClaims]
}

func NewUserTokenIssuer(
	refreshIssuer jwt.TokenIssuer[claims.UserClaims],
	authorizationIssuer jwt.TokenIssuer[claims.UserClaims]) *UserTokenIssuer {
	return &UserTokenIssuer{
		refreshIssuer:       refreshIssuer,
		authorizationIssuer: authorizationIssuer,
	}
}

func (t *UserTokenIssuer) IssueRefreshToken(ctx contexts.LContext, claims claims.UserClaims) (string, error) {
	tken, _, err := t.refreshIssuer.IssueToken(claims.UserID.String(), claims)
	if err != nil {
		return "", err
	}

	return tken, nil
}

func (t *UserTokenIssuer) IssueAuthorizationToken(ctx contexts.LContext, claims claims.UserClaims) (string, error) {
	tken, _, err := t.authorizationIssuer.IssueToken(claims.UserID.String(), claims)
	if err != nil {
		return "", err
	}

	return tken, nil
}

func (t *UserTokenIssuer) DecryptRefreshToken(ctx contexts.LContext, token string) (*gojwt.Token, *claims.UserClaims, error) {
	jwtToken, claims, err := t.refreshIssuer.Decrypt(token)
	if err != nil {
		return nil, nil, err
	}

	return jwtToken, &claims, nil
}

func (t *UserTokenIssuer) DecryptAuthorizationToken(ctx contexts.LContext, token string) (*gojwt.Token, *claims.UserClaims, error) {
	jwtToken, claims, err := t.authorizationIssuer.Decrypt(token)
	if err != nil {
		return nil, nil, err
	}

	return jwtToken, &claims, nil
}
