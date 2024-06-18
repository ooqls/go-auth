package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/braumsmilk/go-registry"
	"github.com/braumsmilk/go-auth/pg"
	"github.com/braumsmilk/go-auth/init/seed"
)

var host string
var port int
var user string
var pw string
var database string
var configPath string

func parseFlags() {
	flag.StringVar(&host, "host", "192.168.49.2", "host of postgres")
	flag.IntVar(&port, "port", 30040, "port of postgres")
	flag.StringVar(&user, "user", "user", "user of the postgres")
	flag.StringVar(&pw, "password", "user100", "password of user")
	flag.StringVar(&database, "db", "postgres", "database for postgres")
	flag.StringVar(&configPath, "config", "./registry.yaml", "path to a registry.yaml")
	flag.Parse()
}

func main() {
	parseFlags()
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		registry.Set(registry.Registry{
			Postgres: &registry.Server{
				Host: host,
				Port: port,
				Auth: registry.Auth{
					Username: user,
					Password: pw,
				},
			},
		})
	} else {
		err = registry.Init(configPath)
		if err != nil {
			panic(err)
		}
	}

	postg := registry.Get().Postgres
	opts := pg.PostgresOptions{
		Host: postg.Host,
		Port: postg.Port,
		User: postg.Auth.Username,
		DB:   database,
		Pw:   postg.Auth.Password,
	}
	log.Printf("user=%s, password=%s, port=%d, host=%s, database=%s",
		opts.User, opts.Pw, opts.Port, opts.Host, opts.DB)
	_, err := pg.Init(opts)
	if err != nil {
		panic(err)
	}

	seed.SeedPostgresDatabase()
}
