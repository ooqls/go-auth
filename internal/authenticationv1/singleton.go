package authenticationv1

import (
	"sync"

	"github.com/ooqls/getset/cache/factory"
)

var (
	instance     Authenticator
	instanceOnce sync.Once
)

func InitAuthenticator(
	userTokenIssuer *UserTokenIssuer,
	cacheFactory factory.CacheFactory,
	challenger Challenger,
	audience []string,
) {
	instanceOnce.Do(func() {
		instance = NewAuthenticatorV1(userTokenIssuer, cacheFactory, challenger, audience)
	})
}

func GetAuthenticatorV1() Authenticator {
	if instance == nil {
		panic("authenticationv1: authenticator not initialized; call InitAuthenticator first")
	}
	return instance
}
