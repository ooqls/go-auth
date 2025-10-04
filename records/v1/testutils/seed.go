package testutils

import (
	"log"
	"path/filepath"

	"github.com/ooqls/go-db/sqlx"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func SeedDatabase() {
	f := "../sqlc/migrations/*.sql"
	files, err := filepath.Glob(f)
	panicIfError(err)
	for _, f := range files {
		log.Printf("seed file %s", f)
		panicIfError(sqlx.SeedSQLXFile(f))
	}
}
