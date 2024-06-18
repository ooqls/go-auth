package seed

import (
	"log"

	"github.com/braumsmilk/go-auth/pg"
	"github.com/braumsmilk/go-auth/pg/tables"
)

func SeedPostgresDatabase() {
	pg.Get().Exec(tables.GetDropTableStmt())

	for _, stmt := range tables.GetCreateTableStmts() {
		log.Printf("%s", stmt)
		_, err := pg.Get().Exec(stmt)
		if err != nil {
			panic(err)
		}
	}

	for _, stmt := range tables.GetCreateIndexStmts() {
		log.Printf("%s", stmt)
		_, err := pg.Get().Exec(stmt)
		if err != nil {
			panic(err)
		}
	}
}
