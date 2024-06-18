package integrationTest

import (
	"testing"

	"github.com/braumsmilk/go-auth/redis/usercachev1"
	"github.com/braumsmilk/go-auth/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutils.InitRedis()
	m.Run()
}

func TestUserCache(t *testing.T) {
	ucache := usercachev1.NewRedisCache()
	ucache.AddUser(1, "user")

	name := ucache.GetUser(1)
	assert.Equalf(t, "user", name, "should get the same name for userid")
}
