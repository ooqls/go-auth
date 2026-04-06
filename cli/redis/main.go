package main

import (
	"context"

	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/getset/registry"
)

func main() {
	err := redis.Init(registry.Database{
		Server: registry.Server{
			Host: "localhost",
			Port: 6379,
			Auth: registry.Auth{
				Username: "admin",
				Password: "admin",
			},
			TLS: &registry.TLSConfig{
				Enabled:  true,
				CertPath: "./tls/server.crt",
				KeyPath:  "./tls/server.key",
				CaPath:   "./tls/ca.crt",
			},
		},
		Database: "0",
	})
	if err != nil {
		panic(err)
	}

	cmd := redis.GetConnection(context.Background()).Ping(context.Background())
	if cmd.Err() != nil {
		panic(cmd.Err())
	}
}
