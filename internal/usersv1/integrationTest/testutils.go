package integrationtest

import (
	"context"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/valkey"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
)

func newWriter() usersv1.Writer {
	return usersv1.NewSQLWriter(datagen.New(pgx.GetPGX()))
}

func newReader(ctx context.Context) usersv1.Reader {
	valkeyC := valkey.GetConnection(ctx)
	c := cache.NewGenericCache("test", cache.NewValkeyCache(valkeyC, time.Second*5))
	return usersv1.NewSQLReader(*c, datagen.New(pgx.GetPGX()))
}

func newKey() ([]byte, []byte) {
	salt := authenticationv1.GenerateSalt("test")

	algo := crypto.NewAESGCMAlgorithm("test", salt)
	b, err := algo.GetKey()
	if err != nil {
		panic(err)
	}
	return b, salt[:]
}
