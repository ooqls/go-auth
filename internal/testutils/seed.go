package testutils

import (
	"context"

	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/schema"
)

func SeedDatabase() {
	pgx.SeedPGX(context.Background(), schema.GetSchemaMigrations(), schema.GetViewMigrations())
}
