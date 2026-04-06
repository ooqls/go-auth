package testutils

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/ooqls/getset/db/pgx"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func SeedDatabase() {
	goAuthPath := os.Getenv("GO_AUTH")
	if goAuthPath == "" {
		panic("GO_AUTH is not set, please source the source file before running the tests")
	}

	f := filepath.Join(goAuthPath, "internal", "schema", "*.sql")
	files, err := filepath.Glob(f)
	panicIfError(err)
	for _, f := range files {
		log.Printf("goAuthPath: %s", goAuthPath)
		log.Printf("seed file %s", f)
		panicIfError(pgx.SeedPGXFile(context.Background(), f))
	}
}
