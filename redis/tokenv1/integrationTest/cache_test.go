package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/braumsmilk/go-auth/redis/tokenv1"
	"github.com/braumsmilk/go-auth/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	c := testutils.InitRedis()
	timeout := time.Second * 15
	defer c.Stop(context.Background(), &timeout)
	m.Run()
}

func TestTokenCache(t *testing.T) {
	tcache := tokenv1.NewRedisTokenCache()
	ctx := context.Background()

	authed, err := tcache.IsAuthenticated(ctx, 1, "abc")
	assert.Nilf(t, err, "should not error when checking if token is authenticated")
	assert.Falsef(t, authed, "should not authenticate token that does not exist")

	token, err := tcache.GetNewAuthToken(ctx, 1)
	assert.Nilf(t, err, "should not error when getting new token")
	assert.NotEmptyf(t, token, "should have gotten a new token")

	authed, err = tcache.IsAuthenticated(ctx, 1, token)
	assert.Nilf(t, err, "should not error when checking if token is authed")
	assert.Truef(t, authed, "new token should be authed")

	curToken, err := tcache.GetCurrentToken(ctx, 1)
	assert.Nilf(t, err, "should not error when getting current token")
	assert.Equalf(t, token, curToken, "new token and current token should be the same")

}
