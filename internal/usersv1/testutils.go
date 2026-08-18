package usersv1

import (
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
)

func newReader() Reader {
	c := cache.NewMemCache()
	cc := cache.NewGenericCache("test", c)
	return NewSQLReader(*cc, datagen.New(pgx.GetPGX()))
}

func newWriter() Writer {
	return NewSQLWriter(datagen.New(pgx.GetPGX()))
}
