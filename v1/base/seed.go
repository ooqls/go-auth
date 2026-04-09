package main

import (
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/go-auth/internal/schema"
)

func Seed() *app.App {
	seedApp := app.New("seed", app.Features{
		SQL: app.PGX(
			app.WithCreateTableStatements(schema.GetSchemaMigrations()),
			app.WithCreateIndexStatements(schema.GetViewMigrations())),
		Registry: app.Registry(),
	})
	return seedApp

}
